package templates

// TODO: add csrf thingie
const Login = `
{{ define "content" }}
			<div class="user signin">
				<div class="ui middle very relaxed page grid">
					<div class="column">
						<form class="ui form" action="/login" method="post">
							<input type="hidden" name="_csrf" value="">
							<h3 class="ui top attached header">
								Sign In using your GIN credentials
							</h3>
							<div class="ui attached segment">
								<div class="required inline field ">
									<label for="username">Username or email</label>
									<input id="username" name="username" value="" autofocus required>
								</div>
								<div class="required inline field ">
									<label for="password">Password</label>
									<input id="password" name="password" type="password" autocomplete="off" value="" required>
								</div>
								<div class="inline field">
									<label></label>
									<button class="ui green button">Sign In</button>
								</div>
							</div>
						</form>
					</div>
				</div>
			</div>
{{ end }}
`


// vim: ft=gohtmltmpl
