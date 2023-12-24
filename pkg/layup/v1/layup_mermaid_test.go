package layupv1_test

import (
	"fmt"
	"strings"
	"testing"

	layupv1 "github.com/picatz/layup/pkg/layup/v1"
)

func TestWriteMermaid(t *testing.T) {
	model, err := layupv1.ParseHCL(strings.NewReader(thisProject))
	if err != nil {
		t.Fatal(err)
	}

	mermaidBuffer := strings.Builder{}

	if err := layupv1.WriteMermiad(&mermaidBuffer, model); err != nil {
		t.Fatal(err)
	}

	fmt.Println(mermaidBuffer.String())
}
