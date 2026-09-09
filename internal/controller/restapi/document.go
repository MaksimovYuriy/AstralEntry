package restapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/maksimovyuriy/astralentry/internal/controller/restapi/middleware"
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

	ownerID := middleware.UserID(r.Context())

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

func (c *Controller) listDocuments(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	input := request.ListDocuments{
		Login:  query.Get("login"),
		Key:    query.Get("key"),
		Value:  query.Get("value"),
		Limit:  query.Get("limit"),
		Offset: query.Get("offset"),
	}
	input.Normalize()

	if err := input.Validate(); err != nil {
		c.writeError(w, r, invalidRequest("validate document list", err))
		return
	}

	limit, offset, err := input.ParsePagination()
	if err != nil {
		c.writeError(w, r, invalidRequest("validate document list", err))
		return
	}

	requesterID := middleware.UserID(r.Context())

	documents, err := c.documents.List(
		r.Context(),
		requesterID,
		input.Login,
		input.Key,
		input.Value,
		limit,
		offset,
	)
	if err != nil {
		c.writeError(w, r, err)
		return
	}

	body, err := json.Marshal(response.FormatData(
		response.ListDocumentsFromEntities(documents),
	))
	if err != nil {
		c.writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

func (c *Controller) getDocument(w http.ResponseWriter, r *http.Request) {
	input := request.GetDocument{ID: r.PathValue("id")}
	if err := input.Validate(); err != nil {
		c.writeError(w, r, invalidRequest("validate document id", err))
		return
	}

	document, content, file, err := c.documents.Get(
		r.Context(),
		middleware.UserID(r.Context()),
		input.ID,
	)
	if err != nil {
		c.writeError(w, r, err)
		return
	}
	if file != nil {
		defer file.Close()
	}

	if err := writeDocumentContent(w, r, document, content, file); err != nil {
		c.writeError(w, r, err)
	}
}

func (c *Controller) deleteDocument(w http.ResponseWriter, r *http.Request) {
	input := request.DeleteDocument{ID: r.PathValue("id")}
	if err := input.Validate(); err != nil {
		c.writeError(w, r, invalidRequest("validate document id", err))
		return
	}

	if err := c.documents.Delete(
		r.Context(),
		middleware.UserID(r.Context()),
		input.ID,
	); err != nil {
		c.writeError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response.FormatResponse(
		response.DeleteDocument{input.ID: true},
	))
}

func writeDocumentContent(
	w http.ResponseWriter,
	r *http.Request,
	document entity.Document,
	content entity.DocumentContent,
	file io.ReadSeeker,
) error {
	if document.File {
		w.Header().Set("Content-Type", document.Mime)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": document.Name,
		}))
		http.ServeContent(w, r, document.Name, document.Created, file)
		return nil
	}

	body, err := json.Marshal(response.FormatData(content.JSON))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}

	return nil
}
