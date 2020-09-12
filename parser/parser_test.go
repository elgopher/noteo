package parser_test

import (
	"strings"
	"testing"

	"github.com/jacekolszak/noteo/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		content      string
		expectedYaml string
		expectedBody string
	}{
		"empty": {},
		"only body": {
			content:      "text",
			expectedYaml: "",
			expectedBody: "text",
		},
		"two lines body": {
			content:      "one\ntwo",
			expectedYaml: "",
			expectedBody: "one\ntwo",
		},
		"only yaml": {
			content:      "---\ntags: abc\n---",
			expectedYaml: "---\ntags: abc\n---",
			expectedBody: "",
		},
		"only yaml with eol": {
			content:      "---\ntags: abc\n---\n",
			expectedYaml: "---\ntags: abc\n---\n",
			expectedBody: "",
		},
		"two lines yaml": {
			content:      "---\ntags: abc\ncreated:2020-02-02\n---",
			expectedYaml: "---\ntags: abc\ncreated:2020-02-02\n---",
			expectedBody: "",
		},
		"yaml and body": {
			content:      "---\ntags: abc\n---\n\ntext",
			expectedYaml: "---\ntags: abc\n---\n",
			expectedBody: "\ntext",
		},
		"not closed yaml": {
			content:      "---\ntags:abc",
			expectedYaml: "",
			expectedBody: "---\ntags:abc",
		},
		"tag ---": {
			content:      "---\ntags: ---\n---",
			expectedYaml: "---\ntags: ---\n---",
			expectedBody: "",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			yml, body, err := parser.Parse(strings.NewReader(test.content))
			require.NoError(t, err)
			assert.Equal(t, test.expectedYaml, yml)
			assert.Equal(t, test.expectedBody, body)
		})
	}

}
