package qhttp

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

type Context interface {
	Get(key string) (value interface{}, exists bool)
	MustGet(key string) interface{}
	Set(key string, value interface{})
}

type APIResponse interface {
	WithError(err error) APIResponse // use an error as error cause
	IsError(err error) bool          // check if an error is a APIResponse error
	HasError() error                 // convert APIResponse to error, it should just return a copy of APIResponse except the status_code == OK
	Error() string                   // make APIResponse an error, we cannot use it directly but with HasError()
}

type APIRequest interface {
	Rsp(Context) APIResponse
	RspOK(Context) APIResponse
	RspInvalidParam(Context, error) APIResponse
	RspInternalError(Context, error) APIResponse
}

// APIMiddleware 为API中间件，返回非nil的时候停止后续执行
type APIMiddleware func(c Context, req APIRequest) APIResponse

func GinAPIHandler(r APIRequest, middlewares ...APIMiddleware) gin.HandlerFunc {
	typ := reflect.Indirect(reflect.ValueOf(r)).Type()

	return func(c *gin.Context) {
		var rsp APIResponse

		defer func() {
			c.JSON(http.StatusOK, rsp)
		}()

		req := reflect.New(typ).Interface().(APIRequest)
		err := c.ShouldBindJSON(req)

		if err != nil {
			rsp = req.RspInvalidParam(c, err)
			return
		}

		for _, mw := range middlewares {
			rsp = mw(c, req)
			if rsp != nil {
				return
			}
		}

		rsp = req.Rsp(c)
		return
	}
}
