package note

import (
	"github.com/jacekolszak/noteo/parser"
	"os"
	"sync"
)

type originalContent struct {
	path        string
	once        sync.Once
	frontMatter string
	body        string
}

func (c *originalContent) ensureLoaded() error {
	var err error
	c.once.Do(func() {
		var file *os.File
		file, err = os.Open(c.path)
		if err != nil {
			return
		}
		defer file.Close()
		c.frontMatter, c.body, err = parser.Parse(file)
	})
	return err
}

func (c *originalContent) FrontMatter() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.frontMatter, nil
}

func (c *originalContent) Body() (string, error) {
	if err := c.ensureLoaded(); err != nil {
		return "", err
	}
	return c.body, nil
}

func (c *originalContent) Full() (string, error) {
	frontMatter, err := c.FrontMatter()
	if err != nil {
		return "", err
	}
	body, err := c.Body()
	if err != nil {
		return "", err
	}
	return frontMatter + body, nil
}
