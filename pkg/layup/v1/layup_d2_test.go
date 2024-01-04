package layupv1_test

import (
	"fmt"
	"strings"
	"testing"

	layupv1 "github.com/picatz/layup/pkg/layup/v1"
)

func TestWriteD2(t *testing.T) {
	model, err := layupv1.ParseHCL(strings.NewReader(thisProject))
	if err != nil {
		t.Fatal(err)
	}

	d2Buffer := strings.Builder{}

	err = layupv1.WriteD2(&d2Buffer, model)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(d2Buffer.String())
}
