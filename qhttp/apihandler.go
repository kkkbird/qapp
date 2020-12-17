package qhttp

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// Context is the request context, can use *gin.Context directly
type Context interface {
	Get(key string) (value interface{}, exists bool)
	MustGet(key string) interface{}
	Set(key string, value interface{})
}

// APIRequest define the interface for api request
type APIRequest interface {
	Rsp(Context) APIResponse // main response function
	RspOK(Context, ...interface{}) APIResponse
	RspInvalidParam(Context, error) APIResponse
	RspInternalError(Context, error) APIResponse
}

// APIResponse define the interface for api response
type APIResponse interface {
	WithError(err error) APIResponse // use error string as error cause
	IsError(err error) bool          // check if an error is a APIResponse error
	HasError() error                 // convert APIResponse to error, it should just return a copy of APIResponse except the status_code == OK
	Error() string                   // make APIResponse an error, we cannot use it directly but with HasError()
}

// APIMiddleware use as pre-handle the request after unmarshal
type APIMiddleware func(c Context, req APIRequest) (APIResponse, APIMiddlewareDefer)

// APIMiddlewareDefer used if APIMiddleware need something to free or call
type APIMiddlewareDefer func(c Context, req APIRequest, rsp APIResponse)

// GinAPIHandler implement the api handler using gin
func GinAPIHandler(r APIRequest, middlewares ...APIMiddleware) gin.HandlerFunc {
	typ := reflect.Indirect(reflect.ValueOf(r)).Type()

	return func(c *gin.Context) {
		var rsp APIResponse

		defer func() {
			c.JSON(http.StatusOK, rsp)
		}()

		req := reflect.New(typ).Interface().(APIRequest)

		var err error

		if len(c.ContentType()) == 0 { // try bind json if content-type absent
			err = c.ShouldBindJSON(req)
		} else {
			err = c.ShouldBind(req)
		}

		if err != nil {
			rsp = req.RspInvalidParam(c, err)
			return
		}

		defer func() {
			if err := recover(); err != nil {
				rsp = req.RspInternalError(c, fmt.Errorf("panic: %v", err))
			}
		}()

		dfs := make([]APIMiddlewareDefer, 0)

		defer func() {
			for i := len(dfs) - 1; i >= 0; i-- {
				dfs[i](c, req, rsp)
			}
		}()

		var df APIMiddlewareDefer

		for _, mw := range middlewares {
			rsp, df = mw(c, req)

			if df != nil {
				dfs = append(dfs, df)
			}

			if rsp != nil {
				return
			}
		}

		rsp = req.Rsp(c)
		return
	}
}

// an example to implete the base api struct
const (
	StatusOK            = iota
	StatusInternalError = 1000
	StatusInvalidParams = 1001
)

// CommonResponse is an example to implement the APIResponse
type CommonResponse struct {
	Status     int    `json:"status"`
	StatusInfo string `json:"status_info,omitempty"`
	ErrorCause string `json:"error_cause,omitempty"`
}

// HasError check and return if response error
func (r CommonResponse) HasError() error {
	if r.Status != StatusOK {
		return r
	}

	return nil
}

func (r CommonResponse) Error() string {
	if r.Status == StatusOK {
		panic("cannot use ok respone as err")
	}
	return fmt.Sprintf("status code=%d, info=%s, cause=%s", r.Status, r.StatusInfo, r.ErrorCause)
}

// WithError set the err as error cause
func (r CommonResponse) WithError(err error) APIResponse {
	return CommonResponse{
		Status:     r.Status,
		StatusInfo: r.StatusInfo,
		ErrorCause: err.Error(),
	}
}

// IsError can check if an err is a api response error,
func (r CommonResponse) IsError(err error) bool {
	if err == nil {
		return false
	}

	err2, ok := err.(CommonResponse)
	if !ok {
		return false
	}

	return err2.Status == r.Status
}

// CommonRequest is an example to implement the APIRequest without the Rsp() function
type CommonRequest struct{}

// RspOK return the rsp ok
func (r *CommonRequest) RspOK(c Context, data ...interface{}) APIResponse {
	return RspOk
}

// RspInvalidParam return the RspInvalidParam response
func (r *CommonRequest) RspInvalidParam(c Context, err error) APIResponse {
	return RspInvalidParam.WithError(err)
}

// RspInternalError return the RspInternalError response
func (r *CommonRequest) RspInternalError(c Context, err error) APIResponse {
	return RspInternalError.WithError(err)
}

// Predfined response
var (
	RspOk            = CommonResponse{StatusOK, "OK", ""}
	RspInternalError = CommonResponse{StatusInternalError, "server internal error", ""}
	RspInvalidParam  = CommonResponse{StatusInvalidParams, "invalid params", ""}
)
