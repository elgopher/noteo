package cmd

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/jacekolszak/noteo/date"
	"github.com/jacekolszak/noteo/notes"
	"github.com/jacekolszak/noteo/output/jayson"
	"github.com/jacekolszak/noteo/output/quiet"
	"github.com/jacekolszak/noteo/output/table"
	"github.com/jacekolszak/noteo/output/yml"
	"github.com/spf13/cobra"
)

type formatter interface {
	Header() string
	Note(note notes.Note) string
}

type lsCommand struct {
	// projection
	quietMode    bool
	outputFormat string
	date         string
	// filtering
	tagFilter      []string
	notagFilter    []string
	tagGrep        []string
	tagGreater     []string
	tagLower       []string
	tagAfter       []string
	tagBefore      []string
	noTags         bool
	modifiedAfter  string
	modifiedBefore string
	createdAfter   string
	createdBefore  string
	grep           string
	// sorting and limiting
	limit           int
	sortByCreated   bool
	sortByTagDate   string
	sortByTagNumber string
	reverse         bool
}

func ls() *cobra.Command {
	c := &lsCommand{}
	ls := &cobra.Command{
		Use:   "ls",
		Short: "List notes summary",
		Args:  cobra.RangeArgs(0, 1),
		RunE:  c.RunE,
		Example: `
  # List notes matching regular expression
  noteo ls --grep regex

  # List notes created after 2020-08-30 (midnight)
  noteo ls --created-after "2020-08-30" 

  # List only file names
  noteo ls -q

  # List notes with both "task" and "link" tags
  noteo ls -t task -t link

  # List notes with tag name "priority" and value greater than 1
  noteo ls --tag-greater priority:1

  # List notes with tag name "deadline" and value (which is a date) after 2020-08-30, sorted by date taken from this tag
  noteo ls --tag-after deadline:2020-08-30 --sort-by-tag-date deadline

  # List specific columns
  noteo ls -o table=file,tags`,
	}
	ls.Flags().BoolVarP(&c.quietMode, "quiet", "q", false, "")
	ls.Flags().StringVarP(&c.outputFormat, "output", "o", "table=file,beginning,modified,tags", "")
	ls.Flags().StringVar(&c.date, "date", "", "")
	// filtering
	ls.Flags().StringArrayVarP(&c.tagFilter, "tag", "t", nil, "")
	ls.Flags().StringArrayVar(&c.notagFilter, "no-tag", nil, "")
	ls.Flags().StringArrayVar(&c.tagGrep, "tag-grep", nil, "")
	ls.Flags().StringArrayVar(&c.tagGreater, "tag-greater", nil, "")
	ls.Flags().StringArrayVar(&c.tagLower, "tag-lower", nil, "")
	ls.Flags().StringArrayVar(&c.tagAfter, "tag-after", nil, "")
	ls.Flags().StringArrayVar(&c.tagBefore, "tag-before", nil, "")
	ls.Flags().BoolVar(&c.noTags, "no-tags", false, "")
	ls.Flags().StringVar(&c.modifiedAfter, "modified-after", "", "")
	ls.Flags().StringVar(&c.modifiedBefore, "modified-before", "", "")
	ls.Flags().StringVar(&c.createdAfter, "created-after", "", "")
	ls.Flags().StringVar(&c.createdBefore, "created-before", "", "")
	ls.Flags().StringVar(&c.grep, "grep", "", "")
	// sorting and limiting
	ls.Flags().IntVarP(&c.limit, "limit", "l", math.MaxInt32, "")
	ls.Flags().BoolVar(&c.sortByCreated, "sort-by-created", false, "")
	ls.Flags().StringVarP(&c.sortByTagDate, "sort-by-tag-date", "", "", "")
	ls.Flags().StringVarP(&c.sortByTagNumber, "sort-by-tag-number", "", "", "")
	ls.Flags().BoolVar(&c.reverse, "reverse", false, "")
	ls.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

Filtering flags:
      --created-after string        filter notes created after given date
      --created-before string       filter notes created before given date
      --grep string                 grep text using regular expression
      --modified-after string       filter notes modified after given date
      --modified-before string      filter notes modified before given date
      --no-tag stringArray          filter notes not having tag
      --no-tags                     filter notes not having tags at all
  -t, --tag stringArray             filter notes having tag
      --tag-after stringArray       filter notes having tag with value date after specified date, i.e "foo:2010-08-01"
      --tag-before stringArray      filter notes having tag with value date before specified date, i.e "foo:2010-08-01"
      --tag-greater stringArray     filter notes having tag with value number greater than specified number i.e "foo:2"
      --tag-grep stringArray        filter notes having tag matching regular expression 
      --tag-lower stringArray       filter notes having tag with value number greater than specified number i.e "foo:2"

Sorting and limiting flags:
  -l, --limit int                   limits number of notes returned (default 2147483647)
      --reverse                     makes sorting ascending
      --sort-by-created             sorts by created date descending
      --sort-by-tag-date string     
      --sort-by-tag-number string   

Other flags:
      --date string                 show dates in given format: relative (default), iso8601 or rfc2822. For now used in table and wide output only.
  -h, --help                        help for ls
  -o, --output string               Specify output format: table using given columns, wide, json or yaml
                                    (default "table=file,beginning,modified,tags")
  -q, --quiet                       Show only file names{{if .HasAvailableInheritedFlags}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
	ls.Flags().Usage = func() {

	}
	return ls
}

func (c *lsCommand) RunE(cmd *cobra.Command, args []string) error {
	repo, err := repo(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	dirNotes, notesErrors := repo.Notes(ctx)

	predicates, err := c.filterPredicates()
	if err != nil {
		return err
	}
	filtered, filterErrors := notes.Filter(ctx, toNotes(dirNotes), predicates...)
	sortedNotes, topErrors := notes.Top(ctx, c.limit, filtered, c.sort())

	printErrors(ctx, notesErrors, filterErrors, topErrors)

	out, err := c.formatter()
	if err != nil {
		return err
	}
	fmt.Print(out.Header())
	for note := range sortedNotes {
		fmt.Print(out.Note(note))
	}
	return nil
}

func (c *lsCommand) filterPredicates() ([]notes.Predicate, error) {
	var predicates []notes.Predicate
	for _, t := range c.tagFilter {
		predicates = append(predicates, notes.Tag(t))
	}
	for _, t := range c.notagFilter {
		predicates = append(predicates, notes.NoTag(t))
	}
	for _, grep := range c.tagGrep {
		regex, err := regexp.Compile(grep)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, notes.TagGrep(regex))
	}
	for _, greater := range c.tagGreater {
		p, err := notes.TagGreater(greater)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	for _, lower := range c.tagLower {
		p, err := notes.TagLower(lower)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	for _, after := range c.tagAfter {
		p, err := notes.TagAfter(after)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	for _, before := range c.tagBefore {
		p, err := notes.TagBefore(before)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	if c.noTags {
		predicates = append(predicates, notes.NoTags())
	}
	if c.modifiedAfter != "" {
		p, err := notes.ModifiedAfter(c.modifiedAfter)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	if c.modifiedBefore != "" {
		p, err := notes.ModifiedBefore(c.modifiedBefore)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	if c.createdAfter != "" {
		p, err := notes.CreatedAfter(c.createdAfter)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	if c.createdBefore != "" {
		p, err := notes.CreatedBefore(c.createdBefore)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	if c.grep != "" {
		p, err := notes.Grep(c.grep)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, p)
	}
	return predicates, nil
}

func (c *lsCommand) sort() notes.Less {
	sort := notes.ModifiedDesc
	if c.reverse {
		sort = notes.ModifiedAsc
	}
	switch {
	case c.sortByCreated && !c.reverse:
		sort = notes.CreatedDesc
	case c.sortByCreated && c.reverse:
		sort = notes.CreatedAsc
	case c.sortByTagDate != "" && !c.reverse:
		sort = notes.TagDateDesc(c.sortByTagDate)
	case c.sortByTagDate != "" && c.reverse:
		sort = notes.TagDateAsc(c.sortByTagDate)
	case c.sortByTagNumber != "" && !c.reverse:
		sort = notes.TagNumberDesc(c.sortByTagNumber)
	case c.sortByTagNumber != "" && c.reverse:
		sort = notes.TagNumberAsc(c.sortByTagNumber)
	}
	return sort
}

func (c *lsCommand) dateFormat(defaultFormat date.Format) (date.Format, error) {
	switch strings.ToLower(c.date) {
	case "rfc", "rfc2822":
		return date.RFC2822, nil
	case "iso", "iso8601":
		return date.ISO8601, nil
	case "relative":
		return date.Relative, nil
	case "":
		return defaultFormat, nil
	default:
		return "", fmt.Errorf("unsupported date format: %s. Supported formats are: rfc2822 (or rfc), iso8601 (or iso), relative", c.date)
	}
}

func (c *lsCommand) formatter() (formatter, error) {
	outputFormat := strings.ToLower(c.outputFormat)
	var err error
	var out formatter
	dateFormat, err := c.dateFormat(date.Relative)
	if err != nil {
		return nil, err
	}
	switch {
	case c.quietMode:
		out = quiet.Formatter{}
	case outputFormat == "wide":
		out, err = table.NewFormatter([]string{"file", "beginning", "modified", "created", "tags"}, dateFormat)
	case strings.HasPrefix(outputFormat, "table="):
		columns := strings.Split(strings.TrimPrefix(outputFormat, "table="), ",")
		out, err = table.NewFormatter(columns, dateFormat)
	case outputFormat == "json":
		out = jayson.Formatter{}
	case outputFormat == "yaml":
		out = yml.Formatter{}
	default:
		err = fmt.Errorf("unsupported output format in --output flag: %s", outputFormat)
	}
	return out, err
}
