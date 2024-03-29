package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elgopher/noteo/note"
	"github.com/elgopher/noteo/repository"
)

func TestRepository_Add(t *testing.T) {
	t.Run("should generate name", func(t *testing.T) {
		tests := map[string]struct {
			content          string
			expectedFilename string
		}{
			"note with YAML front matter black": {
				content: `---
Tags: ""
---

yaml`,
				expectedFilename: "yaml.md",
			},
			"note with YAML without blank line:": {
				content: `---
Tags: ""
---
yaml-without-blank-line`,
				expectedFilename: "yaml-without-blank-line.md",
			},
			"should remove illegal chars": {
				content:          "chars`~!@#$%^&*()_+[{]}\\|;:'\",<>/?",
				expectedFilename: "chars.md",
			},
			"should replace diacritics with ascii": {
				content:          "ąćęłńóśżź",
				expectedFilename: "acelnoszz.md",
			},
			"should replace space with dash": {
				content:          "foo bar",
				expectedFilename: "foo-bar.md",
			},
			"should trim spaces": {
				content:          " trimspaces ",
				expectedFilename: "trimspaces.md",
			},
			"should trim spaces after removing illegal chars": {
				content:          "% removedchars @",
				expectedFilename: "removedchars.md",
			},
			"should trim newlines": {
				content:          "\nnewlines\n",
				expectedFilename: "newlines.md",
			},
			"should lowercase": {
				content:          "LOWERCASE",
				expectedFilename: "lowercase.md",
			},
			"unknown": {
				content:          " @#",
				expectedFilename: "unknown.md",
			},
			"unknown with front matter": {
				content: `---
Tags: ""
---

`,
				expectedFilename: "unknown.md",
			},
			"yaml without break line": {
				content: `---
Tags: ""
---`,
				expectedFilename: "unknown.md",
			},
			"should use only first line of body as filename": {
				content: `---
Tags: ""
---

first
second`,
				expectedFilename: "first.md",
			},
			"should use capital letter word for very long text": {
				content:          "Very long text and only This word will be used as title",
				expectedFilename: "this.md",
			},
			"should trim very long text": {
				content:          "this is very long text without any capital letter words",
				expectedFilename: "this-is-very-long-text-without.md",
			},
			"should remove carriage return": {
				content:          "return\r",
				expectedFilename: "return.md",
			},
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				_, repo := repo(t)
				// when
				file, err := repo.Add(test.content)
				// then
				require.NoError(t, err)
				assert.Equal(t, test.expectedFilename, file)
			})
		}
	})
}

func TestRepository_Move(t *testing.T) {
	t.Run("should rename file", func(t *testing.T) {
		dir, repo := repo(t)
		require.NoError(t, os.Chdir(dir))
		writeFile(t, "source.md", "source")
		linkFile := filepath.Join(dir, "link.md")
		writeFile(t, linkFile, "[link](source.md)")
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		// when
		notes, success, errors := repo.Move(ctx, "source.md", "target.md")
		// then
		assertSuccess(t, ctx, notes, success, errors)
		assert.NoFileExists(t, "source.md")
		assert.FileExists(t, "target.md")
		assertFileEquals(t, linkFile, "[link](target.md)")
	})

	t.Run("should move file to directory", func(t *testing.T) {
		dir, repo := repo(t)
		require.NoError(t, os.Chdir(dir))
		require.NoError(t, os.MkdirAll("target", os.ModePerm))
		writeFile(t, "source.md", "source")
		linkFile := filepath.Join(dir, "link.md")
		writeFile(t, linkFile, "[link](source.md)")
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		// when
		notes, success, errors := repo.Move(ctx, "source.md", "target")
		// then
		assertSuccess(t, ctx, notes, success, errors)
		assert.NoFileExists(t, "source.md")
		assert.FileExists(t, filepath.Join("target", "source.md"))
		assertFileEquals(t, linkFile, "[link](target/source.md)")
	})

	t.Run("should move whole dir", func(t *testing.T) {
		dir, repo := repo(t)
		require.NoError(t, os.Chdir(dir))
		require.NoError(t, os.MkdirAll("source", os.ModePerm))
		require.NoError(t, os.MkdirAll("target", os.ModePerm))
		sourceDir := filepath.Join(dir, "source")
		sourceFile := filepath.Join(sourceDir, "foo.md")
		writeFile(t, sourceFile, "foo")
		linkFile := filepath.Join(dir, "link.md")
		writeFile(t, linkFile, "[link](source/foo.md)")
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		// when
		notes, success, errors := repo.Move(ctx, "source", "target")
		// then
		assertSuccess(t, ctx, notes, success, errors)
		assert.NoDirExists(t, sourceDir)
		assert.DirExists(t, filepath.Join(dir, "target", "source"))
		assert.FileExists(t, filepath.Join(dir, "target", "source", "foo.md"))
		assertFileEquals(t, linkFile, "[link](target/source/foo.md)")
	})
}

func assertSuccess(t *testing.T, ctx context.Context, notes <-chan *note.Note, success <-chan bool, errors <-chan error) {
	var successClosed, errorClosed, notesClosed bool
	for !successClosed || !errorClosed || !notesClosed {
		select {
		case _, ok := <-notes:
			if !ok {
				notesClosed = true
				continue
			}
		case e, ok := <-errors:
			if !ok {
				errorClosed = true
				continue
			}
			require.FailNowf(t, "error received", "%v", e)
		case <-ctx.Done():
			require.FailNow(t, "timeout")
		case s, ok := <-success:
			if !ok {
				successClosed = true
				continue
			}
			assert.True(t, s)
		}
	}
}

func assertFileEquals(t *testing.T, file, expected string) {
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, expected, string(content))
}

func repo(t *testing.T) (dir string, repo *repository.Repository) {
	dir, err := os.MkdirTemp("", "noteo-test")
	require.NoError(t, err)
	_, err = repository.Init(dir)
	require.NoError(t, err)
	repo, err = repository.ForWorkDir(dir)
	require.NoError(t, err)
	return dir, repo
}

func writeFile(t *testing.T, filename, content string) {
	require.NoError(t, os.WriteFile(filename, []byte(content), os.ModePerm))
}
