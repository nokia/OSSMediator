/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package OssCollector

import (
	"collector/pkg/config"
	"collector/pkg/ndacapis"
	"collector/pkg/utils"
	"collector/pkg/validator"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	//confFile     string
	certFile     string
	skipTLS      bool
	logDir       string
	logLevel     int
	runningMutex sync.Mutex
	isRunning    bool
	stopCh       chan struct{}
)

func StartCollector(confFile string, running bool, stopCh chan struct{}, goroutine *sync.WaitGroup) {
	/***
	runningMutex.Lock()
	defer runningMutex.Unlock()
	if isRunning {
		return
	}

	isRunning = true

	stopCh = make(chan struct{})
	go func() {
		for {
			select {
			case <-stopCh:
				log.Info("****Collector stopped****")
				return
			default:
				collectorRun(confFile)
				time.Sleep(2 * time.Second)
			}
		}
	}()
	***/

	fmt.Println("Starting Collector from startCollector func")
	running = true
	internalStopCh := make(chan struct{})
	var goroutine1 sync.WaitGroup
	goroutine1.Add(1)
	go collectorRun(confFile, running, internalStopCh, &goroutine1)
	for {
		select {
		case <-stopCh:
			fmt.Println("Stopping startCollector...")
			close(internalStopCh)
			goroutine.Done()
			goroutine1.Wait()
			return
		default:
			fmt.Println("Inside startCollector()...")
			time.Sleep(time.Second * 2)
		}
	}
}

func collectorRun(confFile string, running bool, stopCh chan struct{}, goroutine *sync.WaitGroup) {
	skipTLS = true

	/**
	log.Info("Starting DA OSS Collector...")

	fmt.Println("config file is :", confFile)
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
		fmt.Println("Inside refreshing token called from osscollector")
		go ndacapis.RefreshToken(user)
		//Create the sub response directory for the API under the user's base response directory.
		utils.CreateResponseDirectory(user.ResponseDest, config.Conf.ListNhGAPI.API)
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

	**/
	fmt.Println("Starting collectorRun...")
	running = true
	internalStopCh := make(chan struct{})
	var goroutine1 sync.WaitGroup

	//call other logic

	log.Info("Starting DA OSS Collector...")

	fmt.Println("config file is :", confFile)
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
		fmt.Printf("\nInside refreshing token called from osscollector\n")
		//Create the sub response directory for the API under the user's base response directory.
		utils.CreateResponseDirectory(user.ResponseDest, config.Conf.ListNhGAPI.API)
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
		goroutine1.Add(1)
		go ndacapis.RefreshToken(user, running, internalStopCh, &goroutine1)
	}
	//shutdownHook()
	goroutine1.Add(1)
	//start data collection from configured APIs
	go ndacapis.StartDataCollection(running, internalStopCh, &goroutine1)

	for {
		select {
		case <-stopCh:
			fmt.Println("Stopping CollectorRun...")
			//Perform logout when program terminates using os.Interrupt(ctrl+c)
			//shutdownHook()
			var checkrun = StopCollector()
			fmt.Println("running? ", checkrun)
			close(internalStopCh)
			goroutine.Done()
			goroutine1.Wait()
			return
		default:
			fmt.Println("CollectorRun is running...")
			time.Sleep(time.Second * 2)
		}
	}
}

func StopCollector() bool {
	//runningMutex.Lock()
	//defer runningMutex.Unlock()

	//if !isRunning {
	//	return isRunning
	//}

	fmt.Println("******Stopping OSSCollector*****")
	for _, user := range config.Conf.Users {
		err := ndacapis.Logout(user)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Errorf("Logout failed for %s", user.Email)
		}
		//close(stopCh)
	}
	//close(stopCh)
	isRunning = false
	log.Info("OSSCollector stopped successfully!!")
	return isRunning
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
