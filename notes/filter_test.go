package notes_test

import (
	"context"
	"testing"
	"time"

	"github.com/jacekolszak/noteo/notes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	expectedNote := &noteMock{}
	filteredOutNote1 := &noteMock{}
	filteredOutNote2 := &noteMock{}

	t.Run("should remove not matched notes", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		notesChannel := make(chan notes.Note, 3)
		notesChannel <- filteredOutNote1
		notesChannel <- expectedNote
		notesChannel <- filteredOutNote2
		// when
		filtered, errors := notes.Filter(ctx, notesChannel, func(note notes.Note) (bool, error) {
			if expectedNote == note {
				return true, nil
			}
			return false, nil
		})
		close(notesChannel)
		// then
		var output []notes.Note
	main:
		for {
			select {
			case n, ok := <-filtered:
				if !ok {
					break main
				}
				output = append(output, n)
			case err := <-errors:
				require.NoError(t, err)
			}
		}
		assert.Equal(t, []notes.Note{expectedNote}, output)
	})
}
