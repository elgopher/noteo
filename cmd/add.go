package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jacekolszak/noteo/repository"
	"github.com/spf13/cobra"
)

var editor = "vim +"

var add = &cobra.Command{
	Use:   "add [TEXT]",
	Short: "Add a new note",
	Long:  "Add a new note in a current working directory",
	Aliases: []string{
		"create", "new",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		repo, err := repository.ForWorkDir(wd)
		if err != nil {
			return err
		}
		var text string
		if len(cmd.Flags().Args()) == 0 {
			template := newFileTemplate(time.Now()) + "\n"
			tmpFile := filepath.Join(os.TempDir(), uuid.New().String()+" .md")
			if err := ioutil.WriteFile(tmpFile, []byte(template), 0664); err != nil {
				return err
			}
			text, err = textFromEditor(tmpFile)
			if err != nil {
				return err
			}
			if text == template {
				fmt.Println("no new file added")
				return nil
			}
		} else {
			template := newFileTemplate(time.Now())
			text = template + strings.Join(cmd.Flags().Args(), " ")
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

func textFromEditor(file string) (string, error) {
	editorNameWithArgs := strings.Split(editor, " ")
	name := editorNameWithArgs[0]
	args := append(editorNameWithArgs[1:], file)
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	text, err := ioutil.ReadFile(file)
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
