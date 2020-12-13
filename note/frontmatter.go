package note

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/tag"
	"gopkg.in/yaml.v2"
)

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
			tagsSlice, e := parseTags(tags)
			if e != nil {
				err = e
				return
			}
			h.tags = append(h.tags, tagsSlice...)
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

func parseTags(tags interface{}) ([]tag.Tag, error) {
	var result []tag.Tag
	for _, t := range stringTags(tags) {
		t = strings.Trim(t, " ")
		ta, err := tag.New(t)
		if err != nil {
			return nil, err
		}
		result = append(result, ta)
	}
	return result, nil
}

func stringTags(tags interface{}) []string {
	tagSeparator := regexp.MustCompile(`[,\s]+`)
	var stringTags []string
	switch v := tags.(type) {
	case string:
		stringTags = tagSeparator.Split(v, -1)
		if stringTags[0] == "" {
			stringTags = stringTags[1:]
		}
	case []interface{}:
		for _, s := range v {
			stringTags = append(stringTags, fmt.Sprintf("%s", s))
		}
	}
	return stringTags
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
	normalizedTag, err := newTag.MakeDateAbsolute()
	if err == nil {
		newTag = normalizedTag
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

func (h *frontMatter) removeTag(newTag tag.Tag) error {
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

func (h *frontMatter) removeTagRegex(regex *regexp.Regexp) error {
	if err := h.ensureParsed(); err != nil {
		return err
	}
	for i, oldTag := range h.tags {
		if regex.MatchString(oldTag.String()) {
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
		stringTags = append(stringTags, t.String())
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
