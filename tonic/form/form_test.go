package form

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"github.com/G-Node/tonic/templates"
	"golang.org/x/net/html"
)

var allElementTypes = []ElementType{CheckboxInput, ColorInput, DateInput, DateTimeInput, EmailInput, FileInput, HiddenInput, MonthInput, NumberInput, PasswordInput, RadioInput, RangeInput, SearchInput, TelInput, TextInput, TimeInput, URLInput, WeekInput, TextArea, Select}

func TestFormElementsHTML(t *testing.T) {
	testElements := make([]Element, len(allElementTypes))
	for idx := range allElementTypes {
		elemType := allElementTypes[idx]
		elem := Element{
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
		testElements[idx] = elem
	}

	testPage := Page{
		Description: "One of each element supported by Tonic",
		Elements:    testElements,
	}
	testForm := Form{
		Pages:       []Page{testPage},
		Name:        "Tonic example form",
		Description: "",
	}

	// render form and check if it's valid HTML
	tmpl := template.New("layout")
	tmpl, err := tmpl.Parse(templates.Layout)
	if err != nil {
		t.Fatalf("Failed to parse Layout template: %s", err.Error())
	}
	tmpl, err = tmpl.Parse(templates.Form)
	if err != nil {
		t.Fatalf("Failed to parse Form template: %s", err.Error())
	}

	data := make(map[string]interface{})
	data["form"] = testForm

	formHTML := new(bytes.Buffer)
	if err := tmpl.Execute(formHTML, data); err != nil {
		t.Fatalf("Failed to render form: %v", err.Error())
	}

	if _, err := html.Parse(formHTML); err != nil {
		t.Fatalf("Bad HTML when rendering form: %v", err.Error())
	}
}
