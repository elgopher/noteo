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
some-short-idea.md created
$ noteo ls
FILE                   BEGINNING          MODIFIED        TAGS                                    
some-short-idea.md     Some short idea    9 seconds ago          	        
$ noteo tag add -n short some-short-idea.md
some-short-idea.md updated
$ cat some-short-idea.md
---
Created: Fri Sep  4 00:54:19 CEST 2020
Tags: short
---

Some short idea 
$ mkdir ideas
$ noteo mv some-short-idea.md ideas/some-short-idea.md
File moved
$ noteo ls
FILE                       BEGINNING         MODIFIED        TAGS                                    
ideas/some-short-idea.md   Some short idea   2 minutes ago   short
```
