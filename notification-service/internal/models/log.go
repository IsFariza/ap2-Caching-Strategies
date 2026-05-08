package models

type NotificationLog struct {
	Time    string      `json:"time"`
	Subject string      `json:"subject"`
	Event   interface{} `json:"event"`
}
