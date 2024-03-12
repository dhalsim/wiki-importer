{{.Title}}{{if not (eq .Title .OriginalTitle)}} (original {{.OriginalTitle}}){{end}} is a movie{{if eq .Status "Released"}} released in {{.ReleaseDate}}{{end}}.
{{if .PosterPath}}
![poster](https://media.themoviedb.org/t/p/w300_and_h450_bestv2{{.PosterPath}})
{{if .Tagline}}_{{.Tagline}}_{{end}}
{{end}}
## Plot

{{.Overview}}

## Characters

{{range .Cast}}  - [[{{.Name}}]] as **{{.Character}}**
{{end}}

## Information

Runtime
: {{.Runtime}} minutes

Produced in
{{- range .ProductionCountries}}
: [[{{.Name}}|{{.Iso31661}}]]
{{- end}}

{{- if .Budget -}}
Budget
: {{.Budget}} USD
{{- end}}

{{- if .Revenue -}}
Revenue
: {{.Revenue}} USD
{{- end}}

Popularity
: {{.Popularity}}

Genres
{{- range .Genres}}
: [[{{.Name}}]]
{{- end}}

{{- if .Homepage -}}Homepage
: {{.Homepage}}
{{- end}}

IMDB
: https://www.imdb.com/title/{{.ImdbID}}

Production companies
{{- range .ProductionCompanies}}
: [[{{.Name}}]] ({{.OriginCountry}})
{{- end}}
