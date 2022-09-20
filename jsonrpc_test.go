/*  Copyright (C) 2022  Kosmas Valianos (kosmas.valianos@gmail.com)

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
	"bytes"
	"reflect"
	"testing"
)

func Test_ParseNotification(t *testing.T) {
	tests := []struct {
		name                 string
		rawBytes             []byte
		expectedNotification *notification
		wantErr              bool
	}{
		{
			name:     "Valid notification",
			rawBytes: []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]}`),
			expectedNotification: &notification{
				JsonRPC: jsonRPCProtocol,
				Method:  "subtract",
				Params:  []byte(`[42, 23]`),
			},
		},
		{
			name:     "Parse error",
			rawBytes: []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]`),
			wantErr:  true,
		},
		{
			name:     "Invalid notification - \"jsonrpc\" value wrong",
			rawBytes: []byte(`{"jsonrpc": "1.0", "method": "subtract", "params": [42, 23]}`),
			wantErr:  true,
		},
		{
			name:     "Invalid notification - \"method\" value has prefix rpc.",
			rawBytes: []byte(`{"jsonrpc": "2.0", "method": "rpc.subtract", "params": [42, 23]}`),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notification, err := ParseNotification(tt.rawBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNotification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(notification, tt.expectedNotification) {
				t.Errorf("ParseNotification() = %v, want %v", notification, tt.expectedNotification)
			}
		})
	}
}

func TestNewNotification(t *testing.T) {
	type args struct {
		method string
		params any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid parameters - with params object",
			args: args{
				method: "subtract",
				params: []int{42, 43},
			},
			want: []byte(`{"jsonrpc":"2.0","method":"subtract","params":[42,43]}` + "\n"),
		},
		{
			name: "Valid parameters - no params object",
			args: args{
				method: "subtract",
			},
			want: []byte(`{"jsonrpc":"2.0","method":"subtract"}` + "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewNotification(tt.args.method, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNotification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NewNotification() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func equalRequests(r1, r2 *request) bool {
	return (r1 == r2 || (r1.JsonRPC == r2.JsonRPC && r1.Method == r2.Method && bytes.Equal(r1.Params, r2.Params)))
}

func equalJsonRPCErrors(e1, e2 *jsonRPCError) bool {
	return (e1 == e2 || (e1.Code == e2.Code && e1.Message == e2.Message && bytes.Equal(e1.Data, e2.Data)))
}

func Test_ParseRequest(t *testing.T) {
	tests := []struct {
		name                 string
		rawBytes             []byte
		expectedRequest      *request
		expectedJsonRPCError *jsonRPCError
	}{
		{
			name:     "Valid request",
			rawBytes: []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`),
			expectedRequest: &request{
				JsonRPC: jsonRPCProtocol,
				Method:  "subtract",
				Params:  []byte(`[42, 23]`),
				ID:      1,
			},
		},
		{
			name:                 "Parse error",
			rawBytes:             []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1`),
			expectedJsonRPCError: &JsonParseError,
		},
		{
			name:                 "Invalid request - \"jsonrpc\" value wrong",
			rawBytes:             []byte(`{"jsonrpc": "1.0", "method": "subtract", "params": [42, 23], "id": 1}`),
			expectedJsonRPCError: &JsonInvalidRequest,
		},
		{
			name:                 "Invalid request - \"method\" value has prefix rpc.",
			rawBytes:             []byte(`{"jsonrpc": "2.0", "method": "rpc.subtract", "params": [42, 23], "id": 1}`),
			expectedJsonRPCError: &JsonInvalidRequest,
		},
		{
			name:                 "Invalid request - \"id\" missing",
			rawBytes:             []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]}`),
			expectedJsonRPCError: &JsonInvalidRequest,
		},
		{
			name:                 "Invalid request - \"id\" value has invalid type",
			rawBytes:             []byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": {"test": 1}}`),
			expectedJsonRPCError: &JsonInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, jsonRPCError := ParseRequest(tt.rawBytes)
			if !equalRequests(request, tt.expectedRequest) {
				t.Errorf("ParseRequest() error = %v, wantErr %v", jsonRPCError, tt.expectedJsonRPCError)
				return
			}
			if !equalJsonRPCErrors(jsonRPCError, tt.expectedJsonRPCError) {
				t.Errorf("ParseRequest() = %v, want %v", request, tt.expectedRequest)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	type args struct {
		method string
		params any
		id     string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid parameters",
			args: args{
				method: "database",
				params: struct {
					Count int      `json:"count"`
					Names []string `json:"names"`
				}{
					Count: 2,
					Names: []string{"foo", "bar"},
				},
				id: "84dca59c-d3c2-4a0b-9ec7-627e810aeab7",
			},
			want: []byte(`{"jsonrpc":"2.0","method":"database","params":{"count":2,"names":["foo","bar"]},"id":"84dca59c-d3c2-4a0b-9ec7-627e810aeab7"}` + "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequest(tt.args.method, tt.args.params, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRequest() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestNewJsonRPCError(t *testing.T) {
	type args struct {
		code    int
		message string
		data    any
	}
	tests := []struct {
		name    string
		args    args
		want    *jsonRPCError
		wantErr bool
	}{
		{
			name: "Valid parameters",
			args: args{
				code:    -32000,
				message: "Database error",
				data: struct {
					ServerName     string `json:"server-name"`
					ServerProtocol string `json:"server-protocol"`
				}{
					ServerName:     "example.com",
					ServerProtocol: "http",
				},
			},
			want: &jsonRPCError{
				Code:    -32000,
				Message: "Database error",
				Data:    []byte(`{"server-name":"example.com","server-protocol":"http"}`),
			},
		},
		{
			name: "Invalid parameters",
			args: args{
				code:    -32100,
				message: "Database error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.args.data)
			got, err := NewJsonRPCError(tt.args.code, tt.args.message, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJsonRPCError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewJsonRPCError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		id      any
		code    int
		message string
		data    any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid parameters",
			args: args{
				id:      1,
				code:    -32000,
				message: "Database error",
				data: struct {
					ServerName     string `json:"server-name"`
					ServerProtocol string `json:"server-protocol"`
				}{
					ServerName:     "example.com",
					ServerProtocol: "http",
				},
			},
			want: []byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"Database error","data":{"server-name":"example.com","server-protocol":"http"}},"id":1}` + "\n"),
		},
		{
			name: "Invalid parameters",
			args: args{
				// Trigger a nil *jsonRPCError object
				code: -32100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonError, _ := NewJsonRPCError(tt.args.code, tt.args.message, tt.args.data)
			got, err := NewErrorResponse(tt.args.id, jsonError)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewErrorResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewErrorResponse() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}

func TestNewResultResponse(t *testing.T) {
	type args struct {
		id     string
		result any
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid parameters",
			args: args{
				id: "84dca59c-d3c2-4a0b-9ec7-627e810aeab7",
				result: struct {
					Count int      `json:"count"`
					Names []string `json:"names"`
				}{
					Count: 2,
					Names: []string{"foo", "bar"},
				},
			},
			want: []byte(`{"jsonrpc":"2.0","result":{"count":2,"names":["foo","bar"]},"id":"84dca59c-d3c2-4a0b-9ec7-627e810aeab7"}` + "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResultResponse(tt.args.id, tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResultResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResultResponse() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
