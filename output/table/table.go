package table

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/output"
	"github.com/juju/ansiterm"
	"golang.org/x/crypto/ssh/terminal"
)

var mapping = map[string]column{
	"FILE":      fileColumn{},
	"BEGINNING": beginningColumn{},
	"MODIFIED":  modifiedColumn{},
	"CREATED":   createdColumn{},
	"TAGS":      tagsColumn{},
}

func NewFormatter(columns []string, dateFormat date.Format) (*Formatter, error) {
	w, h, err := terminal.GetSize(0)
	if err != nil {
		w = 80
		h = 25
	}
	var cols []string
	for _, c := range columns {
		c = strings.ToUpper(c)
		if _, ok := mapping[c]; !ok {
			return nil, fmt.Errorf("unsupported output column: %s", c)
		}
		cols = append(cols, c)
	}
	buffer := bytes.NewBuffer([]byte{})
	writer := ansiterm.NewTabWriter(buffer, 0, 8, 1, '\t', 0)
	writer.SetColorCapable(true)
	return &Formatter{
			columns:    cols,
			dateFormat: dateFormat,
			width:      w,
			height:     h,
			buffer:     buffer,
			writer:     writer,
		},
		nil
}

type Formatter struct {
	columns    []string
	dateFormat date.Format
	width      int
	height     int
	line       int
	writer     *ansiterm.TabWriter
	buffer     *bytes.Buffer
}

func (o *Formatter) flush() string {
	_ = o.writer.Flush()
	out := o.buffer.String()
	o.buffer.Reset()
	o.line = 0
	return out
}

func (o *Formatter) Header() string {
	o.line++
	for _, c := range o.columns {
		options := opts{dateFormat: o.dateFormat}
		column := mapping[c]
		column.printHeader(options, o.writer)
		_, _ = o.writer.Write([]byte("\t"))
	}
	_, _ = o.writer.Write([]byte("\n"))
	return ""
}

func (o *Formatter) Footer() string {
	return o.flush()
}

func (o *Formatter) Note(note notes.Note) string {
	out := ""
	if o.line == o.height {
		out = o.flush()
	}
	o.line++
	for _, c := range o.columns {
		options := opts{dateFormat: o.dateFormat}
		column := mapping[c]
		column.printValue(note, options, o.writer)
		_, _ = o.writer.Write([]byte("\t"))
	}
	_, _ = o.writer.Write([]byte("\n"))
	return out
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

type column interface {
	printHeader(opts opts, writer *ansiterm.TabWriter)
	printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter)
}

type opts struct {
	dateFormat date.Format
}

type fileColumn struct{}

func (f fileColumn) printHeader(opts opts, writer *ansiterm.TabWriter) {
	_, _ = writer.Write([]byte("FILE"))
}

func (f fileColumn) printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter) {
	writer.SetForeground(ansiterm.BrightBlue)
	defer writer.Reset()
	_, _ = fmt.Fprint(writer, note.Path())
}

type beginningColumn struct{}

func (s beginningColumn) printHeader(opts opts, writer *ansiterm.TabWriter) {
	_, _ = fmt.Fprint(writer, format("BEGINNING", 34))
}
func (s beginningColumn) printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter) {
	text, _ := note.Text()
	writer.SetStyle(ansiterm.Bold)
	defer writer.Reset()
	_, _ = fmt.Fprint(writer, format(beginning(text), 34))
}

type modifiedColumn struct{}

func (m modifiedColumn) printHeader(_ opts, writer *ansiterm.TabWriter) {
	_, _ = writer.Write([]byte("MODIFIED"))
}

func (m modifiedColumn) printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter) {
	modified, err := note.Modified()
	if err != nil {
		errString := err.Error()
		errString = strings.ReplaceAll(errString, "\t", " ")
		_, _ = fmt.Fprint(writer, errString)
		return
	}
	formatted := date.FormatWithType(modified, opts.dateFormat)
	_, _ = fmt.Fprint(writer, formatted)
}

type createdColumn struct{}

func (c createdColumn) printHeader(_ opts, writer *ansiterm.TabWriter) {
	_, _ = writer.Write([]byte("CREATED"))
}

func (c createdColumn) printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter) {
	created, err := note.Created()
	if err != nil {
		errString := err.Error()
		errString = strings.ReplaceAll(errString, "\t", " ")
		_, _ = fmt.Fprint(writer, errString)
		return
	}
	modified := date.FormatWithType(created, opts.dateFormat)
	_, _ = fmt.Fprint(writer, modified)
}

type tagsColumn struct{}

func (t tagsColumn) printHeader(_ opts, writer *ansiterm.TabWriter) {
	_, _ = writer.Write([]byte("TAGS"))
}

func (t tagsColumn) printValue(note notes.Note, opts opts, writer *ansiterm.TabWriter) {
	tags, _ := output.StringTags(note)
	tagsString := strings.Join(tags, " ")
	_, _ = fmt.Fprint(writer, tagsString)
}
