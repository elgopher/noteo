package cmd

import (
	"github.com/spf13/cobra"
)

func tag() *cobra.Command {
	var tag = &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}
	tag.AddCommand(tagSet())
	tag.AddCommand(tagRm())
	tag.AddCommand(tagLs)
	return tag
}
