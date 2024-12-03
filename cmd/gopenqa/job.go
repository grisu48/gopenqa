package main

import (
	"fmt"
	"strconv"
)

func runJobState(args []string) error {
	var id int64
	if len(args) < 1 {
		return fmt.Errorf("missing argument: job")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	id, _ = strconv.ParseInt(args[0], 10, 64)
	if id <= 0 {
		return fmt.Errorf("invalid ID")
	}

	state, err := instance.GetJobState(id)
	if err != nil {
		return err
	}
	if state.BlockedBy > 0 {
		fmt.Printf("Blocked by %d\n", state.BlockedBy)
	}
	if state.State == "done" {
		fmt.Println(state.Result)
	} else {
		fmt.Println(state.State)
	}
	return nil
}

func runJob(args []string) error {

	var id int64
	if len(args) < 1 {
		return fmt.Errorf("missing argument: job")
	} else if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	id, _ = strconv.ParseInt(args[0], 10, 64)
	if id <= 0 {
		return fmt.Errorf("invalid ID")
	}

	job, err := instance.GetJobFollow(id)
	if err != nil {
		return err
	}
	fmt.Println(job.String())
	return nil
}

func runJobs(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing argument: jobs")
	}
	ids := make([]int64, 0)

	for _, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if id <= 0 || err != nil {
			return fmt.Errorf("invalid ID")
		}
		ids = append(ids, id)
	}

	jobs, err := instance.GetJobsFollow(ids)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		fmt.Printf("%s\n", job.String())
	}
	return nil
}
