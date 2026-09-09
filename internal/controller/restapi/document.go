package restapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/request"
	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/response"
	"github.com/maksimovyuriy/astralentry/internal/entity"
)

const multipartMemoryLimit = 8 << 20

func (c *Controller) createDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(multipartMemoryLimit); err != nil {
		c.writeError(w, r, invalidRequest("parse multipart form", err))
		return
	}
	defer r.MultipartForm.RemoveAll()

	var input request.CreateDocument
	if err := json.Unmarshal([]byte(r.FormValue("meta")), &input.Meta); err != nil {
		c.writeError(w, r, invalidRequest("decode meta", err))
		return
	}
	input.Normalize()

	ownerID, err := c.auth.Authorize(r.Context(), input.Meta.Token)
	if err != nil {
		c.writeError(w, r, err)
		return
	}

	if value := r.FormValue("json"); value != "" {
		input.JSON = json.RawMessage(value)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		c.writeError(w, r, invalidRequest("read file", err))
		return
	}
	if err == nil {
		defer file.Close()
		input.File = fileHeader
	}
	if err := input.Validate(); err != nil {
		c.writeError(w, r, invalidRequest("validate document", err))
		return
	}

	_, err = c.documents.Create(
		r.Context(),
		entity.Document{
			OwnerID: ownerID,
			Name:    input.Meta.Name,
			Mime:    input.Meta.Mime,
			File:    input.Meta.File,
			Public:  input.Meta.Public,
		},
		entity.DocumentContent{JSON: input.JSON},
		input.Meta.Grant,
		file,
	)
	if err != nil {
		c.writeError(w, r, err)
		return
	}

	result := response.CreateDocument{JSON: input.JSON}
	if input.File != nil {
		result.File = input.Meta.Name
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response.FormatData(result))
}
