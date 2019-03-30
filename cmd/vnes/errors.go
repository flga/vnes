package main

import (
	"fmt"
	"strings"
)

type errorList struct {
	l []error
}

func (e *errorList) Errorf(format string, args ...interface{}) error {
	if e == nil {
		return nil
	}

	return fmt.Errorf(format, args...)
}

func (e *errorList) Add(errors ...error) *errorList {
	for _, err := range errors {
		if err == nil {
			continue
		}
		if e == nil {
			e = &errorList{}
		}
		e.l = append(e.l, err)
	}

	return e
}

func (e *errorList) Error() string {
	var slist []string
	for _, err := range e.l {
		slist = append(slist, err.Error())
	}
	return strings.Join(slist, ", ")
}
