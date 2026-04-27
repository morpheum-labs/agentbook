package api

import "embed"

//go:embed openapi.json
var openapiRoot embed.FS

func readEmbeddedOpenapi() ([]byte, error) {
	return openapiRoot.ReadFile("openapi.json")
}
