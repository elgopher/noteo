package parser

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
)

var yamlDividerRegex = regexp.MustCompile(`^---`)

func Parse(reader io.Reader) (yml string, body string, err error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(spltLinesIncludeEOL)
	if scanner.Scan() {
		firstLine := scanner.Text()
		yamel := ""
		if yamlDividerRegex.MatchString(firstLine) {
			yamel += firstLine
			for scanner.Scan() {
				ymlLine := scanner.Text()
				yamel += ymlLine
				if yamlDividerRegex.MatchString(ymlLine) {
					yml = yamel
					for scanner.Scan() {
						bodyLine := scanner.Text()
						body += bodyLine
					}
					return
				}
			}
			body = yamel
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
