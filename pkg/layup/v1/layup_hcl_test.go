package layupv1_test

import (
	"fmt"
	"strings"
	"testing"

	layupv1 "github.com/picatz/layup/pkg/layup/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var thisProject = `uri = "layup://example"

layer "github" {
    node "my_account" {
        url = "https://github.com/picatz"
    }

    node "this_repository" {
        url = "https://github.com/picatz/layup"
    }

    node "buf_organization" {
        url = "https://github.com/bufbuild"
    }

    link "owner" {
        from = "my_account"
        to = "this_repository"
    }
}

layer "go" {
    node "owner" {
        url = "https://google.com"
    }

    node "language" {
        url = "https://golang.org"
    }

    node "runtime" {
        url = "https://golang.org/pkg/runtime"
    }

    link "stewardship" {
        from = "owner"
        to = "language"
    }

    link "implementation" {
        from = "language"
        to = "runtime"
    }
}

layer "buf" {
    node "cli" {
        url = "https://buf.build/docs/installation"
    }

    link "maintenance" {
        from = "cli"
        to = layer.github.node.buf_organization
    }

    link "uses" {
        from = "cli"
        to = layer.go.node.runtime
    }
}

layer "layup" {
    node "schema" {}

    node "hcl" {}

    node "cli" {}

    link "conversion" {
        from = "hcl"
        to = "schema"
    }

    link "schmea_source_code_genration" {
        from = "schema"
        to = layer.buf.node.cli
    }

    link "schema_source_code" {
        from = "schema"
        to = layer.github.node.this_repository
    }

    link "uses" {
        from = "cli"
        to = layer.go.node.runtime
    }
}`

func TestParseHCL_this_project(t *testing.T) {
	m, err := layupv1.ParseHCL(strings.NewReader(thisProject))
	if err != nil {
		t.Fatal(err)
	}

	b, err := protojson.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}

func TestParseHCL_basic(t *testing.T) {
	config := `
uri = "layup://test"

layer "1" {
	node "a" {
		example = 1234
	}
	node "b" {
		example = "abcd"
	}

	link "within" {
		from = node.a
		to = node.b
	}

	link "loop_within" {
		from =  node.a
		to = node.b
	}

	link "outside" {
		from = "a" # can also reference via string
		to = "github://picatz/layup"
	}
}

layer "2" {
	node "a" {}

	link "across" {
		from = node.a
		to = layer.1.node.a
	}
}
`

	m, err := layupv1.ParseHCL(strings.NewReader(config))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(m)
}
