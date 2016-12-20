package common

import (
	"encoding/json"
)

type Message struct {
	Content  string `json:"message"`
	Username string `json:"user,omitempty"`
}

// Encodes message.
// Current encoding scheme is JSON.
func Encode(message *Message) []byte {
	b, _ := json.Marshal(message)
	return b
}

func Decode(data []byte) (*Message, error) {
	var m Message

	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
