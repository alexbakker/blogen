module github.com/alexbakker/blogen

go 1.23

require (
	github.com/alecthomas/chroma/v2 v2.14.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/gorilla/feeds v1.2.0
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/spf13/cobra v1.8.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.27.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/alecthomas/chroma/v2 => github.com/alexbakker/chroma/v2 v2.0.2-0.20221112115940-ab29907878eb
