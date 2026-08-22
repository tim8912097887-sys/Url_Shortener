package envelope

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

func NewErrorResponse(errorBody Error) ErrorResponse {
	return ErrorResponse{State: "error", Data: nil, Error: errorBody, MetaData: MetaData{Timestamp: time.Now()}}
}