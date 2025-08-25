package rest

import "github.com/gin-gonic/gin"

type HttpMethod int

const (
	GET HttpMethod = iota
	POST
	PUT
	PATCH
)

type Route struct {
	Method      HttpMethod
	Path        string
	HandlerFunc gin.HandlerFunc
	Group       string
}

func NewRoute(method HttpMethod, group, path string, handler gin.HandlerFunc) Route {
	return Route{
		Method:      method,
		Path:        path,
		Group:       group,
		HandlerFunc: handler,
	}
}
