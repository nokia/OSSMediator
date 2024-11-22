module elasticsearchplugin

go 1.23

replace golang.org/x/sys => golang.org/x/sys v0.27.0

require (
	github.com/fsnotify/fsnotify v1.8.0
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require golang.org/x/sys v0.13.0 // indirect
