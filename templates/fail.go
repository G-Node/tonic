package templates

var Fail = `
{{ define "content" }}

<br><br>
<h1>{{ .StatusCode }}: {{ .StatusText }}</h1>
<div style="color: red; font-weight: bold">
{{ .Message }}
</div>

{{ end }}
`
