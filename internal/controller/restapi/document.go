package restapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/request"
)

const multipartMemoryLimit = 8 << 20

func (c *Controller) createDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(multipartMemoryLimit); err != nil {
		c.writeError(w, errInvalidRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	var input request.CreateDocument
	if err := json.Unmarshal([]byte(r.FormValue("meta")), &input.Meta); err != nil {
		c.writeError(w, errInvalidRequest)
		return
	}

	ownerID, err := c.auth.Authorize(r.Context(), input.Meta.Token)
	if err != nil {
		c.writeError(w, err)
		return
	}

	if value := r.FormValue("json"); value != "" {
		input.JSON = json.RawMessage(value)
		if !json.Valid(input.JSON) {
			c.writeError(w, errInvalidRequest)
			return
		}
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		c.writeError(w, errInvalidRequest)
		return
	}
	if err == nil {
		defer file.Close()
		input.File = fileHeader
	}

	_ = ownerID
	_ = input
	_ = file
	c.writeError(w, errNotImplemented)
}
