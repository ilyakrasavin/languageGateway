package languageGateway

import (
	"encoding/json"
	"errors"
	"log"
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

	log.SetPrefix("SendToModerations: ")

	mr, err := buildModerationsRequest(um.M, um.MN, k)

	if err != nil {
		return false, errors.Join(errors.New("error while calling buildModerationsRequest: "), err)
	}

	// Prepares and makes a POST request to moderations API
	flag, err := moderationEndpoint(mr)

	if err != nil {

		return false, errors.Join(errors.New("error while calling moderationEndpoint: "), err)
	}

	// Reports violation to backend for further handling
	if flag {
		return true, nil
	}

	return false, nil
}

// Prepares the prompt and sends it to the model
func SendToModel(um UserMessage, sp SystemPrompt, k string) (ModelResponse, error) {

	log.SetPrefix("SendToModel: ")

	cr, err := buildCompletionsRequest(um.MN, sp, um.M, k)

	if err != nil {
		return ModelResponse{}, errors.Join(errors.New("error while calling buildCompletionsRequest: "), err)
	}

	// Submits POST to API
	response, err := chatCompletionsEndpoint(cr)
	if err != nil {
		return ModelResponse{}, errors.Join(errors.New("error while calling chatCompletionsEndpoint: "), err)
	}

	var mr ModelResponse
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &mr); err != nil {
		return ModelResponse{}, errors.Join(errors.New("error while unmarshalling into ModelReponse struct: "), err)

	}

	return mr, err

}
