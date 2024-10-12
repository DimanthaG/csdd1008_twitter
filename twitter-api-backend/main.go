package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dghubble/oauth1"
)

// RequestData struct handles the incoming data from the frontend
type RequestData struct {
	APIKey            string `json:"apiKey"`
	APISecretKey      string `json:"apiSecretKey"`
	AccessToken       string `json:"accessToken"`       // OAuth 1.0a Access Token
	AccessTokenSecret string `json:"accessTokenSecret"` // OAuth 1.0a Access Token Secret
	Action            string `json:"action"`
	Content           string `json:"content,omitempty"`
	TweetID           string `json:"tweetID,omitempty"`
}

func main() {
	// Serve static files from the frontend directory (React build files)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	// Handle API requests from the React frontend
	http.HandleFunc("/api/twitter", handleTwitter)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handle Twitter-related API requests
func handleTwitter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	var data RequestData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Create an OAuth1 client using user-provided API credentials and Access Token/Secret
	config := oauth1.NewConfig(data.APIKey, data.APISecretKey)
	token := oauth1.NewToken(data.AccessToken, data.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Handle the different actions: post, delete
	switch data.Action {
	case "post":
		postTweet(httpClient, w, data.Content)
	case "delete":
		deleteTweet(httpClient, w, data.TweetID)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// Post a new tweet using Twitter API v2 with OAuth 1.0a User Context
func postTweet(httpClient *http.Client, w http.ResponseWriter, tweetContent string) {
	endpoint := "https://api.twitter.com/2/tweets"
	jsonBody := fmt.Sprintf(`{"text": "%s"}`, tweetContent)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to post tweet", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

// Delete a tweet using Twitter API v2 with OAuth 1.0a User Context
func deleteTweet(httpClient *http.Client, w http.ResponseWriter, tweetID string) {
	endpoint := fmt.Sprintf("https://api.twitter.com/2/tweets/%s", tweetID)

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to delete tweet", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}
