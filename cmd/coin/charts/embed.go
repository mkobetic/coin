// +build ignore

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"text/template"

	"github.com/mkobetic/coin/check"
)

var output = template.Must(template.New("").Parse(`package main

// Code generated from files in charts subdirectory. DO NOT EDIT.
//
//go:generate go run charts/embed.go

var charts = map[string][]byte{}

func init() {
	for file, encoded := range map[string]string{
	{{- range $file, $contents := . }}
    	"{{ $file }}": "{{ $contents }}",
	{{- end }}
	} {
		charts[file] = decode(file, encoded)
	}
}
`))

func main() {
	_, thisFile, _, _ := runtime.Caller(0)

	dir := path.Dir(thisFile)
	fis, err := ioutil.ReadDir(dir)
	check.NoError(err, "reading charts directory")
	files := map[string]string{}
	for _, fi := range fis {
		if ext := path.Ext(fi.Name()); ext != ".js" && ext != ".html" {
			continue
		}
		contents, err := ioutil.ReadFile(path.Join(dir, fi.Name()))
		check.NoError(err, "reading charts/%s\n", fi.Name())
		files[fi.Name()] = encode(fi.Name(), contents)
	}
	w, err := os.Create(path.Join(path.Dir(dir), "charts.go"))
	check.NoError(err, "opening charts.go\n")
	err = output.Execute(w, files)
	check.NoError(err, "writing charts.go\n")
}

func encode(name string, decoded []byte) (encoded string) {
	var buf bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &buf)
	gz := gzip.NewWriter(b64)
	_, err := gz.Write(decoded)
	check.NoError(err, "writing charts/%s", name)
	err = gz.Close()
	check.NoError(err, "closing gzip charts/%s", name)
	err = b64.Close()
	check.NoError(err, "closing b64 charts/%s", name)
	return buf.String()
}
