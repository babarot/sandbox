package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	DirPath string `validate:"dirpath"`
}

func main() {
	validate := validator.New()

	paths := []string{
		"C:\\Users\\runneradmin\\.gomi",
		"C:\\Users\\runneradmin\\.gomi\\",
		"/home/user/.gomi",
		"/home/user/.gomi/",
		"./relative/path",
		"./relative/path/",
	}

	for _, path := range paths {
		cfg := &Config{DirPath: path}
		err := validate.Struct(cfg)
		if err != nil {
			if _, ok := err.(*validator.InvalidValidationError); ok {
				fmt.Printf("Internal validator error: %v\n", err)
				continue
			}
			for _, err := range err.(validator.ValidationErrors) {
				fmt.Printf("Path %q is invalid: %v\n", path, err)
			}
		} else {
			fmt.Printf("Path %q is valid\n", path)
		}
	}
}
