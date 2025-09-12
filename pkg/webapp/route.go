package webapp

import "github.com/gobwas/glob"

// RequestMethod is a type for HTTP request method
type RequestMethod string

const (
	GET     RequestMethod = "GET"
	POST    RequestMethod = "POST"
	PUT     RequestMethod = "PUT"
	DELETE  RequestMethod = "DELETE"
	PATCH   RequestMethod = "PATCH"
	HEAD    RequestMethod = "HEAD"
	OPTIONS RequestMethod = "OPTIONS"
)

// ANY is a slice of all HTTP request methods
var ANY = []RequestMethod{GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS}

// Route is a struct for a route definition
type Route struct {
	Path    string
	Methods []RequestMethod
}

// RouteKey is a struct for a route key definition
type RouteKey struct {
	Path   string
	Method string
	Glob   glob.Glob // 预编译
}




