/*
* Copyright 2018 Nokia
* Licensed under BSD 3-Clause Clear License,
* see LICENSE file for details.
 */

package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	//LastReceivedFile keeps the timestamp of last file received from the GET API call
	//It will be sent as query parameter to the API to send response after the LastReceivedFile timestamp.
	lastReceivedFile = "./LastReceivedFile"

	//Timeout duration for HTTP calls
	timeout = time.Duration(30 * time.Second)

	//Query Params for GET APIS
	startTimeQueryParam = "start_timestamp"
	endTimeQueryParam   = "end_timestamp"

	//Headers
	authorizationHeader = "Authorization"

	//regex for PM file name
	pmFileNameFormat = `PM(?P<file_create_date>\d{8}\d{4}[\+-]\d{4}).*\.xml`

	timeFormat = "200601021504Z0700"

	//Success status code from response
	successStatusCode = "SUCCESS"

	//Backoff duration for retrying refresh token.
	initialBackoff = 5 * time.Minute
)

var (
	//Conf keeps the config from json and console
	Conf Config

	//HTTP client for all API calls
	client *http.Client
	// transport *http.Transport
)

//GetAPIResponse keeps track of response recevied from PM/FM API
type GetAPIResponse struct {
	FileName    string `json:"filename"`     //tar file name
	MD5CheckSum string `json:"md5_checksum"` //MD5sum value
	EncodedFile string `json:"encoded_file"` //Encode file
	Status      Status `json:"status"`       // Status of the response
}

//Status keeps track of status from response
type Status struct {
	StatusCode        string            `json:"status_code"`
	StatusDescription StatusDescription `json:"status_description"`
}

//StatusDescription keeps track of status description from response
type StatusDescription struct {
	DescriptionCode string `json:"description_code"`
	Description     string `json:"description"`
}

//sessionToken struct tracks the access_token, refresh_token and expiry_time of the token
//As the session token will be shared by multiple APIs.
type sessionToken struct {
	access       sync.Mutex
	accessToken  string
	refreshToken string
	expiryTime   time.Time
}

//Trigger the APIs periodically at specified interval
//  and writes response to responseDest directory.
func Trigger(ticker *time.Ticker, apiURL string, user *User) {
	for t := range ticker.C {
		log.Infof("Triggered %s for %s at %v", apiURL, user.Email, t)
		response := callAPI(apiURL, user)
		if response != nil {
			log.Infof("Writting response for %s to %s", user.Email, user.ResponseDest+"/"+path.Base(apiURL))
			writeResponse(response, user, apiURL)
		}
	}
}

//CreateHTTPClient creates HTTP client for all the GET/POST API calls, if certFile is empty and skipTLS is false TLS authentication will be done using root certificates.
//certFile keeps the server certificate file path
//skipTLS if true all API calls will skip TLS auth.
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
		caCert, err := ioutil.ReadFile(certFile)
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

//CallAPI calls the API, adds authorization, query params and returns response.
//If successful it returns response as array of byte, if there is any error it return nil.
func callAPI(apiURL string, user *User) *GetAPIResponse {
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil
	}
	//Lock the SessionToken object before calling API
	user.sessionToken.access.Lock()
	request.Header.Set(authorizationHeader, user.sessionToken.accessToken)
	user.sessionToken.access.Unlock()

	//Adding query params
	query := request.URL.Query()
	query.Add(endTimeQueryParam, truncateSeconds(time.Now()).Format(time.RFC3339))
	//Reading start time value from file
	apiLastReceivedFile := lastReceivedFile + "_" + user.Email + "_" + path.Base(apiURL)
	data, err := ioutil.ReadFile(apiLastReceivedFile)
	if err != nil || len(data) == 0 {
		//sending start time = time.Now() - 15 minutes
		query.Add(startTimeQueryParam, truncateSeconds(time.Now().Add(time.Duration(-1*15)*time.Minute)).Format(time.RFC3339))
	} else {
		query.Add(startTimeQueryParam, strings.TrimSpace(string(data)))
	}
	request.URL.RawQuery = query.Encode()
	log.Info(startTimeQueryParam, ": ", query[startTimeQueryParam])
	log.Info(endTimeQueryParam, ": ", query[endTimeQueryParam])
	log.Info("URL:", request.URL)

	response, err := doRequest(request)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Error while calling %s for %s", apiURL, user.Email)
		return nil
	}

	//Map the received response to getAPIResponse struct
	resp := new(GetAPIResponse)
	err = json.NewDecoder(bytes.NewReader(response)).Decode(resp)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to decode response")
		return nil
	}

	//check response for status code
	err = checkStatusCode(resp.Status)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Errorf("Invalid status code received while calling %s for %s", apiURL, user.Email)
		return nil
	}
	log.Infof("%s called successfully for %s.", apiURL, user.Email)
	return resp
}

//Executes the request.
//If successful returns response and nil, if there is any error it return error.
func doRequest(request *http.Request) ([]byte, error) {
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if 200 != response.StatusCode {
		return nil, fmt.Errorf("%d: %s", response.StatusCode, response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

//Validates the response's status code.
//If status code = SUCCESS it returns nil, else it returns invalid status code error.
func checkStatusCode(status Status) error {
	if status.StatusCode != successStatusCode {
		return fmt.Errorf("Error while validating response status: Status Code: %s, Status Message: %s", status.StatusCode, status.StatusDescription.Description)
	}
	return nil
}

//writeResponse reads the response, decode the encoded tar file from response
// and untars the decoded tar file to responseDest directory.
func writeResponse(response *GetAPIResponse, user *User, apiURL string) {
	//decode the encoded tar file
	decodedData, err := base64.StdEncoding.DecodeString(response.EncodedFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to decode response")
		return
	}

	//store the decoded tar file to responseDest directory
	responseDest := user.ResponseDest + "/" + path.Base(apiURL)
	fileName := responseDest + "/" + response.FileName
	err = writeFile(fileName, decodedData)
	if err != nil {
		log.Error(err)
		return
	}

	//validate MD5Sum
	err = validateCheckSum(fileName, response.MD5CheckSum)
	if err != nil {
		log.Error(err)
		return
	}

	//untar the received tar file
	err = untar(fileName, responseDest)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to unzip")
	}
	//deleting the received tar file
	err = os.Remove(fileName)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Unable to delete received tar file")
	}
	//storing LastReceivedFile timestamp value to file
	err = storeLastReceivedFileTime(user, apiURL)
	if err != nil {
		log.Error(err)
	}
}

//writes data to file
func writeFile(fileName string, data []byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("File creation failed: %v", err)
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Error while writting response to file: %v", err)
	}
	return nil
}

//Validates the integrity of received file using MD5SUM value.
//It calculates the MD5SUM of the received file and matches with the received check sum value.
//If successful it returns nil, if there is any error it return error.
//In case of validation fail it returns error with message check sum validation failed.
func validateCheckSum(fileName string, checkSum string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("Error while calculating checksum: %v", err)
	}
	defer file.Close()

	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("Error while calculating checksum: %v", err)
	}

	calculatedCheckSum := hex.EncodeToString(h.Sum(nil))
	if checkSum != calculatedCheckSum {
		return fmt.Errorf("CheckSum Validation failed: Recieved MD5Sum: %s, Calculated MD5Sum: %s", checkSum, calculatedCheckSum)
	}
	log.Info("Checksum validated")
	return nil
}

//Untars the received tar file from response and store the untarred files in responseDest directory.
// If successful it returns nil and if there is any error, it will return the error.
func untar(filename string, responseDest string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(responseDest, header.Name)
		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			defer f.Close()
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}

//Writes the last received file time to a file so that next time that tme stamp will be used as start_time for api calls.
//Creates file for each user and each APIs.
//For FM API it will store time.Now() - 15 minutes to the file.
//returns error if writing to or reading from the response directory fails.
func storeLastReceivedFileTime(user *User, apiURL string) error {
	responseDest := user.ResponseDest + "/" + path.Base(apiURL)
	files, err := ioutil.ReadDir(responseDest)
	if err != nil {
		return fmt.Errorf("Error while getting last received file: %v", err)
	}
	//return error if no files found in directory
	if len(files) == 0 {
		return fmt.Errorf("No files found inside %s", responseDest)
	}
	apiLastReceivedFile := lastReceivedFile + "_" + user.Email + "_" + path.Base(apiURL)
	log.Debug("Reading from ", apiLastReceivedFile)
	if strings.Contains(path.Base(apiURL), "fm") {
		err = writeFile(apiLastReceivedFile, []byte(truncateSeconds(time.Now().Add(time.Duration(-1*15)*time.Minute)).Format(time.RFC3339)))
		if err != nil {
			return fmt.Errorf("Unable to write last received file create date to file, error: %v", err)
		}
		return nil
	}
	//getting last file name
	lastFileName := files[len(files)-1].Name()
	//getting file create date from file name using regex
	r := regexp.MustCompile(pmFileNameFormat)
	captures := make(map[string]string)
	match := r.FindStringSubmatch(lastFileName)
	if match == nil {
		return fmt.Errorf("Error while getting last received file, %s din't match with the file format", lastFileName)
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}
		captures[name] = match[i]
	}

	fileCreateTime, err := time.Parse(timeFormat, captures["file_create_date"])
	if err != nil {
		return fmt.Errorf("Unable to parse last received file create date to date, error: %v", err)
	}
	err = writeFile(apiLastReceivedFile, []byte(truncateSeconds(fileCreateTime).Format(time.RFC3339)))
	if err != nil {
		return fmt.Errorf("Unable to write last received file create date to file, error: %v", err)
	}
	return nil
}

//CreateResponseDirectory creates directory named path, along with any necessary parents.
// If the directory creation fails it will terminate the program.
func CreateResponseDirectory(basePath string, api string) {
	path := basePath + "/" + path.Base(api)
	log.Infof("Creating %s directory", path)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalf("Error while creating %s", path)
	}
}

//truncates seconds part from time
func truncateSeconds(t time.Time) time.Time {
	duration := 60 * time.Second
	return t.Truncate(duration)
}
