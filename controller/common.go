package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOk        = 0
	InvalidParams = 17001
	DbQueryError  = 17002
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
