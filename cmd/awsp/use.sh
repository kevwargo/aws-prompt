{{ range $name, $value := . -}}
export {{ $name }}={{ $value }};
{{ end }}
