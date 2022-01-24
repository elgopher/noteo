package yml

import (
	"time"

	"github.com/elgopher/noteo/notes"
	"github.com/elgopher/noteo/output"
	"gopkg.in/yaml.v2"
)

type Formatter struct{}

func (f Formatter) Header() string {
	return ""
}

func (f Formatter) Footer() string {
	return ""
}

func (f Formatter) Note(note notes.Note) string {
	body, err := note.Body()
	if err != nil {
		return err.Error()
	}
	created, err := note.Created()
	if err != nil {
		return err.Error()
	}
	modified, err := note.Modified()
	if err != nil {
		return err.Error()
	}
	tags, err := output.StringTags(note)
	if err != nil {
		return err.Error()
	}
	n := noteToMarshal{
		File:     note.Path(),
		Modified: modified,
		Created:  created,
		Text:     body,
		Tags:     tags,
	}
	bytes, err := yaml.Marshal(n)
	if err != nil {
		return "error marshalling note: " + err.Error()
	}
	return string(bytes) + "\n"
}

type noteToMarshal struct {
	File     string    `yaml:"file"`
	Modified time.Time `yaml:"modified"`
	Created  time.Time `yaml:"created"`
	Tags     []string  `yaml:"tags"`
	Text     string    `yaml:"text"`
}
