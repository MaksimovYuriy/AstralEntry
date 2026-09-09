package request

import (
	"encoding/json"
	"mime/multipart"
)

type CreateDocument struct {
	Meta CreateDocumentMeta
	JSON json.RawMessage
	File *multipart.FileHeader
}

type CreateDocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}
