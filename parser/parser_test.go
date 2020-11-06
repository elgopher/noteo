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
		content             string
		expectedFrontMatter string
		expectedBody        string
	}{
		"empty": {},
		"only body": {
			content:             "text",
			expectedFrontMatter: "",
			expectedBody:        "text",
		},
		"two lines body": {
			content:             "one\ntwo",
			expectedFrontMatter: "",
			expectedBody:        "one\ntwo",
		},
		"only front matter": {
			content:             "---\ntags: abc\n---",
			expectedFrontMatter: "---\ntags: abc\n---",
			expectedBody:        "",
		},
		"only front matter with eol": {
			content:             "---\ntags: abc\n---\n",
			expectedFrontMatter: "---\ntags: abc\n---\n",
			expectedBody:        "",
		},
		"two lines front matter": {
			content:             "---\ntags: abc\ncreated:2020-02-02\n---",
			expectedFrontMatter: "---\ntags: abc\ncreated:2020-02-02\n---",
			expectedBody:        "",
		},
		"front matter and body": {
			content:             "---\ntags: abc\n---\n\ntext",
			expectedFrontMatter: "---\ntags: abc\n---\n",
			expectedBody:        "\ntext",
		},
		"not closed front matter": {
			content:             "---\ntags:abc",
			expectedFrontMatter: "",
			expectedBody:        "---\ntags:abc",
		},
		"tag ---": {
			content:             "---\ntags: ---\n---",
			expectedFrontMatter: "---\ntags: ---\n---",
			expectedBody:        "",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			frontMatter, body, err := parser.Parse(strings.NewReader(test.content))
			require.NoError(t, err)
			assert.Equal(t, test.expectedFrontMatter, frontMatter)
			assert.Equal(t, test.expectedBody, body)
		})
	}

}
