package credentials

import "strings"

// Supported material_kind values for credential_secret_versions.
const (
	MaterialAPIKey                     = "api_key"
	MaterialBearerToken                = "bearer_token"
	MaterialGitHubPAT                  = "github_pat"
	MaterialOAuthClient                = "oauth_client"
	MaterialOAuthTokens                = "oauth_tokens"
	MaterialOAuthAuthorizationPending  = "oauth_authorization_pending"
	MaterialTOTPSeed                   = "totp_seed"
	MaterialRecoveryCodeHashes         = "recovery_code_hashes"
)

// KnownMaterialKinds is the allowlist for create/rotate.
var KnownMaterialKinds = map[string]struct{}{
	MaterialAPIKey:                    {},
	MaterialBearerToken:               {},
	MaterialGitHubPAT:                 {},
	MaterialOAuthClient:               {},
	MaterialOAuthTokens:               {},
	MaterialOAuthAuthorizationPending: {},
	MaterialTOTPSeed:                  {},
	MaterialRecoveryCodeHashes:        {},
}

// ValidMaterialKind reports whether kind is allowed.
func ValidMaterialKind(kind string) bool {
	_, ok := KnownMaterialKinds[strings.TrimSpace(kind)]
	return ok
}
