package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func extractIntegers(args []string) ([]int, []string) {
	ids := make([]int, 0)
	rem := make([]string, 0)

	for _, arg := range args {
		if id, err := strconv.Atoi(arg); err == nil {
			ids = append(ids, id)
		} else {
			rem = append(rem, arg)
		}
	}
	return ids, rem
}

func prompt(msg string) string {
	if msg != "" {
		fmt.Print(msg)
		os.Stdout.Sync()
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(line)
}

// Replace some magic by their remote
func magicRemote(remote string) string {
	// Skip if remote
	if strings.HasPrefix(remote, "http://") || strings.HasPrefix(remote, "https://") {
		return remote
	}
	if remote == "" || remote == "ooo" || remote == "o3" {
		return "https://openqa.opensuse.org"
	} else if remote == "osd" {
		return "http://openqa.suse.de"
	} else if remote == "duck" {
		return "http://duck-norris.qam.suse.de"
	}
	return remote
}

func printJson(data interface{}) error {
	// Print as json
	if buf, err := json.Marshal(data); err != nil {
		return err
	} else {
		fmt.Println(string(buf))
		return nil
	}
}
