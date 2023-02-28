package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/elgopher/noteo/config"
)

var add = &cobra.Command{
	Use:   "add [TEXT]",
	Short: "Add a new note",
	Long:  "Add a new note in a current working directory",
	Aliases: []string{
		"create", "new",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := workingDirRepository()
		if err != nil {
			return err
		}
		repoConfig, err := repo.Config()
		if err != nil {
			return err
		}
		cfg := config.New(repoConfig)
		text, err := readNoteText(cmd.Flags(), cfg)
		if err != nil {
			return err
		}
		if text == "" {
			fmt.Println("no new file added")
			return nil
		}
		f, err := repo.Add(text)
		if err != nil {
			return err
		}
		printer := NewPrinter()
		printer.PrintFile(f)
		printer.Println(" created")
		return nil
	},
}

func readNoteText(flags *pflag.FlagSet, cfg *config.Config) (string, error) {
	var text string
	if len(flags.Args()) == 0 {
		template := newFileTemplate(time.Now()) + "\n"
		tmpFile := filepath.Join(os.TempDir(), uuid.New().String()+" .md")
		if err := os.WriteFile(tmpFile, []byte(template), 0664); err != nil {
			return "", err
		}
		text, err := textFromEditor(tmpFile, cfg.EditorCommand())
		if err != nil {
			return "", err
		}
		if text == template {
			return "", nil
		}
		return text, nil
	} else {
		template := newFileTemplate(time.Now())
		text = template + strings.Join(flags.Args(), " ")
	}
	return text, nil
}

func textFromEditor(file, editorCommand string) (string, error) {
	editorNameWithArgs := strings.Split(editorCommand, " ")
	name := editorNameWithArgs[0]
	args := append(editorNameWithArgs[1:], file)
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	text, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return string(text), nil
}

func newFileTemplate(created time.Time) string {
	return `---
Created: ` + created.Format(time.UnixDate) + `
Tags: 
---

`
}
