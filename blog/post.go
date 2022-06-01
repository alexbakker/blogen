package blog

import (
	"html/template"
	"time"
)

const (
	PostDateFormat = "2006-01-02"
)

type Post struct {
	Name        string
	Filename    string
	Title       string   `yaml:"title"`
	Date        PostDate `yaml:"date"`
	Draft       bool     `yaml:"draft"`
	TOC         template.HTML
	Content     template.HTML
	Summary     template.HTML
	SummaryText string
}

type PostDate time.Time

type PostInfo struct {
	*PageInfo
	Post *Post
}

type postSlice []*Post

func (s postSlice) Len() int {
	return len(s)
}

func (s postSlice) Less(i, j int) bool {
	return time.Time(s[i].Date).After(time.Time(s[j].Date))
}

func (s postSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (d PostDate) MarshalText() ([]byte, error) {
	return []byte(d.RFC3339()), nil
}

func (d *PostDate) UnmarshalText(data []byte) error {
	parsedDate, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return err
	}

	*d = PostDate(parsedDate)
	return nil
}

func (d PostDate) Format(layout string) string {
	return time.Time(d).Format(layout)
}

func (d PostDate) Year() int {
	return time.Time(d).Year()
}

func (d PostDate) RFC3339() string {
	return time.Time(d).Format(time.RFC3339)
}
