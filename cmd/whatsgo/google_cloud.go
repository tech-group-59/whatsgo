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
	"strings"
)

// https://developers.google.com/sheets/api/quickstart/go

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *GoogleCloudConfig, gConfig *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := config.TokenFile
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(gConfig)
		saveToken(tokFile, tok)
	}
	return gConfig.Client(context.Background(), tok)
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

	client := getClient(&config.GoogleCloud, gConfig)

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

func (tracker *CloudTracker) getOrCreateFolder(parentFolderId string, path string) (string, error) {
	// Split the path into folders
	folders := strings.Split(path, "/")

	// Create the root folder if it does not exist
	if parentFolderId == "" {
		rootFolder, err := tracker.driveService.Files.Get("root").Do()
		if err != nil {
			log.Errorf("Unable to get root folder: %v", err)
			return "", err
		}
		parentFolderId = rootFolder.Id
	}

	// Create the folders
	folderId := parentFolderId
	for _, folder := range folders {
		// Search for the folder
		searchResult, err := tracker.driveService.Files.List().Q(fmt.Sprintf("name='%s' and '%s' in parents", folder, folderId)).Do()
		if err != nil {
			log.Errorf("Unable to search for folder: %v", err)
			return "", err
		}
		if len(searchResult.Files) == 0 {
			// Create the folder
			newFolder, err := tracker.driveService.Files.Create(&drive.File{
				Name:     folder,
				Parents:  []string{folderId},
				MimeType: "application/vnd.google-apps.folder",
			}).Do()
			if err != nil {
				log.Errorf("Unable to create folder: %v", err)
				return "", err
			}
			folderId = newFolder.Id
		} else {
			folderId = searchResult.Files[0].Id
		}
	}

	return folderId, nil
}

func (tracker *CloudTracker) TrackMessage(message *TrackableMessage) error {
	path := fmt.Sprintf("%s/%s", message.Metadata.Folder, message.Metadata.Date)
	folderId, err := tracker.getOrCreateFolder(tracker.folderID, path)
	if err != nil {
		return err
	}

	// Store all files into a Google Drive folder
	fileLinks := make([]string, len(message.Files))
	for i, filePath := range message.Files {
		link, err := tracker.storeFile(filePath, folderId)
		if err != nil {
			return err
		}
		fileLinks[i] = link
	}

	// Get or create the spreadsheet for the chat
	spreadsheet, err := tracker.getOrCreateSpreadsheet(message.Chat, folderId)
	if err != nil {
		return err
	}

	// Insert the data about the message
	fileLinksInterface := make([]interface{}, len(fileLinks)*2)
	for i, v := range fileLinks {
		fileLinksInterface[i] = fmt.Sprintf("=IMAGE(\"%s\")", v)
		fileLinksInterface[i+1] = v
	}
	values := append([]interface{}{
		message.MessageID, message.Timestamp, message.Sender, message.Content, message.ParsedContent,
	}, fileLinksInterface...)
	err = tracker.insertRow(spreadsheet, values)
	if err != nil {
		return err
	}

	return nil
}

// StoreFile stores a file in Google Cloud

func (tracker *CloudTracker) storeFile(filePath string, folderId string) (string, error) {
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
		Parents:  []string{folderId},
	}).Media(f).Do()
	if err != nil {
		log.Errorf("Unable to create file on Drive: %v", err)
		return "", err
	}

	// Return the link to the file
	//return "https://drive.google.com/open?id=" + file.Id, nil

	// Share the file
	_, err = tracker.driveService.Permissions.Create(file.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		log.Errorf("Unable to share file: %v", err)
		return "", err

	}
	return "https://drive.google.com/uc?id=" + file.Id, nil
}

func (tracker *CloudTracker) getOrCreateSpreadsheet(chat string, folderId string) (*sheets.Spreadsheet, error) {
	// Check if a spreadsheet exists for the chat inside the specified folder
	searchResult, err := tracker.driveService.Files.List().Q(fmt.Sprintf("name='%s' and '%s' in parents", chat, folderId)).Do()
	if err != nil {
		log.Errorf("Unable to search for file: %v", err)
		return nil, err
	}
	log.Infof("Search result: %v", searchResult)

	if len(searchResult.Files) != 0 {
		// If the spreadsheet exists, get it
		spreadsheet, err := tracker.sheetsService.Spreadsheets.Get(searchResult.Files[0].Id).Do()
		if err != nil {
			log.Errorf("Unable to get spreadsheet: %v", err)
			return nil, err
		}
		log.Infof("Found existing spreadsheet for chat %s", chat)
		return spreadsheet, nil
	}
	log.Infof("Creating new spreadsheet for chat %s", chat)
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
	_, err = tracker.driveService.Files.Update(spreadsheet.SpreadsheetId, &drive.File{}).AddParents(folderId).Do()
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
	log.Infof("Inserted row into spreadsheet %s", spreadsheet.SpreadsheetId)

	return nil
}
