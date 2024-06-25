/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package main

import (
	"flag"
	"fmt"
	"io"
	logger "log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"collector/pkg/config"
	"collector/pkg/ndacapis"
	"collector/pkg/utils"
	"collector/pkg/validator"

	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	confFile         string
	certFile         string
	skipTLS          bool
	logDir           string
	logLevel         int
	enableConsoleLog bool
	version          bool
	appVersion       string
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
	err := config.ReadConfig(confFile)
	if err != nil {
		log.Fatal(err)
	}

	//validate config
	err = validator.ValidateConf(config.Conf)
	if err != nil {
		log.Fatal(err)
	}

	//Create HTTP client for all the GET/POST API calls
	ndacapis.CreateHTTPClient(certFile, skipTLS)

	// Authenticating the users
	for _, user := range config.Conf.Users {
		//if usertype is ABAC, no need to login
		authType := strings.ToUpper(user.AuthType)
		if authType == "PASSWORD" {
			user.Password, err = utils.ReadPassword(user.Email)
			if err != nil {
				fmt.Printf("\nUnable to read password for: %s\nError: %v", user.Email, err)
				log.WithFields(log.Fields{"error": err}).Fatalf("Login Failed for %s", user.Email)
			}
			err = ndacapis.Login(user)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Fatalf("Login Failed for %s", user.Email)
			}
		} else {
			sessionToken, err := utils.ReadSessionToken(user.Email)
			if err != nil {
				fmt.Printf("\nUnable to read sessionToken for: %s\nError: %v", user.Email, err)
				log.WithFields(log.Fields{"error": err}).Fatalf("Token Authorization Failed for %s", user.Email)
			}
			err = ndacapis.TokenAuthorize(user, sessionToken)
			if err != nil {
				fmt.Printf("\nToken Authorization Failed for %s...\nError: %v", user.Email, err)
				log.WithFields(log.Fields{"error": err}).Fatalf("Token Authorization Failed for %s", user.Email)
			}
		}
	}

	checkpointPath := "./checkpoints"
	log.Info("Creating checkpoints directory")
	err = os.MkdirAll(checkpointPath, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to create checkpoints folder")
	}

	//refreshing access token before expiry
	for _, user := range config.Conf.Users {
		go ndacapis.RefreshToken(user)
		//Create the sub response directory for the API under the user's base response directory.
		utils.CreateResponseDirectory(user.ResponseDest, config.Conf.ListNetworkAPI.NhgAPI)
		if config.Conf.ListNetworkAPI.GngAPI != "" {
			utils.CreateResponseDirectory(user.ResponseDest, config.Conf.ListNetworkAPI.GngAPI)
		}
		if strings.ToUpper(user.AuthType) == "ADTOKEN" {
			utils.CreateResponseDirectory(user.ResponseDest, config.Conf.UserAGAPIs.ListOrgUUID)
			utils.CreateResponseDirectory(user.ResponseDest, config.Conf.UserAGAPIs.ListAccUUID)
		}

		for _, api := range config.Conf.MetricAPIs {
			utils.CreateResponseDirectory(user.ResponseDest, api.API)
		}
		for _, api := range config.Conf.SimAPIs {
			utils.CreateResponseDirectory(user.ResponseDest, api.API)
		}
	}

	//start data collection from configured APIs
	ndacapis.StartDataCollection()

	//Perform logout when program terminates using os.Interrupt(ctrl+c)
	shutdownHook()
}

// Reads command line options
func parseFlags() {
	//read command line arguments
	flag.StringVar(&confFile, "conf_file", "../resources/conf.json", "config file path")
	flag.StringVar(&certFile, "cert_file", "", "certificate file path")
	flag.BoolVar(&skipTLS, "skip_tls", false, "skip TLS authentication")
	flag.StringVar(&logDir, "log_dir", "../log", "Log directory")
	flag.IntVar(&logLevel, "log_level", 4, "Log level")
	flag.BoolVar(&enableConsoleLog, "enable_console_log", false, "Enable console logging, if true logs won't be written to file")
	flag.BoolVar(&version, "v", false, "Prints OSSMediator's version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./collector [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message and exit.\n")
		fmt.Fprintf(os.Stderr, "\t-conf_file string\n\t\tConfig file path (default \"../resources/conf.json\")\n")
		fmt.Fprintf(os.Stderr, "\t-cert_file string\n\t\tCertificate file path (if cert_file is not passed then it will establish TLS auth using root certificates.)\n")
		fmt.Fprintf(os.Stderr, "\t-log_dir string\n\t\tLog Directory (default \"../log\"), logs will be stored in collector.log file.\n")
		fmt.Fprintf(os.Stderr, "\t-log_level int\n\t\tLog Level (default 4). Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)\n")
		fmt.Fprintf(os.Stderr, "\t-skip_tls\n\t\tSkip TLS Authentication\n")
		fmt.Fprintf(os.Stderr, "\t-enable_console_long\n\t\tEnable console logging, if true logs won't be written to file\n")
		fmt.Fprintf(os.Stderr, "\t-v\n\t\tPrints OSSMediator's version\n")
	}
	flag.Parse()
}

// create log file (collector.log) within logDir (in case of failure logs will be written to console)
// if console logs is enabled then logs are written to stdout instead of file.
func initLogger(logDir string, logLevel int) {
	if enableConsoleLog {
		log.SetOutput(os.Stdout)
		log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.Level(logLevel))
		return
	}
	var err error
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warningf("Unable to create log directory %s", logDir)
		log.Info("Failed to log to file, using default stderr")
		log.SetOutput(os.Stdout)
		log.SetFormatter(&log.TextFormatter{})
		log.SetLevel(log.Level(logLevel))
		return
	}

	logFile := logDir + "/collector.log"
	_, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		lumberjackLogrotate := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100, // Max megabytes before log is rotated
			MaxBackups: 10,  // Max number of old log files to keep
			MaxAge:     20,  // Max number of days to retain log files
			Compress:   true,
		}
		log.SetOutput(lumberjackLogrotate)
	} else {
		log.Info("Failed to log to file, using default stderr")
		log.SetOutput(os.Stdout)
	}
	logger.SetOutput(io.Discard)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.Level(logLevel))
}

func shutdownHook() {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	<-osSignal
	log.Info("Received shutdown signal...Logging out...")
	for _, user := range config.Conf.Users {
		err := ndacapis.Logout(user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Logout failed for %s", user.Email)
		}
	}
	log.Info("Terminating DA OSS Collector...")
	os.Exit(0)
}
