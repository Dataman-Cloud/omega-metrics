package main

import (
	"net/http"
	"time"

	"github.com/Dataman-Cloud/omega-metrics/cache"
	"github.com/Dataman-Cloud/omega-metrics/util"
	log "github.com/cihub/seelog"
	redis "github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
)

const (
	HeaderToken = "Authorization"
	Key         = "UserID"
)

const (
	StatusFailed    = 1
	StatusDeploying = 2
	StatusRunning   = 3
	StatusStopped   = 4
	StatusDeleted   = 5
)

type Application struct {
	Id      int64
	Cid     int64
	Name    string
	Status  uint8
	Json    []byte
	Created time.Time
	Updated time.Time
}

func init() {
	conn := cache.Open()
	defer conn.Close()
	conn.Do("HSET", Key, "123456", "user123")
}

func authenticate(ctx *gin.Context) {
	authenticated := false
	token := util.Header(ctx, HeaderToken)

	if len(token) > 0 {
		conn := cache.Open()
		defer conn.Close()
		log.Debug("token: ", token)
		uid, err := redis.String(conn.Do("HGET", Key, token))
		if err == nil {
			authenticated = true
			ctx.Set("uid", uid)
			log.Debug("uid: ", uid)
		} else if err != redis.ErrNil {
			log.Error("[app] got error1 ", err)
		} else {
			log.Error("[app] got err2 ", err)
		}
	}

	log.Debug("header: ", ctx.Request.Header)
	log.Debug("token: ", token)

	if authenticated {
		ctx.Next()
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": -1})
		ctx.Abort()
	}
}
