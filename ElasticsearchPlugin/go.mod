module elasticsearchplugin

go 1.26.2

replace golang.org/x/sys => golang.org/x/sys v0.43.0

require (
	github.com/fsnotify/fsnotify v1.9.0
	github.com/sirupsen/logrus v1.9.4
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require golang.org/x/sys v0.13.0 // indirect
