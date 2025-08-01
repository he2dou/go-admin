package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/he2dou/go-admin/internal/config"
	"github.com/he2dou/go-admin/internal/pkg/ginx"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Copy body to context bytes array
func CopyBodyMiddleware(skippers ...SkipperFunc) gin.HandlerFunc {
	var maxMemory int64 = 64 << 20 // 64 MB
	if v := config.App.HTTP.MaxContentLength; v > 0 {
		maxMemory = v
	}

	return func(c *gin.Context) {
		if SkipHandler(c, skippers...) || c.Request.Body == nil {
			c.Next()
			return
		}

		var requestBody []byte
		isGzip := false
		safe := &io.LimitedReader{R: c.Request.Body, N: maxMemory}

		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(safe)
			if err == nil {
				isGzip = true
				requestBody, _ = ioutil.ReadAll(reader)
			}
		}

		if !isGzip {
			requestBody, _ = ioutil.ReadAll(safe)
		}

		c.Request.Body.Close()
		bf := bytes.NewBuffer(requestBody)
		c.Request.Body = http.MaxBytesReader(c.Writer, ioutil.NopCloser(bf), maxMemory)
		c.Set(ginx.ReqBodyKey, requestBody)

		c.Next()
	}
}
