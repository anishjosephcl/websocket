package handler

const tmpl = `
<!DOCTYPE html>
<html>
<head>
	<title>Messages</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #4CAF50; color: white; }
		tr:nth-child(even) { background-color: #f2f2f2; }
	</style>
</head>
<body>
	<h1>Messages (TraceID: {{.TraceID}})</h1>
	<p>Total messages: {{.Count}}</p>
	<table>
		<tr>
			<th>User ID</th>
			<th>Message</th>
		</tr>
		{{range .Messages}}
		<tr>
			<td>{{.UserID}}</td>
			<td>{{.Message}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>
`
