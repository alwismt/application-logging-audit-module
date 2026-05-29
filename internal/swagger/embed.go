package swagger

import _ "embed"

//go:embed openapi.yaml
var spec []byte

// Spec returns the embedded OpenAPI document.
func Spec() []byte {
	return spec
}
