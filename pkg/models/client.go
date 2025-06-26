package models

type Client struct {
	SDKKey   string
	AppID    string
	Messages chan Event
}
