package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"syscall"
)

type Config struct {
	Users []struct {
		EmailID  string `json:"email_id"`
		AuthType string `json:"auth_type"`
	} `json:"users"`
}

var (
	confFile   string
	version    bool
	appVersion string
	secretDir  = ".secret"
)

func main() {
	//read command line arguments
	flag.StringVar(&confFile, "c", "../resources/conf.json", "config file path")
	flag.BoolVar(&version, "v", false, "Prints OSSMediator's version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./storesecret [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message.\n")
		fmt.Fprintf(os.Stderr, "\t-c string\n\t\tConfig file path (default \"../resources/conf.json\")\n")
		fmt.Fprintf(os.Stderr, "\t-v\n\t\tPrints OSSMediator's version\n")
	}
	flag.Parse()
	if version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	conf, err := readConfig(confFile)
	if err != nil {
		log.Fatal("Unable to read config file, error: ", err)
	}

	err = os.MkdirAll(secretDir, 0600)
	if err != nil {
		log.Fatalf("Error while creating %s", secretDir)
	}
	readPassword(conf)
}

func readConfig(confFile string) (*Config, error) {
	contents, err := os.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = json.Unmarshal(contents, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func readPassword(conf *Config) {
	for _, user := range conf.Users {
		authType := strings.ToUpper(user.AuthType)
		if authType == "PASSWORD" || authType == "" {
			fmt.Printf("Enter password for %s: ", user.EmailID)
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatalf("Error in reading password for %v: %v", user.EmailID, err)
			}
			storePassword(user.EmailID, bytePassword)
		} else if authType == "ADTOKEN" {
			fmt.Printf("Enter access token for %s: ", user.EmailID)
			byteAccessToken, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatalf("Error in reading password for %v: %v", user.EmailID, err)
			}
			fmt.Printf("\nEnter refresh token for %s: ", user.EmailID)
			byteRefreshToken, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatalf("Error in reading token for %v: %v", user.EmailID, err)
			}
			storeToken(user.EmailID, byteAccessToken, byteRefreshToken)
		}
	}
}

func storePassword(user string, password []byte) {
	fileName := secretDir + "/." + user
	encodedPassword := base64.StdEncoding.EncodeToString(password)
	err := os.WriteFile(fileName, []byte(encodedPassword), 0600)
	if err != nil {
		log.Fatalf("Unable to store password for %v to %v, error: %v", user, fileName, err)
	}
	fmt.Printf("\nPassword stored for %v\n", user)
}

func storeToken(user string, accessToken []byte, refreshToken []byte) {
	fileName := secretDir + "/." + user
	encodedPassword := base64.StdEncoding.EncodeToString(accessToken) + "\n" + base64.StdEncoding.EncodeToString(refreshToken)
	err := os.WriteFile(fileName, []byte(encodedPassword), 0600)
	if err != nil {
		log.Fatalf("Unable to store password for %v to %v, error: %v", user, fileName, err)
	}
	fmt.Printf("\nToken stored for %v\n", user)
}
