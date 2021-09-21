package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type YamlFile struct {
	Name    string
	Traefik bool
	Istio   bool
	Common  bool
	Job     bool
	App     bool
	Content string
}

type context struct {
	Yamls []YamlFile
}

var (
	yamlsTemplate = `
package templates
// Code generated by templates generator; DO NOT EDIT.


type Yamls struct {
	TraefikYamls map[string]string
	IstioYamls map[string]string
	JobYamls map[string]string
	AppYamls map[string]string
}

var GeneratedYamls = Yamls{
  TraefikYamls: map[string]string {
{{- range $_, $yaml := .Yamls }}
{{- if or $yaml.Traefik $yaml.Common }} 
    "{{ $yaml.Name }}": 
{{ $yaml.Content }},
{{- end }}
{{- end }}
	},
  IstioYamls: map[string]string {
{{- range $_, $yaml := .Yamls }}
{{- if or $yaml.Istio $yaml.Common }} 
    "{{ $yaml.Name }}": 
{{ $yaml.Content }},
{{- end }}
{{- end }}
},
  JobYamls: map[string]string {
{{- range $_, $yaml := .Yamls }}
{{- if $yaml.Job  }} 
    "{{ $yaml.Name }}": 
{{ $yaml.Content }},
{{- end }}
{{- end }}
},
  AppYamls: map[string]string {
{{- range $_, $yaml := .Yamls }}
{{- if $yaml.App  }} 
    "{{ $yaml.Name }}": 
{{ $yaml.Content }},
{{- end }}
{{- end }}
},
}
`
)
var templatesDir = "./internal/templates"

func main() {
	yamls := readDir("common")
	yamls = append(yamls, readDir("traefik")...)
	yamls = append(yamls, readDir("istio")...)
	yamls = append(yamls, readDir("job")...)
	yamls = append(yamls, readDir("app")...)

	tmpl, err := template.New("tpl").Parse(yamlsTemplate)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, context{Yamls: yamls})
	if err != nil {
		panic(err)
	}
	rawFile := buf.Bytes()
	//rawFile = bytes.Replace(rawFile, []byte("\\\n"), []byte{}, -1)
	formatedFile, err := format.Source(rawFile)
	if err != nil {
		panic(err)
	}
	out := filepath.Join(templatesDir, "yamls_result.go")
	file, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.Write(formatedFile)
}

func readDir(dir string) []YamlFile {
	targetDir := filepath.Join(templatesDir, dir, "yamls")
	infos, err := ioutil.ReadDir(targetDir)
	if err != nil {
		panic(err)
	}

	var yamls []YamlFile
	for _, info := range infos {
		if !strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".tpl") {
			continue
		}
		filename := filepath.Join(targetDir, info.Name())
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		yamls = append(yamls, YamlFile{
			Name:    info.Name(),
			Traefik: dir == "traefik",
			Istio:   dir == "istio",
			Common:  dir == "common",
			Job:     dir == "job",
			App:     dir == "app",
			Content: fmt.Sprintf("`%s`", string(content)),
		})
	}
	return yamls
}
