package output

import "github.com/elgopher/noteo/notes"

func StringTags(note notes.Note) ([]string, error) {
	var ret []string
	tags, err := note.Tags()
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		ret = append(ret, t.String())
	}
	return ret, nil
}
