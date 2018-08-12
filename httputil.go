package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	maxPostSize = 1e6 //1 MB
)

//ErrorRequestBodyTooLarge is returned when a requests body has an unsupported size.
var ErrorRequestBodyTooLarge = errors.New("Request body too large")

//SendError is used for sending an error message and accompanying status code.
func SendError(w http.ResponseWriter, message string, statusCode int) (int, error) {
	if message == "" {
		msg := "message must not be an empty string"
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	je := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	payload, err := json.Marshal(je)
	if err != nil {
		msg := fmt.Sprintf("Could not marshal message into payload, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	bytesWritten, err := w.Write(payload)
	if err != nil {
		msg := fmt.Sprintf("There was an error sending the response, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	return bytesWritten, nil
}

//SendSuccess sends success to the client with a message.
func SendSuccess(w http.ResponseWriter, message string) (int, error) {
	data := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		msg := fmt.Sprintf("Could not marshal data into payload, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytesWritten, err := w.Write(payload)
	if err != nil {
		msg := fmt.Sprintf("There was an error sending the response, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	log.Printf("Sent %s", string(payload))

	return bytesWritten, nil
}

//SendData sends an object encoded as JSON to the requestor
func SendData(w http.ResponseWriter, data interface{}, statusCode int) (int, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		msg := fmt.Sprintf("Could not marshal data into payload, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	bytesWritten, err := w.Write(payload)
	if err != nil {
		msg := fmt.Sprintf("There was an error sending the response, %v", err)
		log.Printf(msg)
		return 0, errors.New(msg)
	}

	log.Printf("Sent %s", string(payload))

	return bytesWritten, nil
}

//DecodeData decodes a response payload, assumes the payload is JSON, writes errors to w,
//returns the decoded payload. If error is not nil, data is not usable.
func DecodeData(data interface{}, r *http.Request) (interface{}, error) {
	if r.ContentLength > maxPostSize {
		return nil, ErrorRequestBodyTooLarge
	}

	fullPost := new(bytes.Buffer)
	limitedBody := io.LimitReader(r.Body, maxPostSize)
	_, err := io.Copy(fullPost, limitedBody)
	if err != nil {
		return nil, fmt.Errorf("Could not read body %v", err)
	}

	err = json.Unmarshal(fullPost.Bytes(), data)
	if err != nil {
		return nil, fmt.Errorf("Bad request %v", err)
	}

	return data, nil
}
