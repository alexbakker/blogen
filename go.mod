module github.com/alexbakker/blogen

go 1.17

require (
	github.com/alecthomas/chroma/v2 v2.0.1
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gorilla/feeds v1.1.1
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/spf13/cobra v1.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/alecthomas/chroma/v2 => github.com/alexbakker/chroma/v2 v2.0.2-0.20220529203822-b157ed683eb7
