package formatter

type Body struct {
	Error    *ErrorData `json:"error,omitempty"`
	Response any        `json:"response,omitempty"`
	Data     any        `json:"data,omitempty"`
}

type ErrorData struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func FormatResponse(value any) Body {
	return Body{Response: value}
}

func FormatData(value any) Body {
	return Body{Data: value}
}

func FormatError(code int, text string) Body {
	return Body{
		Error: &ErrorData{
			Code: code,
			Text: text,
		},
	}
}
