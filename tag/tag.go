package tag

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jacekolszak/noteo/date"
)

var validName = regexp.MustCompile(`^\S+$`)

func New(name string) (Tag, error) {
	if !validName.MatchString(name) {
		return "", fmt.Errorf("%s is not a valid tag", name)
	}
	return Tag(name), nil
}

type Tag string

func (t Tag) Name() string {
	s := string(t)
	if strings.Contains(s, ":") {
		return s[:strings.Index(s, ":")]
	}
	return s
}

func (t Tag) Value() (string, error) {
	s := string(t)
	if !strings.Contains(s, ":") {
		return "", fmt.Errorf("tag %s is not name:value", t)
	}
	val := s[strings.Index(s, ":")+1:]
	return val, nil
}

func (t Tag) Number() (int, error) {
	value, err := t.Value()
	if err != nil {
		return 0, err
	}
	return parseNumber(value)
}

func parseNumber(value string) (int, error) {
	num, err := strconv.ParseInt(value, 10, 32)
	return int(num), err
}

func (t Tag) AbsoluteDate() (time.Time, error) {
	value, err := t.Value()
	if err != nil {
		return time.Time{}, err
	}
	return date.ParseAbsolute(value)
}

func (t Tag) RelativeDate() (time.Time, error) {
	value, err := t.Value()
	if err != nil {
		return time.Time{}, err
	}
	return date.Parse(value)
}
