package middleware

import (
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"time"
)

// GinLogger 接收gin框架默认的日志
func GinLogger(log *logs.BeeLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logs.Info(`"%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v" "%s=%v"`,
			"client_ip", c.ClientIP(),
			"req_method", c.Request.Method,
			"http_version", c.Request.Proto,
			"http_code", c.Writer.Status(),
			"response_body_size", c.Writer.Size(),
			"response_time", time.Since(start).Milliseconds(),
			"req_id", c.Request.Header.Get("request_id"),
			"req_uri", c.Request.RequestURI,
			"referer", c.Request.Referer(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
