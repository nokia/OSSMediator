module collector

go 1.25.6

replace golang.org/x/sys => golang.org/x/sys v0.40.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
