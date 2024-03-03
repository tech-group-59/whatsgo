package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"path/filepath"
)

// https://developers.google.com/sheets/api/quickstart/go

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	fmt.Print("Authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Errorf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Errorf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// CloudTracker is a struct for tracking messages in Google Cloud
type CloudTracker struct {
	driveService  *drive.Service
	sheetsService *sheets.Service
	folderID      string
}

func (tracker *CloudTracker) Init(config *Config) error {
	ctx := context.Background()

	// Read the credentials from the file
	b, err := os.ReadFile(config.GoogleCloud.CredentialsFile)
	if err != nil {
		log.Errorf("Unable to read client secret file: %v", err)
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	gConfig, err := google.ConfigFromJSON(b, drive.DriveScope, sheets.SpreadsheetsScope)
	if err != nil {
		log.Errorf("Unable to parse client secret file to config: %v", err)
		return err
	}

	client := getClient(gConfig)

	// Create a new drive service
	driveService, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Errorf("Unable to retrieve Drive client: %v", err)
		return err
	}
	tracker.driveService = driveService
	tracker.folderID = config.GoogleCloud.FolderID

	// Create a new sheets service
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Errorf("Unable to retrieve Sheets client: %v", err)
		return err
	}
	tracker.sheetsService = sheetsService

	return nil
}

func (tracker *CloudTracker) StoreMessageWithFiles(messageID string, sender string, chat string, content string, timestamp string, files []string) error {
	// Store all files into a Google Drive folder
	fileLinks := make([]string, len(files))
	for i, filePath := range files {
		link, err := tracker.storeFile(filePath)
		if err != nil {
			return err
		}
		fileLinks[i] = link
	}

	// Get or create the spreadsheet for the chat
	spreadsheet, err := tracker.getOrCreateSpreadsheet(chat)
	if err != nil {
		return err
	}

	// Insert the data about the message
	fileLinksInterface := make([]interface{}, len(fileLinks))
	for i, v := range fileLinks {
		fileLinksInterface[i] = v
	}
	values := append([]interface{}{timestamp, sender, content}, fileLinksInterface...)
	err = tracker.insertRow(spreadsheet, values)
	if err != nil {
		return err
	}

	return nil
}

// StoreMessage stores a message in Google Cloud
func (tracker *CloudTracker) StoreMessage(messageID string, sender string, chat string, content string, timestamp string) error {
	// Get or create the spreadsheet for the chat
	spreadsheet, err := tracker.getOrCreateSpreadsheet(chat)
	if err != nil {
		return err
	}

	// Insert the data about the message
	err = tracker.insertRow(spreadsheet, []interface{}{timestamp, sender, content})
	if err != nil {
		return err
	}

	return nil
}

// StoreFile stores a file in Google Cloud

func (tracker *CloudTracker) storeFile(filePath string) (string, error) {
	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Unable to open file: %v", err)
		return "", err
	}
	defer f.Close()

	// Create a new file on Google Drive
	file, err := tracker.driveService.Files.Create(&drive.File{
		Name:     filepath.Base(filePath),
		MimeType: "application/octet-stream",
		Parents:  []string{tracker.folderID},
	}).Media(f).Do()
	if err != nil {
		log.Errorf("Unable to create file on Drive: %v", err)
		return "", err
	}

	// Return the link to the file
	return "https://drive.google.com/open?id=" + file.Id, nil
}
func (tracker *CloudTracker) getOrCreateSpreadsheet(chat string) (*sheets.Spreadsheet, error) {
	// Check if a spreadsheet exists for the chat inside the specified folder
	searchResult, err := tracker.driveService.Files.List().Q(fmt.Sprintf("name='%s' and '%s' in parents", chat, tracker.folderID)).Do()
	if err != nil {
		log.Errorf("Unable to search for file: %v", err)
		return nil, err
	}

	if len(searchResult.Files) != 0 {
		// If the spreadsheet exists, get it
		spreadsheet, err := tracker.sheetsService.Spreadsheets.Get(searchResult.Files[0].Id).Do()
		if err != nil {
			log.Errorf("Unable to get spreadsheet: %v", err)
			return nil, err
		}
		return spreadsheet, nil
	}
	// If the spreadsheet does not exist, create a new one
	spreadsheet, err := tracker.sheetsService.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: chat,
		},
	}).Do()
	if err != nil {
		log.Errorf("Unable to create spreadsheet: %v", err)
		return nil, err
	}

	// Move the spreadsheet to the specified folder
	_, err = tracker.driveService.Files.Update(spreadsheet.SpreadsheetId, &drive.File{}).AddParents(tracker.folderID).Do()
	if err != nil {
		log.Errorf("Unable to move spreadsheet to folder: %v", err)
		return nil, err
	}

	return spreadsheet, nil
}

func (tracker *CloudTracker) insertRow(spreadsheet *sheets.Spreadsheet, values []interface{}) error {
	// Insert a new row at the top of the document
	_, err := tracker.sheetsService.Spreadsheets.Values.Append(spreadsheet.SpreadsheetId, "A1", &sheets.ValueRange{
		Values: [][]interface{}{values},
	}).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		log.Errorf("Unable to insert data: %v", err)
		return err
	}

	return nil
}
