package templates

// Form and Job view page template
const Form = `
{{ define "content" }}
			<div class="ginform">
				<div class="ui middle very relaxed page grid">
					<div class="column">
						<form class="ui form" action="/" method="post">
							<input type="hidden" name="_csrf" value="">
							<h3 class="ui top attached header">
								Demo form
							</h3>
							<div class="ui attached segment">
								{{with .elements}}
									{{range $idx, $elem := .}}
										<div class="inline {{if $elem.Required}}required{{end}} field ">
											<label for="{{$elem.ID}}">{{$elem.Label}}</label>
											<input id="{{$elem.ID}}" name="{{$elem.Name}}" value="{{$elem.Value}}" autofocus {{if $elem.Required}}required{{end}} {{if $.readonly}}readonly{{end}}>
											<span class="help">{{$elem.Description}}</span>
										</div>
									{{end}}
								{{end}}
								{{if not .readonly}}
									<div class="inline field">
										<label></label>
										<button class="ui green button">Submit</button>
									</div>
								{{end}}
								{{if .submit_time}}
									<div>
										Submitted {{.submit_time}}
									</div>
									<div>
										{{if .end_time}}
											Finished {{.end_time}}
										{{else}}
											In queue
										{{end}}
									</div>
								{{end}}
								{{if .message}}
									<div>
										{{.message}}
									</div>
								{{end}}
							</div>
						</form>
					</div>
				</div>
			</div>
{{ end }}
`
