package response

type Body struct {
	Error    *ErrorData `json:"error,omitempty"`
	Response any        `json:"response,omitempty"`
	Data     any        `json:"data,omitempty"`
}

type ErrorData struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}
