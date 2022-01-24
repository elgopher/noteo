package cmd

import (
	"bufio"
	"fmt"

	"github.com/elgopher/noteo/repository"
	"github.com/spf13/cobra"
)

func tagRm() *cobra.Command {
	var (
		name  string
		stdin bool
		grep  string
	)
	tagRm := &cobra.Command{
		Use:   "rm",
		Short: "Remove tags from notes",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := workingDirRepository()
			if err != nil {
				return err
			}
			untagFile, err := untagFileFunc(name, grep, repo)
			if err != nil {
				return err
			}
			removeFileTags := func(file string) {
				updated, err := untagFile(file)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "skipping:", err)
				} else if updated {
					printer := NewPrinter()
					printer.PrintFile(file)
					printer.Println(" updated")
				}
			}
			if stdin {
				scanner := bufio.NewScanner(cmd.InOrStdin())
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					file := scanner.Text()
					removeFileTags(file)
				}
				return nil
			}
			for _, file := range args {
				removeFileTags(file)
			}
			return nil
		},
	}
	tagRm.Flags().BoolVarP(&stdin, "stdin", "", false, "read file names from standard input")
	tagRm.Flags().StringVarP(&name, "name", "n", "", "short name without space. Can have form of name:number-or-date")
	tagRm.Flags().StringVar(&grep, "grep", "", "name regular expression")
	return tagRm
}

type untagFile func(file string) (bool, error)

func untagFileFunc(name string, grep string, repo *repository.Repository) (untagFile, error) {
	if name == "" && grep == "" {
		return nil, fmt.Errorf("no name given using -n flag or regex with --grep flag")
	}
	if name != "" {
		return func(file string) (bool, error) {
			return repo.UntagFile(file, name)
		}, nil
	} else {
		return func(file string) (bool, error) {
			return repo.UntagFileRegex(file, grep)
		}, nil
	}
}
