package blog

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

type IndexInfo struct {
	Blog  *Config
	Posts []*Post
}

type Blog struct {
	config    Config
	theme     *Theme
	dir       string
	templates map[string]*template.Template
}

func New(config Config, dir string) (*Blog, error) {
	themeDir := path.Join(dir, "theme")
	theme, err := loadTheme(themeDir)
	if err != nil {
		return nil, err
	}

	b := Blog{
		config: config,
		theme:  theme,
		dir:    dir,
	}

	b.templates, err = b.loadTemplates(path.Join(themeDir, "templates"))
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (b *Blog) Generate(dir string) error {
	posts, err := b.renderPosts()
	if err != nil {
		return err
	}

	// create the directory if needed
	if err = mkdir(dir); err != nil {
		return err
	}
	// clear the directory
	if err = clearDir(dir); err != nil {
		return err
	}

	// copy extra files/folders over
	for _, path := range b.config.Files {
		dst := filepath.Join(dir, path)
		src := filepath.Join(b.dir, path)

		if err := copyFileOrDir(dst, src); err != nil {
			return err
		}
	}

	// generate theme files
	staticDir := filepath.Join(dir, "static")
	if err = mkdir(staticDir); err != nil {
		return err
	}
	if err = b.theme.Generate(staticDir); err != nil {
		return err
	}

	// generate the index page
	info := IndexInfo{Blog: &b.config}
	for _, post := range posts {
		if post.Exclude || post.Draft {
			continue
		}
		info.Posts = append(info.Posts, post)
	}
	if err = b.generatePage(filepath.Join(dir, "index.html"), "index.html", &info); err != nil {
		return err
	}

	// generate the post pages
	postDir := filepath.Join(dir, "post")
	if err = mkdir(postDir); err != nil {
		return err
	}
	for _, post := range posts {
		if post.Exclude {
			continue
		}

		info := PostInfo{Blog: &b.config, Post: post}
		if err = b.generatePage(filepath.Join(postDir, post.Filename), "post.html", &info); err != nil {
			return err
		}
	}

	// generate rss feed
	if b.hasFeature("rss") {
		feed := &feeds.Feed{
			Title:       b.config.Title,
			Link:        &feeds.Link{Href: b.config.URL},
			Description: b.config.Description,
		}

		for _, post := range posts {
			if post.Draft || post.Exclude {
				continue
			}

			url, err := url.Parse(b.config.URL)
			if err != nil {
				return err
			}
			url.Path = path.Join(url.Path, "post", post.Filename)

			item := feeds.Item{
				Title:       post.Title,
				Link:        &feeds.Link{Href: url.String()},
				Description: string(post.Content),
				Created:     time.Time(post.Date),
			}

			feed.Items = append(feed.Items, &item)
		}

		rss, err := feed.ToRss()
		if err != nil {
			return err
		}

		if err = ioutil.WriteFile(filepath.Join(dir, "feed.xml"), []byte(rss), 0666); err != nil {
			return err
		}
	}

	return nil
}

func (b *Blog) renderPosts() ([]*Post, error) {
	var posts []*Post
	dir := filepath.Join(b.dir, "posts")

	// parse and render blog posts
	err := walkFiles(dir, func(file os.FileInfo) error {
		filename := filepath.Join(dir, file.Name())
		name := strings.TrimSuffix(file.Name(), ".md")
		post := Post{
			Name:     name,
			Filename: name + ".html",
		}

		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		if err = b.renderPost(&post, bytes); err != nil {
			return err
		}

		posts = append(posts, &post)
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

func (b *Blog) hasFeature(feature string) bool {
	for _, f := range b.config.Features {
		if f == feature {
			return true
		}
	}

	return false
}
