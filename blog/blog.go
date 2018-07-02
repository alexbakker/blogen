package blog

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"github.com/gorilla/feeds"
)

type BlogInfo struct {
	Author Author
	Posts  []*PostInfo
}

type Blog struct {
	config    Config
	dir       string
	templates map[string]*template.Template
}

func New(config Config, dir string) (*Blog, error) {
	templates, err := loadTemplates(path.Join(dir, "templates"))
	if err != nil {
		return nil, err
	}

	return &Blog{
		config:    config,
		dir:       dir,
		templates: templates,
	}, nil
}

func (b *Blog) Generate(dir string) error {
	posts, err := b.renderPosts()
	if err != nil {
		return err
	}

	// remove the directory and recreate it
	dir = path.Join(b.dir, dir)
	if err = os.RemoveAll(dir); err != nil {
		return err
	}
	if err = mkdir(dir); err != nil {
		return err
	}

	// copy the static files over
	staticDir := path.Join(dir, "static")
	if err = mkdir(staticDir); err != nil {
		return err
	}
	if err = copyDir(staticDir, path.Join(b.dir, "static")); err != nil {
		return err
	}

	// generate syntax highlighting css file
	styleFile, err := os.Create(path.Join(staticDir, "css", "code.css"))
	if err != nil {
		return err
	}
	defer styleFile.Close()

	style := styles.Get(b.config.CodeStyle)
	if style == nil {
		return errors.New("style not found")
	}
	if err = html.New(html.WithClasses()).WriteCSS(styleFile, style); err != nil {
		return err
	}

	// generate the index page
	info := BlogInfo{
		Author: b.config.Author,
	}
	for _, post := range posts {
		if post.Info.Exclude || post.Info.Draft {
			continue
		}
		info.Posts = append(info.Posts, post.Info)
	}
	if err = b.generatePage(path.Join(dir, "index.html"), "index.html", &info); err != nil {
		return err
	}

	// generate the post pages
	postDir := path.Join(dir, "post")
	if err = mkdir(postDir); err != nil {
		return err
	}
	for _, post := range posts {
		if post.Info.Exclude {
			continue
		}
		if err = b.generatePage(path.Join(postDir, post.Info.Filename), "post.html", post); err != nil {
			return err
		}
	}

	// generate rss feed
	if b.config.EnableRSS {
		feed := &feeds.Feed{
			Title:       b.config.Title,
			Link:        &feeds.Link{Href: b.config.URL},
			Description: b.config.Description,
		}

		for _, post := range posts {
			if post.Info.Draft || post.Info.Exclude {
				continue
			}

			url, err := url.Parse(b.config.URL)
			if err != nil {
				return err
			}
			url.Path = path.Join(url.Path, "post", post.Info.Filename)

			item := feeds.Item{
				Title:       post.Info.Title,
				Link:        &feeds.Link{Href: url.String()},
				Description: string(post.Content),
				Created:     time.Time(post.Info.Date),
			}

			feed.Items = append(feed.Items, &item)
		}

		rss, err := feed.ToRss()
		if err != nil {
			return err
		}

		if err = ioutil.WriteFile(path.Join(dir, "feed.xml"), []byte(rss), 0666); err != nil {
			return err
		}
	}

	return nil
}

func (b *Blog) renderPosts() ([]*Post, error) {
	var posts []*Post
	dir := path.Join(b.dir, "posts")

	// parse and render blog posts
	err := walkFiles(dir, func(file os.FileInfo) error {
		filename := path.Join(dir, file.Name())
		name := strings.TrimSuffix(file.Name(), ".md")
		info := PostInfo{
			Name:     name,
			Filename: name + ".html",
		}

		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		post, err := b.renderPost(&info, bytes)
		if err != nil {
			return err
		}

		posts = append(posts, post)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// sort posts by publish date
	sort.Sort(postSlice(posts))
	return posts, nil
}

func (b *Blog) renderTemplate(w io.Writer, name string, data interface{}) error {
	tmpl, exists := b.templates[name]
	if !exists {
		return fmt.Errorf("template %s does not exist", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

func (b *Blog) generatePage(filename string, name string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return b.renderTemplate(file, name, data)
}
