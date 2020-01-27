package errors

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
)

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}

type base struct {
	Err         error
	Files       map[string]*hcl.File
	Diagnostics hcl.Diagnostics
}

func (e *base) Error() string {
	switch {
	case e.Diagnostics.HasErrors():
		if len(e.Files) == 0 {
			return e.Diagnostics.Error()
		}
		var b bytes.Buffer
		wr := hcl.NewDiagnosticTextWriter(
			&b,      // writer to send messages to
			e.Files, // the parser's file cache, for source snippets
			100,     // wrapping width
			true,    // generate colored/highlighted output
		)
		wr.WriteDiagnostics(e.Diagnostics)
		return b.String()
	case e.Err != nil:
		return e.Err.Error()
	}
	return ""
}

// New is
func New(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	e := &base{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			e.Err = &errorString{arg}
		// case *Error:
		// 	// Make a copy
		// 	copy := *arg
		// 	e.Err = &copy
		case hcl.Diagnostics:
			e.Diagnostics = arg
		case map[string]*hcl.File:
			e.Files = arg
		case error:
			e.Err = arg
		default:
			panic(args)
			// _, file, line, _ := runtime.Caller(1)
			// log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
			// return Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	if len(e.Diagnostics) == 0 && e.Err == nil {
		return nil
	}

	return e
}
