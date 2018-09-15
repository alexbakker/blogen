package blog

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	yaml "gopkg.in/yaml.v2"
)

const (
	themeFilename = "theme.yml"
)

type Style struct {
	Syntax string   `yaml:"syntax"`
	Output string   `yaml:"output"`
	Files  []string `yaml:"files"`
}

type Theme struct {
	Name   string   `yaml:"name"`
	Static []string `yaml:"static"`
	Style  Style    `yaml:"style"`
	dir    string
}

func loadTheme(dir string) (*Theme, error) {
	bytes, err := ioutil.ReadFile(filepath.Join(dir, themeFilename))
	if err != nil {
		return nil, err
	}

	var theme Theme
	if err = yaml.Unmarshal(bytes, &theme); err != nil {
		return nil, err
	}
	theme.dir = dir

	return &theme, nil
}

func (t *Theme) Generate(dir string) error {
	// copy the static files over
	for _, path := range t.Static {
		dst := filepath.Join(dir, path)
		src := filepath.Join(t.dir, "static", path)

		if err := copyFileOrDir(dst, src); err != nil {
			return err
		}
	}

	// generate the combined css file
	dst := filepath.Join(dir, t.Style.Output)
	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	files := make([]string, 0, len(t.Style.Files))
	for _, file := range t.Style.Files {
		files = append(files, filepath.Join(t.dir, "static", file))
	}
	if err := mergeFiles(dstFile, files); err != nil {
		return err
	}

	// generate syntax highlighting css file and merge it into the style file
	style := styles.Get(t.Style.Syntax)
	if style == nil {
		return errors.New("style not found")
	}
	if err = html.New(html.WithClasses()).WriteCSS(dstFile, style); err != nil {
		return err
	}

	return nil
}
