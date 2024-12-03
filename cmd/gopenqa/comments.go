package main

import (
	"fmt"
	"strconv"
	"strings"
)

func runComments(args []string) error {
	method := "GET"

	if len(args) < 1 {
		return fmt.Errorf("not enough arguments")
	}
	// get method
	var id int64
	if len(args) == 1 {
		method = "GET"
		id, _ = strconv.ParseInt(args[0], 10, 64)
	} else {
		method = strings.ToUpper(strings.TrimSpace(args[0]))
		id, _ = strconv.ParseInt(args[1], 10, 64)
		args = args[2:]
	}
	if id <= 0 {
		return fmt.Errorf("invalid ID")
	}

	if method == "GET" {
		comments, err := instance.GetComments(id)
		if err != nil {
			return err
		}
		if err := printJson(comments); err != nil {
			return err
		}
		return nil
	} else {
		return fmt.Errorf("Method %s is not (yet) supported", method)
	}
}
