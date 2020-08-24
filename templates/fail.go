package templates

// Fail template for displaying error messages to the user.
// Use for 4xx and 5xx responses.
var Fail = `
{{ define "content" }}

<br><br>
<h1>{{ .StatusCode }}: {{ .StatusText }}</h1>
<div style="color: red; font-weight: bold">
{{ .Message }}
</div>

{{ end }}
`
