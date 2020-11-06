package notes

import (
	"context"
	"fmt"
	"sort"
	"time"
)

func Top(ctx context.Context, limit int, notes <-chan Note, less Less) (note <-chan Note, errors <-chan error) {
	out := make(chan Note)
	errs := make(chan error)

	go func() {
		defer close(out)
		defer close(errs)
		slice := collectNotes(ctx, notes)
		sortNotes(slice, less, errs)
		if len(slice) > limit {
			slice = slice[:limit]
		}
		for _, note := range slice {
			out <- note
		}
	}()
	return out, errs
}

func collectNotes(ctx context.Context, notes <-chan Note) []Note {
	var collectedNotes []Note
	for {
		select {
		case note, ok := <-notes:
			if !ok {
				return collectedNotes
			}
			collectedNotes = append(collectedNotes, note)
		case <-ctx.Done():
			return nil
		}
	}
}

func sortNotes(slice []Note, less Less, errs chan<- error) {
	sort.Slice(slice, func(i, j int) bool {
		l, err := less(slice[i], slice[j])
		if err != nil {
			errs <- fmt.Errorf("comparing notes failed %s and %s: %v", slice[i].Path(), slice[j].Path(), err)
		}
		return l
	})
}

type Less func(i, j Note) (bool, error)

var ModifiedDesc Less = func(i, j Note) (bool, error) {
	firstModified, e := i.Modified()
	if e != nil {
		return false, e
	}
	secondModified, e := j.Modified()
	if e != nil {
		return false, e
	}
	return firstModified.After(secondModified), nil
}

var ModifiedAsc Less = func(i, j Note) (bool, error) {
	firstModified, e := i.Modified()
	if e != nil {
		return false, e
	}
	secondModified, e := j.Modified()
	if e != nil {
		return false, e
	}
	return firstModified.Before(secondModified), nil
}

var CreatedDesc Less = func(first, second Note) (bool, error) {
	firstCreated, e := first.Created()
	if e != nil {
		return false, e
	}
	secondCreated, e := second.Created()
	if e != nil {
		return false, e
	}
	return firstCreated.After(secondCreated), nil
}

var CreatedAsc Less = func(first, second Note) (bool, error) {
	firstCreated, e := first.Created()
	if e != nil {
		return false, e
	}
	secondCreated, e := second.Created()
	if e != nil {
		return false, e
	}
	return firstCreated.Before(secondCreated), nil
}

func TagDateDesc(name string) Less {
	return tagDateLess(name, func(first, second time.Time) bool {
		return first.After(second)
	})
}

func TagDateAsc(name string) Less {
	return tagDateLess(name, func(first, second time.Time) bool {
		return first.Before(second)
	})
}

func tagDateLess(name string, less func(first, second time.Time) bool) Less {
	return func(first, second Note) (bool, error) {
		firstTag, found, err := FindTagByName(first, name)
		if err != nil || !found {
			return false, err
		}
		secondTag, found, err := FindTagByName(second, name)
		if err != nil || !found {
			return true, err
		}
		firstDate, err := firstTag.AbsoluteDate()
		if err != nil {
			return false, err
		}
		secondDate, err := secondTag.AbsoluteDate()
		if err != nil {
			return false, err
		}
		return less(firstDate, secondDate), nil
	}
}

func TagNumberDesc(name string) Less {
	return tagNumberLess(name, func(first, second int) bool {
		return first > second
	})
}

func TagNumberAsc(name string) Less {
	return tagNumberLess(name, func(first, second int) bool {
		return first < second
	})
}

func tagNumberLess(name string, less func(first, second int) bool) Less {
	return func(first, second Note) (bool, error) {
		firstTag, found, err := FindTagByName(first, name)
		if err != nil || !found {
			return false, err
		}
		secondTag, found, err := FindTagByName(second, name)
		if err != nil || !found {
			return true, err
		}
		firstNumber, err := firstTag.Number()
		if err != nil {
			return false, err
		}
		secondNumber, err := secondTag.Number()
		if err != nil {
			return false, err
		}
		return less(firstNumber, secondNumber), nil
	}
}
