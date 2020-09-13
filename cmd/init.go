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
		cfgFile, err := repository.Init(".")
		if err != nil {
			return err
		}
		fmt.Printf("Repository initialized. Configuration file saved at %s\n", cfgFile)
		return nil
	},
}
