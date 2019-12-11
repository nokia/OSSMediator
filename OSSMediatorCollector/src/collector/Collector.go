/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	logger "log"
	"os"
	"os/signal"
	"syscall"

	"collector/util"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	confFile   string
	certFile   string
	skipTLS    bool
	logDir     string
	logLevel   int
	version    bool
	appVersion string
)

func main() {
	//Read command line options
	parseFlags()
	if version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	//initialize logger
	initLogger(logDir, logLevel)

	log.Info("Starting DA OSS Collector...")
	//Reading config from json file
	err := util.ReadConfig(confFile)
	if err != nil {
		log.Fatal(err)
	}

	//Create HTTP client for all the GET/POST API calls
	util.CreateHTTPClient(certFile, skipTLS)

	// Authenticating the users
	for _, user := range util.Conf.Users {
		err = util.Login(user)
		if err != nil {
			fmt.Printf("\nLogin Failed for %s...\nError: %v", user.Email, err)
			log.WithFields(log.Fields{"error": err}).Fatalf("Login Failed for %s", user.Email)
		}
	}

	//refreshing access token before expiry
	for _, user := range util.Conf.Users {
		go util.RefreshToken(user)
		//Create the sub response directory for the API under the user's base response directory.
		for _, api := range util.Conf.APIs {
			util.CreateResponseDirectory(user.ResponseDest, api.API)
		}
	}

	//start data collection from PM/FM APIs
	util.StartDataCollection()

	//Perform logout when program terminates using os.Interrupt(ctrl+c)
	shutdownHook()
}

//Reads command line options
func parseFlags() {
	//read command line arguments
	flag.StringVar(&confFile, "conf_file", "../resources/conf.json", "config file path")
	flag.StringVar(&certFile, "cert_file", "", "certificate file path")
	flag.BoolVar(&skipTLS, "skip_tls", false, "skip TLS authentication")
	flag.StringVar(&logDir, "log_dir", "../log", "Log directory")
	flag.IntVar(&logLevel, "log_level", 4, "Log level")
	flag.BoolVar(&version, "v", false, "Prints OSSMediator's version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./collector [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message and exit.\n")
		fmt.Fprintf(os.Stderr, "\t-conf_file string\n\t\tConfig file path (default \"../resources/conf.json\")\n")
		fmt.Fprintf(os.Stderr, "\t-cert_file string\n\t\tCertificate file path (if cert_file is not passed then it will establish TLS auth using root certificates.)\n")
		fmt.Fprintf(os.Stderr, "\t-log_dir string\n\t\tLog Directory (default \"../log\"), logs will be stored in collector.log file.\n")
		fmt.Fprintf(os.Stderr, "\t-log_level int\n\t\tLog Level (default 4), logger level in collector.log file. Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)\n")
		fmt.Fprintf(os.Stderr, "\t-skip_tls\n\t\tSkip TLS Authentication\n")
		fmt.Fprintf(os.Stderr, "\t-v\n\t\tPrints OSSMediator's version\n")
	}
	flag.Parse()
}

//create log file (collector.log) within logDir (in case of failure logs will be written to console)
func initLogger(logDir string, logLevel int) {
	var err error
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warningf("Unable to create log directory %s", logDir)
		log.Info("Failed to log to file, using default stderr")
		log.SetOutput(os.Stdout)
		log.SetFormatter(&log.TextFormatter{})
		return
	}

	logFile := logDir + "/collector.log"
	_, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		lumberjackLogrotate := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    2,  // Max megabytes before log is rotated
			MaxBackups: 10, // Max number of old log files to keep
			MaxAge:     20, // Max number of days to retain log files
			Compress:   true,
		}
		log.SetOutput(lumberjackLogrotate)
	} else {
		log.Info("Failed to log to file, using default stderr")
		log.SetOutput(os.Stdout)
	}
	logger.SetOutput(ioutil.Discard)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.Level(logLevel))
}

func shutdownHook() {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	<-osSignal
	log.Info("Received shutdown signal...Logging out...")
	for _, user := range util.Conf.Users {
		err := util.Logout(user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Logout failed for %s", user.Email)
		}
	}
	log.Info("Terminating DA OSS Collector...")
	os.Exit(0)
}
