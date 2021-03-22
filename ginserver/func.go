package ginserver

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/arch3754/toolbox/errors"
)

func Md5sum(body string) string {
	Md5 := md5.New()
	io.WriteString(Md5, body)
	return fmt.Sprintf("%x", Md5.Sum(nil))
}

//结构体去空格
func TrimStructSpace(p interface{}) {
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	for k := 0; k < t.Elem().NumField(); k++ {
		if t.Elem().Field(k).Type.Name() == "string" {
			v.Elem().Field(k).SetString(strings.TrimSpace(v.Elem().Field(k).String()))
		}
	}
}

func MustQueryStr(c *gin.Context, key string) string {
	val := c.Query(key)
	if val == "" {
		errors.Bomb("arg[%s] not found", key)
	}

	return val
}
func QueryStr(c *gin.Context, key string, defaultVal string) string {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func MustQueryInt(c *gin.Context, key string) int {
	strv := MustQueryStr(c, key)

	intv, err := strconv.Atoi(strv)
	if err != nil {
		errors.Bomb("cannot convert [%s] to int", strv)
	}

	return intv
}

func RenderMessage(c *gin.Context, v interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys["response"] = v
	if v == nil {
		c.JSON(200, gin.H{"err": "", "request_id": c.GetHeader("request_id")})
		return
	}

	switch t := v.(type) {
	case string:
		c.JSON(200, gin.H{"err": t, "request_id": c.GetHeader("request_id")})
	case error:
		c.JSON(200, gin.H{"err": t.Error(), "request_id": c.GetHeader("request_id")})
	}
}

func RenderData(c *gin.Context, data interface{}, err error) {
	if err == nil {
		if c.Keys == nil {
			c.Keys = make(map[string]interface{})
		}
		c.Keys["response"] = data
		c.JSON(200, gin.H{"dat": data, "err": "", "request_id": c.GetHeader("request_id")})
		return
	}
	RenderMessage(c, err.Error())
}
