package google

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SheetsClient struct {
	httpClient *http.Client
}

func NewSheetsClient(_ context.Context, _ string) (*SheetsClient, error) {
	return &SheetsClient{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

func (c *SheetsClient) ReadRows(ctx context.Context, sheetURL, worksheetName string) ([][]string, error) {
	csvURL, err := buildPublicCSVURL(sheetURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, csvURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public google sheet csv: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch public google sheet csv: status %s", resp.Status)
	}

	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse csv: %w", err)
	}

	_ = worksheetName
	return rows, nil
}

func buildPublicCSVURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("sheet url is empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid google sheet url: %w", err)
	}

	if strings.Contains(strings.ToLower(u.RawQuery), "output=csv") || strings.Contains(strings.ToLower(u.RawQuery), "format=csv") {
		return raw, nil
	}

	if !strings.Contains(u.Host, "docs.google.com") {
		return "", fmt.Errorf("unsupported host: %s", u.Host)
	}

	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(pathParts) >= 3 && pathParts[0] == "spreadsheets" && pathParts[1] == "d" && pathParts[2] != "" && pathParts[2] != "e" {
		spreadsheetID := pathParts[2]
		gid := extractGID(u)
		return fmt.Sprintf(
			"https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%s",
			spreadsheetID,
			gid,
		), nil
	}

	if len(pathParts) >= 4 && pathParts[0] == "spreadsheets" && pathParts[1] == "d" && pathParts[2] == "e" && pathParts[3] != "" {
		publishedID := pathParts[3]
		gid := extractGID(u)
		return fmt.Sprintf(
			"https://docs.google.com/spreadsheets/d/e/%s/pub?output=csv&gid=%s&single=true",
			publishedID,
			gid,
		), nil
	}

	return "", fmt.Errorf("unsupported google sheet url format")
}

func extractGID(u *url.URL) string {

	if gid := strings.TrimSpace(u.Query().Get("gid")); gid != "" {
		return gid
	}

	if fragment := strings.TrimSpace(u.Fragment); fragment != "" {
		if strings.HasPrefix(fragment, "gid=") {
			return strings.TrimPrefix(fragment, "gid=")
		}
	}

	return "0"
}
