package languageGateway

import (
	"encoding/json"
)

type UserMessage struct {
	uid string

	mn string
	m  string
}

type ModelResponse struct {
	Sender      string `json:"sender"`
	MessageType string `json:"type"`
	Content     string `json:"content"`
}

type SystemPrompt struct {
	s  string
	ss string
	l  string
	ut string

	r  string
	rf string

	mt string

	h string
}

func SendToModerations(um UserMessage, k string) (bool, error) {

	var mr moderationsRequest = buildModerationsRequest(um.m, um.mn, k)

	// Prepares and makes a POST request to moderations API
	flag, err := moderationEndpoint(mr)

	if err != nil {

		return true, err
	}

	// Reports violation to backend for further handling
	if flag {

		return true, nil
	}

	// No errors, message not flagged by moderations.

	return false, nil
}

// Prepares the prompt and sends it to the model
func SendToModel(um UserMessage, sp SystemPrompt, k string) (ModelResponse, error) {

	var cr completionsRequest = buildCompletionsRequest(um.mn, sp, um.m, k)

	// Submits POST to API
	response, err := chatCompletionsEndpoint(cr)

	var mr ModelResponse
	json.Unmarshal([]byte(response.Choices[0].Message.Content), &mr)

	return mr, err

}
