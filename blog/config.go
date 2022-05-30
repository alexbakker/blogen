package blog

type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
	About string `yaml:"about"`
}

type Social struct {
	URL  string `yaml:"url"`
	Text string `yaml:"text"`
	Icon string `yaml:"icon"`
}

type Link struct {
	URL         string `yaml:"url"`
	Text        string `yaml:"text"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
}

type License struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Year int    `yaml:"year"`
}

type Config struct {
	ExcludeDrafts bool
	VersionInfo   string
	Title         string   `yaml:"title"`
	Description   string   `yaml:"description"`
	URL           string   `yaml:"url"`
	PageSize      int      `yaml:"page_size"`
	Features      []string `yaml:"features"`
	Files         []string `yaml:"files"`
	Author        Author   `yaml:"author"`
	License       License  `yaml:"license"`
}
