package buildinfo

import (
	"html/template"
	"io"
	"runtime"
	"runtime/debug"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

var versionTemplate = `
Version:		{{.Version}}
Commit:			{{.Commit}}
Date: 			{{.Date}}
Go version:		{{.GoVersion}}
OS/Arch:		{{.Os}}/{{.Arch}}
`

func PrintVersion(wr io.Writer) error {
	version, commit, date := getVersionInfo()

	tmpl, err := template.New("").Parse(versionTemplate)
	if err != nil {
		return err
	}

	v := struct {
		Version   string
		Commit    string
		Date      string
		GoVersion string
		Os        string
		Arch      string
	}{
		Version:   version,
		Commit:    commit,
		Date:      date,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	return tmpl.Execute(wr, v)
}

func getVersionInfo() (string, string, string) {
	if version != "" && commit != "" && date != "" {
		return version, commit, date
	}

	v, c, d := "dev", "none", "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		v = info.Main.Version
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				c = s.Value[:7]
			case "vcs.time":
				d = s.Value
			}
		}
	}
	return v, c, d
}
