package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/jacekolszak/noteo/repository"
	"github.com/spf13/cobra"
)

// mv represents the mv command
var mv = &cobra.Command{
	Use:   "mv",
	Args:  cobra.ExactArgs(2),
	Short: "Move note - EXPERIMENTAL (update links if necessary)",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repo, err := repository.ForWorkDir(wd)
		if err != nil {
			return err
		}
		ctx := context.Background()
		updated, success, errors := repo.Move(ctx, args[0], args[1])
		printErrors(ctx, errors)
		for {
			select {
			case note, ok := <-updated:
				if ok {
					fmt.Printf("%s updated\n", color.CyanString(note.Path()))
				}
			case succ, ok := <-success:
				if !ok {
					return nil
				}
				if succ {
					fmt.Println("File moved")
				} else {
					return fmt.Errorf("move failed")
				}
			}
		}
	},
}
