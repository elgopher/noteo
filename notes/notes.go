package notes

import (
	"time"

	"github.com/elgopher/noteo/tag"
)

type Note interface {
	Modified() (time.Time, error)
	Created() (time.Time, error)
	Path() string
	Tags() ([]tag.Tag, error)
	Body() (string, error)
}

func FindTagByName(note Note, name string) (tag.Tag, bool, error) {
	tags, err := note.Tags()
	if err != nil {
		return tag.Tag{}, false, err
	}
	for _, t := range tags {
		if t.Name() == name {
			return t, true, nil
		}
	}
	return tag.Tag{}, false, nil
}
