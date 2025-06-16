/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package ndacapis

import (
	"collector/pkg/config"
	"collector/pkg/utils"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	//Timeout duration for HTTP calls
	timeout = 120 * time.Second

	//Maximum no. of retry attempt for API call
	maxRetryAttempts = 3

	//Time interval at which the 1st API call should start
	interval = 15

	//Query Params for GET APIS
	startTimeQueryParam      = "start_timestamp"
	endTimeQueryParam        = "end_timestamp"
	limitQueryParam          = "limit"
	indexQueryParam          = "index"
	alarmTypeQueryParam      = "alarm_type"
	metricTypeQueryParam     = "metric_type"
	searchAfterKeyQueryParam = "search_after_key"
	aggregationQueryParam    = "aggr"
	orgIDQueryParam          = "user_info.org_uuid"
	accIDQueryParam          = "user_info.account_uuid"

	//Headers
	authorizationHeader = "Authorization"

	//Success status code from response
	successStatusCode = "SUCCESS"

	//response type
	fmResponseType = "fmdata"
)

var (
	//HTTP client for all API calls
	client *http.Client

	//used for logging
	txnID uint64 = 1000

	activeAPIs = map[string]struct{}{}
	mux        = sync.RWMutex{}
)

type fn func(*config.APIConf, *config.User, uint64, bool)

// StartDataCollection starts the tickers for PM/FM APIs.
func StartDataCollection() {
	for _, user := range config.Conf.Users {
		if config.Conf.ListNetworkAPI != nil {
			getNhgDetails(&config.APIConf{API: config.Conf.ListNetworkAPI.NhgAPI, Interval: config.Conf.ListNetworkAPI.Interval}, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			if config.Conf.ListNetworkAPI.GngAPI != "" {
				getGngDetails(&config.APIConf{API: config.Conf.ListNetworkAPI.GngAPI, Interval: config.Conf.ListNetworkAPI.Interval}, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			}
			ticker := time.NewTicker(time.Duration(config.Conf.ListNetworkAPI.Interval) * time.Minute)
			go triggerNetworkAPI(ticker, config.Conf.ListNetworkAPI, user, config.Conf.PrettyResponse)
		}
	}

	currentTime := utils.CurrentTime()
	diff := currentTime.Minute() - (currentTime.Minute() / interval * interval) - config.Conf.Delay
	begTime := currentTime.Add(time.Duration(-1*diff) * time.Minute)
	if currentTime.After(begTime) {
		begTime = begTime.Add(time.Duration(interval) * time.Minute)
		for _, user := range config.Conf.Users {
			for _, api := range config.Conf.MetricAPIs {
				go fetchMetricsData(api, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			}
			for _, api := range config.Conf.SimAPIs {
				go fetchSimData(api, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			}
		}
	}

	timer := time.NewTimer(time.Until(begTime))
	<-timer.C
	//For each API creates ticker to trigger the API periodically at specified interval.
	for _, user := range config.Conf.Users {
		for _, api := range config.Conf.MetricAPIs {
			go fetchMetricsData(api, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			ticker := time.NewTicker(time.Duration(api.Interval) * time.Minute)
			go trigger(ticker, api, user, config.Conf.PrettyResponse, fetchMetricsData)
		}
		for _, api := range config.Conf.SimAPIs {
			go fetchSimData(api, user, atomic.AddUint64(&txnID, 1), config.Conf.PrettyResponse)
			ticker := time.NewTicker(time.Duration(api.Interval) * time.Minute)
			go trigger(ticker, api, user, config.Conf.PrettyResponse, fetchSimData)
		}
	}
}

// triggers the network apis periodically at specified interval.
func triggerNetworkAPI(ticker *time.Ticker, api *config.ListNetworkAPIConf, user *config.User, prettyResponse bool) {
	for {
		<-ticker.C
		getNhgDetails(&config.APIConf{API: api.NhgAPI, Interval: api.Interval}, user, atomic.AddUint64(&txnID, 1), prettyResponse)
		if config.Conf.ListNetworkAPI.GngAPI != "" {
			getGngDetails(&config.APIConf{API: api.GngAPI, Interval: api.Interval}, user, atomic.AddUint64(&txnID, 1), prettyResponse)
		}
	}
}

// triggers the method periodically at specified interval.
func trigger(ticker *time.Ticker, api *config.APIConf, user *config.User, prettyResponse bool, method fn) {
	for {
		<-ticker.C
		method(api, user, atomic.AddUint64(&txnID, 1), prettyResponse)
	}
}

// CreateHTTPClient creates HTTP client for all the GET/POST API calls, if certFile is empty and skipTLS is false TLS authentication will be done using root certificates.
// certFile keeps the server certificate file path
// skipTLS if true all API calls will skip TLS auth.
func CreateHTTPClient(certFile string, skipTLS bool) {
	if skipTLS {
		//skipping certificates
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr, Timeout: timeout}
		log.Debugf("Skipping TLS authentication")
	} else if certFile == "" {
		client = &http.Client{Timeout: timeout}
		log.Debugf("TLS authentication using root certificates")
	} else {
		//Load CA cert
		caCert, err := os.ReadFile(certFile)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error while reading server certificate file")
			client = &http.Client{Timeout: timeout}
			log.Debugf("TLS authentication using root certificates")
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		//Setup HTTPS client
		tlsConfig := &tls.Config{
			RootCAs: caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		tr := &http.Transport{TLSClientConfig: tlsConfig}
		client = &http.Client{Transport: tr, Timeout: timeout}
		log.Debugf("Using CA certificate %s", certFile)
	}
}

// Executes the request.
// If successful returns response and nil, if there is any error it return error.
func doRequest(request *http.Request) ([]byte, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		errResp := new(ErrorResponse)
		_ = json.NewDecoder(response.Body).Decode(errResp)
		return nil, fmt.Errorf("%d: %s", response.StatusCode, errResp.Detail)
	}

	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
	default:
		reader = response.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Validates the response's status code.
// If status code = SUCCESS it returns nil, else it returns invalid status code error.
func checkStatusCode(status Status) error {
	if status.StatusCode != successStatusCode {
		return fmt.Errorf("error while validating response status: Status Code: %s, Status Message: %s", status.StatusCode, status.StatusDescription.Description)
	}
	return nil
}
