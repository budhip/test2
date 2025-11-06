package dddnotification

const SERVICE_NAME string = "ddd-notification"

type (
	PayloadNotification struct {
		Title        string      `json:"title"`
		Service      string      `json:"service"`
		SlackChannel string      `json:"slackChannel"`
		Data         MessageData `json:"data"`
	}
	MessageData struct {
		Operation string `json:"Operation"`
		Message   string `json:"Message"`
	}
)

type ResponseSendMessage struct {
	Status  int                    `json:"status"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
