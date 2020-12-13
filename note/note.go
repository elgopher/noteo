package note

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacekolszak/noteo/parser"
	"github.com/jacekolszak/noteo/tag"
)

var tagSeparator = regexp.MustCompile(`[,\s]+`)

type Note struct {
	path            string
	modified        func() (time.Time, error)
	originalContent *originalContent
	frontMatter     *frontMatter
	body            *body
}

func New(path string) *Note {
	return newWithModifiedFunc(path, readModifiedFunc(path))
}

func readModifiedFunc(path string) func() (time.Time, error) {
	return func() (modTime time.Time, err error) {
		var stat os.FileInfo
		stat, err = os.Stat(path)
		if err != nil {
			return
		}
		modTime = stat.ModTime()
		return
	}
}

func newWithModifiedFunc(path string, modified func() (time.Time, error)) *Note {
	original := &originalContent{path: path}
	return &Note{
		path:            path,
		modified:        modified,
		originalContent: original,
		frontMatter: &frontMatter{
			path:     path,
			original: original.FrontMatter,
			mapSlice: mapSlice{},
		},
		body: &body{
			original: original.Body,
		},
	}
}

func NewWithModified(path string, modified time.Time) *Note {
	return newWithModifiedFunc(path, func() (time.Time, error) {
		return modified, nil
	})
}

func (n *Note) Path() string {
	return n.path
}

func (n *Note) Modified() (time.Time, error) {
	return n.modified()
}

func (n *Note) Created() (time.Time, error) {
	return n.frontMatter.Created()
}

func (n *Note) Tags() ([]tag.Tag, error) {
	return n.frontMatter.Tags()
}

func (n *Note) Body() (string, error) {
	return n.body.text()
}

func (n *Note) SetTag(newTag tag.Tag) error {
	return n.frontMatter.setTag(newTag)
}

func (n *Note) RemoveTag(newTag tag.Tag) error {
	return n.frontMatter.removeTag(newTag)
}

func (n *Note) RemoveTagRegex(regex *regexp.Regexp) error {
	return n.frontMatter.removeTagRegex(regex)
}

func (n *Note) UpdateLink(from, to string) error {
	body, err := n.body.text()
	if err != nil {
		return err
	}

	relativeFrom, err := n.relativePath(from)
	if err != nil {
		return err
	}

	var returnedError error
	markdownLinkRegexp := regexp.MustCompile(`(\[[^][]+])\(([^()]+)\)`) // TODO does not take into account code fences
	body = markdownLinkRegexp.ReplaceAllStringFunc(body, func(s string) string {
		linkPath := markdownLinkRegexp.FindStringSubmatch(s)[2]
		relativeLinkPath, err := n.relativePath(linkPath)
		if err != nil {
			returnedError = err
			return s
		}
		if relativeFrom == relativeLinkPath {
			return markdownLinkRegexp.ReplaceAllString(s, `$1(`+to+`)`)
		}
		if isAncestorPath(relativeFrom, relativeLinkPath) {
			newTo := to + strings.TrimPrefix(relativeLinkPath, relativeFrom)
			newTo = filepath.ToSlash(newTo) // always use slashes, even on Windows
			return markdownLinkRegexp.ReplaceAllString(s, `$1(`+newTo+`)`)
		}
		return s
	})
	if returnedError != nil {
		return returnedError
	}
	n.body.setText(body)
	return nil
}

func isAncestorPath(ancestor string, descendant string) bool {
	rel, err := filepath.Rel(ancestor, descendant)
	if err != nil {
		return false
	}
	return filepath.Dir(rel) == "."
}

// Returns relative path to note path
func (n *Note) relativePath(p string) (string, error) {
	dir := filepath.Dir(n.path)
	relativePath := p
	if !filepath.IsAbs(relativePath) {
		relativePath = filepath.Join(dir, p)
	}
	return filepath.Rel(dir, relativePath)
}

func (n *Note) Save() (bool, error) {
	frontMatter, err := n.frontMatter.marshal()
	if err != nil {
		return false, err
	}
	text, err := n.Body()
	if err != nil {
		return false, err
	}
	newContent := frontMatter + text
	original, err := n.originalContent.Full()
	if err != nil {
		return false, err
	}
	if original == newContent {
		return false, nil
	}
	newBytes := []byte(newContent)
	if err := ioutil.WriteFile(n.path, newBytes, 0664); err != nil {
		return false, err
	}
	return true, nil
}

type originalContent struct {
	path        string
	once        sync.Once
	frontMatter string
	body        string
}

func (c *originalContent) ensureLoaded() error {
	var err error
	c.once.Do(func() {
		var file *os.File
		file, err = os.Open(c.path)
		if err != nil {
			return
		}
		defer file.Close()
		c.frontMatter, c.body, err = parser.Parse(file)
	})
	return err
}

func (c *originalContent) FrontMatter() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.frontMatter, nil
}

func (c *originalContent) Body() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.body, nil
}

func (c *originalContent) Full() (string, error) {
	frontMatter, err := c.FrontMatter()
	if err != nil {
		return "", err
	}
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	return frontMatter + body, nil
}

// TODO this code has multiple problems:
// 1. When two go-routines runs text() and setText() then the result is unknown
// 2. When setText is executed first and then text() overrides already modified body
type body struct {
	body     string
	once     sync.Once
	original func() (string, error)
}

func (t *body) text() (string, error) {
	var err error
	t.once.Do(func() {
		var body string
		body, err = t.original()
		if err != nil {
			return
		}
		t.body = body
	})
	return t.body, err
}

func (t *body) setText(body string) {
	t.body = body
}
