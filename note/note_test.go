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

func TestNote_Path(t *testing.T) {
	t.Run("should return path", func(t *testing.T) {
		n := note.New("path")
		// expect
		assert.Equal(t, "path", n.Path())
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

func TestNote_Created(t *testing.T) {
	t.Run("should return zero value Time (1 Jan 1970) for note without Created tag", func(t *testing.T) {
		filename := writeTempFile(t, "body")
		n := note.New(filename)
		// when
		created, err := n.Created()
		// then
		require.NoError(t, err)
		assert.Equal(t, time.Time{}, created)
	})

	t.Run("should return time from Created tag", func(t *testing.T) {
		filename := writeTempFile(t, "---\nCreated: 2006-01-02\n---\nbody")
		n := note.New(filename)
		// when
		created, err := n.Created()
		// then
		require.NoError(t, err)
		expectedTime, err := date.Parse("2006-01-02")
		require.NoError(t, err)
		assert.Equal(t, expectedTime, created)
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

	t.Run("should update existing tag", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: tag:1\n---\nbody")
		n := note.New(filename)
		newTag, err := tag.New("tag:2")
		require.NoError(t, err)
		// when
		err = n.SetTag(newTag)
		// then
		require.NoError(t, err)
		assertTags(t, n, "tag:2")
	})
}

func TestNote_RemoveTag(t *testing.T) {
	t.Run("should remove tag", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: tag another\n---\ntext")
		n := note.New(filename)
		tagToRemove := newTag(t, "tag")
		// when
		err := n.RemoveTag(tagToRemove)
		// then
		require.NoError(t, err)
		assertTags(t, n, "another")
	})

	t.Run("should remove last remaining tag", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: tag\n---\ntext")
		n := note.New(filename)
		tagToRemove := newTag(t, "tag")
		// when
		err := n.RemoveTag(tagToRemove)
		// then
		require.NoError(t, err)
		assertNoTags(t, n)
	})

	t.Run("removing missing tag does nothing", func(t *testing.T) {
		filename := writeTempFile(t, "content")
		n := note.New(filename)
		missingTag := newTag(t, "missing")
		// when
		err := n.RemoveTag(missingTag)
		// then
		require.NoError(t, err)
		assertNoTags(t, n)
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

	t.Run("should not save file if nothing changed", func(t *testing.T) {
		filename := writeTempFile(t, "text")
		n := note.New(filename)
		// when
		saved, err := n.Save()
		// then
		assert.False(t, saved)
		assert.NoError(t, err)
	})
}

func TestNote_Body(t *testing.T) {
	t.Run("should return body when note does not have a front matter", func(t *testing.T) {
		filename := writeTempFile(t, "body")
		n := note.New(filename)
		// when
		actual, err := n.Body()
		// then
		require.NoError(t, err)
		assert.Equal(t, "body", actual)
	})

	t.Run("should return body when note has a front matter", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: tag\n---\nbody")
		n := note.New(filename)
		// when
		actual, err := n.Body()
		// then
		require.NoError(t, err)
		assert.Equal(t, "body", actual)
	})

	t.Run("should return empty body", func(t *testing.T) {
		filename := writeTempFile(t, "---\nTags: tag\n---\n")
		n := note.New(filename)
		// when
		actual, err := n.Body()
		// then
		require.NoError(t, err)
		assert.Empty(t, actual)
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

func assertNoTags(t *testing.T, n *note.Note) {
	assertTags(t, n)
}

func assertTags(t *testing.T, n *note.Note, expectedTags ...string) {
	tags, err := n.Tags()
	require.NoError(t, err)
	require.Equal(t, len(expectedTags), len(tags), "different tags len")
	for i, expectedTag := range expectedTags {
		assert.Equal(t, newTag(t, expectedTag), tags[i])
	}
}

func newTag(t *testing.T, name string) tag.Tag {
	createdTag, err := tag.New(name)
	require.NoError(t, err)
	return createdTag
}
