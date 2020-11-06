package note_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/jacekolszak/noteo/note"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should return note for missing file", func(t *testing.T) {
		assert.NotNil(t, note.New("missing"))
	})
}

func TestNewWithModified(t *testing.T) {
	t.Run("should return note for missing file", func(t *testing.T) {
		assert.NotNil(t, note.NewWithModified("missing", time.Now()))
	})
}

func TestNote_Modified(t *testing.T) {
	t.Run("should return error for missing file", func(t *testing.T) {
		n := note.New("missing")
		// when
		_, err := n.Modified()
		// then
		assert.Error(t, err)
	})

	t.Run("should not error for existing file", func(t *testing.T) {
		file, err := ioutil.TempFile("", "noteo-test")
		require.NoError(t, err)
		n := note.New(file.Name())
		// when
		_, err = n.Modified()
		// then
		assert.NoError(t, err)
	})

	t.Run("should return passed modified time", func(t *testing.T) {
		givenTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
		require.NoError(t, err)
		n := note.NewWithModified("path", givenTime)
		// when
		modified, err := n.Modified()
		// then
		require.NoError(t, err)
		assert.Equal(t, givenTime, modified)
	})
}
