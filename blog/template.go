package blog

import (
	"html/template"
	"io/ioutil"
	"os"
	"path"
)

func loadTemplates(dir string) (map[string]*template.Template, error) {
	baseBytes, err := ioutil.ReadFile(path.Join(dir, "base.html"))
	if err != nil {
		return nil, err
	}

	baseTemplate := string(baseBytes)
	pageDir := path.Join(dir, "pages")
	templates := map[string]*template.Template{}

	err = walkFiles(pageDir, func(file os.FileInfo) error {
		filename := path.Join(pageDir, file.Name())

		// parse the child layout
		childTmpl, err := template.New(file.Name()).ParseFiles(filename)
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
