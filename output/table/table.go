package table

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/output"
)

var mapping = map[string]column{
	"FILE":      fileColumn{},
	"BEGINNING": beginningColumn{},
	"MODIFIED":  modifiedColumn{},
	"CREATED":   createdColumn{},
	"TAGS":      tagsColumn{},
}

func NewFormatter(columns []string, dateFormat date.Format) (*Formatter, error) {
	var cols []string
	for _, c := range columns {
		c = strings.ToUpper(c)
		if _, ok := mapping[c]; !ok {
			return nil, fmt.Errorf("unsupported output column: %s", c)
		}
		cols = append(cols, c)
	}
	return &Formatter{columns: cols, dateFormat: dateFormat}, nil
}

type Formatter struct {
	columns    []string
	dateFormat date.Format
}

func (o Formatter) Header() string {
	header := ""
	for _, c := range o.columns {
		header += mapping[c].header(opts{dateFormat: o.dateFormat}) + "\t"
	}
	header += "\n"
	return header
}

func (o Formatter) Note(note notes.Note) string {
	header := ""
	for _, c := range o.columns {
		header += mapping[c].cell(note, opts{dateFormat: o.dateFormat}) + "\t"
	}
	header += "\n"
	return header
}

func format(text string, limit int) string {
	runes := []rune(text) // cast is need to make len work
	if len(runes) > limit {
		runes = runes[:limit-1]
		runes = append(runes, '…')
	}
	return fmt.Sprintf("%-*s", limit, string(runes))
}

func beginning(text string) string {
	t := strings.Trim(text, "\n")
	if strings.Contains(t, "\n") {
		t = t[:strings.IndexRune(t, '\n')]
	}
	t = strings.ReplaceAll(t, "\t", " ")
	for i := 0; i < 5; i++ {
		t = strings.TrimPrefix(t, "#")
	}
	t = strings.TrimPrefix(t, "*")
	t = strings.ReplaceAll(t, "\r", "")
	t = strings.Trim(t, " ")
	return t
}

// File is formatted a slightly different
func formatNoCut(text string, limit int) string {
	return fmt.Sprintf("%-*s", limit, text)
}

type column interface {
	header(opts opts) string
	cell(note notes.Note, opts opts) string
}

type opts struct {
	dateFormat date.Format
}

type fileColumn struct{}

func (f fileColumn) header(opts) string {
	return color.CyanString(format("FILE", 35))
}
func (f fileColumn) cell(note notes.Note, opts opts) string {
	return color.CyanString(formatNoCut(note.Path(), 35))
}

type beginningColumn struct{}

func (s beginningColumn) header(opts) string {
	return format("BEGINNING", 34)
}
func (s beginningColumn) cell(note notes.Note, opts opts) string {
	text, _ := note.Text()
	return format(beginning(text), 34)
}

type modifiedColumn struct{}

func (m modifiedColumn) header(opts opts) string {
	return format("MODIFIED", dateLimit(opts.dateFormat))
}

func dateLimit(dateFormat date.Format) int {
	limit := 40
	switch dateFormat {
	case date.Relative:
		limit = 18
	case date.ISO8601:
		limit = 35
	case date.RFC2822:
		limit = 35
	}
	return limit
}

func (m modifiedColumn) cell(note notes.Note, opts opts) string {
	modified := date.FormatWithType(note.Modified(), opts.dateFormat)
	return format(modified, dateLimit(opts.dateFormat))
}

type createdColumn struct{}

func (c createdColumn) header(opts opts) string {
	return format("CREATED", dateLimit(opts.dateFormat))
}

func (c createdColumn) cell(note notes.Note, opts opts) string {
	created, err := note.Created()
	if err != nil {
		errString := err.Error()
		errString = strings.ReplaceAll(errString, "\t", " ")
		return errString
	}
	modified := date.FormatWithType(created, opts.dateFormat)
	return format(modified, dateLimit(opts.dateFormat))
}

type tagsColumn struct{}

func (t tagsColumn) header(opts) string {
	return format("TAGS", 40)
}

func (t tagsColumn) cell(note notes.Note, opts opts) string {
	tags, _ := output.StringTags(note)
	tagsString := strings.Join(tags, " ")
	return formatNoCut(tagsString, 40)
}
