// Copyright © 2018 Ray Johnson <ray.johnson@gmail.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cobraman

func init() {
	RegisterTemplate("mdoc", "-", "use_section", mdocManTemplate)
}

// mdocManTemplate is a template what will use the mdoc macro package.
// TODO: The Dt macro can take one additonal arg - what does it do?
const mdocManTemplate = `.\" Man page for {{.CommandPath}}
.Dd {{ .Date.Format "January 2006"}}
{{ if .CenterHeader -}}
.Dt {{.CommandPath | dashify | backslashify | upper}} \&{{ .Section }} "{{.CenterHeader}}" 
{{- else -}}
.Dt {{.CommandPath | dashify | backslashify | upper}} {{ .Section }}
{{- end }}
.Sh NAME
.Nm {{ .CommandPath | dashify | backslashify }}
{{- if .ShortDescription }}
.Nd {{ .ShortDescription }}
{{- end }}
.Sh SYNOPSIS
{{- if .SubCommands }}
{{- range .SubCommands }}
.Nm {{ .CommandPath }} Op Fl flags Op args
{{- end }}
{{- else }}
.Nm {{ .CommandPath }}
{{- range .AllFlags }}
.Op Fl {{ if .Shorthand }}{{ .Shorthand | backslashify }} | {{ end -}}
{{ print "-" .Name | backslashify }}
{{- end }}
{{ if not .NoArgs }}.Op Fl <args>
{{- end }}
{{- end }}
.Ek
.Sh DESCRIPTION
.Nm
{{ .Description | simpleToMdoc }}
{{- if .AllFlags }}
.Pp
The options are as follows:
.Pp
.Bl -tag -width Ds -compact
{{ range .AllFlags -}}
.Pp
.It {{ if .Shorthand }}Fl {{ .Shorthand | backslashify }}, {{ end -}}
Fl {{ print "-" .Name | backslashify }}
{{- if not .NoOptDefVal }} Ar {{if .ArgHint }} {{ .ArgHint }}{{ else }} {{ .DefValue }}{{ end }}{{ end }}
{{ .Usage | backslashify }}
{{ end }}
.El
{{- end }}
{{- if .Environment }}
.Sh ENVIRONMENT
{{ .Environment | simpleToMdoc }}
{{- end }}
{{- if .Files }}
.Sh FILES
{{ .Files | simpleToMdoc }}
{{- end }}
{{- if .Bugs }}
.Sh BUGS
{{ .Bugs | simpleToMdoc }}
{{- end }}
{{- if .Examples }}
.Sh EXAMPLES
{{ .Examples | simpleToMdoc }}
{{- end }}
{{- if .Author }}
.Sh AUTHOR
{{ .Author }}
{{- end }}
{{- if .SeeAlsos }}
.Sh SEE ALSO
{{- range $index, $element := .SeeAlsos}}
{{- if $index}} ,{{end}}
.Xr {{$element.CmdPath}} {{$element.Section}}
{{- end }}
{{- end }}
`
