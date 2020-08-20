package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/G-Node/tonic/tonic"
)

func main() {

	elems := []tonic.Element{
		{
			Name:  "name",
			Label: "Name",
		},
		{
			Name:  "description",
			Label: "Description",
		},
		{
			Name:        "duration",
			Label:       "Duration",
			Description: "Seconds to wait before finishing the job.  Use for simulating long-running jobs.",
		},
	}
	ut := tonic.NewService(elems, exampleFunc)
	ut.Config = &tonic.Config{GINServer: "https://gin.dev.g-node.org", CookieName: "utonic-example"}
	ut.Start()
	defer ut.Stop()
	ut.WaitForInterrupt()
}

func exampleFunc(values map[string]string) ([]string, error) {
	fail := false
	msgs := make([]string, 0)
	for k, v := range values {
		msg := fmt.Sprintf("Example function got %s: %q\n", k, v)
		msgs = append(msgs, msg)
		if v == "error" {
			msgs = append(msgs, "Found 'error' value. Stopping.")
			fail = true
		}
	}

	duration := values["duration"]
	if duration != "" {
		d, err := strconv.Atoi(duration)
		if err != nil {
			return msgs, fmt.Errorf("Duration not an integer: %s", err.Error())
		}
		msgs = append(msgs, fmt.Sprintf("Waiting %d seconds", d))
		time.Sleep(time.Second * time.Duration(d))
	}

	if fail {
		return msgs, fmt.Errorf("Failed to run: error detected")
	}
	msgs = append(msgs, "All OK. Example function finished successfully.")
	return msgs, nil
}
