package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"os"
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
	fmt.Printf("Go to the following link in your browser then type the " +  "authorization code: \n%v\n", authURL)

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


func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Printf("hogehoge")
	}

	client := getClient(config)

	srv, err := sheets.New(client)

	ssID := ""
	writeRange := "シート2!A2:E"
	values := [][]interface{}{{"hogehoge", "fugafuga"},{"hogehoge"}}

	data := []*sheets.ValueRange{
		{
			Range: writeRange,
			Values: values,
		},
	}

	updateValueReq := sheets.BatchUpdateValuesRequest{
		Data: data,
		ValueInputOption: "RAW",
	}

	resp, err := srv.Spreadsheets.Values.BatchUpdate(ssID, &updateValueReq).Do()
	log.Printf("%v,\n %v", resp, err)
	//if len(resp.Values) == 0 {
	//	fmt.Println("No data found.")
	//} else {
	//	fmt.Println("Name, Major:")
	//	for _, row := range resp.Values {
	//		// Print columns A and E, which correspond to indices 0 and 4.
	//		fmt.Printf("%s\n", row[0])
	//	}
	//}
}
