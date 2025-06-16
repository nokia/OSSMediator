module collector

go 1.24

replace golang.org/x/sys => golang.org/x/sys v0.32.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
