package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elgopher/noteo/note"
	"github.com/elgopher/noteo/notes"
	"github.com/elgopher/noteo/repository"
)

func repo(commandArgs []string) (*repository.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if len(commandArgs) > 0 {
		dir := commandArgs[0]
		if filepath.IsAbs(dir) {
			wd = dir
		} else {
			wd = filepath.Join(wd, dir)
		}
	}
	return repository.ForWorkDir(wd)
}

func workingDirRepository() (*repository.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repo, err := repository.ForWorkDir(wd)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func toNotes(all <-chan *note.Note) <-chan notes.Note {
	ret := make(chan notes.Note)
	go func() {
		defer close(ret)
		for {
			note, ok := <-all
			if !ok {
				return
			}
			ret <- note
		}

	}()
	return ret
}

func printErrors(ctx context.Context, errors ...<-chan error) {
	for _, ch := range errors {
		go func(channel <-chan error) {
			for {
				select {
				case <-ctx.Done():
					return
				case warning, ok := <-channel:
					if !ok {
						return
					}
					_, _ = fmt.Fprintln(os.Stderr, warning)
				}
			}
		}(ch)
	}
}
