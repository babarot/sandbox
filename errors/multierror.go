package errors

// Copied from https://github.com/hashicorp/go-multierror/

import (
	"fmt"
	"strings"
)

// Stack is an error type to track multiple errors. This is used to
// accumulate errors in cases and return them as a single "error".
type Stack struct {
	Errors      []error
	ErrorFormat ErrorFormatFunc
}

func (e *Stack) Error() string {
	fn := e.ErrorFormat
	if fn == nil {
		// fn = ListFormatFunc
		fn = RichListFormatFunc
	}

	return fn(e.Errors)
}

// ErrorOrNil returns an error interface if this Error represents
// a list of errors, or returns nil if the list of errors is empty. This
// function is useful at the end of accumulation to make sure that the value
// returned represents the existence of errors.
func (e *Stack) ErrorOrNil() error {
	if e == nil {
		return nil
	}
	if len(e.Errors) == 0 {
		return nil
	}

	return e
}

// WrappedErrors returns the list of errors that this Error is wrapping.
// It is an implementation of the errwrap.Wrapper interface so that
// multierror.Error can be used with that library.
//
// This method is not safe to be called concurrently and is no different
// than accessing the Errors field directly. It is implemented only to
// satisfy the errwrap.Wrapper interface.
func (e *Stack) WrappedErrors() []error {
	return e.Errors
}

// ErrorFormatFunc is a function callback that is called by Error to
// turn the list of errors into a string.
type ErrorFormatFunc func([]error) string

// RichListFormatFunc is a basic formatter that outputs the number of errors
// that occurred along with a bullet point list of the errors.
func RichListFormatFunc(es []error) string {
	format := func(text string) string {
		var s string
		lines := strings.Split(text, "\n")
		switch len(lines) {
		default:
			s += lines[0]
			for _, line := range lines[1:] {
				// if line == "" {
				// 	continue
				// }
				s += fmt.Sprintf("\n\t  %s", line)
			}
		case 0:
			s = es[0].Error()
		}
		return s
	}

	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", format(es[0].Error()))
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", format(err.Error()))
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(es), strings.Join(points, "\n\t"))
}

// ListFormatFunc is a basic formatter that outputs the number of errors
// that occurred along with a bullet point list of the errors.
func ListFormatFunc(es []error) string {
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", es[0])
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(es), strings.Join(points, "\n\t"))
}

// Append is a helper function that will append more errors
// onto an Error in order to create a larger multi-error.
//
// If err is not a multierror.Error, then it will be turned into
// one. If any of the errs are multierr.Error, they will be flattened
// one level into err.
func Append(err error, errs ...error) *Stack {
	switch err := err.(type) {
	case *Stack:
		// Typed nils can reach here, so initialize if we are nil
		if err == nil {
			err = new(Stack)
		}

		// Go through each error and flatten
		for _, e := range errs {
			switch e := e.(type) {
			case *Stack:
				if e != nil {
					err.Errors = append(err.Errors, e.Errors...)
				}
			default:
				if e != nil {
					err.Errors = append(err.Errors, e)
				}
			}
		}

		return err
	default:
		newErrs := make([]error, 0, len(errs)+1)
		if err != nil {
			newErrs = append(newErrs, err)
		}
		newErrs = append(newErrs, errs...)

		return Append(&Stack{}, newErrs...)
	}
}
