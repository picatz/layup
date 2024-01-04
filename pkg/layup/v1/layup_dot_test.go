package layupv1_test

import (
	"fmt"
	"strings"
	"testing"

	layupv1 "github.com/picatz/layup/pkg/layup/v1"
)

func TestWriteDOT(t *testing.T) {
	model, err := layupv1.ParseHCL(strings.NewReader(thisProject))
	if err != nil {
		t.Fatal(err)
	}

	dotBuffer := strings.Builder{}

	err = layupv1.WriteDOT(&dotBuffer, model)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(dotBuffer.String())
}
