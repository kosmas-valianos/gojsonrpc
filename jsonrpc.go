/*
 * gojsonrpc is Go package to parse and create JSON-RPC 2.0 requests/notifications and send JSON-RPC 2.0 responses
 * Copyright (C) 2022  Kosmas Valianos (kosmas.valianos@gmail.com)
 *
 * The gojsonrpc package is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The gojsonrpc package is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const jsonRPCProtocol = "2.0"

type notification struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// ParseNotification parses a JSON-RPC notification from raw bytes
// Returns a notification object or an error
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

// NewNotification creates a notification using the method and the params
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
	return json.Marshal(notification)
}

type request struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

// ParseRequest parses a JSON-RPC request from raw bytes
// Returns a request object or a jsonRPCError error object
func ParseRequest(requestRaw []byte) (*request, *jsonRPCError) {
	var request request
	err := json.Unmarshal(requestRaw, &request)
	if err != nil {
		return nil, &JsonParseError
	}

	if request.JsonRPC != jsonRPCProtocol {
		return nil, &JsonInvalidRequest
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
		return nil, &JsonInvalidRequest
	}
}

// NewRequest creates a request using the method, the id and the params
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
	return json.Marshal(request)
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
	JsonInvalidRequest          = jsonRPCError{Code: ParseError, Message: "Invalid Request"}
	JsonMethodNotFound          = jsonRPCError{Code: ParseError, Message: "Method not found"}
	JsonInvalidMethodParameters = jsonRPCError{Code: ParseError, Message: "Invalid method parameters"}
)

// Error implements Error() of error interface
func (j *jsonRPCError) Error() string {
	return fmt.Sprintf("Code: %v Message: %v Data: %v", j.Code, j.Message, string(j.Data))
}

// NewJsonRPCError creates a jsonRPCError
// Returns a jsonRPCError object or an error
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
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
	ID      any             `json:"id,omitempty"`
}

// NewErrorResponse creates a response from a jsonRPCError object using the id if applicable
// Returns the raw bytes of the response or an error
func NewErrorResponse(id any, jsonError *jsonRPCError) ([]byte, error) {
	if jsonError == nil {
		return nil, errors.New("no JSON-RPC error passed as parameter")
	}

	response := response{
		Jsonrpc: jsonRPCProtocol,
		Error:   jsonError,
	}

	if jsonError.Code != ParseError && jsonError.Code != InvalidRequest {
		response.ID = id
	}

	responseRaw, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}
	return responseRaw, nil
}

type idInterface interface {
	~int | ~float64 | ~string
}

// NewResultResponse creates a response from a result object using the id
// Returns the raw bytes of the response or an error
func NewResultResponse[I idInterface](id I, result any) ([]byte, error) {
	response := response{
		Jsonrpc: jsonRPCProtocol,
		ID:      id,
	}

	var err error
	response.Result, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}

	responseRaw, err := json.Marshal(&response)
	if err != nil {
		return nil, err
	}
	return responseRaw, nil
}
