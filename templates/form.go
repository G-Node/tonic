package templates

// Temporary form for demonstration with fixed fields
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

const LogView = `
{{define "content"}}
	<div class="repository file list">
		<div class="header-wrapper">
			<div class="ui container">
				<div class="ui vertically padded grid head">
					<div class="column">
						<div class="ui header">
							<div class="ui huge breadcrumb">
								<i class="mega-octicon octicon-repo"></i>
							</div>
						</div>
					</div>
				</div>
			</div>
			<div class="ui tabs container">
			</div>
			<div class="ui tabs divider"></div>
		</div>
		<div class="ui container">
			<p id="repo-desc">
			<span class="description has-emoji">Work log</span>
			<a class="link" href=""></a>
			</p>
			<table id="repo-files-table" class="ui unstackable fixed single line table">
				<tbody>
					{{range $job := .}}
						<tr>
							<td class="name two wide">J{{$job.ID}}</td>
							<td class="name text bold nine wide"><a href="/log/{{$job.ID}}">{{$job.Label}}</a></td>
							<td class="name four wide">{{$job.SubmitTime}}</td>
							<td class="name four wide">{{$job.EndTime}}</td>
						</tr>
					{{end}}
				</tbody>
			</table>
		</div>
	</div>
{{end}}
`
