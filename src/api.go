package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"errors"
)

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type moderationsRequest struct {
	k          string
	model_name string
	input      string
}

type moderationsResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Categories struct {
			Hate            bool `json:"hate"`
			HateThreatening bool `json:"hate/threatening"`
			SelfHarm        bool `json:"self-harm"`
			Sexual          bool `json:"sexual"`
			SexualMinors    bool `json:"sexual/minors"`
			Violence        bool `json:"violence"`
			ViolenceGraphic bool `json:"violence/graphic"`
		} `json:"categories"`
		CategoryScores struct {
			Hate            float64 `json:"hate"`
			HateThreatening float64 `json:"hate/threatening"`
			SelfHarm        float64 `json:"self-harm"`
			Sexual          float64 `json:"sexual"`
			SexualMinors    float64 `json:"sexual/minors"`
			Violence        float64 `json:"violence"`
			ViolenceGraphic float64 `json:"violence/graphic"`
		} `json:"category_scores"`
		Flagged bool `json:"flagged"`
	} `json:"results"`
}

type completionsRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	k        string
}

type completionsResponse struct {
	Choices []struct {
		FinishReason string      `json:"finish_reason"`
		Index        int         `json:"index"`
		Message      message     `json:"message"`
		Logprobs     interface{} `json:"logprobs"` // Use interface{} to handle null values
	} `json:"choices"`
	Created int64  `json:"created"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// // Checks OpenAI Completions endpoint for reachability
// func StatusCheck() (int, err) {

// 	return 200, nil
// }

// m - message
// mm - Moderations model name
// k - API key
func buildModerationsRequest(m string, mn string, k string) moderationsRequest {

	return moderationsRequest{
		k:          k,
		model_name: mn,
		input:      m,
	}

}

// k - API Key
// mn - Completions model name
// sp - System prompt struct
// um - User message (for existing conversation)
// mt - max number of tokens for response
func buildCompletionsRequest(mn string, sp SystemPrompt, um string, k string) completionsRequest {

	// Convert System Prompt & User Message into structs.

	// Supplies sytem prompt only on activity initialization
	// Supplies System prompt, Message history, New user message for existing activity

	var m []message

	if strings.Compare(um, "") == 0 {

		m = []message{
			{Role: "system", Content: sp.mt + sp.ut + sp.r + sp.rf},
		}

	} else {

		m = []message{
			{Role: "system", Content: sp.mt + sp.ut + sp.r + sp.rf},
			{Role: "user", Content: "History: " + sp.h},
			{Role: "user", Content: "New Message: " + um},
		}

	}

	return completionsRequest{
		Model:    mn,
		Messages: m,
		k:        k,
	}

}

// POST request to Moderations endpoint

// The following naked curl request is converted using stdlib's net/http
// request := `curl https://api.openai.com/v1/moderations \
// 	-X POST \
// 	-H "Content-Type: application/json" \
// 	-H "Authorization: Bearer ` + r.API_KEY + ` \
// 	-d '{
// 		"input":"` + r.input + `"
// 		"model":"` + r.model_name + `"
// 		}'`

func moderationEndpoint(r moderationsRequest) (bool, error) {

	requestBody, err := json.Marshal(map[string]string{
		"input": r.input,
		"model": r.model_name,
	})

	if err != nil {
		log.Fatalf("Error encoding request body: %v", err)
		return true, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/moderations", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return true, err
	}

	// Sets request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.k))

	// Performs request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
		return true, err
	}
	defer resp.Body.Close()

	// Reads response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
		return true, err
	}

	var response moderationsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error decoding response body: %v", err)
		return true, err
	}

	if len(response.Results) == 0 {
		return true, errors.New("no result returned by moderations api")
	}

	return response.Results[len(response.Results)-1].Flagged, nil

}

// Issues POST request to Chat Completions Endpoint
// Naked curl request of the following format is converted using stdlib's net/http

// curl https://api.openai.com/v1/chat/completions \
//   -H "Content-Type: application/json" \
//   -H "Authorization: Bearer $OPENAI_API_KEY" \
//   -d '{
//     "model": "gpt-4o-mini",
//     "messages": [
//       {
//         "role": "system",
//         "content": "You are a helpful assistant."
//       },
//       {
//         "role": "user",
//         "content": "Who won the world series in 2020?"
//       },
//       {
//         "role": "assistant",
//         "content": "The Los Angeles Dodgers won the World Series in 2020."
//       },
//       {
//         "role": "user",
//         "content": "Where was it played?"
//       }
//     ]
//   }'

func chatCompletionsEndpoint(r completionsRequest) (completionsResponse, error) {

	requestBodyMap := map[string]interface{}{
		"model":    r.Model,
		"messages": r.Messages,
	}

	// Marshal the request to JSON
	requestBody, err := json.Marshal(requestBodyMap)
	if err != nil {
		log.Fatalf("Error encoding request body: %v", err)
	}

	fmt.Println("Request to Completions looks like: ", r)

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.k))

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Unmarshal the response to the struct
	var chatCompletionResponse completionsResponse
	err = json.Unmarshal(body, &chatCompletionResponse)

	if err != nil {
		log.Fatalf("Error decoding response body: %v", err)

		if err != nil {
			log.Fatalf("Error decoding response body: %v", err)
		}

	}

	return chatCompletionResponse, nil
}
