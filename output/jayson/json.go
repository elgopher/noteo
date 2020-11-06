package jayson

import (
	"encoding/json"
	"time"

	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/output"
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
	bytes, err := json.Marshal(n)
	if err != nil {
		return "error marshalling note: " + err.Error()
	}
	return string(bytes) + "\n"
}

type noteToMarshal struct {
	File     string    `json:"file"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
	Tags     []string  `json:"tags"`
	Text     string    `json:"text"`
}
