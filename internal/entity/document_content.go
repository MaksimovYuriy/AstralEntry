package entity

import "encoding/json"

type DocumentContent struct {
	DocumentID string
	JSON       json.RawMessage
	FilePath   string
}
