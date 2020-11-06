package note

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/parser"
	"github.com/jacekolszak/noteo/tag"
	"gopkg.in/yaml.v2"
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

func (n *Note) Text() (string, error) {
	return n.body.text()
}

func (n *Note) SetTag(newTag tag.Tag) error {
	return n.frontMatter.setTag(newTag)
}

func (n *Note) UnsetTag(newTag tag.Tag) error {
	return n.frontMatter.unsetTag(newTag)
}

func (n *Note) UnsetTagRegex(regex *regexp.Regexp) error {
	return n.frontMatter.unsetTagRegex(regex)
}

func (n *Note) UpdateLink(from, to string) error {
	body, err := n.body.text()
	if err != nil {
		return err
	}
	markdownLinkRegexp := regexp.MustCompile(`(\[[^][]+])\(([^()]+)\)`) // TODO does not take into account code fences
	body = markdownLinkRegexp.ReplaceAllStringFunc(body, func(s string) string {
		linkPath := markdownLinkRegexp.FindStringSubmatch(s)[2]
		fullLinkPath := filepath.Join(filepath.Dir(n.path), linkPath)
		if strings.HasPrefix(fullLinkPath, from) {
			newTo, err := filepath.Rel(filepath.Dir(n.path), to)
			if err != nil {
				panic(err)
			}
			rel, err := filepath.Rel(from, fullLinkPath)
			if err != nil {
				panic(err)
			}
			newTo = filepath.Join(newTo, rel)
			newTo = filepath.ToSlash(newTo)
			return markdownLinkRegexp.ReplaceAllString(s, `$1(`+newTo+`)`)
		}
		return s
	})
	n.body.setText(body)
	return nil
}

func (n *Note) Save() (bool, error) {
	frontMatter, err := n.frontMatter.marshal()
	if err != nil {
		return false, err
	}
	text, err := n.Text()
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

type frontMatter struct {
	path     string
	original func() (string, error)
	once     sync.Once
	mapSlice mapSlice
	created  time.Time
	tags     []tag.Tag
}

type mapSlice yaml.MapSlice

func (s mapSlice) at(name string) (interface{}, bool) {
	nameLowerCase := strings.ToLower(name)
	for _, item := range s {
		key := fmt.Sprintf("%v", item.Key)
		key = strings.ToLower(key)
		if key == nameLowerCase {
			return item.Value, true
		}
	}
	return nil, false
}

func (s mapSlice) set(name string, val interface{}) mapSlice {
	nameLowerCase := strings.ToLower(name)
	for i, item := range s {
		key := fmt.Sprintf("%v", item.Key)
		key = strings.ToLower(key)
		if key == nameLowerCase {
			s[i] = yaml.MapItem{
				Key:   item.Key,
				Value: val,
			}
			return s
		}
	}
	return append(s, yaml.MapItem{
		Key:   name,
		Value: val,
	})
}

func (s mapSlice) isEmpty() bool {
	return len(s) == 0
}

func (h *frontMatter) ensureParsed() error {
	var err error
	h.once.Do(func() {
		frontMatter, e := h.original()
		if e != nil {
			err = e
			return
		}
		if e := yaml.Unmarshal([]byte(frontMatter), &h.mapSlice); e != nil {
			err = fmt.Errorf("%s YAML front matter unmarshal failed: %v", h.path, e)
			return
		}
		tags, ok := h.mapSlice.at("Tags")
		if ok {
			var tagsSlice []string
			switch v := tags.(type) {
			case string:
				tagsSlice = tagSeparator.Split(v, -1)
				if tagsSlice[0] == "" {
					tagsSlice = tagsSlice[1:]
				}
			case []string:
				tagsSlice = v
			}
			for _, t := range tagsSlice {
				t = strings.Trim(t, " ")
				ta, e := tag.New(t)
				if err != nil {
					err = e
					return
				}
				h.tags = append(h.tags, ta)
			}
		}
		created, ok := h.mapSlice.at("Created")
		if ok {
			createdTime, e := date.ParseAbsolute(created.(string))
			if e == nil {
				h.created = createdTime
			} else {
				err = fmt.Errorf("%s parse failed: %v", h.path, e)
			}
		}
	})
	return err
}

func (h *frontMatter) Created() (time.Time, error) {
	if err := h.ensureParsed(); err != nil {
		return time.Time{}, err
	}
	return h.created, nil
}

func (h *frontMatter) Tags() ([]tag.Tag, error) {
	if err := h.ensureParsed(); err != nil {
		return nil, err
	}
	return h.tags, nil
}

func (h *frontMatter) setTag(newTag tag.Tag) error {
	if err := h.ensureParsed(); err != nil {
		return err
	}
	for i, oldTag := range h.tags {
		if oldTag.Name() == newTag.Name() {
			h.tags[i] = newTag
			return nil
		}
	}
	h.tags = append(h.tags, newTag)
	return nil
}

func (h *frontMatter) unsetTag(newTag tag.Tag) error {
	if err := h.ensureParsed(); err != nil {
		return err
	}
	for i, oldTag := range h.tags {
		if oldTag == newTag {
			h.tags = append(h.tags[:i], h.tags[i+1:]...)
			return nil
		}
	}
	return nil
}

func (h *frontMatter) unsetTagRegex(regex *regexp.Regexp) error {
	if err := h.ensureParsed(); err != nil {
		return err
	}
	for i, oldTag := range h.tags {
		if regex.MatchString(string(oldTag)) {
			h.tags = append(h.tags[:i], h.tags[i+1:]...)
			return nil
		}
	}
	return nil
}

func (h *frontMatter) marshal() (string, error) {
	tags, err := h.Tags()
	if err != nil {
		return "", err
	}
	var stringTags []string
	for _, t := range tags {
		stringTags = append(stringTags, string(t))
	}
	serializedTags := strings.Join(stringTags, " ")
	_, tagsWereGivenBefore := h.mapSlice.at("Tags")
	if len(serializedTags) != 0 || tagsWereGivenBefore {
		h.mapSlice = h.mapSlice.set("Tags", serializedTags)
	}
	marshaledBytes, err := yaml.Marshal(h.mapSlice)
	if err != nil {
		return "", err
	}
	if h.mapSlice.isEmpty() {
		return "", nil
	}
	return "---\n" + string(marshaledBytes) + "---\n", nil
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
