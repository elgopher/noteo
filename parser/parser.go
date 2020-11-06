package parser

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

var frontMatterDividerRegex = regexp.MustCompile(`^---`)

func Parse(reader io.Reader) (frontMatter string, body string, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(spltLinesIncludeEOL)
	if scanner.Scan() {
		firstLine := scanner.Text()
		lines := ""
		if frontMatterDividerRegex.MatchString(firstLine) {
			lines += firstLine
			for scanner.Scan() {
				line := scanner.Text()
				lines += line
				if frontMatterDividerRegex.MatchString(line) {
					frontMatter = lines
					for scanner.Scan() {
						bodyLine := scanner.Text()
						body += bodyLine
					}
					return
				}
			}
			body = lines
		} else {
			body += firstLine
			for scanner.Scan() {
				body += scanner.Text()
			}
		}
	}
	return
}

func spltLinesIncludeEOL(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
