package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/term"
	"io/ioutil"
	"log"
	"os"
	"syscall"
)

type Config struct {
	Users []struct {
		EmailID string `json:"email_id"`
	} `json:"users"`
}

var (
	confFile  string
	secretDir = ".secret"
)

func main() {
	//read command line arguments
	flag.StringVar(&confFile, "c", "../resources/conf.json", "config file path")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./storesecret [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message.\n")
		fmt.Fprintf(os.Stderr, "\t-c string\n\t\tConfig file path (default \"../resources/conf.json\")\n")
	}
	flag.Parse()

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
	contents, err := ioutil.ReadFile(confFile)
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
		fmt.Printf("Enter password for %s: ", user.EmailID)
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Error in reading password for %v: %v", user.EmailID, err)
		}
		storePassword(user.EmailID, bytePassword)
	}
}

func storePassword(user string, password []byte) {
	fileName := secretDir + "/." + user
	encodedPassword := base64.StdEncoding.EncodeToString(password)
	err := ioutil.WriteFile(fileName, []byte(encodedPassword), 0600)
	if err != nil {
		log.Fatalf("Unable to store password for %v to %v, error: %v", user, fileName, err)
	}
	fmt.Printf("\nPassword stored for %v\n", user)
}
