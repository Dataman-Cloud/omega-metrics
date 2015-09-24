package util

import "github.com/gin-gonic/gin"

func Header(c *gin.Context, key string) string {
	if values, _ := c.Request.Header[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}
