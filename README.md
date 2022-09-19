# gojsonrpc
A Go package to **parse** and **create** [JSON-RPC 2.0](https://www.haproxy.org/download/2.7/doc/proxy-protocol.txt) requests/notifications and send JSON-RPC 2.0 responses.

## Installation
`go get github.com/kosmas-valianos/gojsonrpc`

## API/Usage

### Create a JSON-RPC 2.0 request/notification
Use the `NewNotification()`, `NewRequest()` respectively by passing the `method`, the `params`, and the `id` in case of request. The `params` can be `any` and if it shall be omitted then `nil` shall be passed. The `id` must be `int`, `float64` or `string`. Both functions return either a `[]bytes` slice with the raw data or an `error`.

```
params := struct {
	Count int      `json:"count"`
	Names []string `json:"names"`
}{
	Count: 2,
	Names: []string{"foo", "bar"},
}
rawRequest, err := NewRequest("mymethod", params, 5)
if err != nil {
	fmt.Println(err)
}
rawNotification, err := NewNotification("mymethod", params)
if err != nil {
	fmt.Println(err)
}
```

### Parse a JSON-RPC 2.0 request/notification
Use the `ParseRequest()`, `ParseNotification` respectively by passing a raw `[]bytes` slice. Both functions return either a `*request`/`*notification` object or an `error`. In case of `ParseRequest()` the error is a `*jsonRPCError` object which can then be used to create a response with `NewErrorResponse()`.

```
notification, err := ParseNotification([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23]}`))
if err != nil {
	fmt.Println(err)
}

request, jsonRPCError := ParseRequest([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`))
if jsonRPCError != nil {
  // jsonRPCError implements Error() of error interface so a nice error will be printed
	fmt.Println(jsonRPCError)
}
```

### Create a JSON-RPC 2.0 response
Use the `NewResultResponse()` by passing the `id` and the `result` object to create a response with a result. The `result` can be `any` while the `id` must be `int`, `float64` or `string`. It returns a `[]bytes` slice with the raw data or an `error`.

```
result := struct {
	Count int      `json:"count"`
	Names []string `json:"names"`
}{
	Count: 2,
	Names: []string{"foo", "bar"},
}
rawBytes, err := NewResultResponse(5, result)
if err != nil {
	fmt.Println(err)
}
```

Use the `NewErrorResponse()` similarly but instead of a `result` object use a `*jsonRPCError` object. It can be given one from `ParseRequest()` in case there was a failure there, or a custom one can be created with `NewJsonRPCError()`. It returns a `[]bytes` slice with the raw data or an `error`.

```
request, jsonRPCError := ParseRequest([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}`))
if jsonRPCError != nil {
  rawBytes, err := NewErrorResponse(5, jsonRPCError)
  if err != nil {
    fmt.Println(err)
  }
}
```

Use the `NewJsonRPCError` by passing a `code`, a `message` and optionally a `data` object to create a custom `*jsonRPCError` object which can then be used in `NewErrorResponse()`. Note that according to the specification the `code` in case of custom error must be between `-32099` and `-32000`. It returns a `*jsonRPCError` object or an `error`.

```
data := struct {
  ServerName     string `json:"server-name"`
  ServerProtocol string `json:"server-protocol"`
}{
  ServerName:     "example.com",
  ServerProtocol: "http",
}
jsonError, err := NewJsonRPCError(-32000, "Database error", data)
if err != nil {
  fmt.Println(err)
}
```
