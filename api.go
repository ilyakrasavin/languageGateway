package languageGateway

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
	k         string
	ModelName string
	Input     string
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
		Logprobs     interface{} `json:"logprobs"` // Uses interface{} to handle null values
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

// m - message
// mm - Moderations model name
// k - API key
func buildModerationsRequest(m string, mn string, k string) (moderationsRequest, error) {

	log.SetPrefix("buildModerationsRequest: ")

	if m == "" || mn == "" || k == "" {
		return moderationsRequest{}, errors.New("one or more argument values are missing, check input")
	}

	return moderationsRequest{
		k:         k,
		ModelName: mn,
		Input:     m,
	}, nil

}

// k - API Key
// mn - Completions model name
// sp - System prompt struct
// um - User message (for existing conversation)
// mt - max number of tokens for response
func buildCompletionsRequest(mn string, sp SystemPrompt, um string, k string) (completionsRequest, error) {

	log.SetPrefix("buildCompletionsRequest: ")

	if mn == "" || k == "" {
		return completionsRequest{}, errors.New("one or more argument values are missing, check input")
	}

	var m []message

	if strings.Compare(um, "") == 0 {

		m = []message{
			{Role: "system", Content: sp.MT + sp.UT + sp.R + sp.RF},
		}

	} else {

		m = []message{
			{Role: "system", Content: sp.MT + sp.UT + sp.R + sp.RF},
			{Role: "user", Content: "History: " + sp.H},
			{Role: "user", Content: "New Message: " + um},
		}

	}

	return completionsRequest{
		Model:    mn,
		Messages: m,
		k:        k,
	}, nil

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

	log.SetPrefix("moderationEndpoint: ")

	requestBody, err := json.Marshal(map[string]string{
		"input": r.Input,
		"model": r.ModelName,
	})

	if err != nil {
		return false, errors.Join(errors.New("error while encoding request body"), err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/moderations", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, errors.Join(errors.New("error while creating request: "), err)
	}

	// Sets request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.k))

	// Performs request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.Join(errors.New("error while making a request: "), err)
	}
	defer resp.Body.Close()

	// Reads response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Join(errors.New("error while reading response body: "), err)
	}

	var response moderationsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, errors.Join(errors.New("error while decoding response body: "), err)
	}

	if len(response.Results) == 0 {
		return false, errors.New("no results returned from moderations api")
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

	log.SetPrefix("chatCompletionsEndpoint: ")

	requestBodyMap := map[string]interface{}{
		"model":    r.Model,
		"messages": r.Messages,
	}

	// Marshals the request to JSON
	requestBody, err := json.Marshal(requestBodyMap)
	if err != nil {
		return completionsResponse{}, errors.Join(errors.New("error while masrshalling requestBodyMap into JSON: "), err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return completionsResponse{}, errors.Join(errors.New("error while creating a request: "), err)

	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.k))

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return completionsResponse{}, errors.Join(errors.New("error while making a request: "), err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return completionsResponse{}, errors.Join(errors.New("error while reading response body: "), err)
	}

	// Unmarshal the response to the struct
	var chatCompletionResponse completionsResponse
	err = json.Unmarshal(body, &chatCompletionResponse)

	if err != nil {
		return completionsResponse{}, errors.Join(errors.New("error while unmarshalling response body into completionsResponse struct: "), err)
	}

	return chatCompletionResponse, nil
}
