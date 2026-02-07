package gzip

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CompressWriter struct {
	w  gin.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w gin.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(b []byte) (int, error) {
	if strings.Contains(c.Header().Get("Content-Type"), "application/json") || strings.Contains(c.Header().Get("Content-Type"), "text/html") {
		c.Header().Set("Content-Encoding", "gzip")
		return c.zw.Write(b)
	}
	return c.w.Write(b)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

func (c *CompressWriter) CloseNotify() <-chan bool {
	return c.w.CloseNotify()
}

func (c *CompressWriter) Flush() {
	c.w.Flush()
}

func (c *CompressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return c.w.Hijack()
}

func (c *CompressWriter) Pusher() http.Pusher {
	return c.w.Pusher()
}

func (c *CompressWriter) Status() int {
	return c.w.Status()
}

func (c *CompressWriter) Size() int {
	return c.w.Size()
}

func (c *CompressWriter) WriteString(s string) (int, error) {
	return c.zw.Write([]byte(s))
}

func (c *CompressWriter) Written() bool {
	return c.w.Written()
}

func (c *CompressWriter) WriteHeaderNow() {
	c.w.WriteHeaderNow()
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
