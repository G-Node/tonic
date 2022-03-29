package templates

// TODO: Switch Login || Logout

// Layout is the main site template. It includes the header and footer and
// embeds the content for every other page.
var Layout = `
{{ define "layout" }}
<html>
	<!DOCTYPE html>
	<head data-suburl="">
		<link rel="shortcut icon" href="https://gindata.biologie.hu-berlin.de/img/favicon.png" />
		<link rel="stylesheet" href="/assets/font-awesome-4.6.3/css/font-awesome.min.css">
		<link rel="stylesheet" href="/assets/octicons-4.3.0/octicons.min.css">
		<link rel="stylesheet" href="/assets/semantic-2.3.1.min.css">
		<link rel="stylesheet" href="/assets/gogs.css">
		<link rel="stylesheet" href="/assets/custom.css">
		<title>Project creator</title>
		<meta name="twitter:card" content="summary" />
		<meta name="twitter:site" content="@gnode" />
		<meta name="twitter:title" content="GIN Valid"/>
		<meta name="twitter:description" content="G-Node GIN Validation service"/>
		<meta name="twitter:image" content="https://gindata.biologie.hu-berlin.de/img/favicon.png" />
	</head>
	<body>
		<div class="full height">
			<div class="following bar light">
				<div class="ui container">
					<div class="ui grid">
						<div class="column">
							<div class="ui top secondary menu">
								<a class="item brand" href="https://gindata.biologie.hu-berlin.de">
									<img class="ui mini image" src="https://gindata.biologie.hu-berlin.de/img/favicon.png">
								</a>
								<a class="item" href="/">New</a>
								<a class="item" href="/log">Jobs</a>
							</div>
						</div>
					</div>
				</div>
			</div>
			{{ template "content" . }}
		</div>
		<footer>
			<div class="ui container">
				<div class="ui center links item brand footertext">
					<a href="https://gindata.biologie.hu-berlin.de"><img class="ui mini footericon" src="https://projects.g-node.org/assets/gnode-bootstrap-theme/1.2.0-snapshot/img/gnode-icon-50x50-transparent.png"/>© tonic team 2022</a>
					<a href="https://gindata.biologie.hu-berlin.de/G-Node/Info/wiki/about">About</a>
					<a href="https://gindata.biologie.hu-berlin.de/G-Node/Info/wiki/imprint">Imprint</a>
					<a href="https://gindata.biologie.hu-berlin.de/G-Node/Info/wiki/contact">Contact</a>
					<a href="https://gindata.biologie.hu-berlin.de/G-Node/Info/wiki/Terms+of+Use">Terms of Use</a>
					<a href="https://gindata.biologie.hu-berlin.de/G-Node/Info/wiki/Datenschutz">Datenschutz</a>
				</div>
				<div class="ui center links item brand footertext">
				</div>
			</div>
		</footer>
	</body>
</html>
{{ end }}
`
