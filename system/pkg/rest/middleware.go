package rest

import "github.com/gin-gonic/gin"

type Middleware struct {
	Handler gin.HandlerFunc
	Group   string
}

func NewMiddleware(group string, handler gin.HandlerFunc) Middleware {
	return Middleware{
		Group:   group,
		Handler: handler,
	}
}
