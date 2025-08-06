package model

type Status string

type JsonResponse struct {
	RequestId string `json:"request_id"`
	Status    int    `json:"status_code"`
	Messages  string `json:"messages"`
	Data      any    `json:"data"`
}

type JsonResponseTotal struct {
	RequestId string `json:"request_id"`
	Status    int    `json:"status_code"`
	Messages  string `json:"messages"`
	Total     int    `json:"total"`
	Data      any    `json:"data"`
}

type Pagination struct {
	Limit       int      `query:"limit" json:"limit"`
	Page        int      `query:"page" json:"page"`
	Search      string   `query:"search" json:"search"`
	Order       string   `query:"order" json:"order"`
	Translation string   `query:"translation" json:"translation"`
	OrderFields []string `query:"order_fields" json:"order_fields"`
}

type JsonResponsError struct {
	RequestId        string `json:"request_id"`
	StatusCode       int    `json:"status_code"`
	ErrorCode        any    `json:"error_code"`
	ErrorMessage     string `json:"error_message"`
	DeveloperMessage any    `json:"developer_message"`
}

func (p *Pagination) GetOrder() string {
	switch p.Order {
	case "DESC":
		return "DESC"
	default:
		return "ASC"
	}
}
