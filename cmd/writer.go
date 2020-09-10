package cmd

import (
	"fmt"
	"os"

	"github.com/juju/ansiterm"
)

type Printer ansiterm.Writer

func NewPrinter() *Printer {
	return (*Printer)(ansiterm.NewWriter(os.Stdout))
}

func (w Printer) PrintFile(file string) {
	writer := ansiterm.Writer(w)
	writer.SetForeground(ansiterm.BrightBlue)
	_, _ = fmt.Fprint(writer, file)
	writer.Reset()
}

func (w Printer) PrintComand(file string) {
	writer := ansiterm.Writer(w)
	writer.SetStyle(ansiterm.Bold)
	_, _ = fmt.Fprint(writer, file)
	writer.Reset()
}

func (w Printer) Print(a ...interface{}) {
	_, _ = fmt.Fprint(w, a...)
}

func (w Printer) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(w, a...)
}
