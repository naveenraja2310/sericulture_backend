package model

import "time"

type SuccessResponse struct {
	StatusCode    int         `json:"statusCode"`
	StatusMessage string      `json:"statusMessage"`
	Data          interface{} `json:"data"`
}

type ErrorResponse struct {
	ApiPath      string    `json:"apiPath"`
	ErrorCode    int       `json:"errorCode"`
	ErrorMessage string    `json:"errorMessage"`
	ErrorTime    time.Time `json:"errorTime"`
}
