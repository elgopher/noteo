## Installation

1. Download the [release](https://github.com/jacekolszak/noteo/releases)
1. Copy to your `bin` folder

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