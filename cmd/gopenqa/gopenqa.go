/*
 * This is a example CLI tool for demonstrating the usage of gopenqa. This has no further purposes and is not intended for productive usage
 */
package main

import (
	"fmt"
	"os"

	"github.com/os-autoinst/gopenqa"
)

var cf Config
var instance gopenqa.Instance

func usage() {
	var cf Config // Create new config with defaults to display them
	cf.ApplyDefaults()
	fmt.Printf("Usage: %s [OPTIONS] ENTITY [METHOD] [COMMAND]\n", os.Args[0])
	fmt.Println("OPTIONS")
	fmt.Printf("  -r, --remote INSTANCE                           Define the openQA instance (default: %s)\n", cf.Remote)
	fmt.Println("  -k, --apikey KEY                               Set APIKEY for instance")
	fmt.Println("  -s, --apisecret SECRET                         Set APISECRET for instance")
	fmt.Println("  -v, --verbose                                  Verbose run")
	fmt.Println("  -y                                             No prompt")
	fmt.Println("")
	fmt.Println("ENTITY")
	fmt.Println("")
	fmt.Println("  job [ID]")
	fmt.Println("  jobs [IDS...]")
	fmt.Println("  jobgroup(s)")
	fmt.Println("  machine(s)")
	fmt.Println("  product(s) | medium(s)")
	fmt.Println("  parentgroup(s)")
	fmt.Println("  comments")
	fmt.Println("  jobstate")
}

func parseArgs(args []string) (string, []string, error) {
	entity := ""
	commands := make([]string, 0)

	n := len(args)
	for i := 0; i < n; i++ {
		arg := args[i]
		if arg == "" {
			continue
		}
		if arg[0] == '-' {
			if arg == "-h" || arg == "--help" {
				usage()
				os.Exit(0)
			} else if arg == "-r" || arg == "--remote" || arg == "--openqa" {
				i++
				cf.Remote = magicRemote(args[i])

			} else if arg == "-k" || arg == "--apikey" {
				i++
				cf.ApiKey = args[i]
			} else if arg == "-s" || arg == "--apisecret" {
				i++
				cf.ApiSecret = args[i]
			} else if arg == "-v" || arg == "--verbose" {
				cf.Verbose = true
			} else if arg == "-y" || arg == "--yes" {
				cf.NoPrompt = true
			} else {
				return entity, args, fmt.Errorf("Invalid argument: %s", arg)
			}
		} else {
			commands = append(commands, arg)
		}
	}

	if len(commands) == 0 {
		return "", commands, fmt.Errorf("no entity given")
	} else {
		entity := commands[0]
		commands = commands[1:]
		return entity, commands, nil
	}
}

func main() {
	cf.ApplyDefaults()

	if len(os.Args) <= 1 {
		usage()
		os.Exit(1)
	}

	entity, command, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	instance = gopenqa.CreateInstance(cf.Remote)
	instance.SetApiKey(cf.ApiKey, cf.ApiSecret)
	instance.SetVerbose(cf.Verbose)
	if entity == "h" || entity == "help" {
		usage()
		os.Exit(0)
	} else if entity == "jobtemplates" || entity == "templates" {
		if err := runJobTemplates(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "machines" {
		if err := runMachines(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "machine" {
		if err := runMachine(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "products" || entity == "mediums" {
		if err := runProducts(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "product" || entity == "medium" {
		if err := runProduct(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "jobgroups" || entity == "job_groups" {
		if err := runJobGroups(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "jobgroup" || entity == "job_group" {
		if err := runJobGroup(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "parentgroup" || entity == "parent_group" || entity == "parent_job_group" || entity == "parentjobgroup" {
		if err := runParentGroup(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "parentgroups" || entity == "parent_groups" || entity == "parent_job_groups" || entity == "parentjobgroups" {
		if err := runParentGroups(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "comments" {
		if err := runComments(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "jobstate" || entity == "state" {
		if err := runJobState(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "job" {
		if err := runJob(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else if entity == "jobs" {
		if err := runJobs(command); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Invalid entity: %s\n", entity)
		os.Exit(1)
	}

}
