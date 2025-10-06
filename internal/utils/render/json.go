package render

import (
	"encoding/json"
	"net/http"
)

type container struct {
	Meta any `json:"meta,omitempty"`
	Data any `json:"data"`
}

type Pagination struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
	Q      string `json:"q,omitempty"`
}

func JSON(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(&container{
		Data: v,
	})
}

func JSONWithMeta(w http.ResponseWriter, data, meta any) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(&container{
		Data: data,
		Meta: meta,
	})
}
