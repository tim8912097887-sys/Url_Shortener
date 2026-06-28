package response

import "time"

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	State    string   `json:"state"`
	Data     any      `json:"data"`
	Error    Error    `json:"error"`
	MetaData MetaData `json:"metaData"`
}

func NewErrorResponse(code string, message string) ErrorResponse {
	return ErrorResponse{State: "error", Data: nil, Error: Error{Code: code, Message: message}, MetaData: MetaData{Timestamp: time.Now()}}
}