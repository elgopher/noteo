// Package config provides configuration from merging environment and repository config
package config

import (
	"os"
	"runtime"
)

type RepoConfig interface {
	EditorCommand() string
}

type Config struct {
	repoConfig RepoConfig
}

func New(config RepoConfig) *Config {
	return &Config{repoConfig: config}
}

func (c *Config) EditorCommand() string {
	if c.repoConfig.EditorCommand() != "" {
		return c.repoConfig.EditorCommand()
	}
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vim +"
}
