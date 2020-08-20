package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/repository"
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

func toNotes(all <-chan *repository.Note) <-chan notes.Note {
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
