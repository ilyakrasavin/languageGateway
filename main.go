package languageGateway

import (
	"encoding/json"
)

type UserMessage struct {
	UID string

	MN string
	M  string
}

type ModelResponse struct {
	Sender      string `json:"sender"`
	MessageType string `json:"type"`
	Content     string `json:"content"`
}

type SystemPrompt struct {
	SC  string
	SS  string
	LVL string
	UT  string

	R  string
	RF string

	MT string

	H string
}

func SendToModerations(um UserMessage, k string) (bool, error) {

	var mr moderationsRequest = buildModerationsRequest(um.M, um.MN, k)

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

	var cr completionsRequest = buildCompletionsRequest(um.MN, sp, um.M, k)

	// Submits POST to API
	response, err := chatCompletionsEndpoint(cr)

	var mr ModelResponse
	json.Unmarshal([]byte(response.Choices[0].Message.Content), &mr)

	return mr, err

}
