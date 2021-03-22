package middleware

import (
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func AddRequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		//为进来的每个请求，赋予一个唯一的uuid，便于查询访问日志，定位问题
		c.Request.Header.Set("request_id", uuid.NewV1().String())
		c.Next()
	}
}
