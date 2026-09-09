package response

import "encoding/json"

type CreateDocument struct {
	JSON json.RawMessage `json:"json,omitempty"`
	File string          `json:"file,omitempty"`
}
