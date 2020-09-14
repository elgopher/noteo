Noteo is a command line note-taking assistant which can be helpful in all 3 stages of note-taking:

* brainstorming (quickly adding notes)
* reviewing and refining notes (tagging, organizing notes in folders)
* searching for notes (using advanced filtering and sorting)

## Installation

Download the [release archive](https://github.com/jacekolszak/noteo/releases) and extract the binary to your `bin` folder.

If your OS/arch is not available you can try to build it manually using [Go](https://golang.org/):

```bash
go get -u github.com/jacekolszak/noteo
```

## Demo

![Demo](demo/demo.gif)

## File format

Each note is a file in standard [Markdown](https://en.wikipedia.org/wiki/Markdown) format with meta information provided in the beginning of the file in the form of [YAML front matter](https://jekyllrb.com/docs/front-matter/)

```md
---
Created: Sat Sep  5 12:30:05 CEST 2020
Tags: space separated tags
---

Some Markdown content here.
```

Noteo extracts information from this header to filter out and sort notes. Some commands such as `tag set` and `tag rm` may update the header too. If the YAML front matter is missing, Noteo uses default values such as empty `Tags` or `Created` equal to file modification date.

### Tag format

Each tag is a string without whitespaces (space, tab, new line), for example `idea`, `task`

Tag might have a special form of `name:value`, for example `deadline:2020-09-30` or `priority:1`. Value can be a date or integer.
