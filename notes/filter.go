package notes

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/elgopher/noteo/date"
	"github.com/elgopher/noteo/tag"
)

func Filter(ctx context.Context, notes <-chan Note, predicates ...Predicate) (note <-chan Note, errors <-chan error) {
	out := make(chan Note)
	errs := make(chan error)
	go func() {
		defer close(out)
		defer close(errs)
		filterLoop(ctx, notes, predicates, out, errs)
	}()
	return out, errs
}

func filterLoop(ctx context.Context, notes <-chan Note, predicates []Predicate, out chan<- Note, errs chan<- error) {
	for {
		select {
		case note, ok := <-notes:
			if !ok {
				return
			}
			if noteMatches(note, predicates, errs) {
				out <- note
			}
		case <-ctx.Done():
			return
		}
	}
}

func noteMatches(note Note, predicates []Predicate, errs chan<- error) bool {
	for _, predicate := range predicates {
		matches, err := predicate(note)
		if err != nil {
			errs <- fmt.Errorf("executing predicate failed on note %s: %v", note.Path(), err)
			return false
		}
		if !matches {
			return false
		}
	}
	return true
}

type Predicate func(note Note) (bool, error)

func Tag(t string) Predicate {
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		for _, anotherTag := range tags {
			if anotherTag.String() == t {
				return true, nil
			}
		}
		return false, nil
	}
}

func NoTag(t string) Predicate {
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		for _, anotherTag := range tags {
			if anotherTag.String() == t {
				return false, nil
			}
		}
		return true, nil
	}
}

func TagGrep(regex *regexp.Regexp) Predicate {
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		for _, t := range tags {
			if regex.MatchString(t.String()) {
				return true, nil
			}
		}
		return false, nil
	}
}

func TagGreater(tagNameValue string) (Predicate, error) {
	return tagNumber(tagNameValue, func(anotherNumber, number int) bool {
		return anotherNumber > number
	})
}

func TagLower(tagNameValue string) (Predicate, error) {
	return tagNumber(tagNameValue, func(anotherNumber, number int) bool {
		return anotherNumber < number
	})
}

func tagNumber(tagNameValue string, f func(anotherNumber, number int) bool) (Predicate, error) {
	kv, err := tag.New(tagNameValue)
	if err != nil {
		return nil, err
	}
	number, err := kv.Number()
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		for _, another := range tags {
			if another.Name() == kv.Name() {
				anotherNumber, err := another.Number()
				if err != nil {
					return false, fmt.Errorf("error getting number from tag \"%s\": %v", another, err)
				}
				return f(anotherNumber, number), nil
			}
		}
		return false, nil
	}, nil
}

func TagAfter(tagNameValue string) (Predicate, error) {
	return tagDate(tagNameValue, func(anotherDate, date time.Time) bool {
		return anotherDate.After(date)
	})
}

func TagBefore(tagNameValue string) (Predicate, error) {
	return tagDate(tagNameValue, func(anotherDate, date time.Time) bool {
		return anotherDate.Before(date)
	})
}

func tagDate(tagNameValue string, f func(anotherDate, date time.Time) bool) (Predicate, error) {
	kv, err := tag.New(tagNameValue)
	if err != nil {
		return nil, err
	}
	relativeDate, err := kv.RelativeDate()
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		for _, another := range tags {
			if another.Name() == kv.Name() {
				anotherDate, err := another.AbsoluteDate()
				if err != nil {
					return false, fmt.Errorf("error getting date from tag \"%s\": %v", another, err)
				}
				return f(anotherDate, relativeDate), nil
			}
		}
		return false, nil
	}, nil
}

func NoTags() Predicate {
	return func(note Note) (bool, error) {
		tags, err := note.Tags()
		if err != nil {
			return false, err
		}
		return len(tags) == 0, nil
	}
}

func ModifiedAfter(modifiedAfter string) (Predicate, error) {
	t, err := date.Parse(modifiedAfter)
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		modified, err := note.Modified()
		if err != nil {
			return false, err
		}
		return modified.After(t), nil
	}, nil
}

func ModifiedBefore(modifiedBefore string) (Predicate, error) {
	t, err := date.Parse(modifiedBefore)
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		modified, err := note.Modified()
		if err != nil {
			return false, err
		}
		return modified.Before(t), nil
	}, nil
}

func CreatedAfter(createdAfter string) (Predicate, error) {
	t, err := date.Parse(createdAfter)
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		created, err := note.Created()
		if err != nil {
			return false, err
		}
		return created.After(t), nil
	}, nil
}

func CreatedBefore(createdBefore string) (Predicate, error) {
	t, err := date.Parse(createdBefore)
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		created, err := note.Created()
		if err != nil {
			return false, err
		}
		return created.Before(t), nil
	}, nil
}

func Grep(expr string) (Predicate, error) {
	regex, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return func(note Note) (bool, error) {
		body, err := note.Body()
		if err != nil {
			return false, err
		}
		return regex.MatchString(body), nil
	}, nil
}
