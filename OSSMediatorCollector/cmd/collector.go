package main

import (
	"collector/OssCollector"
	"collector/pkg/utils"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	logger "log"
	"net/http"
	"os"
	"sync"
)

var (
	confFile         string
	certFile         string
	skipTLS          bool
	logDir           string
	logLevel         int
	enableConsoleLog bool
	running          bool
	goroutine        sync.WaitGroup
	stopCh           chan struct{}
	internalRunning  bool
)

var (
	//clientID     = "4748e241-b573-422c-a423-e2cd0af95de2"     //grafana-trial
	//clientSecret = "Ayh8Q~MhHURqNFE7aYGVdMHT6Oyie.FMDTCrobd9" //grafana-trial
	clientID     = "634d076a-8d0f-4ce3-bf02-ed1a28fbb60a"     //dev2
	clientSecret = "I0V8Q~RkEMzw0qSqQFPz1D4V~x7zr3HIq9b.hdgS" //dev2
	redirectURI  = "https://10.183.35.228:9000/callback"
	//redirectURI  = "http://localhost:8080/callback"
	oauth2Config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     microsoft.AzureADEndpoint("c8c2a43b-31cc-430e-87b1-394e4cc06d9b"),
		Scopes:       []string{"openid", "profile", "offline_access", "User.Read", "api://634d076a-8d0f-4ce3-bf02-ed1a28fbb60a/token"},
	}
)

type User struct {
	Username     string `json:"user"`
	AuthType     string `json:"auth_type"`
	Password     string `json:"password"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	BaseUrl      string `json:"base_url"`
}

type RequestBodyOSSTrigger struct {
	Action string `json:"action"`
}

type Oss struct {
	Status string `json:"status"`
	Action string `json:"action"`
}

var users map[string]User
var oss Oss

func handleAddABACUserRequest(w http.ResponseWriter, r *http.Request) {
	if users == nil {
		users = make(map[string]User)
	}
	fmt.Println("Request to Configure Azure Token")
	w.Header().Set("Access-Control-Allow-Origin", "https://10.183.35.228:3000") // Update with your Grafana frontend URL.
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request to Configure Azure Token")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Update with your Grafana frontend URL.
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin,Authorization, Accept")

	fmt.Println("Inside callback")

	if users == nil {
		users = make(map[string]User)
	}

	code := r.URL.Query().Get("code")

	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Error exchanging code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	client := oauth2Config.Client(r.Context(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// Read the user's profile data
	var userData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		http.Error(w, "Failed to read user data", http.StatusInternalServerError)
		return
	}

	// Extract the email from the user's profile
	user, ok := userData["mail"].(string)
	if !ok {
		http.Error(w, "Email not found", http.StatusInternalServerError)
		return
	}

	accessToken := token.AccessToken
	refreshToken := token.RefreshToken

	fmt.Println("\n\naccess token : ", accessToken)
	fmt.Println("\n\nRefresh token : ", refreshToken)

	// Store or use the token as needed.
	// You can use token.AccessToken for API requests.

	err = utils.StoreToken(user, []byte(accessToken), []byte(refreshToken))
	if err != nil {
		fmt.Println("Error storing token:", err)
		return
	}

	//save in conf file
	err = utils.StoreConf(user, "ADTOKEN", "")
	if err != nil {
		fmt.Println("Error updating config file : ", err)
		return
	}

	// Respond with a success message and a 200 OK status
	response := map[string]string{
		"message": "Success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	fmt.Fprintf(w, "ABAC User added...you can close the window")
	// Optionally, you can also retrieve the user's profile information.
	// See: https://docs.microsoft.com/en-us/azure/active-directory/develop/quickstart-v2-go#step-4-authenticate-and-authorize
}

func handleAddUserRequest(w http.ResponseWriter, r *http.Request) {
	if users == nil {
		users = make(map[string]User)
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse the request body
	var requestBody User
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Validate the required fields
	if requestBody.Username == "" || requestBody.AuthType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	//validate fields bases on auth type
	if requestBody.AuthType == "password" {
		if requestBody.Password == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
	}

	if requestBody.AuthType == "azure_token" {
		if requestBody.AccessToken == "" || requestBody.RefreshToken == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
	}

	fmt.Println(requestBody)

	// Process the request (store user details in the map)
	users[requestBody.Username] = requestBody

	//save the user in conf file and storesecret directory as well
	user := requestBody.Username
	password := requestBody.Password
	accessToken := requestBody.AccessToken
	refreshToken := requestBody.RefreshToken
	if requestBody.AuthType == "password" {
		err = utils.StorePassword(user, []byte(password))
		if err != nil {
			fmt.Println("Error storing password:", err)
			return
		}
	} else if requestBody.AuthType == "azure_token" {
		err = utils.StoreToken(user, []byte(accessToken), []byte(refreshToken))
		if err != nil {
			fmt.Println("Error storing token:", err)
			return
		}
	}

	//save in conf file
	err = utils.StoreConf(user, requestBody.AuthType, requestBody.BaseUrl)
	if err != nil {
		fmt.Println("Error updating config file : ", err)
		return
	}

	// Respond with a success message and a 200 OK status
	response := map[string]string{
		"message": "Success",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func handleDeleteRequest(w http.ResponseWriter, r *http.Request) {
	if users == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse the request body
	var requestBody User
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	//
	if requestBody.Username == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Process the request (perform necessary actions here)
	fmt.Println(requestBody)

	username, ok := users[requestBody.Username]
	if !ok {
		http.Error(w, "Username not found in request body.", http.StatusBadRequest)
		return
	}

	_, exists := users[username.Username]
	if !exists {
		http.Error(w, "User not found in the database", http.StatusNotFound)
		return
	}

	//delete user from conf file and storesecret directory
	err = utils.DeleteConf(username.Username)

	if err != nil {
		http.Error(w, "Error deleting from config", http.StatusNotFound)
		return
	}
	delete(users, username.Username)
	response := map[string]string{
		"message": "Success",
	}
	// Respond with 200 OK on successful deletion
	w.WriteHeader(http.StatusOK)
	fmt.Println("User deleted")
	//fmt.Fprintf(w, "User with ID %d deleted successfully.", requestBody.Username)
	json.NewEncoder(w).Encode(response)
}

func handleListUsersRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var formattedUsers []map[string]string

	// Iterate over the existing user map and extract the desired fields
	for _, user := range users {
		formattedUser := map[string]string{
			"auth_type": user.AuthType,
			"user":      user.Username,
		}
		formattedUsers = append(formattedUsers, formattedUser)
	}

	// Create a map with a "users" key and the formatted user data
	response := map[string][]map[string]string{
		"users": formattedUsers,
	}

	// Convert the response to JSON and send it in the HTTP response
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("users: ", users)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func handleOSSStatusRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Responding OSS Status")
	fmt.Println("oss status ", oss.Status)

	//remove this later for security reasons
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(oss)
}

func handleOSSTriggerRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("Here in trigger handle")
	//remove this later for security reasons
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse the request body
	var requestBodyOSSTrigger RequestBodyOSSTrigger
	err := json.NewDecoder(r.Body).Decode(&requestBodyOSSTrigger)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println("request : ", requestBodyOSSTrigger)
	// Validate the required fields
	if requestBodyOSSTrigger.Action == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	fmt.Println(requestBodyOSSTrigger)

	// Process the request -> start or stop OSSMediator
	//go TriggerOSSMediator(requestBodyOSSTrigger.Action)

	//validate user request
	if requestBodyOSSTrigger.Action == "start" && oss.Status == "running" {
		http.Error(w, "OSSMediator already running", http.StatusBadRequest)
		return
	}
	if requestBodyOSSTrigger.Action == "stop" && oss.Status == "stopped" {
		http.Error(w, "OSSMediator already stopped", http.StatusBadRequest)
		return
	}
	/***
		switch requestBodyOSSTrigger.Action {
		case "start":
			if !running {
				fmt.Println("Starting OSS Collector from server")
				goroutine.Add(1)
				running = true
				oss.Status = "running"
				fmt.Println("OSS status : ", oss.Status)
				go OssCollector.StartCollector(confFile, internalRunning, stopCh, &goroutine)
			} else {
				http.Error(w, "OSSMediator already running", http.StatusConflict)
				return
			}
		case "stop":
			if running {
				fmt.Println("Stopping OSSCollector")
				close(stopCh)
				goroutine.Wait()
				running = false
			} else {
				http.Error(w, "OSSMediator already stopped", http.StatusConflict)
				return
			}
		}
	***/

	if requestBodyOSSTrigger.Action == "start" {
		//newConfig, err := loadConfig(fileToWatch)
		//if err != nil {
		//	log.Println("Error reloading configuration:", err)
		//	} else {
		//		config = newConfig
		//		log.Printf("Updated Configuration: %+v\n", config)
		//	}

		oss.Status = "running"
		fmt.Println("OSS status : ", oss.Status)
		fmt.Println("starting collector...")
		goroutine.Add(1)
		running = true
		stopCh = make(chan struct{})
		go OssCollector.StartCollector(confFile, running, stopCh, &goroutine)
	} else if requestBodyOSSTrigger.Action == "stop" {
		//isRunning := OssCollector.StopCollector()
		oss.Status = "stopped"
		fmt.Println("OSS status from status: ", oss.Status)
		//fmt.Println("oss status from collector: ", isRunning)
		//if !isRunning {
		oss.Status = "stopped"
		//}
		fmt.Println("stopping collector")
		close(stopCh)
		goroutine.Wait()
		running = false
	}

	// Respond with a success message and a 200 OK status
	response := map[string]string{
		"message": "Success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	//Read command line options
	parseFlags()
	//initialize logger
	initLogger(logDir, logLevel)

	// Respond with oss status
	oss.Status = "stopped"
	running = false
	mux := http.NewServeMux()
	// Setup CORS options
	corsOpts := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://10.183.35.228:3000", "http://10.183.35.228:9000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	mux.HandleFunc("/add_user", handleAddUserRequest)
	//mux.HandleFunc("/add_abac", handleAddUserRequest)
	mux.HandleFunc("/list_users", handleListUsersRequest) // Add a new route for handling GET requests
	mux.HandleFunc("/delete_user", handleDeleteRequest)
	mux.HandleFunc("/status", handleOSSStatusRequest)
	mux.HandleFunc("/trigger", handleOSSTriggerRequest)
	mux.HandleFunc("/add_abac", handleAddABACUserRequest)
	mux.HandleFunc("/callback", handleCallback)

	// Wrap the original mux with the CORS middleware
	handler := corsOpts.Handler(mux)

	//check if config file has users
	// if users exist, populate users in the map and start oss, else wait for admin to add users
	//and wait for mediator trigger

	isUserConfigured := populateUsers()
	if isUserConfigured {
		//fmt.Println("Starting OSS Collector from config")
		//goroutine.Add(1)
		//running = true
		//var startCollectorRunning bool
		//go OssCollector.StartCollector(confFile, startCollectorRunning, stopCh, &goroutine)
		//oss.Status = "running"
		fmt.Println("OSS status : ", oss.Status)
	}

	certPath := "../resources/grafana.crt"
	keyFile := "../resources/grafana.key"

	// Start the server on port 8080
	fmt.Println("Server listening on https://<ip>:9000")
	err := http.ListenAndServeTLS("0.0.0.0:9000", certPath, keyFile, handler)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func populateUsers() bool {
	users = make(map[string]User)
	filePath := "../resources/conf.json"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	//unmarshal json into map
	var jsonData map[string]interface{}
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return false
	}

	//check if users key exists in the json
	if usersList, ok := jsonData["users"].([]interface{}); ok {
		//extract users
		for _, userItem := range usersList {
			if userMap, isUserMap := userItem.(map[string]interface{}); isUserMap {
				var user User
				if emailID, hasEmailID := userMap["email_id"].(string); hasEmailID {
					user.Username = emailID
				}
				if authType, hasAuthType := userMap["auth_type"].(string); hasAuthType {
					user.AuthType = authType
				}
				users[user.Username] = user
			}
		}
	}
	if len(users) == 0 {
		fmt.Println("no users configured")
		return false
	} else {
		fmt.Println("no. of users configured initially: ", len(users))
		for _, user2 := range users {
			fmt.Println(user2)
		}
		return true
	}
}

/*
func loadConfig(filePath string) (*Configuration, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Configuration
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err	}
	return &config, nil
}
*/

// Reads command line options
func parseFlags() {
	//read command line arguments
	flag.StringVar(&confFile, "conf_file", "../resources/conf.json", "config file path")
	flag.StringVar(&certFile, "cert_file", "", "certificate file path")
	flag.BoolVar(&skipTLS, "skip_tls", false, "skip TLS authentication")
	flag.StringVar(&logDir, "log_dir", "../log", "Log directory")
	flag.IntVar(&logLevel, "log_level", 4, "Log level")
	flag.BoolVar(&enableConsoleLog, "enable_console_log", false, "Enable console logging, if true logs won't be written to file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ./collector [options]\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "\t-h, --help\n\t\tOutput a usage message and exit.\n")
		fmt.Fprintf(os.Stderr, "\t-conf_file string\n\t\tConfig file path (default \"../resources/conf.json\")\n")
		fmt.Fprintf(os.Stderr, "\t-cert_file string\n\t\tCertificate file path (if cert_file is not passed then it will establish TLS auth using root certificates.)\n")
		fmt.Fprintf(os.Stderr, "\t-log_dir string\n\t\tLog Directory (default \"../log\"), logs will be stored in collector.log file.\n")
		fmt.Fprintf(os.Stderr, "\t-log_level int\n\t\tLog Level (default 4). Values: 0 (PANIC), 1 (FATAl), 2 (ERROR), 3 (WARNING), 4 (INFO), 5 (DEBUG)\n")
		fmt.Fprintf(os.Stderr, "\t-skip_tls\n\t\tSkip TLS Authentication\n")
		fmt.Fprintf(os.Stderr, "\t-enable_console_log\n\t\tEnable console logging, if true logs won't be written to file\n")
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
