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
		return Tag{}, fmt.Errorf("%s is not a valid tag", name)
	}
	return Tag{tag: name}, nil
}

type Tag struct {
	tag string
}

func (t Tag) Name() string {
	s := t.tag
	if strings.Contains(s, ":") {
		return s[:strings.Index(s, ":")]
	}
	return s
}

func (t Tag) Value() (string, error) {
	s := t.tag
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

func (t Tag) MakeDateAbsolute() (Tag, error) {
	relativeDate, err := t.RelativeDate()
	if err != nil {
		return t, err
	}
	if relativeDate.Hour() == 0 && relativeDate.Minute() == 0 && relativeDate.Second() == 0 && relativeDate.Nanosecond() == 0 {
		return Tag{
			tag: t.Name() + ":" + relativeDate.Format("2006-01-02"),
		}, nil
	}
	return Tag{
		tag: t.Name() + ":" + relativeDate.Format(time.RFC3339),
	}, nil
}

func (t Tag) String() string {
	return t.tag
}
