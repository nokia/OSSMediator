module elasticsearchplugin

go 1.22

replace golang.org/x/sys => golang.org/x/sys v0.25.0

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/sirupsen/logrus v1.9.3
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require golang.org/x/sys v0.4.0 // indirect
