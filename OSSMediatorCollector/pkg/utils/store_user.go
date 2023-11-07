package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type OSSUser struct {
	EmailId      string `json:"email_id"`
	ResponseDest string `json:"response_dest"`
	AuthType     string `json:"auth_type"`
}

type Config struct {
	OSSUsers []OSSUser `json:"users"`
	BaseUrl  string    `json:"base_url"`
}

var (
	confFile  string
	SecretDir = "./.secret"
)

/**
func StoreConf(user string, authType string, baseUrl string) error {
	filePath := "../resources/conf.json"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}

	var dataMap map[string]interface{}

	err = json.Unmarshal(file, &dataMap)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}
	//if dataMap["users"] == nil {
	//	dataMap["users"] = []OSSUser{} // Initialize the users field if it doesn't exist
	//}

	// Initialize the "users" field if it doesn't exist
	if dataMap["users"] == nil {
		dataMap["users"] = []interface{}{}
	}

	if user1, ok := dataMap["users"].([]interface{}); ok {
		newUser := OSSUser{EmailId: user, ResponseDest: "/reports", AuthType: authType}
		user1 = append(user1, newUser)
		dataMap["users"] = user1
	}

	if baseUrl != "" {
		dataMap["base_url"] = baseUrl
	}

	// Serialize the updated data to JSON
	updatedData, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Write the updated JSON data back to the file
	if err := ioutil.WriteFile(filePath, updatedData, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	fmt.Println("New user has been added and the data has been updated in the file.")
	return nil
}
**/

func StoreConf(user string, authType string, baseUrl string) error {
	filePath := "../resources/conf.json"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}

	var dataMap map[string]interface{}

	err = json.Unmarshal(file, &dataMap)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	// Ensure the "users" field is initialized as an empty array if not present
	if dataMap["users"] == nil {
		dataMap["users"] = []interface{}{}
	}

	// Type assertion to []interface{} for users array
	users, ok := dataMap["users"].([]interface{})
	if !ok {
		fmt.Println("Error type asserting users array")
		return errors.New("error type asserting users array")
	}

	// Create a new user object
	newUser := map[string]interface{}{
		"email_id":      user,
		"response_dest": "/reports",
		"auth_type":     authType,
	}

	// Append the new user to the existing users array
	users = append(users, newUser)

	// Update the "users" field in dataMap
	dataMap["users"] = users

	if baseUrl != "" {
		dataMap["base_url"] = baseUrl
	}

	// Serialize the updated data to JSON
	updatedData, err := json.MarshalIndent(dataMap, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	// Write the updated JSON data back to the file
	if err := ioutil.WriteFile(filePath, updatedData, 0644); err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	fmt.Println("New user has been added and the data has been updated in the file.")
	return nil
}

func DeleteConf(user string) error {
	filePath := "../resources/conf.json"
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	var jsonData map[string]interface{}
	err = json.Unmarshal(file, &jsonData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return err
	}

	if users, ok := jsonData["users"].([]interface{}); ok {
		var updatedUsers []interface{}
		for _, user3 := range users {
			if userMap, isUserMap := user3.(map[string]interface{}); isUserMap {
				if email, hasEmail := userMap["email_id"].(string); hasEmail && email == user {
					continue
				}
			}
			updatedUsers = append(updatedUsers, user3)
		}
		jsonData["users"] = updatedUsers
	}

	// Write the modified data back to the JSON file
	newData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}
	err = ioutil.WriteFile(filePath, newData, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	fmt.Println("User deleted successfully.")
	return nil
}

func StorePassword(user string, password []byte) error {
	fmt.Println("In here")
	fileName := SecretDir + "/." + user
	encodedPassword := base64.StdEncoding.EncodeToString(password)
	err := os.WriteFile(fileName, []byte(encodedPassword), 0600)
	if err != nil {
		log.Fatalf("Unable to store password for %v to %v, error: %v", user, fileName, err)
		return err
	}
	fmt.Printf("\nPassword stored for %v\n", user)
	return nil
}

func StoreToken(user string, accessToken []byte, refreshToken []byte) error {
	fmt.Println("In StoreToken")
	fileName := SecretDir + "/." + user
	//fileName := filepath.Join(SecretDir, user)
	encodedPassword := base64.StdEncoding.EncodeToString(accessToken) + "\n" + base64.StdEncoding.EncodeToString(refreshToken)
	err := ioutil.WriteFile(fileName, []byte(encodedPassword), 0644)
	if err != nil {
		log.Fatalf("Unable to store password for %v to %v, error: %v", user, fileName, err)
		return err
	}
	fmt.Printf("\nToken stored for %v\n", user)
	return nil
}
