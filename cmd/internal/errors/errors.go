package errors

import (
	"fmt"
	"strings"
)

func NewList(errors ...error) List {
	return List.Add(nil, errors...)
}

type List []error

func (e List) Add(errors ...error) List {
	for _, err := range errors {
		if err == nil {
			continue
		}

		e = append(e, err)
	}

	return e
}

func (e List) Errorf(format string, args ...interface{}) error {
	if e == nil {
		return nil
	}

	return fmt.Errorf(format, args...)
}

func (e List) Error() string {
	var slist []string
	for _, err := range e {
		slist = append(slist, err.Error())
	}
	return strings.Join(slist, ", ")
}
