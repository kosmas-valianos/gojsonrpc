/*  Copyright 2022  Kosmas Valianos (kosmas.valianos@gmail.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const jsonRPCProtocol = "2.0"

type idInterface interface {
	~int | ~float64 | ~string
}

type notification struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// ParseNotification parses a JSON-RPC notification from raw bytes.
// Returns a *notification object or an error
func ParseNotification(notificationRaw []byte) (*notification, error) {
	var notification notification
	err := json.Unmarshal(notificationRaw, &notification)
	if err != nil {
		return nil, err
	}

	if notification.JsonRPC != jsonRPCProtocol {
		return nil, errors.New("invalid notification")
	}

	if strings.HasPrefix(notification.Method, "rpc.") {
		return nil, errors.New("invalid notification")
	}

	return &notification, nil
}

// NewNotification creates a notification using the method and the params.
// Returns the raw bytes of the notification or an error
func NewNotification(method string, params any) ([]byte, error) {
	notification := notification{
		JsonRPC: "2.0",
		Method:  method,
	}

	if params != nil {
		var err error
		notification.Params, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	notificationRaw, err := json.Marshal(&notification)
	if err != nil {
		return nil, err
	}
	return append(notificationRaw, '\n'), nil
}

type request struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

// NewResultResponse creates a result response using a result object (nil for omitting).
// Returns the raw bytes of the response or an error
func (r *request) NewResultResponse(result any) ([]byte, error) {
	response := response{
		JsonRPC: jsonRPCProtocol,
		ID:      r.ID,
	}
	return marshalResultResponse(response, result)
}

// ParseRequest parses a JSON-RPC request from raw bytes.
// Returns a *request object or a *jsonRPCError error object
func ParseRequest(requestRaw []byte) (*request, *jsonRPCError) {
	jsonRPCError := &JsonParseError
	var request request
	err := json.Unmarshal(requestRaw, &request)
	if err != nil {
		return nil, jsonRPCError
	}
	jsonRPCError = &JsonInvalidRequest

	if request.JsonRPC != jsonRPCProtocol {
		return nil, jsonRPCError
	}

	if strings.HasPrefix(request.Method, "rpc.") {
		return nil, &JsonInvalidRequest
	}

	switch request.ID.(type) {
	case float64:
		// This is the type which json.Unmarshal() uses for JSON number
		return &request, nil
	case string:
		return &request, nil
	default:
		return nil, jsonRPCError
	}
}

// NewRequest creates a request using the method, the params and the id.
// Returns the raw bytes of the request or an error
func NewRequest[I idInterface](method string, params any, id I) ([]byte, error) {
	request := request{
		JsonRPC: "2.0",
		Method:  method,
		ID:      id,
	}

	if params != nil {
		var err error
		request.Params, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}
	}

	requestRaw, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}
	return append(requestRaw, '\n'), nil
}

type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Const error codes
const (
	ParseError              = -32700
	InvalidRequest          = -32600
	MethodNotFound          = -32601
	InvalidMethodParameters = -32602
	InternalError           = -32603
)

// Common error objects
var (
	JsonParseError              = jsonRPCError{Code: ParseError, Message: "Parse error"}
	JsonInvalidRequest          = jsonRPCError{Code: InvalidRequest, Message: "Invalid Request"}
	JsonMethodNotFound          = jsonRPCError{Code: MethodNotFound, Message: "Method not found"}
	JsonInvalidMethodParameters = jsonRPCError{Code: InvalidMethodParameters, Message: "Invalid method parameters"}
	JsonInternalError           = jsonRPCError{Code: InternalError, Message: "Internal error"}
)

// Error implements Error() of error interface
func (j *jsonRPCError) Error() string {
	return fmt.Sprintf("Code: %v Message: %v Data: %v", j.Code, j.Message, string(j.Data))
}

// AddData adds a data object using an existing jsonRPCError object.
// Returns a new *jsonRPCError object or an error.
// It's useful when a data object needs to be added in a common jsonRPCError object.
// e.g. jsonRPCError, _ = jsonrpc.JsonInvalidMethodParameters.AddData(message)
func (j jsonRPCError) AddData(data any) (*jsonRPCError, error) {
	jsonRPCError := jsonRPCError{
		Code:    j.Code,
		Message: j.Message,
	}

	var err error
	jsonRPCError.Data, err = json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &jsonRPCError, nil
}

// NewJsonRPCError creates a jsonRPCError.
// Returns a *jsonRPCError object or an error
func NewJsonRPCError(code int, message string, data any) (*jsonRPCError, error) {
	if code < -32099 || code > -32000 {
		return nil, errors.New("code must be between  -32099 and -32000")
	}

	jsonRPCError := jsonRPCError{
		Code:    code,
		Message: message,
	}

	var err error
	jsonRPCError.Data, err = json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &jsonRPCError, nil
}

type response struct {
	JsonRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      any             `json:"id"`
}

// ParseResponse parses a JSON-RPC request from raw bytes.
// Returns a *response object or a error
func ParseResponse(responseRaw []byte) (*response, error) {
	var response response
	err := json.Unmarshal(responseRaw, &response)
	if err != nil {
		return nil, err
	}

	if response.JsonRPC != jsonRPCProtocol {
		return nil, fmt.Errorf("jsonrpc must be exactly \"%v\"", jsonRPCProtocol)
	}

	if len(response.Result) == 0 && response.Error == nil {
		return nil, errors.New("response must have a \"result\" or an \"error\"")
	} else if len(response.Result) > 0 && response.Error != nil {
		return nil, errors.New("response must not have a \"result\" and an \"error\"")
	}

	if response.ID == nil {
		if response.Error == nil {
			return nil, errors.New("response's ID must not be null when error does not exist")
		} else if response.Error.Code != ParseError && response.Error.Code != InvalidRequest {
			return nil, fmt.Errorf("response's ID must be null only when error's code is %v or %v", ParseError, InvalidRequest)
		}
	}

	return &response, nil
}

// NewErrorResponse creates a response from a *jsonRPCError object using the id if it's applicable and not nil.
// Returns the raw bytes of the response or an error
func NewErrorResponse(id any, jsonError *jsonRPCError) ([]byte, error) {
	if jsonError == nil {
		return nil, errors.New("no JSON-RPC error passed as parameter")
	}

	response := response{
		JsonRPC: jsonRPCProtocol,
		Error:   jsonError,
	}

	if jsonError.Code != ParseError && jsonError.Code != InvalidRequest {
		if id == nil {
			return nil, errors.New("id must be present unless the error is ParseError or InvalidRequest")
		}
		switch id.(type) {
		case int:
			response.ID = id
		case float64:
			response.ID = id
		case string:
			response.ID = id
		default:
			return nil, errors.New("id must be of type int, float64 or string")
		}
	}

	responseRaw, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}
	return append(responseRaw, '\n'), nil
}

// NewResultResponse creates a response from a result object using the id.
// Returns the raw bytes of the response or an error
func NewResultResponse[I idInterface](id I, result any) ([]byte, error) {
	response := response{
		JsonRPC: jsonRPCProtocol,
		ID:      id,
	}

	return marshalResultResponse(response, result)
}

func marshalResultResponse(response response, result any) ([]byte, error) {
	var err error
	response.Result, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}

	responseRaw, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}
	return append(responseRaw, '\n'), nil
}
