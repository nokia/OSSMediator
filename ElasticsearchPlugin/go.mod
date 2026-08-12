module elasticsearchplugin

go 1.26.5

replace golang.org/x/sys => golang.org/x/sys v0.47.0

require (
	github.com/fsnotify/fsnotify v1.10.1
	github.com/sirupsen/logrus v1.9.4
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require golang.org/x/sys v0.47.0 // indirect
