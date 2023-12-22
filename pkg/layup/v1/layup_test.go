package layupv1_test

import (
	"fmt"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	layupv1 "github.com/picatz/layup/pkg/layup/v1"
	"google.golang.org/protobuf/encoding/protojson"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func TestBasics(t *testing.T) {
	v, err := protovalidate.New()
	if err != nil {
		fmt.Println("failed to initialize validator:", err)
	}

	t.Run("valid model", func(t *testing.T) {
		model := &layupv1.Model{
			Uri: "layup://test",
			Layers: []*layupv1.Layer{
				{
					Id: "1",
					Attributes: map[string]*structpb.Value{
						"test": structpb.NewStringValue("test"),
					},
					Nodes: []*layupv1.Node{
						{
							Id: "a",
						},
						{
							Id: "b",
						},
					},
					Links: []*layupv1.Link{
						{
							Id:   "loop",
							From: "a",
							To:   "a",
						},
						{
							Id:   "soup",
							From: "a",
							To:   "github://picatz/layup",
						},
					},
				},
				{
					Id: "2",
					Attributes: map[string]*structpb.Value{
						"test": structpb.NewStringValue("test"),
					},
				},
			},
		}

		err := v.Validate(model)
		if err != nil {
			t.Fatal(err)
		}

		b, err := protojson.Marshal(model)
		if err != nil {
			t.Fatal("failed to marshal model:", err)
		}

		fmt.Println(string(b))
	})

	t.Run("invalid model", func(t *testing.T) {
		model := &layupv1.Model{
			Uri: "layup://test",
			Layers: []*layupv1.Layer{
				// layup://test/layers/1
				{
					Id: "1",
					Attributes: map[string]*structpb.Value{
						"test": structpb.NewStringValue("test"),
					},
					Nodes: []*layupv1.Node{
						{
							Id: "a",
						},
						{
							Id: "a",
						},
					},
					Links: []*layupv1.Link{
						{
							Id:   "%",
							From: "???",
							To:   "///",
						},
					},
				},
				{
					Id: "1",
					Attributes: map[string]*structpb.Value{
						"test": structpb.NewStringValue("test"),
					},
				},
			},
		}

		err := v.Validate(model)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		t.Log(err)

		b, err := protojson.Marshal(model)
		if err != nil {
			t.Fatal("failed to marshal model:", err)
		}

		fmt.Println(string(b))
	})
}
