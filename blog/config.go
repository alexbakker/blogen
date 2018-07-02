package blog

type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Config struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Author      Author `yaml:"author"`
	URL         string `yaml:"url"`
	EnableRSS   bool   `yaml:"enable_rss"`
	CodeStyle   string `yaml:"code_style"`
}
