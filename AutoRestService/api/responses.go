package api

import (
	"encoding/json"
	"net/http"
)

/*
Msg writes a response with a message as json
*/
func Msg(w http.ResponseWriter, code int, message string) {
	msg, err := json.Marshal(struct {
		Message string `json:"message"`
	}{message})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(msg)
}
