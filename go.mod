module github.com/jcwillox/dotbot

go 1.17

require (
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/creasty/defaults v1.5.2
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/jcwillox/emerald v0.3.1
	github.com/k0kubun/pp/v3 v3.0.7
	github.com/spf13/cobra v1.2.1
	golang.org/x/sys v0.0.0-20211113001501-0c823b97ae02
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/k0kubun/pp/v3 v3.0.7 => github.com/k0kubun/pp/v3 v3.0.8-0.20210415165650-b87d88f85b84
