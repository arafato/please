package artifacts

import (
	"html/template"
	"os"

	"github.com/arafat/please/schema"
)

type StandardScript struct {
	schema.ContainerArgs
	ApplicationArgs []string
	Image           string
	Version         string
	Application     string
	Platform        string
}

const standardScriptTemplate = `#!/usr/bin/env bash
container run{{range .DNS}} --dns {{.}}{{end}}{{range .AdditionalFlags}} {{.}}{{end}} --rm{{range .Volumes}} \
  --volume {{.}}{{end}}{{if .WorkDir}} \
  --workdir {{.WorkDir}}{{end}} \
  --platform {{.Platform}} \
  {{.Image}}:{{.Version}}{{if .ApplicationArgs}}{{range .ApplicationArgs}} \
  {{.}}{{end}}{{end}} "$@" 2>/dev/null
`

func (s *StandardScript) WriteScript(path string) error {
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
