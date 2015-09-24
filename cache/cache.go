package cache

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/Sirupsen/logrus"
	redis "github.com/garyburd/redigo/redis"
)

var pool *redis.Pool

func Open() redis.Conn {
	log.Debug("cahce :", pool)
	if pool != nil {
		return pool.Get()
	}

	mutex := &sync.Mutex{}
	mutex.Lock()
	InitCache()
	defer mutex.Unlock()
	log.Debug("cahce1 :", pool)

	return pool.Get()
}

func initConn() (redis.Conn, error) {
	conf := config.Pairs()
	addr := fmt.Sprintf("%s:%d", conf.Cache.Host, conf.Cache.Port)
	c, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return c, err
}

func InitCache() {
	conf := config.Pairs()
	pool = redis.NewPool(initConn, conf.Cache.PoolSize)
	conn := Open()
	defer conn.Close()
	pong, err := conn.Do("ping")
	if err != nil {
		log.Error("got err", err)
		log.Error("can't connect cache server: ", conf.Cache)
		panic(-1)
	}
	log.Debug("reach cache server ", pong)
	log.Debug("initialized cache: ", conf.Cache)
}

func DestroyCache() {
	log.Info("destroying Cache")
	if pool != nil {
		pool.Close()
		log.Info("cache was closed")
	}
}
