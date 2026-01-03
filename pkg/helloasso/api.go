// Package helloasso provides a client for the HelloAsso API.
package helloasso

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	token        string
	formsCache   []Form
	formsCacheTS int64
)

func addRequestHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token)
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
	State                       string `json:"-"`
	Title                       string `json:"title"`
	WidgetButtonURL             string `json:"-"`
	WidgetFullURL               string `json:"-"`
	WidgetVignetteHorizontalURL string `json:"-"`
	WidgetVignetteVerticalURL   string `json:"-"`
	FormSlug                    string `json:"formSlug"`
	FormType                    string `json:"-"`
	URL                         string `json:"url"`
	OrganizationSlug            string `json:"-"`
}

func login() error {
	resp, err := http.PostForm("https://api.helloasso.com/oauth2/token", url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {os.Getenv("CLIENT_ID")},
		"client_secret": {os.Getenv("CLIENT_SECRET")},
	})
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: received non-200 response code:", resp.StatusCode)
		fmt.Println("Response body:", string(body))
		return fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}
	var loginResponse LoginResponse
	err = json.Unmarshal(body, &loginResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return err
	}
	token = loginResponse.AccessToken

	return nil
}

func GetForms() ([]Form, error) {
	now := time.Now().Unix()
	if formsCache != nil && now-formsCacheTS < 60*10 {
		return formsCache, nil
	}

	err := login()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(
		"GET",
		"https://api.helloasso.com/v5/organizations/"+os.Getenv("ORGANIZATION")+"/forms?states=Public",
		nil,
	)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	addRequestHeaders(req, token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var listResponse ListResponse[Form]
	err = json.Unmarshal(body, &listResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return nil, err
	}
	formsCache = listResponse.Data
	formsCacheTS = now

	return listResponse.Data, nil
}
