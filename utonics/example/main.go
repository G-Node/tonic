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

var allElementTypes = []form.ElementType{form.CheckboxInput, form.ColorInput, form.DateInput, form.DateTimeInput, form.EmailInput, form.FileInput, form.HiddenInput, form.MonthInput, form.NumberInput, form.PasswordInput, form.RadioInput, form.RangeInput, form.SearchInput, form.TelInput, form.TextInput, form.TimeInput, form.URLInput, form.WeekInput, form.TextArea, form.Select}

func main() {
	pageOneElems := []form.Element{
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
	examplePage := form.Page{
		Description: "Page 1 of example form",
		Elements:    pageOneElems,
	}

	demoElements := make([]form.Element, len(allElementTypes))
	for idx := range allElementTypes {
		elemType := allElementTypes[idx]
		elem := form.Element{
			ID:          fmt.Sprintf("id%s", elemType),
			Name:        string(elemType),
			Label:       fmt.Sprintf("Element type %s", elemType),
			Description: fmt.Sprintf("An element of type %s", elemType),
			ValueList: []string{ // will only have effect on the types where it's valid
				fmt.Sprintf("%s option one", elemType),
				fmt.Sprintf("%s option two", elemType),
				fmt.Sprintf("%s option three", elemType),
			},
			Type: elemType,
		}
		demoElements[idx] = elem
	}

	elementDemoPage := form.Page{
		Description: "One of each element supported by Tonic",
		Elements:    demoElements,
	}
	exampleForm := form.Form{
		Pages:       []form.Page{examplePage, elementDemoPage},
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
	ut, err := tonic.NewService(exampleForm, nil, exampleFunc, config)
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
