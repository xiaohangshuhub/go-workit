package workit

import "github.com/gobwas/glob"

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

var ANY = []RequestMethod{GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS}

type Route struct {
	Path    string
	Methods []RequestMethod
}

type RouteKey struct {
	Path   string
	Method string
	Glob   glob.Glob // 预编译
}
