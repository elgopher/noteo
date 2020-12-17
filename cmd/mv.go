package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// mv represents the mv command
var mv = &cobra.Command{
	Use:   "mv",
	Args:  cobra.ExactArgs(2),
	Short: "Move note - EXPERIMENTAL (update links if necessary)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := workingDirRepository()
		if err != nil {
			return err
		}
		ctx := context.Background()
		updated, success, errors := repo.Move(ctx, args[0], args[1])
		printErrors(ctx, errors)
		printer := NewPrinter()
		for {
			select {
			case note, ok := <-updated:
				if ok {
					printer.PrintFile(note.Path())
					printer.Println(" updated")
				}
			case succ, ok := <-success:
				if !ok {
					return nil
				}
				if succ {
					printer.Println("File moved")
				} else {
					return fmt.Errorf("move failed")
				}
			}
		}
	},
}
