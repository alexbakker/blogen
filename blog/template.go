package blog

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (b *Blog) loadTemplatesDir(baseTemplate string, dir string) (map[string]*template.Template, error) {
	funcs := template.FuncMap{
		"hasFeature": b.hasFeature,
		"readFile":   b.readFile,
		"inc": func(i int) int {
			return i + 1
		},
		"dec": func(i int) int {
			return i - 1
		},
	}

	templates := map[string]*template.Template{}
	err := walkFiles(dir, func(file os.FileInfo) error {
		filename := filepath.Join(dir, file.Name())
		b.log("loading %s", filename)

		// parse the child layout
		childTmpl, err := template.New(file.Name()).Funcs(funcs).ParseFiles(filename)
		if err != nil {
			return err
		}

		// and finally also parse the base layout
		tmpl, err := childTmpl.Parse(baseTemplate)
		if err != nil {
			return err
		}

		templates[file.Name()] = tmpl
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

func (b *Blog) loadTemplates(dir string) error {
	baseBytes, err := ioutil.ReadFile(filepath.Join(dir, "base.html"))
	if err != nil {
		return err
	}

	baseTemplate := string(baseBytes)
	templates, err := b.loadTemplatesDir(baseTemplate, filepath.Join(dir, "components"))
	if err != nil {
		return err
	}
	pageTemplates, err := b.loadTemplatesDir(baseTemplate, filepath.Join(dir, "pages"))
	if err != nil {
		return err
	}

	b.templates = templates
	b.pageTemplates = pageTemplates
	return nil
}

func (b *Blog) readFile(filename string) template.HTML {
	filename = filepath.Join(b.dir, "theme", filename)

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		b.fatal("error reading %s: %s", filename, err)
		return ""
	}

	return template.HTML(bytes)
}
