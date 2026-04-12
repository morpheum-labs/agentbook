package httpapi

import _ "embed"

//go:embed static/SKILL.md
var EmbeddedSkill []byte

//go:embed static/openapi.json
var embeddedOpenAPISpec []byte
