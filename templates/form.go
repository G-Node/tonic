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
								{{.form.Name}}
							</h3>
							{{.form.Description}}
							<div class="ui attached segment">
								{{range $page := .form.Pages}}
									 <p>{{$page.Description}}</p> 
									{{range $elem := $page.Elements}}
										<div class="inline {{if $elem.Required}}required{{end}} field">
											{{$elem.HTML $.readonly}}
										</div>
									{{end}}
									<div class="ui divider"></div>
								{{end}}
								{{if not .readonly}}
									<div class="inline field">
										<label></label>
										<button class="ui green button">Submit</button>
									</div>
								{{end}}
							</div>

							{{if .readonly}}
								<h3 class="ui attached header">Status</h3>
								<div class="ui attached segment">
									{{if not .end_time}}
										<div class="ui message">
											Job is in queue
										</div>
									{{else if .error}}
										<div class="ui negative message">
											<b>Job failed with error:</b> {{.error}}
										</div>
									{{else}}
										<div class="ui positive message">
											Job completed <b>successfully</b>
										</div>
									{{end}}
									<ul class="list">
										<li><b>Submitted</b> {{.submit_time}}</li>
										{{if .end_time}}
											<li><b>Finished</b> {{.end_time}}</li>
										{{end}}
									</ul>
								<h3 class="ui attached header">Job log</h3>
									<ol class="list" start="0">
									{{range $msg := .messages}}
										<li>{{$msg}}</li>
									{{end}}
									</ol>
								{{end}}
							</div>
						</form>
					</div>
				</div>
			</div>
{{ end }}
`
