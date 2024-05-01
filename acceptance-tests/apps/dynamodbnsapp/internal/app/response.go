package app

import (
	"encoding/json"
	"net/http"
)

func NewErrResponse(err error) ErrResponse {
	return ErrResponse{err}
}

type ErrResponse struct {
	error
}

func (e ErrResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"status": e.Error()})
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
