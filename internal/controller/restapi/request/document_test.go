package request

import "testing"

func TestListDocumentsParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		request    ListDocuments
		wantLimit  int
		wantOffset int
		wantErr    bool
	}{
		{name: "default", request: ListDocuments{}, wantLimit: 20},
		{name: "minimum", request: ListDocuments{Limit: "1"}, wantLimit: 1},
		{name: "maximum", request: ListDocuments{Limit: "100"}, wantLimit: 100},
		{name: "offset", request: ListDocuments{Offset: "40"}, wantLimit: 20, wantOffset: 40},
		{name: "zero offset", request: ListDocuments{Offset: "0"}, wantLimit: 20},
		{name: "zero", request: ListDocuments{Limit: "0"}, wantErr: true},
		{name: "above maximum", request: ListDocuments{Limit: "101"}, wantErr: true},
		{name: "not a number", request: ListDocuments{Limit: "many"}, wantErr: true},
		{name: "negative offset", request: ListDocuments{Offset: "-1"}, wantErr: true},
		{name: "invalid offset", request: ListDocuments{Offset: "many"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limit, offset, err := test.request.ParsePagination()
			if (err != nil) != test.wantErr {
				t.Fatalf("ParsePagination() error = %v, wantErr %v", err, test.wantErr)
			}
			if limit != test.wantLimit {
				t.Errorf("ParsePagination() limit = %d, want %d", limit, test.wantLimit)
			}
			if offset != test.wantOffset {
				t.Errorf("ParsePagination() offset = %d, want %d", offset, test.wantOffset)
			}
		})
	}
}

func TestListDocumentsValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ListDocuments
		wantErr bool
	}{
		{name: "without filter", request: ListDocuments{}},
		{name: "complete filter", request: ListDocuments{Key: "name", Value: "photo.jpg"}},
		{name: "key without value", request: ListDocuments{Key: "name"}, wantErr: true},
		{name: "value without key", request: ListDocuments{Value: "photo.jpg"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestGetDocumentValidate(t *testing.T) {
	tests := []struct {
		name    string
		request GetDocument
		wantErr bool
	}{
		{
			name:    "valid UUID",
			request: GetDocument{ID: "550e8400-e29b-41d4-a716-446655440000"},
		},
		{name: "empty ID", request: GetDocument{}, wantErr: true},
		{name: "invalid ID", request: GetDocument{ID: "document"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
