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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"elasticsearchplugin/pkg/config"
	"elasticsearchplugin/pkg/elasticsearch"
	"elasticsearchplugin/pkg/util"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	confFile         string
	logDir           string
	logLevel         int
	enableConsoleLog bool
)

func main() {
	//Read command line options
	parseFlags()

	//initialize logger
	initLogger(logDir, logLevel)

	log.Info("Starting ElasticsearchPlugin...")
	conf, err := config.ReadConfig(confFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Error while reading config")
	}

	//add elasticsearch mapping for core-pm index
	elasticsearch.AddPMMapping(conf.ElasticsearchConf, "core-pm")
	elasticsearch.AddPMMapping(conf.ElasticsearchConf, "ixr-pm")

	//Add watcher to the PM/FM source directory
	err = util.AddWatcher(conf)
	if err != nil {
		log.Fatal(err)
	}

	//watch file creation events in the source directory
	go util.WatchEvents(conf)

	//retry pushing data to elasticsearch that was failed earlier
	elasticsearch.PushFailedData(conf.ElasticsearchConf)

	//remove old data and indices from elasticsearch
	if conf.ElasticsearchConf.DataRetentionDuration > 0 {
		go elasticsearch.CleanUp(conf.ElasticsearchConf)
	}

	//cleanup old response
	if conf.CleanupDuration > 0 {
		for _, sourceDir := range conf.SourceDirs {
			//cleanup of files collected by collector
			files, err := ioutil.ReadDir(sourceDir)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Errorf("Error while reading source directory %s", sourceDir)
			}

			for _, f := range files {
				if f.IsDir() {
					dirPath := filepath.Join(sourceDir, f.Name())
					util.CleanUp(conf.CleanupDuration, dirPath)
				}
			}
		}
	}

	shutdownHook()
}

// Reads command line options
func parseFlags() {
	//read command line arguments
	flag.StringVar(&confFile, "conf_file", "../resources/conf.json", "config file path")
	flag.StringVar(&logDir, "log_dir", "../log", "Log directory")
	flag.IntVar(&logLevel, "log_level", 4, "Log level")
	flag.BoolVar(&enableConsoleLog, "enable_console_log", false, "Enable console logging, if true logs won't be written to file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./elasticsearchplugin [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message and exit.\n")
		fmt.Fprintf(os.Stderr, "\t-conf_file\n\t\tConfig file path (default \"../resources/conf.json\")\n")
		fmt.Fprintf(os.Stderr, "\t-log_dir\n\t\tLog Directory (default \"../log\"), logs will be stored in ElasticsearchPlugin.log file.\n")
		fmt.Fprintf(os.Stderr, "\t-log_level\n\t\tLog Level (default 4). Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)\n")
		fmt.Fprintf(os.Stderr, "\t-enable_console_log\n\t\tEnable console logging, if true logs won't be written to file\n")
	}
	flag.Parse()
}

// create log file (ElasticsearchPlugin.log) within logDir (in case of failure logs will be written to console)
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
		return
	}

	logFile := logDir + "/ElasticsearchPlugin.log"
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
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.Level(logLevel))
}

func shutdownHook() {
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	<-osSignal
	log.Info("Terminating ElasticsearchPlugin...")
	os.Exit(0)
}
