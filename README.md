# gojsonrpc
A Go package to **parse** and **create** [JSON-RPC 2.0](https://www.jsonrpc.org/specification) requests/notifications/responses

## Installation
`go get github.com/kosmas-valianos/gojsonrpc`

## API/Usage

### Create a JSON-RPC 2.0 request/notification
Use the `NewNotification()`, `NewRequest()` respectively by passing the `method`, the `params`, and the `id` in case of request. The `params` can be `any` and if it shall be omitted then `nil` shall be passed. The `id` must be `int`, `float64` or `string`. Both functions return either a `[]bytes` slice with the raw data or an `error`.

```golang
params := struct {
	Count int      `json:"count"`
	Names []string `json:"names"`
}{
	Count: 2,
	Names: []string{"foo", "bar"},
}
jsonRPCRequestRaw, err := NewRequest("mymethod", params, 5)
if err != nil {
	fmt.Println(err)
}
jsonRPCNotificationRaw, err := NewNotification("mymethod", params)
if err != nil {
	fmt.Println(err)
}
```

### Parse a JSON-RPC 2.0 request/notification
Use the `ParseRequest()`, `ParseNotification` respectively by passing a raw `[]bytes` slice. Both functions return either a `*request`/`*notification` object or an `error`. In case of `ParseRequest()` the error is a `*jsonRPCError` object which can then be used to create a response with `NewErrorResponse()`.

```golang
jsonRPCnotification, err := ParseNotification([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]}`))
if err != nil {
	fmt.Println(err)
}

jsonRPCrequest, jsonRPCError := ParseRequest([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`))
if jsonRPCError != nil {
	// jsonRPCError implements Error() of error interface so a nice error will be printed
	fmt.Println(jsonRPCError)
}
```

### Create a JSON-RPC 2.0 response
Use the `NewResultResponse()` by passing the `id` and the `result` object to create a response with a result. The `result` can be `any` while the `id` must be `int`, `float64` or `string`. It returns a `[]bytes` slice with the raw data or an `error`.

```golang
result := struct {
	Count int      `json:"count"`
	Names []string `json:"names"`
}{
	Count: 2,
	Names: []string{"foo", "bar"},
}
jsonRPCResponseRaw, err := NewResultResponse(5, result)
if err != nil {
	fmt.Println(err)
}
```

Use the `NewErrorResponse()` similarly but instead of a `result` object use a `*jsonRPCError` object. In case the error code is not `ParseError` or `InvalidRequest`, an `id` must be passed which must be `int`, `float64` or `string`. It returns a `[]bytes` slice with the raw data or an `error`.

```golang
jsonRPCRequest, jsonRPCError := ParseRequest([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]}`))
if jsonRPCError != nil {
  jsonRPCResponseRaw, err := NewErrorResponse(nil, jsonRPCError)
  if err != nil {
    fmt.Println(err)
  }
}
```

Use the `NewJsonRPCError` by passing a `code`, a `message` and optionally a `data` object to create a custom `*jsonRPCError` object which can then be used in `NewErrorResponse()`. Note that according to the specification the `code` in case of a custom error must be between `-32099` and `-32000`. It returns a `*jsonRPCError` object or an `error`.

```golang
data := struct {
  ServerName     string `json:"server-name"`
  ServerProtocol string `json:"server-protocol"`
}{
  ServerName:     "example.com",
  ServerProtocol: "http",
}
jsonRPCError, err := NewJsonRPCError(-32000, "Database error", data)
if err != nil {
  fmt.Println(err)
}
```

### Parse a JSON-RPC 2.0 response
Use the `ParseResponse()` by passing a raw `[]bytes` slice. It returns a `*response` object or an `error`

```golang
jsonRPCResponse, err = ParseResponse(jsonRPCResponseRaw)
if err != nil {
	t.Error(err)
}
```
