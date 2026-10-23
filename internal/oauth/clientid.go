package oauth

// clientIDTBD marks a tenant whose "Qdrant Cloud CLI" Auth0 application does not exist yet.
const clientIDTBD = "TBD"

// cliClientIDs are the client IDs of the "Qdrant Cloud CLI" native application, one per Auth0
// tenant, keyed by issuer. Each tenant generates its own, and they are public (a native app has
// no secret), so they ship with the CLI and users never pass one.
var cliClientIDs = map[string]string{
	"https://login.cloud.qdrant.io/":             clientIDTBD,
	"https://login.staging-cloud.qdrant.io/":     clientIDTBD,
	"https://login.development-cloud.qdrant.io/": "CjeXCZKv1plzIF9h3vZBGAhM6MLE68fI",
}

// DefaultClientID returns the built-in client ID for issuer, if that tenant has one.
func DefaultClientID(issuer string) (string, bool) {
	id, ok := cliClientIDs[canonicalizeIssuer(issuer)]
	if !ok || id == clientIDTBD {
		return "", false
	}

	return id, true
}
