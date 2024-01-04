package layupv1

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

// WriteDOT writes a DOT graph to the given writer using the
// given Layup model's data.
func WriteDOT(w io.Writer, m *Model) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	bw := bufio.NewWriter(tw)
	defer bw.Flush()

	bw.WriteString(fmt.Sprintf("digraph %q {\n", m.GetUri()))
	bw.WriteString(fmt.Sprintf("\tlabel=%q\n", m.GetUri()))
	bw.WriteString("\tcompound=true\n")
	bw.WriteString("\tnode [shape=box]\n")

	for _, layer := range m.Layers {
		bw.WriteString("\tsubgraph cluster_" + layer.Id + " {\n")
		bw.WriteString(fmt.Sprintf("\t\tlabel=%q\n", layer.Id))

		for _, n := range layer.Nodes {
			bw.WriteString("\t\t" + layer.Id + "_" + n.Id + " [\n")
			bw.WriteString("\t\t\tlabel=" + fmt.Sprintf("%q", n.Id) + "\n")
			for k, v := range n.Attributes {
				var attrStr string

				switch v.GetKind().(type) {
				case *structpb.Value_NumberValue:
					attrStr = fmt.Sprintf("%s=%f", k, v.GetNumberValue())
				case *structpb.Value_StringValue:
					attrStr = fmt.Sprintf("%s=%q", k, v.GetStringValue())
				case *structpb.Value_BoolValue:
					attrStr = fmt.Sprintf("%s=%t", k, v.GetBoolValue())
				}

				bw.WriteString("\t\t\t" + attrStr + "\n")
			}
			bw.WriteString("\t\t]\n\n")
		}

		for _, link := range layer.Links {
			linkToID := link.To

			// If the link has an ID, use that as the label, otherwise
			// if it's using a URI, we need to clean it up a bit to match
			// DOT's syntax. We control the ID of each node in the subgraph,
			// so this is fairly safe.
			if strings.Contains(linkToID, "://") {
				linkToID = strings.TrimPrefix(linkToID, m.GetUri())

				// The link is likely to be a URI to a different layer, so we need
				// to remove the layer name from the URI.
				linkToID = strings.TrimPrefix(linkToID, "/layers/")

				// Replace the "/nodes/" with _ to match the node ID.
				linkToID = strings.Replace(linkToID, "/nodes/", "_", 1)
			} else {
				linkToID = fmt.Sprintf("%s_%s", layer.Id, link.To)
			}

			bw.WriteString("\t\t" + fmt.Sprintf("%s_%s -> %s [label=%q]", layer.Id, link.From, linkToID, link.Id) + "\n")
		}

		bw.WriteString("\t}\n")
	}

	bw.WriteString("}\n")

	return nil
}
