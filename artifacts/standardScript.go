package artifacts

import (
	"html/template"
	"os"

	"github.com/arafat/please/schema"
)

type StandardScript struct {
	schema.ContainerArgs
	HostEnvs        map[string]string
	ApplicationArgs []string
	Image           string
	Version         string
	Application     string
	Platform        string
	Executable      string
}

const standardScriptTemplate = `#!/usr/bin/env bash
{{- range $key, $value := .HostEnvs }}
export {{ $key }}={{ $value }}
{{- end }}
exec container run -i --rm \
{{- range .DNS }}
  --dns {{.}} \
{{- end }}
{{- if .AdditionalFlags }}
{{- range .AdditionalFlags }}
  {{.}} \
{{- end }}
{{- end }}
{{- range .Volumes }}
  --volume {{.}} \
{{- end }}
{{- if .WorkDir }}
  --workdir {{.WorkDir}} \
{{- end }}
  --platform {{.Platform}} \
{{- range $key, $value := .ContainerEnvVars }}
{{- if $value }}
  -e {{ $key }}={{ $value }} \
{{- end }}
{{- end }}
  {{.Image}}:{{.Version}} \
{{- if .Executable }}
  {{.Executable}} \
{{- end }}
{{- if .ApplicationArgs }}
{{- range .ApplicationArgs }}
  {{.}} \
{{- end }}
{{- end }}
  "$@"
`

func (s *StandardScript) Deploy(path string) error {
	tmpl, err := template.New("script").Parse(standardScriptTemplate)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, s)
}
