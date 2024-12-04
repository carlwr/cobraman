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
	RegisterTemplate("troff", "-", "use_section", troffManTemplate)
}

// troffManTemplate generates a man page with only basic troff macros.
// nolint:lll // this is a template
const troffManTemplate = `.TH "{{.CommandPath | dashify | backslashify | upper}}" "{{ .Section }}" "{{.CenterFooter}}" "{{.LeftFooter}}" "{{.CenterHeader}}" 
.nh    {{/* disable hyphenation */}}
.ad l  {{/* disable justification (adjust text to left margin only) */}}
.SH NAME
{{ .CommandPath | dashify | backslashify }}
{{- if .ShortDescription }} - {{ .ShortDescription }}
 {{- end }}
.SH SYNOPSIS
.sp
{{- if .SubCommands }}
{{- range .SubCommands }}
\fB{{ .CommandPath }}\fR [ flags ]
.br{{ end }}
{{- else }}
\fB{{ .CommandPath }} \fR
{{- range .AllFlags -}}
[{{ if .Shorthand }}\fI{{ print "-" .Shorthand | backslashify }}\fP|{{ end -}}
\fI{{ print "--" .Name | backslashify }}\fP] {{ end }}
{{- if not .NoArgs }}[<args>]{{ end }}
{{- end }}
.SH DESCRIPTION
.PP
{{ .Description | simpleToTroff }}
{{- if .AllFlags }}
.SH OPTIONS
{{ range .AllFlags -}}
.TP
{{ if .Shorthand }}\fB{{ print "-" .Shorthand | backslashify }}\fP, {{ end -}}
\fB{{ print "--" .Name | backslashify }}\fP{{ if not .NoOptDefVal }} =
{{- if .ArgHint }} <{{ .ArgHint }}>{{ else }} {{ .DefValue }}{{ end }}{{ end }}
{{ .Usage | backslashify }}
{{ end }}
{{- end -}}
{{- if .Environment }}
.SH ENVIRONMENT
.PP
{{ .Environment | simpleToTroff }}
{{- end }}
{{- if .Files }}
.SH FILES
.PP
{{ .Files | simpleToTroff }}
{{- end }}
{{- if .Bugs }}
.SH BUGS
.PP
{{ .Bugs | simpleToTroff }}
{{- end }}
{{- if .Examples }}
.SH EXAMPLES
.PP
{{ .Examples | simpleToTroff }}
{{- end }}
.SH AUTHOR
{{- if .Author }}
{{ .Author }}
{{- end }}
.PP
{{- if .SeeAlsos }}
.SH SEE ALSO
{{- range .SeeAlsos }}
.BR {{ .CmdPath | dashify | backslashify }} ({{ .Section }})
{{- end }}
{{- end }}
`
