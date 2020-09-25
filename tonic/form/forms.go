package form

import (
	"fmt"
	"html/template"
	"strings"
)

const (
	// CheckboxInput is a fieldset that groups a number of "checkbox" inputs for the same variable.
	CheckboxInput ElementType = "checkbox"
	// ColorInput is an input element with type "color".
	ColorInput ElementType = "color"
	// DateInput is an input element with type "date"."
	DateInput ElementType = "date"
	// DateTimeInput is an input element with type "datetime-local".
	DateTimeInput ElementType = "datetime-local"
	// EmailInput is an input element with type "email".
	EmailInput ElementType = "email"
	// FileInput is an input element with type "file".
	FileInput ElementType = "file"
	// HiddenInput is an input element with type "hidden".
	HiddenInput ElementType = "hidden"
	// MonthInput is an input element with type "month".
	MonthInput ElementType = "month"
	// NumberInput is an input element with type "number".
	NumberInput ElementType = "number"
	// PasswordInput is an input element with type "password".
	PasswordInput ElementType = "password"
	// RadioInput is a fieldset that groups a number of "radio" inputs for the same variable.
	RadioInput ElementType = "radio"
	// RangeInput is an input element with type "range".
	RangeInput ElementType = "range"
	// SearchInput is an input element with type "search".
	SearchInput ElementType = "search"
	// TelInput is an input element with type "tel".
	TelInput ElementType = "tel"
	// TextInput is an input element with type "text".
	TextInput ElementType = "text"
	// TimeInput is an input element with type "time".
	TimeInput ElementType = "time"
	// URLInput is an input element with type "url".
	URLInput ElementType = "url"
	// WeekInput is an input element with type "week".
	WeekInput ElementType = "week"
	// TextArea is an "textarea" element.
	TextArea ElementType = "textarea"
	// Select is an input element with type "select".  Requires a ValueList.
	Select ElementType = "select"
)

// ElementType defines the type of a form input element:
// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input
type ElementType string

// Form is the top level type for defining the web form for user input.
type Form struct {
	// The Name appears at the top of all pages in the form and in the HTML
	// title.
	Name string
	// The Description appears under the Name on every page.
	Description string
	// Each Page creates a form page with the included elements.  The last page
	// contains the submit button.
	Pages []Page
}

// Page represents a single page of a multi-page web form.
type Page struct {
	// The Description appears under the form description.  Use it to provide
	// information about the elements of the specific page.
	Description string
	// Each element creates an input field on the form.
	Elements []Element
}

// Element represents a single form element (field).
type Element struct {
	// ID of the element.  Must be unique.
	ID string
	// Name of the element.  Used as key to retrieve the value on submission.
	Name string
	// The Label of the field as it appears on the rendered form.
	Label string
	// If set, the field will be filled with the given value, or the
	// appropriate option will be selected, when rendered.
	Value string
	// Whether the element represents a required form field.
	Required bool
	// An optional description for the field.  If set will be displayed under
	// the input field.  Can be used to provide extra information such as input
	// constraints.
	Description string
	// Type is the HTML input element type.
	Type ElementType
	// ValueList should contain a set of values that represent the permissible
	// or recommended options available to the element.  For input type
	// elements, it represents suggested values (datalist).  For select
	// elements, it represents the values in the list.
	ValueList []string
	// Read only fields can't be edited.
	ReadOnly bool
}

// HTML returns the HTML representation of this element.
func (e *Element) HTML(ro bool) template.HTML {
	label := fmt.Sprintf("<label for=%q>%s</label>", e.ID, e.Label)
	var field, required, readonly, disabled string
	if e.Required {
		required = "required"
	}
	if e.ReadOnly || ro {
		readonly = "readonly"
		disabled = "disabled"
	}

	// Note that checkbox and radio use fieldset to group multiple elements so
	// use a different layout and return early
	switch e.Type {
	case RadioInput:
		fallthrough
	case CheckboxInput:
		lines := make([]string, 0, len(e.ValueList)*4+4)
		lines = append(lines, fmt.Sprintf("<fieldset %s>", disabled))
		lines = append(lines, fmt.Sprintf("<legend>%s</legend>", e.Label))
		for _, value := range e.ValueList {
			checked := ""
			if sliceContains(strings.Split(e.Value, "\n"), value) {
				checked = "checked"
			}
			lines = append(lines, "<div>")
			field = fmt.Sprintf("<input type=%q id=%q name=%q value=%q %s %s>", e.Type, value, e.ID, value, required, checked)
			lines = append(lines, field)
			lines = append(lines, fmt.Sprintf("<label for=%q>%s</label>", value, value))
			lines = append(lines, "</div>")
		}
		lines = append(lines, "</fieldset")
		description := fmt.Sprintf("<span class=\"help\">%s</span>", e.Description)
		return template.HTML(strings.Join(lines, "\n") + description)
	case TextArea:
		field = fmt.Sprintf("<textarea id=%q name=%q %s %s>%s</textarea>", e.ID, e.Name, required, readonly, e.Value)
	case Select:
		lines := make([]string, 0, len(e.ValueList)+2)
		lines = append(lines, fmt.Sprintf("<select id=%q name=%q>", e.ID, e.Name))
		for _, value := range e.ValueList {
			lines = append(lines, fmt.Sprintf("<option value=%q>%s</option>", value, value))
		}
		lines = append(lines, "</select>")
		field = strings.Join(lines, "\n")
	default:
		lines := make([]string, 0, len(e.ValueList)+3)
		var valueListID string
		if len(e.ValueList) > 0 {
			valueListID = fmt.Sprintf("%s-values", e.ID)
		}
		lines = append(lines, fmt.Sprintf("<input type=%q id=%q name=%q value=%q %s %s list=%q>", e.Type, e.ID, e.Name, e.Value, required, readonly, valueListID))

		lines = append(lines, fmt.Sprintf("<datalist id=%q>", valueListID))
		for _, value := range e.ValueList {
			lines = append(lines, fmt.Sprintf("<option value=%q>", value))
		}
		lines = append(lines, "</datalist>")
		field = strings.Join(lines, "\n")
	}
	description := fmt.Sprintf("<span class=\"help\">%s</span>", e.Description)
	return template.HTML(label + field + description)
}

func sliceContains(strSlice []string, value string) bool {
	for _, slv := range strSlice {
		if slv == value {
			return true
		}
	}
	return false
}
