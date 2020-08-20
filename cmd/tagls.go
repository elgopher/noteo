package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var tagLs = &cobra.Command{
	Use:   "ls",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := repo(args)
		if err != nil {
			return err
		}

		ctx := context.Background()
		tags, errs := repo.Tags(ctx)
		printErrors(ctx, errs)
		for tag := range tags {
			fmt.Println(tag)
		}
		return nil
	},
}
