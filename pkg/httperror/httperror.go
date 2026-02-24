package httperror

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorsWriter struct {
	gin.ResponseWriter
	ctx         *gin.Context
	statusCode  int
	wroteHeader bool
	b           []byte
}

func NewErrorsWriter(w gin.ResponseWriter, ctx *gin.Context) *ErrorsWriter {
	return &ErrorsWriter{
		ResponseWriter: w,
		ctx:            ctx,
		statusCode:     http.StatusCreated,
		wroteHeader:    false,
	}
}

func (ew *ErrorsWriter) AddError(err error) {
	if ew.ctx != nil {
		ew.ctx.Error(err)
	}
}

func (ew *ErrorsWriter) WriteHeader(code int) {
	ew.statusCode = code
	ew.wroteHeader = true
}

func (ew *ErrorsWriter) MyFlush() {
	if !ew.wroteHeader {
		ew.statusCode = http.StatusCreated
	}
	ew.ResponseWriter.WriteHeader(ew.statusCode)
	ew.ResponseWriter.Write(ew.b)
}

func (ew *ErrorsWriter) Write(b []byte) (int, error) {
	ew.b = b
	return len(b), nil
}
