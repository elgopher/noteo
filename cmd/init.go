package cmd

import (
	"fmt"

	"github.com/jacekolszak/noteo/repository"
	"github.com/spf13/cobra"
)

var initialize = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Noteo repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := repository.Init(".")
		if err != nil {
			return err
		}
		fmt.Println("repo initialized")
		return nil
	},
}
