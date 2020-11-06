package note_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/note"
	"github.com/jacekolszak/noteo/tag"
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

func TestNote_SetTag(t *testing.T) {
	t.Run("should add tag for file without front matter", func(t *testing.T) {
		filename := writeTempFile(t, "text")
		n := note.New(filename)
		newTag, err := tag.New("tag")
		require.NoError(t, err)
		// when
		err = n.SetTag(newTag)
		// then
		require.NoError(t, err)
		assertTags(t, n, "tag")
	})

	t.Run("should set tag with relative date", func(t *testing.T) {
		date.SetNow(func() time.Time {
			return time.Date(2020, 9, 10, 16, 30, 11, 0, time.FixedZone("CEST", 60*60*2))
		})
		defer date.SetNow(time.Now)
		filename := writeTempFile(t, "")
		n := note.New(filename)
		// when
		err := n.SetTag(newTag(t, "deadline:now"))
		// then
		require.NoError(t, err)
		assertTags(t, n, "deadline:2020-09-10T16:30:11+02:00")
	})
}

func TestNote_Save(t *testing.T) {
	t.Run("should add yaml front matter for file without it", func(t *testing.T) {
		filename := writeTempFile(t, "text")
		n := note.New(filename)
		require.NoError(t, n.SetTag(newTag(t, "tag")))
		// when
		saved, err := n.Save()
		// then
		require.NoError(t, err)
		assert.True(t, saved)
		// and
		assertFileEquals(t, filename, "---\nTags: tag\n---\ntext")
	})

	t.Run("should update front matter", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: foo\n---\n\ntext")
		n := note.New(filename)
		require.NoError(t, n.SetTag(newTag(t, "tag")))
		// when
		saved, err := n.Save()
		// then
		require.NoError(t, err)
		assert.True(t, saved)
		// and
		assertFileEquals(t, filename, "---\nTags: foo tag\n---\n\ntext")
	})
}

func writeTempFile(t *testing.T, content string) string {
	file, err := ioutil.TempFile("", "noteo-test")
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(file.Name(), []byte(content), os.ModePerm))
	return file.Name()
}

func assertFileEquals(t *testing.T, filename, expectedContent string) {
	bytes, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, string(bytes))
}

func assertTags(t *testing.T, n *note.Note, expectedTags ...string) {
	tags, err := n.Tags()
	require.NoError(t, err)
	for i, expectedTag := range expectedTags {
		assert.Equal(t, newTag(t, expectedTag), tags[i])
	}
}

func newTag(t *testing.T, name string) tag.Tag {
	createdTag, err := tag.New(name)
	require.NoError(t, err)
	return createdTag
}
