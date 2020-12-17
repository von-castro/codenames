package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// encodeAndSendError takes the http status code and a message to be sent to a
//client. It will encode that message in a http response body and write the
//given status code to the header
func encodeAndSendError(writer http.ResponseWriter, request *http.Request, status int, message string) {
	http.Error(writer, http.StatusText(status), status)
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(message)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println("Internal Server error encoding client error message", err)
		return
	}
	log.Println(message)
	return
}
