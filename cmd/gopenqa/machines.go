package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-autoinst/gopenqa"
)

/* Read machines from stdin */
func readMachines(filename string) ([]gopenqa.Machine, error) {
	var data []byte
	var err error

	if filename == "" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			machines := make([]gopenqa.Machine, 0)
			return machines, err
		}
	} else {
		// TODO: Don't use io.ReadAll
		if file, err := os.Open(filename); err != nil {
			return make([]gopenqa.Machine, 0), err
		} else {
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				return make([]gopenqa.Machine, 0), err
			}
		}
	}

	// First try to read a single machine
	var machine gopenqa.Machine
	if err := json.Unmarshal(data, &machine); err == nil {
		machines := make([]gopenqa.Machine, 0)
		machines = append(machines, machine)
		return machines, nil
	}

	// Then try to read a machine array
	var machines []gopenqa.Machine
	if err := json.Unmarshal(data, &machines); err == nil {
		return machines, err
	}

	machines = make([]gopenqa.Machine, 0)
	return machines, fmt.Errorf("invalid input format")
}

func postMachines(args []string) error {
	files := args
	if len(files) == 0 {
		files = append(files, "")
	}

	for _, filename := range files {
		if machines, err := readMachines(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		} else {
			for _, machine := range machines {
				if machine, err := instance.PostMachine(machine); err != nil {
					return err
				} else {
					fmt.Printf("Posted machine %d %s:%s\n", machine.ID, machine.Name, machine.Backend)
				}
			}
		}
	}

	return nil
}

func runMachines(args []string) error {
	method := "GET"

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		if machines, err := instance.GetMachines(); err != nil {
			return err
		} else {
			return printJson(machines)
		}
	} else if method == "POST" {
		return postMachines(args)
	} else if method == "DELETE" {
		ids, _ := extractIntegers(args)
		if len(ids) == 0 {
			fmt.Fprintf(os.Stderr, "Missing machine ids\n")
		} else {
			for _, id := range ids {
				if err := instance.DeleteMachine(id); err != nil {
					return err
				} else {
					fmt.Printf("Deleted machine %d\n", id)
				}
			}
		}
		return nil
	} else if method == "CLEAR" {
		if !cf.NoPrompt {
			fmt.Println("DANGER ZONE !!")
			fmt.Println("Are you sure you want to delete ALL machines? THERE WILL BE NO UNDO, if you are hesitant then stop NOW.")
			if prompt("Type uppercase 'yes' to continue: ") != "YES" {
				return fmt.Errorf("cancelled")
			}
		}

		// Get machines and then delete them one by one
		if cf.Verbose {
			fmt.Println("Fetching machines ... ")
		}
		if machines, err := instance.GetMachines(); err != nil {
			return err
		} else {
			for i, machine := range machines {
				id := machine.ID
				if err := instance.DeleteMachine(id); err != nil {
					return err
				} else {
					fmt.Printf("[%d/%d] Deleted machine %d %s:%s\n", i, len(machines), machine.ID, machine.Name, machine.Backend)
				}
			}
		}

		return nil
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}

func runMachine(args []string) error {
	method := "GET"
	ids, args := extractIntegers(args)

	if len(args) > 0 {
		// get method
		method = args[0]
		args = args[1:]
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "GET" {
		for _, id := range ids {
			if machines, err := instance.GetMachine(id); err != nil {
				return err
			} else {
				if err := printJson(machines); err != nil {
					return err
				}
			}
		}
		return nil
	} else if method == "POST" {
		return postMachines(args)
	} else {
		return fmt.Errorf("invalid method: %s", method)
	}
}
