package main

import (
	"fmt"

	"github.com/b4b4r07/sandbox/errors"
)

func main() {
	var errs *errors.Stack

	if err := run1(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := run2(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := run3(); err != nil {
		errs = errors.Append(errs, err)
	}
	if err := run4(); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to run4"))
	}

	fmt.Printf("%s\n", errs)
}

func run1() error {
	return errors.New("hoge")
}

func run2() error {
	return errors.Detail{
		Head:    "failed to run",
		Summary: "command not found",
		Details: []string{
			"run2",
			"run2",
			"run2",
		},
	}
}

func run3() error {
	return errors.Detail{
		Head:    "failed to run",
		Summary: "command not found",
		Details: []string{
			"run3",
			"run3",
			"run3",
		},
	}
}

func run4() error {
	return errors.New("run4\nerror")
}
