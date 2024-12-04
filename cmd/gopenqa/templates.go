package main

import (
	"fmt"
	"os"
	"strings"
)

func runJobTemplates(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		fmt.Fprintf(os.Stderr, "Getting job templates ... ")
		templates, err := instance.GetJobTemplates()
		if err != nil {
			return err
		}
		return printJson(templates)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}
