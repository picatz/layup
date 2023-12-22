package layupv1

import (
	"fmt"
	"io"

	"github.com/bufbuild/protovalidate-go"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"google.golang.org/protobuf/types/known/structpb"
)

// schema is a struct that represents the top-level HCL schema for Layup.
//
// The rest of the HCL body content is processed manually to allow for natural
// grouping and referencing of nodes and links in the HCL file.
type schema struct {
	URI string `hcl:"uri,attr"`
	// Rest of the HCL body content
	Rest hcl.Body `hcl:",remain"`
}

// ParseHCL parses the given HCL formatted io.Reader into a Model
// by manually reading the HCL's body content (blocks, attributes, etc.)
// and converting it into a layupv1.Model based on Layup's HCL schema(s).
func ParseHCL(r io.Reader) (*Model, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var topLevel schema

	htx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"layer": cty.ObjectVal(map[string]cty.Value{}),
		},
		Functions: map[string]function.Function{},
	}

	err = hclsimple.Decode("layup.hcl", b, htx, &topLevel)
	if err != nil {
		return nil, err
	}

	v, err := protovalidate.New(
		protovalidate.WithMessages(&Model{}),
		protovalidate.WithDisableLazy(true),
	)
	if err != nil {
		return nil, err
	}

	m := &Model{
		Uri: topLevel.URI,
	}

	if body, ok := topLevel.Rest.(*hclsyntax.Body); ok {
		// TODO: consider parsing all layers, then nodes, then links.
		for _, block := range body.Blocks {
			switch block.Type {
			case "layer":
				layer := &Layer{
					Id: block.Labels[0],
				}

				// Create a new eval context for the layer, which is a copy
				// of the top-level eval context, but with the layer's nodes
				// added as variables.
				layerHtx := htx.NewChild()

				layerHtx.Variables = map[string]cty.Value{}
				layerHtx.Functions = map[string]function.Function{}

				// Add a "node" variable which will contain each node namespaced
				// behind it (e.g. node.a, node.b, etc.)
				layerHtx.Variables["node"] = cty.ObjectVal(map[string]cty.Value{})

				// Add a "link" variable which will contain each link namespaced
				// behind it (e.g. link.a, link.b, etc.)
				layerHtx.Variables["link"] = cty.ObjectVal(map[string]cty.Value{})

				// TODO: consider parsing all nodes, then links. This way we can
				//       reference nodes by name in the link blocks that are 'out
				//       of order' in the HCL file.
				for _, node := range block.Body.Blocks {
					switch node.Type {
					case "node":
						pbNode := &Node{
							Id:         node.Labels[0],
							Attributes: map[string]*structpb.Value{},
						}

						hclNode := cty.ObjectVal(map[string]cty.Value{
							"id": cty.StringVal(node.Labels[0]),
						})

						for _, attr := range node.Body.Attributes {
							if attr.Name == "id" {
								continue
							}

							val, diags := attr.Expr.Value(layerHtx)
							if diags.HasErrors() {
								return nil, fmt.Errorf("failed to parse attribute: %w", diags.Errs()[0])
							}

							hclNode.AsValueMap()[attr.Name] = val

							switch val.Type() {
							case cty.String:
								pbNode.Attributes[attr.Name] = structpb.NewStringValue(val.AsString())
							case cty.Number:
								f, _ := val.AsBigFloat().Float64()
								pbNode.Attributes[attr.Name] = structpb.NewNumberValue(f)
							case cty.Bool:
								pbNode.Attributes[attr.Name] = structpb.NewBoolValue(val.True())
							// case cty.List:
							// 	pbNode.Attributes[attr.Name] = structpb.NewListValue(val.AsValueSlice())
							// case cty.Map:
							// 	pbNode.Attributes[attr.Name] = structpb.NewStructValue(val.AsValueMap())
							// case cty.Object:
							// 	pbNode.Attributes[attr.Name] = structpb.NewStructValue(val.AsValueMap())
							default:
								return nil, fmt.Errorf("unknown attribute type: %s", val.Type().FriendlyName())
							}
						}

						// Add the node to the layer's eval context
						vm := layerHtx.Variables["node"].AsValueMap()

						if vm == nil {
							vm = map[string]cty.Value{}
						}

						vm[node.Labels[0]] = hclNode

						layerHtx.Variables["node"] = cty.ObjectVal(vm)

						layer.Nodes = append(layer.Nodes, pbNode)
					case "link":
						link := &Link{
							Id: node.Labels[0],
						}

						for _, attr := range node.Body.Attributes {
							switch attr.Name {
							case "from":
								from, diags := attr.Expr.Value(layerHtx)
								if diags.HasErrors() {
									return nil, fmt.Errorf("failed to parse 'from' attribute: %w", diags.Errs()[0])
								}

								if from.Type() == cty.String {
									link.From = from.AsString()
									continue
								}

								link.From = from.AsValueMap()["id"].AsString()
							case "to":
								// To may be a string or a reference to a node in this layer, or a node in another layer.
								// We can determine this by examining the type of the expression.
								switch expr := attr.Expr.(type) {
								case *hclsyntax.ScopeTraversalExpr:
									switch expr.Traversal.RootName() {
									case "layer":
										// Must be a reference to a node in another layer, so the traversale
										// must be of the form layer.<layer-name>.node.<node-name>; meaning we can check
										// the size before indexing into the traversal.
										if len(expr.Traversal) < 4 {
											return nil, fmt.Errorf("invalid 'to' attribute: %s", expr.Traversal)
										}

										var layerName string

										// Either a Name or Key
										switch t := expr.Traversal[1].(type) {
										case hcl.TraverseAttr:
											layerName = t.Name
										case hcl.TraverseIndex:
											n, _ := t.Key.AsBigFloat().Int64()
											layerName = fmt.Sprintf("%d", n)
										default:
											return nil, fmt.Errorf("unknown traversal type: %#+v", t)
										}

										// otherLater, ok := layerHtx.Variables["layer"].AsValueMap()[layerName]
										// if !ok {
										// 	return nil, fmt.Errorf("unknown layer: %s", layerName)
										// }

										// Get the node name
										var nodeName string
										switch t := expr.Traversal[3].(type) {
										case hcl.TraverseAttr:
											nodeName = t.Name
										case hcl.TraverseIndex:
											n, _ := t.Key.AsBigFloat().Int64()
											nodeName = fmt.Sprintf("%d", n)
										default:
											return nil, fmt.Errorf("unknown traversal type: %#+v", t)
										}

										// otherLayerNodes := otherLater.AsValueMap()["node"].AsValueMap()
										// if otherLayerNodes == nil {
										// 	return nil, fmt.Errorf("unknown layer: %s", layerName)
										// }

										// otherNode, ok := otherLayerNodes[nodeName]
										// if !ok {
										// 	return nil, fmt.Errorf("unknown node: %s", nodeName)
										// }

										link.To = fmt.Sprintf("%s/layers/%s/nodes/%s", m.Uri, layerName, nodeName)
									case "node":
										to, diags := expr.Value(layerHtx)
										if diags.HasErrors() {
											return nil, fmt.Errorf("failed to parse 'to' attribute: %w", diags.Errs()[0])
										}

										if to.Type() == cty.String {
											link.To = to.AsString()
											continue
										}

										link.To = to.AsValueMap()["id"].AsString()
									default:
										return nil, fmt.Errorf("unknown expr root name: %s", expr.Traversal.RootName())
									}
								case *hclsyntax.TemplateExpr:
									to, diags := expr.Value(layerHtx)
									if diags.HasErrors() {
										return nil, fmt.Errorf("failed to parse 'to' attribute: %w", diags.Errs()[0])
									}

									if to.Type() == cty.String {
										link.To = to.AsString()
										continue
									}

									link.To = to.AsValueMap()["id"].AsString()
								default:
									return nil, fmt.Errorf("unknown to type: %#+v", expr)
								}

							}
						}

						layer.Links = append(layer.Links, link)
					default:
						return nil, fmt.Errorf("unknown layer block type: %s", node.Type)
					}
				}

				vm := htx.Variables["layer"].AsValueMap()

				if vm == nil {
					vm = map[string]cty.Value{}
				}

				nodes := map[string]cty.Value{}
				for _, node := range layer.Nodes {
					nodes[node.Id] = cty.ObjectVal(map[string]cty.Value{
						"id": cty.StringVal(node.Id),
					})
				}

				vm[block.Labels[0]] = cty.ObjectVal(map[string]cty.Value{
					"id":   cty.StringVal(block.Labels[0]),
					"node": cty.ObjectVal(nodes),
				})

				htx.Variables["layer"] = cty.ObjectVal(vm)

				m.Layers = append(m.Layers, layer)
			default:
				return nil, fmt.Errorf("unknown block type: %s", block.Type)
			}
		}
	}

	// Apply validation rules to the model after parsing to ensure
	// invalid models are not created through the HCL parser.
	err = v.Validate(m)
	if err != nil {
		return m, err
	}

	return m, nil
}
