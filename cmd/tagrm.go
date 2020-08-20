package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jacekolszak/noteo/repository"
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
			if name == "" && grep == "" {
				return fmt.Errorf("no name given using -n flag or regex with --grep flag")
			}
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			repo, err := repository.ForWorkDir(wd)
			if err != nil {
				return err
			}
			var untagFile func(file string) (bool, error)
			if name != "" {
				untagFile = func(file string) (bool, error) {
					return repo.UntagFile(file, name)
				}
			} else {
				untagFile = func(file string) (bool, error) {
					return repo.UntagFileRegex(file, grep)
				}
			}
			tagFileWith := func(file string) {
				ok, err := untagFile(file)
				if err != nil {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "skipping:", err)
				} else if ok {
					fmt.Printf("%s updated\n", color.CyanString(file))
				}
			}
			if stdin {
				scanner := bufio.NewScanner(cmd.InOrStdin())
				scanner.Split(bufio.ScanLines)
				for scanner.Scan() {
					file := scanner.Text()
					tagFileWith(file)
				}
				return nil
			}
			for _, file := range args {
				tagFileWith(file)
			}
			return nil
		},
	}
	tagRm.Flags().BoolVarP(&stdin, "stdin", "", false, "read file names from standard input")
	tagRm.Flags().StringVarP(&name, "name", "n", "", "short name without space. Can have form of name:number-or-date")
	tagRm.Flags().StringVar(&grep, "grep", "", "name regular expression")
	return tagRm
}
