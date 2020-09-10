package quiet

import (
	"github.com/jacekolszak/noteo/notes"
)

type Formatter struct {
}

func (o Formatter) Header() string {
	return ""
}

func (f Formatter) Footer() string {
	return ""
}

func (o Formatter) Note(note notes.Note) string {
	return note.Path() + "\n"
}
