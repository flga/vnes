package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/bmatcuk/doublestar"
	"github.com/flga/nes/cmd/internal/asset"
)

type tplData struct {
	Pkg    string
	Assets []struct {
		Path []string
		Data string
	}
}

var tpl = template.Must(template.New("").Parse(`// Code generated automatically DO NOT EDIT.

package {{ .Pkg }}

import "github.com/flga/nes/cmd/internal/asset"

var assets = asset.List{
	{{ range .Assets -}}
	asset.New({{- range .Path -}}{{ . | printf "%q"}},{{- end -}}{{- .Data | printf "%q" -}}),
	{{ end }}
}`))

func main() {
	rootPath := flag.String("root", "", "Paths will be rooted here. The resulting path will not be relative. Defaults to the current working directory.")
	outputFile := flag.String("o", "", "Output file.")
	exclude := flag.String("exclude", "", "Comma separated list of glob expressions. Any files that match will be excluded.")
	flag.Parse()

	if *outputFile == "" {
		fmt.Fprintln(os.Stderr, "output file is required")
		os.Exit(1)
	}

	if err := run(*rootPath, *outputFile, flag.Args(), strings.Split(*exclude, ",")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(rootPath, out string, includeGlobs []string, excludeGlobs []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	includeSet := make(map[string]struct{})
	for _, p := range includeGlobs {
		if err := glob(p, includeSet); err != nil {
			return err
		}
	}
	excludeSet := make(map[string]struct{})
	for _, p := range excludeGlobs {
		if err := glob(p, excludeSet); err != nil {
			return err
		}
	}
	for ep := range excludeSet {
		delete(includeSet, ep)
	}

	data, err := parse(includeSet, wd, filepath.Join(wd, rootPath))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, data); err != nil {
		return err
	}

	code, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(wd, out), code, 0666)
}

func glob(pathname string, set map[string]struct{}) error {
	matches, err := doublestar.Glob(pathname)
	if err != nil {
		return err
	}

	for _, m := range matches {
		stat, err := os.Stat(m)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			continue
		}

		set[m] = struct{}{}
	}

	return nil
}

func parse(pathSet map[string]struct{}, wd, root string) (*tplData, error) {
	data := &tplData{
		Pkg: os.Getenv("GOPACKAGE"),
	}

	var paths []string
	for fp := range pathSet {
		paths = append(paths, fp)
	}
	sort.Strings(paths)

	for _, fp := range paths {
		f, err := os.Open(fp)
		if err != nil {
			return nil, err
		}

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		if err := f.Close(); err != nil {
			return nil, err
		}

		encoded, err := asset.Encode(content)
		if err != nil {
			return nil, err
		}

		path := strings.TrimLeft(strings.TrimPrefix(filepath.Join(wd, fp), root), "/")

		data.Assets = append(data.Assets, struct {
			Path []string
			Data string
		}{
			Path: strings.Split(path, string(filepath.Separator)),
			Data: encoded,
		})
	}

	return data, nil
}
