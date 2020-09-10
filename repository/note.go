package repository

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/tag"
	"gopkg.in/yaml.v2"
)

var tagSeparator = regexp.MustCompile(`[,\s]+`)

type Note struct {
	path        string
	modified    time.Time
	content     *content
	frontMatter *frontMatter
}

func newNote(path string, modified time.Time) *Note {
	content := &content{
		path: path,
	}
	frontMatter := &frontMatter{
		path:     path,
		meta:     content.Meta,
		mapSlice: mapSlice{},
	}
	return &Note{
		path:        path,
		modified:    modified,
		content:     content,
		frontMatter: frontMatter,
	}
}

func (n *Note) Path() string {
	return n.path
}

func (n *Note) Modified() time.Time {
	return n.modified
}

func (n *Note) Created() (time.Time, error) {
	return n.frontMatter.Created()
}

func (n *Note) Tags() ([]tag.Tag, error) {
	return n.frontMatter.Tags()
}

func (n *Note) Text() (string, error) {
	return n.content.Body()
}

func (n *Note) setTag(newTag tag.Tag) error {
	return n.frontMatter.setTag(newTag)
}

func (n *Note) unsetTag(newTag tag.Tag) error {
	return n.frontMatter.unsetTag(newTag)
}

func (n *Note) unsetTagRegex(regex *regexp.Regexp) error {
	return n.frontMatter.unsetTagRegex(regex)
}

func (n *Note) updateLink(from, to string) error {
	body, err := n.content.Body()
	if err != nil {
		return err
	}
	linkRegexp := regexp.MustCompile(`(\[[^][]+])\(([^()]+)\)`) // TODO does not take into account code fences
	body = linkRegexp.ReplaceAllStringFunc(body, func(s string) string {
		pth := linkRegexp.FindStringSubmatch(s)[2]
		joinedPath := filepath.Join(filepath.Dir(n.path), pth)
		if from == joinedPath {
			newTo, err := filepath.Rel(filepath.Dir(n.path), to)
			if err != nil {
				panic(err)
			}
			return linkRegexp.ReplaceAllString(s, `$1(`+newTo+`)`)
		}
		return s
	})
	n.content.body = body
	return nil
}

func (n *Note) save() (bool, error) {
	frontMatter, err := n.frontMatter.marshal()
	if err != nil {
		return false, err
	}
	text, err := n.Text()
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(text, "\n") {
		text = "\n\n" + text
	}
	newBytes := []byte(frontMatter + text)
	if bytes.Equal(n.content.bytes, newBytes) {
		return false, nil
	}
	if err := ioutil.WriteFile(n.path, newBytes, 0664); err != nil {
		return false, err
	}
	return true, nil
}

type content struct {
	path  string
	once  sync.Once
	bytes []byte
	meta  string
	body  string
}

func (c *content) ensureLoaded() error {
	var err error
	c.once.Do(func() {
		bytesRead, e := ioutil.ReadFile(c.path)
		if e != nil {
			err = fmt.Errorf("%s ReadFile failed: %v", c.path, e)
			return
		}
		c.bytes = bytesRead
		text := string(bytesRead)
		c.body = text
		yamlDivider := "---"
		yamlDividerLen := len(yamlDivider)
		if strings.HasPrefix(text, yamlDivider) {
			index := strings.Index(text[yamlDividerLen:], yamlDivider)
			c.meta = text[yamlDividerLen : index+yamlDividerLen*2]
			c.body = text[index+yamlDividerLen*2:]
		}
	})
	return err
}

func (c *content) Bytes() ([]byte, error) {
	if err := c.ensureLoaded(); err != nil {
		return nil, err
	}
	return c.bytes, nil
}

func (c *content) Meta() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.meta, nil
}

func (c *content) Body() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.body, nil
}

type frontMatter struct {
	path     string
	meta     func() (string, error)
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

func (h *frontMatter) ensureParsed() error {
	var err error
	h.once.Do(func() {
		meta, e := h.meta()
		if e != nil {
			err = e
			return
		}
		if e := yaml.Unmarshal([]byte(meta), &h.mapSlice); e != nil {
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
		if oldTag == newTag {
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
	h.mapSlice = h.mapSlice.set("Tags", serializedTags)
	marshaledBytes, err := yaml.Marshal(h.mapSlice)
	if err != nil {
		return "", err
	}
	return "---\n" + string(marshaledBytes) + "---", nil
}
