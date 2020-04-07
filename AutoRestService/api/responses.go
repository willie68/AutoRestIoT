package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type SimpleResponseMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

/*
Msg writes a response with a message as json
*/
func MsgResponse(w http.ResponseWriter, code int, message string) {
	m := SimpleResponseMessage{
		Message: message,
		Code:    code,
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(buf.Bytes())
}
