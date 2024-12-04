package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-autoinst/gopenqa"
)

/* Read job groups from stdin */
func readJobGroups(filename string) ([]gopenqa.JobGroup, error) {
	var data []byte
	var err error

	if filename == "" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return make([]gopenqa.JobGroup, 0), err
		}
	} else {
		// TODO: Don't use io.ReadAll
		if file, err := os.Open(filename); err != nil {
			return make([]gopenqa.JobGroup, 0), err
		} else {
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				return make([]gopenqa.JobGroup, 0), err
			}
		}
	}

	// First try to read a single jobgroup
	var jobgroup gopenqa.JobGroup
	if err := json.Unmarshal(data, &jobgroup); err == nil {
		jobgroups := make([]gopenqa.JobGroup, 0)
		jobgroups = append(jobgroups, jobgroup)
		return jobgroups, nil
	}

	// Then try to read a jobgroup array
	var jobgroups []gopenqa.JobGroup
	if err := json.Unmarshal(data, &jobgroups); err == nil {
		return jobgroups, err
	}

	jobgroups = make([]gopenqa.JobGroup, 0)
	return jobgroups, fmt.Errorf("invalid input format")
}

func postJobGroups(args []string) error {
	files := args
	if len(files) == 0 {
		files = append(files, "")
	}

	for _, filename := range files {
		if jobgroups, err := readJobGroups(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		} else {
			for _, jobgroup := range jobgroups {
				if jobgroup, err := instance.PostJobGroup(jobgroup); err != nil {
					return err
				} else {
					fmt.Printf("Posted job group %d %s\n", jobgroup.ID, jobgroup.Name)
				}
			}
		}
	}

	return nil
}

func postParentJobGroups(args []string) error {
	files := args
	if len(files) == 0 {
		files = append(files, "")
	}

	for _, filename := range files {
		if jobgroups, err := readJobGroups(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		} else {
			for _, jobgroup := range jobgroups {
				if jobgroup, err := instance.PostParentJobGroup(jobgroup); err != nil {
					return err
				} else {
					fmt.Printf("Posted parent job group %d %s\n", jobgroup.ID, jobgroup.Name)
				}
			}
		}
	}

	return nil
}

func runJobGroups(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		if jobgroups, err := instance.GetJobGroups(); err != nil {
			return err
		} else {
			return printJson(jobgroups)
		}
	} else if method == "POST" {
		return postJobGroups(args)
	} else if method == "CLEAR" {
		if !cf.NoPrompt {
			fmt.Println("DANGER ZONE !!")
			fmt.Println("Are you sure you want to delete ALL job groups? THERE WILL BE NO UNDO, if you are hesitant then stop NOW.")
			fmt.Println("Deleting job groups also means to delete all attached jobs!")
			if prompt("Type uppercase 'yes' to continue: ") != "YES" {
				return fmt.Errorf("cancelled")
			}
		}

		if jobgroups, err := instance.GetJobGroups(); err != nil {
			return err
		} else {
			fmt.Printf("Delete %d job groups and their jobs ... \n", len(jobgroups))
			for i, jobgroup := range jobgroups {
				id := jobgroup.ID
				if err := instance.DeleteJobTemplate(id); err != nil {
					return err
				}
				if err := instance.DeleteJobGroupJobs(id); err != nil {
					return err
				}
				if err := instance.DeleteJobGroup(id); err != nil {
					return err
				}

				fmt.Printf("[%d/%d] Deleted job group %d %s\n", i, len(jobgroups), jobgroup.ID, jobgroup.Name)

			}
		}

		return nil
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}

func runParentGroups(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		if jobgroups, err := instance.GetParentJobGroups(); err != nil {
			return err
		} else {
			return printJson(jobgroups)
		}
	} else if method == "POST" {
		return postParentJobGroups(args)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}

func runJobGroup(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		ids, args := extractIntegers(args)
		if len(args) > 0 {
			return fmt.Errorf("too many arguments")
		}
		for _, id := range ids {
			if jobgroup, err := instance.GetJobGroup(id); err != nil {
				return err
			} else {
				if err := printJson(jobgroup); err != nil {
					return err
				}
			}
		}
		return nil
	} else if method == "POST" {
		return postJobGroups(args)
	} else if method == "DELETE" {
		ids, args := extractIntegers(args)
		if len(args) > 0 {
			return fmt.Errorf("invalid arguments")
		}
		for _, id := range ids {
			if err := instance.DeleteJobGroup(id); err != nil {
				return err
			}
			fmt.Printf("Deleted job group %d\n", id)
		}
		return nil
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}

func runParentGroup(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		ids, args := extractIntegers(args)
		if len(args) > 0 {
			return fmt.Errorf("too many arguments")
		}
		for _, id := range ids {
			if jobgroup, err := instance.GetParentJobGroup(id); err != nil {
				return err
			} else {
				if err := printJson(jobgroup); err != nil {
					return err
				}
			}
		}
		return nil
	} else if method == "POST" {
		return postParentJobGroups(args)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}
