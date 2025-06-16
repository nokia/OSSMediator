package ndacapis

import (
	"bytes"
	"collector/pkg/config"
	"collector/pkg/utils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	listGngResp = `{
	"status": {
		"status_code": "SUCCESS",
		"status_description": {
			"description_code": "OK",
			"description": "SUCCESS"
		}
	},
	"gng_info": [
		{
			"admin_state": "STAGED",
			"is_gr_setup": false,
			"site_info": [],
			"gng_id": "test_gng_1"
		},
		{
			"admin_state": "NonGR_FULLY_ACTIVATED",
			"is_gr_setup": "true",
			"gng_id": "test_gng_2",
			"slice_id": "test_2"
		}
	]
}`
)

func TestGetGngDetails(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", AuthType: "PASSWORD", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.SliceIDs = map[string]string{}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listGngResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getGngDetail")
	getGngDetails(&config.APIConf{API: "/getGngDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 1 {
		t.Fail()
	}
	if user.NhgIDs[0] != "test_gng_2" {
		t.Fail()
	}
}

func TestGetGngDetailsForInvalidCase(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
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
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getGngDetails(&config.APIConf{API: "/getGngDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetGngDetailsForInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}

	CreateHTTPClient("", true)
	config.Conf = config.Config{
		BaseURL: ":",
	}
	getGngDetails(&config.APIConf{API: "/getGngDetails", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetGngDetailsWithInvalidResponse(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getGngDetails(&config.APIConf{API: "/getGngDetails", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
}

func TestGetGngDetailsABAC(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", AuthType: "ADTOKEN", IsSessionAlive: true, ResponseDest: "./tmp"}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	user.SliceIDs = map[string]string{}
	user.AccountIDsABAC = map[string][]string{
		"org1": {"acc1", "acc2"},
		"org2": {"acc3", "acc4"},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, listGngResp)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}
	CreateHTTPClient("", false)
	utils.CreateResponseDirectory(user.ResponseDest, "/getGngDetail")
	getGngDetails(&config.APIConf{API: "/getGngDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 1 {
		t.Fail()
	}
}

func TestGetGngDetailsABACForInvalidCase(t *testing.T) {
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	user.AccountIDsABAC = map[string][]string{
		"org1": {"acc1", "acc2"},
		"org2": {"acc3", "acc4"},
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
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getGngDetails(&config.APIConf{API: "/getGngDetail", Interval: 15}, &user, 1234, true)
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestGetGngDetailsABACForInvalidURL(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	user.AccountIDsABAC = map[string][]string{
		"org1": {"acc1", "acc2"},
		"org2": {"acc3", "acc4"},
	}

	CreateHTTPClient("", true)
	config.Conf = config.Config{
		BaseURL: ":",
	}
	getGngDetails(&config.APIConf{API: "/getGngDetails", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "missing protocol scheme") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}

func TestGetGngDetailsABACWithInvalidResponse(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()
	user := config.User{Email: "testuser@nokia.com", IsSessionAlive: true}
	user.SessionToken = &config.SessionToken{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
		ExpiryTime:   utils.CurrentTime(),
	}
	user.HwIDsABAC = map[string]config.OrgAccDetails{}
	user.NhgIDsABAC = map[string]config.OrgAccDetails{}
	user.AccountIDsABAC = map[string][]string{
		"org1": {"acc1", "acc2"},
		"org2": {"acc3", "acc4"},
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ``)
	}))
	defer testServer.Close()
	config.Conf = config.Config{
		BaseURL: testServer.URL,
	}

	CreateHTTPClient("", true)
	getGngDetails(&config.APIConf{API: "/getGngDetails", Interval: 15}, &user, 1234, true)
	if !strings.Contains(buf.String(), "Unable to decode response") {
		t.Fail()
	}
	if len(user.NhgIDs) != 0 {
		t.Fail()
	}
	if len(user.NhgIDsABAC) != 0 {
		t.Fail()
	}
}
