package notes

import (
	"time"

	"github.com/jacekolszak/noteo/tag"
)

type Note interface {
	Modified() time.Time
	Created() (time.Time, error)
	Path() string
	Tags() ([]tag.Tag, error)
	Text() (string, error)
}

func FindTagByName(note Note, name string) (tag.Tag, bool, error) {
	tags, err := note.Tags()
	if err != nil {
		return "", false, err
	}
	for _, t := range tags {
		if t.Name() == name {
			return t, true, nil
		}
	}
	return "", false, nil
}
