package blog

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (b *Blog) loadTemplates(dir string) (map[string]*template.Template, error) {
	baseBytes, err := ioutil.ReadFile(filepath.Join(dir, "base.html"))
	if err != nil {
		return nil, err
	}

	baseTemplate := string(baseBytes)
	pageDir := filepath.Join(dir, "pages")
	templates := map[string]*template.Template{}
	funcs := template.FuncMap{
		"hasFeature": b.hasFeature,
	}

	err = walkFiles(pageDir, func(file os.FileInfo) error {
		filename := filepath.Join(pageDir, file.Name())

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
