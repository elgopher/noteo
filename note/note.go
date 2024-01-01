package note

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elgopher/noteo/tag"
)

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
	newBody, err := replaceLinks{notePath: n.path, from: from, to: to}.run(body)
	if err != nil {
		return err
	}
	n.body.setText(newBody)
	return nil
}

type replaceLinks struct {
	notePath string
	from, to string
}

func (u replaceLinks) run(body string) (newBody string, returnedError error) {
	relativeFrom, err := u.relativePath(u.from)
	if err != nil {
		return "", err
	}
	markdownLinkRegexp := regexp.MustCompile(`(\[[^][]+])\(([^()]+)\)`) // TODO does not take into account code fences
	newBody = markdownLinkRegexp.ReplaceAllStringFunc(body, func(s string) string {
		linkPath := markdownLinkRegexp.FindStringSubmatch(s)[2]
		relativeLinkPath, err := u.relativePath(linkPath)
		if err != nil {
			returnedError = err
			return s
		}
		if relativeFrom == relativeLinkPath {
			return markdownLinkRegexp.ReplaceAllString(s, `$1(`+u.to+`)`)
		}
		if isAncestorPath(relativeFrom, relativeLinkPath) {
			newTo := u.to + strings.TrimPrefix(relativeLinkPath, relativeFrom)
			newTo = filepath.ToSlash(newTo) // always use slashes, even on Windows
			return markdownLinkRegexp.ReplaceAllString(s, `$1(`+newTo+`)`)
		}
		return s
	})
	return
}

// Returns relative path to note path
func (u replaceLinks) relativePath(p string) (string, error) {
	dir := filepath.Dir(u.notePath)
	relativePath := p
	if !filepath.IsAbs(relativePath) {
		relativePath = filepath.Join(dir, p)
	}
	return filepath.Rel(dir, relativePath)
}

func isAncestorPath(relativeAncestorPath string, relativeDescendantPath string) bool {
	rel, err := filepath.Rel(relativeAncestorPath, relativeDescendantPath)
	if err != nil {
		return false
	}
	return filepath.Dir(rel) == "."
}

// Save returns true if file was modified.
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
	if err := os.WriteFile(n.path, newBytes, 0664); err != nil {
		return false, err
	}
	return true, nil
}

type body struct {
	body     *string
	original func() (string, error)
	mutex    sync.Mutex
}

func (t *body) text() (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.body == nil {
		body, err := t.original()
		if err != nil {
			return "", err
		}
		t.body = &body
	}
	return *t.body, nil
}

func (t *body) setText(body string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.body = &body
}
