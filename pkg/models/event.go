package models

type Event struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}
