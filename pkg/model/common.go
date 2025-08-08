package model

type Status string

type JsonResponse struct {
	RequestId  string `json:"request_id"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"messages"`
	Data       any    `json:"data"`
}

type JsonResponseTotal struct {
	RequestId  string `json:"request_id"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Total      int    `json:"total"`
	Data       any    `json:"data"`
}

type JsonResponsError struct {
	RequestId        string `json:"request_id"`
	StatusCode       int    `json:"status_code"`
	ErrorMessage     string `json:"error_message"`
	DeveloperMessage any    `json:"developer_message"`
}
