package notes_test

import (
	"context"
	"testing"
	"time"

	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTop(t *testing.T) {
	layout := "2006"
	date2019, err := time.Parse(layout, "2019")
	require.NoError(t, err)
	date2020, err := time.Parse(layout, "2020")
	require.NoError(t, err)
	date2021, err := time.Parse(layout, "2021")
	require.NoError(t, err)
	note2019 := &noteMock{
		modified: date2019,
	}
	note2020 := &noteMock{
		modified: date2020,
	}
	note2021 := &noteMock{
		modified: date2021,
	}

	t.Run("should sort", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		notesChannel := make(chan notes.Note, 3)
		notesChannel <- note2020
		notesChannel <- note2021
		notesChannel <- note2019
		// when
		topNotes, errors := notes.Top(ctx, 10, notesChannel, sortByModified)
		close(notesChannel)
		// then
		output := collectNotes(t, topNotes, errors)
		assert.Equal(t, []notes.Note{note2019, note2020, note2021}, output)
	})

	t.Run("should limit", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		notesChannel := make(chan notes.Note, 2)
		notesChannel <- note2020
		notesChannel <- note2021
		// when
		topNotes, errors := notes.Top(ctx, 1, notesChannel, sortByModified)
		close(notesChannel)
		// then
		output := collectNotes(t, topNotes, errors)
		assert.Equal(t, []notes.Note{note2020}, output)
	})
}

func sortByModified(i, j notes.Note) (bool, error) {
	iModified, _ := i.Modified()
	jModified, _ := j.Modified()
	return iModified.Before(jModified), nil
}

func TestTagDateAsc(t *testing.T) {
	note2019 := &noteMock{
		tags: []string{
			"deadline:2019-01-01",
		},
	}
	note2020 := &noteMock{
		tags: []string{
			"deadline:2020-03-03",
		},
	}

	t.Run("should return Less function sorting dates on specific tag", func(t *testing.T) {
		tests := map[string]struct {
			left, right notes.Note
			expected    bool
		}{
			"note2019 < note2020": {
				left: note2019, right: note2020, expected: true,
			},
			"note2020 < note2019": {
				left: note2020, right: note2019, expected: false,
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				less := notes.TagDateAsc("deadline")
				isLess, err := less(test.left, test.right)
				require.NoError(t, err)
				assert.Equal(t, test.expected, isLess)
			})
		}
	})
}

func TestTagNumberAsc(t *testing.T) {
	noteWithPriority1 := &noteMock{
		tags: []string{
			"priority:1",
		},
	}
	noteWithPriority2 := &noteMock{
		tags: []string{
			"priority:2",
		},
	}

	t.Run("should return Less function sorting dates on specific tag", func(t *testing.T) {
		tests := map[string]struct {
			left, right notes.Note
			expected    bool
		}{
			"priority:1 < priority:2": {
				left: noteWithPriority1, right: noteWithPriority2, expected: true,
			},
			"priority:2 > priority:1": {
				left: noteWithPriority2, right: noteWithPriority1, expected: false,
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				less := notes.TagNumberAsc("priority")
				isLess, err := less(test.left, test.right)
				require.NoError(t, err)
				assert.Equal(t, test.expected, isLess)
			})
		}
	})
}

type noteMock struct {
	modified   time.Time
	created    time.Time
	path       string
	tags       []string
	stringTags []string
	text       string
}

func (n *noteMock) Modified() (time.Time, error) {
	return n.modified, nil
}

func (n *noteMock) Created() (time.Time, error) {
	return n.created, nil
}

func (n *noteMock) Path() string {
	return n.path
}

func (n *noteMock) Tags() ([]tag.Tag, error) {
	tags := make([]tag.Tag, len(n.tags))
	for _, name := range n.tags {
		newTag, err := tag.New(name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, newTag)
	}
	return tags, nil
}

func (n *noteMock) StringTags() ([]string, error) {
	return n.stringTags, nil
}
func (n *noteMock) Text() (string, error) {
	return n.text, nil
}
