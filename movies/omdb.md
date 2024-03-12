{{.Title}} is a {{.Year}} {{.Type}}{{if .Writer}} written by {{.Writer}} and{{end}}{{if .Director}} directed by {{.Director}}{{end}}{{if .Actors}} starring {{.Actors}}{{end}}.

![poster]({{.Poster}})

{{if .BoxOffice -}}
## Box Office

{{.BoxOffice}}
{{- end}}

{{if .Awards -}}
## Awards

{{.Awards}}
{{- end}}

## Plot

{{.Plot}}

## Ratings

| Source | Value |
|   ---  |  ---  |
{{range .Ratings -}}
| {{.Source}} | {{.Value}} |
{{- end}}
