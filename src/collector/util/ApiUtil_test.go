/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
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
	tmpfile, err := ioutil.TempFile(".", "crt")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if err = tmpfile.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}
	defer os.Remove(tmpfile.Name())

	//capturing the logs in buffer for assertion
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(log.Level(5))
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	CreateHTTPClient(tmpfile.Name(), false)
	if !strings.Contains(buf.String(), "Using CA certificate "+tmpfile.Name()) {
		t.Fail()
	}
}

//Checksum validation test
func TestValidateCheckSum(t *testing.T) {
	content := []byte("test file")
	tmpfile, err := ioutil.TempFile(".", "tmp")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write(content); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err = tmpfile.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}

	err = validateCheckSum(tmpfile.Name(), "f20d9f2072bbeb6691c0f9c5099b01f3")
	if err != nil {
		t.Fail()
	}
}

//Checksum validation test with wrong checksum value
func TestValidateCheckSumWithWrongSum(t *testing.T) {
	content := []byte("test tmp file")
	tmpfile, err := ioutil.TempFile(".", "tmp")
	if err != nil {
		t.Log(err)
	}

	defer os.Remove(tmpfile.Name())
	if _, err = tmpfile.Write(content); err != nil {
		t.Log(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Log(err)
	}

	err = validateCheckSum(tmpfile.Name(), "f20d9f2072bbeb6691c0f9c5099b01f3")
	if err == nil {
		t.Fail()
	}
}

func TestValidateCheckSumWithEmptyFile(t *testing.T) {
	err := validateCheckSum("", "f20d9f2072bbeb6691c0f9c5099b01f3")
	if err == nil {
		t.Fail()
	}
}

func TestUntar(t *testing.T) {
	responseDir := "./tmp"
	fileName := "./_testdata/test_file.tgz"
	CreateResponseDirectory(responseDir, "")
	defer os.RemoveAll(responseDir)
	err := untar(fileName, responseDir)
	if err != nil {
		t.Fail()
	}
	files, _ := ioutil.ReadDir(responseDir)
	if len(files) != 10 {
		t.Fail()
	}
}

func TestUntarWithEmptyFile(t *testing.T) {
	err := untar("", "./tmp")
	if err == nil {
		t.Fail()
	}
}

func TestWriteResponseWithWrongEncodedFile(t *testing.T) {
	responseDir := "./tmp"
	CreateResponseDirectory(responseDir, "")
	defer os.RemoveAll(responseDir)
	response := &GetAPIResponse{
		FileName:    "test_file.tgz",
		MD5CheckSum: "checkSum",
		EncodedFile: "encoded",
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	writeResponse(response, user, "")
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
}

func TestWriteResponseWithWrongDir(t *testing.T) {
	fileName := "./_testdata/test_file.tgz"
	content, _ := ioutil.ReadFile(fileName)
	encoded := base64.StdEncoding.EncodeToString(content)
	f, err := os.Open(fileName)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer f.Close()

	response := &GetAPIResponse{
		FileName:    "test_file.tgz",
		MD5CheckSum: "checkSum",
		EncodedFile: encoded,
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	writeResponse(response, user, "")
	if !strings.Contains(buf.String(), "File creation failed") {
		t.Fail()
	}
}

func TestWriteResponseWithWrongCheckSum(t *testing.T) {
	responseDir := "./tmp"
	CreateResponseDirectory(responseDir, "")
	defer os.RemoveAll(responseDir)
	fileName := "./_testdata/test_file.tgz"
	content, _ := ioutil.ReadFile(fileName)
	encoded := base64.StdEncoding.EncodeToString(content)
	f, err := os.Open(fileName)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer f.Close()

	response := &GetAPIResponse{
		FileName:    "test_file.tgz",
		MD5CheckSum: "checkSum",
		EncodedFile: encoded,
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	writeResponse(response, user, "")
	if !strings.Contains(buf.String(), "CheckSum Validation failed") {
		t.Fail()
	}
}

func TestWriteResponse(t *testing.T) {
	responseDir := "./tmp"
	CreateResponseDirectory(responseDir, "")
	fileName := "./_testdata/test_file.tgz"
	content, _ := ioutil.ReadFile(fileName)
	encoded := base64.StdEncoding.EncodeToString(content)
	f, err := os.Open(fileName)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Log(err)
		t.Fail()
	}

	checkSum := fmt.Sprintf("%x", h.Sum(nil))
	response := &GetAPIResponse{
		FileName:    "test_file.tgz",
		MD5CheckSum: checkSum,
		EncodedFile: encoded,
	}
	user := &User{Email: "testuser@nokia.com", ResponseDest: "./tmp"}
	apiLastReceivedFile := lastReceivedFile + "_" + user.Email + "_."
	defer os.RemoveAll(responseDir)
	defer os.Remove(apiLastReceivedFile)
	writeResponse(response, user, "")
	data, _ := ioutil.ReadFile(apiLastReceivedFile)
	if string(data) != "2018-03-14T13:55:00+05:30" {
		t.Fail()
	}
}

//Unit test for callAPI and doRequest
func TestCallAPIForInvalidCase(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"status": {
				"status_code": "FAILURE",
				"status_description": {
					"description_code": "INVALID_ARGUMENT",
					"description": "Token sent is empty. Invalid Token"
				}
			}
		}`)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	response := callAPI(testServer.URL, &user)
	if response != nil {
		t.Fail()
	}
}

func TestCallAPIForInvalidURL(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	CreateHTTPClient("", true)
	response := callAPI(":", &user)
	if response != nil {
		t.Fail()
	}
}

//Unit test for callAPI and doRequest
func TestCallAPI(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"filename":"test_file.tgz",
			"md5_checksum":"checksum",
			"encoded_file":"encoded_file",
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	resp := callAPI(testServer.URL, &user)
	if resp == nil {
		t.Fail()
	}
	if resp.Status.StatusCode != "SUCCESS" || resp.Status.StatusDescription.Description != "Success" || resp.FileName != "test_file.tgz" || resp.EncodedFile != "encoded_file" || resp.MD5CheckSum != "checksum" {
		t.Fail()
	}
}

func TestCallAPIWithErrorStatusCode(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	CreateHTTPClient("", false)
	resp := callAPI(testServer.URL, &user)
	if resp != nil {
		t.Fail()
	}
}

func TestCallAPIWithInvalidResponse(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	CreateHTTPClient("", false)
	resp := callAPI(testServer.URL, &user)
	if resp != nil && strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
}

func TestCallAPIWIthLastReceivedFile(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"filename":"test_file.tgz",
			"md5_checksum":"checksum",
			"encoded_file":"encoded_file",
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	tmpFile := lastReceivedFile + "_" + user.Email + "_" + path.Base(testServer.URL)
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer file.Close()
	_, err = file.Write([]byte("2018-03-14T13:55:00+05:30"))
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	defer testServer.Close()
	defer os.Remove(tmpFile)
	CreateHTTPClient("", false)
	resp := callAPI(testServer.URL, &user)
	if resp == nil {
		t.Fail()
	}
	if resp.Status.StatusCode != "SUCCESS" || resp.Status.StatusDescription.Description != "Success" || resp.FileName != "test_file.tgz" || resp.EncodedFile != "encoded_file" || resp.MD5CheckSum != "checksum" {
		t.Fail()
	}
}
func TestCallAPIWithSkipCert(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"filename":"test_file.tgz",
			"md5_checksum":"checksum",
			"encoded_file":"encoded_file",
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	resp := callAPI(testServer.URL, &user)
	if resp == nil {
		t.Fail()
	}
	if resp.Status.StatusCode != "SUCCESS" || resp.Status.StatusDescription.Description != "Success" || resp.FileName != "test_file.tgz" || resp.EncodedFile != "encoded_file" || resp.MD5CheckSum != "checksum" {
		t.Fail()
	}
}

func TestCallAPIWithCert(t *testing.T) {
	tmpfile, err := ioutil.TempFile(".", "crt")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if err = tmpfile.Close(); err != nil {
		t.Log(err)
		t.Fail()
	}

	CreateHTTPClient(tmpfile.Name(), false)
	user := User{Email: "testuser@nokia.com", password: "1234"}
	user.sessionToken = &sessionToken{
		accessToken:  "accessToken",
		refreshToken: "refreshToken",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"filename":"test_file.tgz",
			"md5_checksum":"checksum",
			"encoded_file":"encoded_file",
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	defer testServer.Close()
	defer os.Remove(tmpfile.Name())
	resp := callAPI(testServer.URL, &user)
	if resp == nil {
		t.Fail()
	}
	if resp.Status.StatusCode != "SUCCESS" || resp.Status.StatusDescription.Description != "Success" || resp.FileName != "test_file.tgz" || resp.EncodedFile != "encoded_file" || resp.MD5CheckSum != "checksum" {
		t.Fail()
	}
}

func TestTrigger(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234", ResponseDest: "./tmp"}
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"filename":"test_file.tgz",
			"md5_checksum":"checksum",
			"encoded_file":"encoded_file",
			"status": {
				"status_code": "SUCCESS",
				"status_description": {
					"description_code": "NOT_SPECIFIED",
					"description": "Success"
				}
			}
		}`)
	}))
	defer testServer.Close()
	CreateHTTPClient("", true)
	ticker := time.NewTicker(500 * time.Millisecond)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	go Trigger(ticker, testServer.URL, &user)
	time.Sleep(1 * time.Second)
	if !strings.Contains(buf.String(), "Triggered "+testServer.URL+" for "+user.Email) {
		t.Fail()
	}
	if !strings.Contains(buf.String(), "Writting response for "+user.Email) {
		t.Fail()
	}
}

func TestTriggerWithWrongURL(t *testing.T) {
	user := User{Email: "testuser@nokia.com", password: "1234", ResponseDest: "./tmp"}
	user.sessionToken = &sessionToken{
		accessToken:  "",
		refreshToken: "",
		expiryTime:   time.Now(),
	}
	CreateHTTPClient("", true)
	ticker := time.NewTicker(500 * time.Millisecond)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	go Trigger(ticker, "http://localhost:8080", &user)
	time.Sleep(1 * time.Second)
	if strings.Contains(buf.String(), "Writting response") {
		t.Fail()
	}
}

func TestStoreLastReceivedFileTimePM(t *testing.T) {
	responseDir := "./tmp/pm"
	fileName := "./_testdata/test_file.tgz"
	err := os.MkdirAll(responseDir, os.ModePerm)
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll("./tmp")
	err = untar(fileName, responseDir)
	if err != nil {
		t.Fail()
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	err = storeLastReceivedFileTime(user, "/pm")
	if err != nil {
		t.Fail()
	}
	//Reading LastReceivedFile value from file
	fileName = lastReceivedFile + "_" + user.Email + "_" + "pm"
	defer os.Remove(fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil || string(data) != "2018-03-14T13:55:00+05:30" {
		t.Fail()
	}
}

func TestStoreLastReceivedFileTimeFM(t *testing.T) {
	responseDir := "./tmp/fm"
	fileName := "./_testdata/test_file.tgz"
	err := os.MkdirAll(responseDir, os.ModePerm)
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll("./tmp")
	err = untar(fileName, responseDir)
	if err != nil {
		t.Fail()
	}
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	err = storeLastReceivedFileTime(user, "/fm")
	if err != nil {
		t.Fail()
	}
	//Reading LastReceivedFile value from file
	fileName = lastReceivedFile + "_" + user.Email + "_" + "fm"
	defer os.Remove(fileName)
	_, err = ioutil.ReadFile(fileName)
	if err != nil {
		t.Fail()
	}
}

func TestStoreLastReceivedFileTimeWithoutFiles(t *testing.T) {
	responseDir := "./tmp"
	err := os.MkdirAll(responseDir, os.ModePerm)
	if err != nil {
		t.Fail()
	}
	defer os.RemoveAll(responseDir)
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	err = storeLastReceivedFileTime(user, "")
	if err == nil {
		t.Fail()
	}
}

func TestStoreLastReceivedFileTimeWithWrongDirectory(t *testing.T) {
	user := &User{Email: "testuser@okia.com", ResponseDest: "./tmp"}
	err := storeLastReceivedFileTime(user, "")
	if err == nil {
		t.Fail()
	}
}

func TestCreateResponseDirectory(t *testing.T) {
	respDir := "./tmp"
	CreateResponseDirectory(respDir, "http://localhost:8080/pmdata")
	defer os.RemoveAll(respDir)
	if _, err := os.Stat(respDir + "/pmdata"); os.IsNotExist(err) {
		t.Fail()
	}
}
