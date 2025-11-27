package dto

type ErrorDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type MessageDTO struct {
	Message string `json:"message"`
}

func NewErrorDTO(err string, message string) *ErrorDTO {
	return &ErrorDTO{
		Error:   err,
		Message: message,
	}
}

func NewMessageDTO(message string) *MessageDTO {
	return &MessageDTO{
		Message: message,
	}
}
