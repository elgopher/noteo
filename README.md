Noteo is a command line note-taking assistant which can be helpful in all 3 stages of note-taking:

* brainstorming (quickly adding notes)
* reviewing and refining notes (tagging, organizing notes in folders)
* searching for notes (using advanced filtering and sorting)

## Installation

Download the [release archive](https://github.com/jacekolszak/noteo/releases) and extract the binary to your `bin` folder.

If your OS/arch is not available you can try to build it manually using Go:

```bash
go get -u github.com/jacekolszak/noteo
```

## Examples

```
$ mkdir mynotes

$ cd mynotes

$ noteo init
  repo initialized

$ noteo add "My fantastic idea"
  my-fantastic-idea.md created

$ noteo ls
  FILE                                    BEGINNING             MODIFIED               TAGS
  my-fantastic-idea.md                    My fantastic idea     About a minute ago

$ noteo tag add -n idea my-fantastic-idea.md
  my-fantastic-idea.md updated

$ cat my-fantastic-idea.md 
  ---
  Created: Sat Sep  5 12:30:05 CEST 2020
  Tags: idea
  ---
  
  My fantastic idea

$ mkdir some-project

$ noteo mv my-fantastic-idea.md some-project/my-fantastic-idea.md
  File moved

$ noteo ls
  FILE                                    BEGINNING             MODIFIED               TAGS
  some-project/my-fantastic-idea.md       My fantastic idea     About a minute ago     idea
```