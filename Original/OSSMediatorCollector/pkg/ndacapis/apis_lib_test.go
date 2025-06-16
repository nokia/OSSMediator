/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package ndacapis

import (
	"bytes"
	"collector/pkg/config"
	"collector/pkg/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateHTTPClientForSkipTLS(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("", true)
	if !strings.Contains(buf.String(), "Skipping TLS authentication") {
		t.Fail()
	}
}

func TestCreateHTTPClientWithRootCRT(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("", false)
	if !strings.Contains(buf.String(), "TLS authentication using root certificates") {
		t.Fail()
	}
}

func TestCreateHTTPClientWithNonExistingCRTFile(t *testing.T) {
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("tmp.crt", false)
	if !(strings.Contains(buf.String(), "Error while reading server certificate file") && strings.Contains(buf.String(), "TLS authentication using root certificates")) {
		t.Fail()
	}
}

func TestCreateHTTPClientWithCRTFile(t *testing.T) {
	tmpfile := createTmpFile(".", "crt", []byte(``))
	if tmpfile == "" {
		t.Fail()
	}
	defer os.Remove(tmpfile)
	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	CreateHTTPClient(tmpfile, false)
	if !strings.Contains(buf.String(), "Using CA certificate "+tmpfile) {
		t.Fail()
	}

}

func createTmpFile(dir string, prefix string, content []byte) string {
	tmpfile, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		log.Error(err)
		return ""
	}

	if _, err = tmpfile.Write(content); err != nil {
		log.Error(err)
		return ""
	}
	if err = tmpfile.Close(); err != nil {
		log.Error(err)
		return ""
	}
	return tmpfile.Name()
}

func TestStartDataCollectionWithInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	config.Conf.BaseURL = "http://localhost:8080"
	config.Conf.Users = []*config.User{
		{
			Email:          "testuser@nokia.com",
			Password:       "MTIzNA==",
			IsSessionAlive: true,
			ResponseDest:   "./tmp",
			SessionToken: &config.SessionToken{
				AccessToken:  "",
				RefreshToken: "",
				ExpiryTime:   utils.CurrentTime(),
			},
		},
	}
	config.Conf.MetricAPIs = []*config.APIConf{{API: "/pmdata", Interval: 15}, {API: "/fmdata", Interval: 15, Type: "HISTORY"}}
	config.Conf.SimAPIs = []*config.APIConf{{API: "/sims", Interval: 15}}
	config.Conf.ListNetworkAPI = &config.ListNetworkAPIConf{GngAPI: "/generic-network-groups", NhgAPI: "/network-hardware-groups", Interval: 60}
	config.Conf.Limit = 10
	config.Conf.MaxConcurrentProcess = 1
	CreateHTTPClient("", true)

	StartDataCollection()
	time.Sleep(2 * time.Millisecond)
	if !strings.Contains(buf.String(), "Triggered http://localhost:8080/network-hardware-groups") {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "Skipping API call for testuser@nokia.com") {
		t.Fail()
	}
	if config.Conf.Users[0].IsSessionAlive {
		t.Fail()
	}
}

func TestStartDataCollection(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.String()
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(url, "network-hardware-groups") {
			fmt.Fprintln(w, listNhgResp)
		} else {
			fmt.Fprintln(w, fmResponse)
		}
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	oldCurrentTime := utils.CurrentTime
	defer func() { utils.CurrentTime = oldCurrentTime }()

	myCurrentTime := func() time.Time {
		return time.Date(2018, 12, 17, 20, 9, 58, 0, time.UTC)
	}
	utils.CurrentTime = myCurrentTime
	config.Conf.BaseURL = testServer.URL
	config.Conf.Users = []*config.User{
		{
			Email:          "testuser@nokia.com",
			Password:       "MTIzNA==",
			ResponseDest:   "./tmp",
			IsSessionAlive: true,
			SessionToken: &config.SessionToken{
				AccessToken:  "",
				RefreshToken: "",
				ExpiryTime:   utils.CurrentTime(),
			},
		},
	}
	config.Conf.MetricAPIs = []*config.APIConf{{API: "/{nhg_id}/fmdata", Interval: 15, Type: "HISTORY", MetricType: "RADIO"}}
	config.Conf.Limit = 10
	config.Conf.ListNetworkAPI = &config.ListNetworkAPIConf{NhgAPI: "/network-hardware-groups", Interval: 60}
	config.Conf.MaxConcurrentProcess = 1
	CreateHTTPClient("", true)

	StartDataCollection()
	time.Sleep(20 * time.Millisecond)
	if !strings.Contains(buf.String(), "Triggered "+testServer.URL) {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "Writing response") {
		t.Fail()
	}
}
