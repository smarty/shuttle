package outputs

// Addition represents the http response, which by default will be serialized to JSON.
type Addition struct {
	A int `json:"a"`
	B int `json:"b"`
	C int `json:"c"`
}
