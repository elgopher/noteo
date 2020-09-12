package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Root() *cobra.Command {
	root := cobra.Command{
		Use:           "noteo",
		Short:         "Command line note-taking assistant",
		Version:       "0.3.0",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(initialize)
	root.AddCommand(add)
	root.AddCommand(ls())
	root.AddCommand(tag())
	root.AddCommand(mv)
	return &root
}

func Execute() {
	root := Root()
	if err := root.Execute(); err != nil {
		if e, ok := err.(repositoryError); ok && e.IsNotRepository() {
			_, _ = fmt.Fprintln(os.Stderr, "not a noteo repository (or any of the parent directories)")
			printer := NewPrinter()
			printer.Print("Please run ")
			printer.PrintComand("note init")
			printer.Println()
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err)
			if err := root.UsageFunc()(root); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}
		os.Exit(1)
	}
}

type repositoryError interface {
	IsNotRepository() bool
}
