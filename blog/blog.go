package blog

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

type IndexInfo struct {
	Blog       *Config
	Posts      []*Post
	Page       int
	TotalPages int
}

type PageInfo struct {
	Blog *Config
}

type Blog struct {
	logger        *log.Logger
	config        Config
	theme         Theme
	dir           string
	templates     map[string]*template.Template
	pageTemplates map[string]*template.Template
}

type tmplRenderer struct {
	log       func(format string, v ...interface{})
	templates map[string]*template.Template
}

func New(config Config, dir string, logger *log.Logger) (*Blog, error) {
	b := Blog{
		logger: logger,
		config: config,
		dir:    dir,
	}

	themeDir := filepath.Join(dir, "theme")
	if err := b.loadTheme(themeDir); err != nil {
		return nil, err
	}

	if err := b.loadTemplates(filepath.Join(themeDir, "templates")); err != nil {
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

	// generate index and pagination
	cmpntRenderer := b.newRenderer(b.templates)
	if b.config.PageSize < 1 {
		return fmt.Errorf("invalid page size: %d", b.config.PageSize)
	}
	pagesDir := filepath.Join(dir, "page")
	if err = mkdir(pagesDir); err != nil {
		return err
	}
	totalPages := (len(posts)-1)/b.config.PageSize + 1
	for i := 0; i < totalPages; i++ {
		info := IndexInfo{
			Blog:       &b.config,
			Posts:      posts[i*b.config.PageSize : min(i*b.config.PageSize+b.config.PageSize, len(posts))],
			Page:       i + 1,
			TotalPages: totalPages,
		}

		if info.Page == 1 {
			err = cmpntRenderer.renderPage(filepath.Join(dir, "index.html"), "index.html", &info)
		} else {
			pageDir := filepath.Join(pagesDir, strconv.Itoa(info.Page))
			if err = mkdir(pageDir); err != nil {
				return err
			}

			err = cmpntRenderer.renderPage(filepath.Join(pageDir, "index.html"), "index.html", &info)
		}

		if err != nil {
			return err
		}
	}

	// generate the post pages
	postDir := filepath.Join(dir, "post")
	if err = mkdir(postDir); err != nil {
		return err
	}
	for _, post := range posts {
		info := PostInfo{Blog: &b.config, Post: post}
		if err = cmpntRenderer.renderPage(filepath.Join(postDir, post.Filename), "post.html", &info); err != nil {
			return err
		}
	}

	// generate the custom pages
	pageRenderer := b.newRenderer(b.pageTemplates)
	for name := range b.pageTemplates {
		info := PageInfo{Blog: &b.config}
		if err = pageRenderer.renderPage(filepath.Join(dir, name), name, &info); err != nil {
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
			if post.Draft {
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
		b.log("rendering %s", filename)

		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}

		if err = b.renderPost(&post, bytes); err != nil {
			return err
		}

		if !post.Draft || !b.config.ExcludeDrafts {
			posts = append(posts, &post)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// sort posts by publish date
	sort.Sort(postSlice(posts))
	return posts, nil
}

func (b *Blog) newRenderer(templates map[string]*template.Template) *tmplRenderer {
	return &tmplRenderer{
		log:       b.log,
		templates: templates,
	}
}

func (r *tmplRenderer) renderTemplate(w io.Writer, name string, data interface{}) error {
	tmpl, exists := r.templates[name]
	if !exists {
		return fmt.Errorf("template %s does not exist", name)
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}

func (r *tmplRenderer) renderPage(filename string, name string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	r.log("rendering %s", filename)
	return r.renderTemplate(file, name, data)
}

func (b *Blog) hasFeature(feature string) bool {
	for _, f := range b.config.Features {
		if f == feature {
			return true
		}
	}

	return false
}

func (b *Blog) log(format string, v ...interface{}) {
	if b.logger != nil {
		b.logger.Printf(format, v...)
	}
}

func (b *Blog) fatal(format string, v ...interface{}) {
	if b.logger != nil {
		b.logger.Fatalf(format, v...)
	}
}
