// Package helloasso provides a client for the HelloAsso API.
package helloasso

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Organization string
	HTTPClient   *http.Client
	token        string
	tokenExpiry  time.Time
	formsCache   []Form
	formsCacheTS int64
}

func NewClient() *Client {
	return &Client{
		BaseURL:      "https://api.helloasso.com",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Organization: os.Getenv("ORGANIZATION"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) login() error {
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	resp, err := c.HTTPClient.PostForm(c.BaseURL+"/oauth2/token", url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var res LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	c.token = res.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
	return nil
}

func (c *Client) GetForms() ([]Form, error) {
	now := time.Now().Unix()
	if c.formsCache != nil && now-c.formsCacheTS < 60*10 {
		return c.formsCache, nil
	}

	if err := c.login(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v5/organizations/%s/forms?states=Public", c.BaseURL, c.Organization), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get forms with status %d: %s", resp.StatusCode, string(body))
	}

	var res ListResponse[Form]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	c.formsCache = res.Data
	c.formsCacheTS = now
	return res.Data, nil
}

// Global default client for compatibility
var defaultClient = NewClient()

// SetDefaultClient sets the global default client (used for testing)
func SetDefaultClient(c *Client) {
	defaultClient = c
}

func GetForms() ([]Form, error) {
	return defaultClient.GetForms()
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type ListResponse[T any] struct {
	Data       []T `json:"data"`
	Pagination struct {
		PageSize          int    `json:"pageSize"`
		TotalCount        int    `json:"totalCount"`
		PageIndex         int    `json:"pageIndex"`
		TotalPages        int    `json:"totalPages"`
		ContinuationToken string `json:"continuationToken"`
	} `json:"pagination"`
}

type Form struct {
	Banner struct {
		FileName  string `json:"fileName"`
		PublicURL string `json:"publicUrl"`
	} `json:"banner"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Meta        struct {
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"meta"`
	Title    string `json:"title"`
	FormSlug string `json:"formSlug"`
	URL      string `json:"url"`
}
