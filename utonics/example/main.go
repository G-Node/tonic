package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/G-Node/tonic/tonic"
	"github.com/G-Node/tonic/tonic/form"
	"github.com/G-Node/tonic/tonic/worker"
)

func main() {
	elems := []form.Element{
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
	page := form.Page{
		Description: "Page 1 of example form",
		Elements:    elems,
	}
	form := form.Form{
		Pages:       []form.Page{page},
		Name:        "Tonic example form",
		Description: "",
	}
	config := tonic.Config{
		GINServer:   "", // not used in example
		GINUsername: "", // not used in example
		GINPassword: "", // not used in example
		CookieName:  "utonic-example",
		Port:        3000,
		DBPath:      "./example.db",
	}
	ut, err := tonic.NewService(form, exampleFunc, config)
	if err != nil {
		log.Fatal(err)
	}
	ut.Start()
	defer ut.Stop()
	ut.WaitForInterrupt()
}

func exampleFunc(values map[string]string, _, _ *worker.Client) ([]string, error) {
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
