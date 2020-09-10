package repository

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jacekolszak/noteo/tag"
	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
)

func Init(dir string) error {
	_, err := ForWorkDir(dir)
	if err == nil {
		return errors.New("repo already initialized")
	}
	e, ok := err.(repoError)
	if ok && e.IsNotRepository() {
		return ioutil.WriteFile(dotFile(dir), []byte(""), 0664)
	}
	return err
}

func ForWorkDir(dir string) (*Repository, error) {
	root, err := findRoot(dir)
	if err != nil {
		return nil, err
	}
	return &Repository{root: root, dir: dir}, nil
}

func (r *Repository) WorkDir() string {
	return r.dir
}

type Repository struct {
	root string
	dir  string
}

func (r *Repository) NewFileTemplate(created time.Time) string {
	return `---
Created: ` + created.Format(time.UnixDate) + `
Tags: ""
---

`
}

func (r *Repository) Add(text string) (string, error) {
	name, err := generateFilename(text)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(r.dir, os.ModePerm); err != nil {
		return "", err
	}
	file := filepath.Join(r.dir, name+".md")
	_, err = os.Stat(file)
	existingFile := err == nil
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if existingFile {
		file = filepath.Join(r.dir, name+"-"+generateUUID()[:7]+".md")
	}
	rel, err := filepath.Rel(r.dir, file)
	if err != nil {
		return file, err
	}
	return rel, ioutil.WriteFile(file, []byte(text), 0664)
}

func (r *Repository) TagFileWith(file string, newTag string) (bool, error) {
	t, err := tag.New(newTag)
	if err != nil {
		return false, err
	}
	normalizedTag, err := t.MakeDateAbsolute()
	if err == nil {
		t = normalizedTag
	}
	if filepath.Ext(file) != ".md" {
		return false, fmt.Errorf("%s has no *.md extension", file)
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(r.dir, file)
	}
	note := newNote(file, time.Now())
	if err := note.setTag(t); err != nil {
		return false, err
	}
	return note.save()
}

func (r *Repository) UntagFile(file string, tagToRemove string) (bool, error) {
	t, err := tag.New(tagToRemove)
	if err != nil {
		return false, err
	}
	if filepath.Ext(file) != ".md" {
		return false, fmt.Errorf("%s has no *.md extension", file)
	}
	note := newNote(file, time.Now())
	if err := note.unsetTag(t); err != nil {
		return false, err
	}
	return note.save()
}

func (r *Repository) UntagFileRegex(file string, tagRegexToRemove string) (bool, error) {
	regex, err := regexp.Compile(tagRegexToRemove)
	if err != nil {
		return false, err
	}
	if filepath.Ext(file) != ".md" {
		return false, fmt.Errorf("%s has no *.md extension", file)
	}
	note := newNote(file, time.Now())
	if err := note.unsetTagRegex(regex); err != nil {
		return false, err
	}
	return note.save()
}

func (r *Repository) Move(ctx context.Context, from, to string) (<-chan *Note, <-chan bool, <-chan error) {
	notes, notesErr := r.AllNotes(ctx)

	updated := make(chan *Note)
	errs := make(chan error)
	success := make(chan bool)

	go func() {
		defer close(updated)
		defer close(errs)
		defer close(success)
		defer func() {
			err := os.Rename(from, to)
			if err != nil {
				errs <- err
				success <- false
				return
			}
			success <- true
		}()
		lstat, err := os.Lstat(to)
		if err != nil && !os.IsNotExist(err) {
			errs <- err
			return
		}
		if !os.IsNotExist(err) && lstat.IsDir() {
			_, file := filepath.Split(from)
			to = filepath.Join(to, file)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-notesErr:
				if !ok {
					continue
				}
				errs <- err
			case note, ok := <-notes:
				if !ok {
					return
				}
				err := note.updateLink(from, to)
				if err != nil {
					errs <- err
					continue
				}
				ok, err = note.save()
				if err != nil {
					errs <- err
					continue
				}
				if ok {
					updated <- note
				}
			}
		}
	}()
	return updated, success, errs
}

func (r *Repository) Notes(ctx context.Context) (<-chan *Note, <-chan error) {
	return r.notes(ctx, r.dir)
}

func (r *Repository) AllNotes(ctx context.Context) (<-chan *Note, <-chan error) {
	return r.notes(ctx, r.root)
}

func (r *Repository) notes(ctx context.Context, dir string) (<-chan *Note, <-chan error) {
	names := make(chan *Note)
	errs := make(chan error)
	go func() {
		defer close(names)
		defer close(errs)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("cancelled")
			default:
				if strings.HasSuffix(path, ".md") {
					relPath, err := filepath.Rel(r.dir, path)
					if err != nil {
						return err
					}
					names <- newNote(relPath, info.ModTime())
				}
			}
			return nil
		})
		if err != nil {
			errs <- err
		}
	}()
	return names, errs
}

func (r *Repository) Tags(ctx context.Context) (<-chan tag.Tag, <-chan error) {
	tags := make(chan tag.Tag)
	errs := make(chan error)
	notes, notesErrs := r.Notes(ctx)
	go func() {
		defer close(tags)
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-notesErrs:
				if !ok {
					return
				}
				errs <- err
			case note, ok := <-notes:
				if !ok {
					return
				}
				noteTags, err := note.Tags()
				if err != nil {
					errs <- err
					continue
				}
				for _, t := range noteTags {
					tags <- t
				}
			}
		}
	}()
	return tags, errs
}

func generateFilename(text string) (string, error) {
	name := stripFrontMatter(text)
	name = strings.TrimLeft(name, "\n")
	if strings.Contains(name, "\n") {
		name = name[:strings.Index(name, "\n")]
	}
	name = strings.Trim(name, "\n")
	// Not yet decided what to do in this case:
	//if name == "" {
	//	return "", errors.New("empty note")
	//}
	name = godiacritics.Normalize(name)
	notAllowedChars := regexp.MustCompile(`[^a-zA-Z0-9.\- ]`)
	name = notAllowedChars.ReplaceAllString(name, "")
	name = strings.Trim(name, " ")
	if len(name) > 30 {
		scanner := bufio.NewScanner(strings.NewReader(name))
		scanner.Split(bufio.ScanWords)
		scanner.Scan() // skip first word
		for scanner.Scan() {
			word := scanner.Text()
			r := rune(word[0])
			if unicode.IsUpper(r) {
				name = word
				break
			}
		}
	}
	if len(name) > 30 {
		name = name[:30]
	}
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	if name == "" {
		name = "unknown"
	}
	return name, nil
}

func stripFrontMatter(text string) string {
	yamlDivider := "---"
	yamlDividerLen := len(yamlDivider)
	if strings.HasPrefix(text, yamlDivider) {
		index := strings.Index(text[yamlDividerLen:], yamlDivider)
		if index > 0 {
			return text[index+yamlDividerLen*2:]
		}
	}
	return text
}

func generateUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func dotFile(dir string) string {
	return filepath.Join(dir, ".noteo")
}

func findRoot(dir string) (string, error) {
	f := dotFile(dir)
	_, err := os.Stat(f)
	if err == nil {
		return dir, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	} else {
		parent := filepath.Dir(dir)
		if dir == parent {
			return "", repoError("repo not initialized: .noteo file note found")
		}
		return findRoot(parent)
	}
}

type repoError string

func (g repoError) Error() string {
	return string(g)
}

func (g repoError) IsNotRepository() bool {
	return true
}
