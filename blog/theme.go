package blog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	yaml "gopkg.in/yaml.v3"
)

const (
	themeFilename = "theme.yml"
)

type Style struct {
	Syntax string `yaml:"syntax"`
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

type Theme struct {
	Name   string   `yaml:"name"`
	Static []string `yaml:"static"`
	Style  Style    `yaml:"style"`
	dir    string
}

func (b *Blog) loadTheme(dir string) error {
	filename := filepath.Join(dir, themeFilename)
	b.log("loading %s\n", filename)

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(bytes, &b.theme); err != nil {
		return err
	}
	b.theme.dir = dir

	return nil
}

func (t *Theme) execSass(input string, w io.Writer) error {
	inputFile, err := os.Open(input)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// generate code syntax highlighting css file
	buf := new(bytes.Buffer)
	style := styles.Get(t.Style.Syntax)
	if style == nil {
		return fmt.Errorf("style %s not found", t.Style.Syntax)
	}
	if err := html.New(html.WithClasses(true)).WriteCSS(buf, style); err != nil {
		return err
	}

	args := []string{"--stdin", "--load-path", filepath.Dir(input), "--style", "compressed"}
	cmd := exec.Command("sassc", args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	// merge the code style into the main style file
	cmd.Stdin = io.MultiReader(inputFile, buf)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("executing sass: %s", err)
	}

	return nil
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

	// generate the combined css file with sass
	dst := filepath.Join(dir, t.Style.Output)
	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if err := t.execSass(filepath.Join(t.dir, t.Style.Input), dstFile); err != nil {
		return err
	}

	return nil
}
