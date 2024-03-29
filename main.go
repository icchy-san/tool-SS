package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

// Request a token from the web
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

// Get token from local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// Save a token you get to a local file.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Save credential information to the local file: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// Post URL to Slack #xl-db更新-medical
func postMsg2Slack(text string) error {
	jsonStr := `{"username": "Differ君", "text":"` + text + `"}`
	req, err := http.NewRequest(
		"POST",
		"",
		bytes.NewBuffer([]byte(jsonStr)),
	)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		log.Println(err)
	}

	// Upload CSV file to Google Drive
	filename := "diffFile.csv"
	createdTimeStamp := time.Now().Unix()
	TimeStampStr := strconv.FormatInt(createdTimeStamp, 10)

	uploadFileName := "DIFF_" + TimeStampStr
	baseMimeType := "text/csv"
	convertedMimeType := "application/vnd.google-apps.spreadsheet"
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	defer f.Close()

	driveFile := &drive.File{
		Name:     uploadFileName,
		MimeType: convertedMimeType,
	}

	client := getClient(config)
	srv, err := drive.New(client)

	res, err := srv.Files.Create(driveFile).Media(f, googleapi.ContentType(baseMimeType)).Do()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	permissiondata := &drive.Permission{
		Type:               "domain",
		Role:               "reader",
		Domain:             "leverages.jp",
		AllowFileDiscovery: true,
	}

	_, err = srv.Permissions.Create(res.Id, permissiondata).Do()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("%s, %s, %s\n", res.Name, res.Id, res.MimeType)
	sheetURL := fmt.Sprintf("SpreadSheet URL: https://docs.google.com/spreadsheets/d/%s\n", res.Id)
	fmt.Printf("%v", sheetURL)

	err = postMsg2Slack(sheetURL)

	if err != nil {
		log.Printf("%s", err)
	}

}
