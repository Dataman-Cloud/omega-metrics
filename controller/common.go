package controller

import (
	"net/http"
	"time"

	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
)

const (
	CodeOk        = 0
	InvalidParams = 17001
	DbQueryError  = 17002

	HeaderToken = "Authorization"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func ReturnOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": CodeOk, "data": data})
}

func ReturnError(c *gin.Context, errorcode int, err error) {
	c.JSON(http.StatusOK, gin.H{"code": errorcode, "data": ErrorMessage{err.Error()}})
}

func Milliseconds(d time.Duration) float64 {
	min := d / 1e6
	nsec := d % 1e6
	return float64(min) + float64(nsec)*(1e-6)
}

// get request token from request header
func GetToken(c *gin.Context) (token string) {
	req := c.Request
	if req == nil {
		log.Error("[Token] Get token failed request is nil")
		return
	}

	token = req.Header.Get(HeaderToken)
	return
}
