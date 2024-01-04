package layupv1

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

// WriteD2 writes a D2 (Terrastruct) graph to the given writer using the
// given Layup model's data.
func WriteD2(w io.Writer, m *Model) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	bw := bufio.NewWriter(tw)
	defer bw.Flush()

	for _, layer := range m.Layers {
		bw.WriteString(layer.Id + ": {\n")

		for _, n := range layer.Nodes {
			bw.WriteString("\t" + n.Id + ": {\n")
			for k, v := range n.Attributes {
				var attrStr string

				switch v.GetKind().(type) {
				case *structpb.Value_NumberValue:
					attrStr = fmt.Sprintf("%s = %f", k, v.GetNumberValue())
				case *structpb.Value_StringValue:
					attrStr = fmt.Sprintf("%s = %q", k, v.GetStringValue())
				case *structpb.Value_BoolValue:
					attrStr = fmt.Sprintf("%s = %t", k, v.GetBoolValue())
				}

				bw.WriteString("\t\t" + attrStr + "\n")
			}
			bw.WriteString("\t}\n\n")
		}

		bw.WriteString("}\n\n")
	}

	for _, layer := range m.Layers {
		for _, link := range layer.Links {
			linkToID := link.To

			// If the link has an ID, use that as the label, otherwise
			// if it's using a URI, we need to clean it up a bit to match
			// mermiad's syntax. We control the ID of each node in the subgraph,
			// so this is fairly safe.
			if strings.Contains(linkToID, "://") {
				linkToID = strings.TrimPrefix(linkToID, m.GetUri())

				// The link is likely to be a URI to a different layer, so we need
				// to remove the layer name from the URI.
				linkToID = strings.TrimPrefix(linkToID, "/layers/")

				// Replace the "/nodes/" with _ to match the node ID.
				linkToID = strings.Replace(linkToID, "/nodes/", ".", 1)
			} else {
				linkToID = fmt.Sprintf("%s.%s", layer.Id, link.To)
			}

			linkFromID := fmt.Sprintf("%s.%s", layer.Id, link.From)

			bw.WriteString("" + fmt.Sprintf("%s -> %s: %q", linkFromID, linkToID, link.Id) + "\n")
		}
	}

	return nil
}
