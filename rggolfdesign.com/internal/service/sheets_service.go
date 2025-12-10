package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type ConsultationRequest struct {
	Name       string
	Email      string
	Phone      string
	Service    string
	Budget     string
	Dimensions string
	Message    string
}

type SheetsService struct {
	spreadsheetID string
	sheetName     string
}

func NewSheetsService(spreadsheetID, sheetName string) *SheetsService {
	return &SheetsService{
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
	}
}

func (s *SheetsService) AppendConsultationRequest(ctx context.Context, req ConsultationRequest) error {
	// Get credentials from environment variable
	credentialsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credentialsJSON == "" {
		return fmt.Errorf("GOOGLE_CREDENTIALS_JSON environment variable not set")
	}

	// Create credentials
	creds, err := google.CredentialsFromJSON(ctx, []byte(credentialsJSON), sheets.SpreadsheetsScope)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Create sheets service
	srv, err := sheets.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to create sheets service: %w", err)
	}

	// Prepare the row data
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	values := []interface{}{
		timestamp,
		req.Name,
		req.Email,
		req.Phone,
		req.Service,
		req.Budget,
		req.Dimensions,
		req.Message,
	}

	// Create the value range
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// Append to the sheet
	appendRange := s.sheetName + "!A:H"
	_, err = srv.Spreadsheets.Values.Append(s.spreadsheetID, appendRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()

	if err != nil {
		return fmt.Errorf("failed to append to sheet: %w", err)
	}

	return nil
}
