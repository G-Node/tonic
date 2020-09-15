package form

const (
	CheckboxInput ElementType = "checkbox"
	ColorInput    ElementType = "color"
	DateInput     ElementType = "date"
	DateTimeInput ElementType = "datetime-local"
	EmailInput    ElementType = "email"
	FileInput     ElementType = "file"
	HiddenInput   ElementType = "hidden"
	ImageInput    ElementType = "image"
	MonthInput    ElementType = "month"
	NumberInput   ElementType = "number"
	PasswordInput ElementType = "password"
	RangeInput    ElementType = "range"
	SearchInput   ElementType = "search"
	TelInput      ElementType = "tel"
	TextInput     ElementType = "text"
	TimeInput     ElementType = "time"
	URLInput      ElementType = "url"
	WeekInput     ElementType = "week"
	TextArea      ElementType = "textarea"
	Select        ElementType = "select"
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
