package util

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

var (
	baseUrl string
	server  *httptest.Server
)

func TestMain(m *testing.M) {

	server := startHttpServer()
	baseUrl = server.URL
	defer server.Close()
	os.Exit(m.Run())
}

func FakeAuthenticate(ctx *gin.Context) {
	ctx.Set("uid", 666)
	ctx.Next()
}

func GetStatus(c *gin.Context) {
	data := make(map[string]interface{})
	item := map[string]interface{}{
		"id":          48,
		"cid":         1,
		"name":        "2048",
		"alias":       "ge5dembuha000000",
		"status":      2,
		"tasks":       1,
		"instances":   1,
		"lastfailure": 0,
	}

	data["48"] = item
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data})
	return
}

func startHttpServer() *httptest.Server {
	router := gin.New()
	groupV3 := router.Group("/api/v3", FakeAuthenticate)
	{
		groupV3.GET("/apps/status", GetStatus)
	}
	return httptest.NewServer(router)
}

func TestQueryAppStatus(t *testing.T) {
	split := strings.Split(baseUrl, "//")[1]
	host := strings.Split(split, ":")[0]
	port, _ := strconv.Atoi(strings.Split(split, ":")[1])

	_, err := QueryAppStatus("4aa9dc6aa9fa4db8bb0a6eb0d5b561ac", "1", "http://" + host, port)
	assert.Nil(t, err)
}
