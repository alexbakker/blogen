package blog

import (
	"html/template"
	"time"
)

const (
	PostDateFormat = "2006-01-02"
)

type PostDate time.Time

type PostInfo struct {
	Name     string
	Filename string
	Title    string   `yaml:"title"`
	Date     PostDate `yaml:"date"`
	Draft    bool     `yaml:"draft"`
	Exclude  bool     `yaml:"exclude"`
	Summary  template.HTML
}

type Post struct {
	Info    *PostInfo
	Content template.HTML
}

type postSlice []*Post

func (s postSlice) Len() int {
	return len(s)
}

func (s postSlice) Less(i, j int) bool {
	return time.Time(s[i].Info.Date).After(time.Time(s[j].Info.Date))
}

func (s postSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (d PostDate) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *PostDate) UnmarshalText(data []byte) error {
	parsedDate, err := time.Parse(PostDateFormat, string(data))
	if err != nil {
		return err
	}

	*d = PostDate(parsedDate)
	return nil
}

func (d PostDate) String() string {
	return time.Time(d).Format(PostDateFormat)
}
