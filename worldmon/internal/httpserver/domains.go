package httpserver

// ServiceNames lists known /api/{service}/… first path segments for GET
// /capabilities. The HTTP proxy still accepts any service name; this is only
// for discovery metadata when registering with agentglobe.
var ServiceNames = []string{
	"news",
}
