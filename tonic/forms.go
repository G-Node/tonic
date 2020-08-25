package tonic

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
}
