package templates

// LogView template for displaying event log in a list.
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
							<td class="name text bold four wide"><a href="/log/{{$job.ID}}">Job {{$job.ID}}</a></td>
							<td class="name four wide">{{$job.SubmitTime}}</td>
							<td class="name four wide">{{$job.EndTime}}</td>
							<td class="name four wide">{{if $job.Error}}{{$job.Error}}{{end}}</td>
						</tr>
					{{end}}
				</tbody>
			</table>
		</div>
	</div>
{{end}}
`
