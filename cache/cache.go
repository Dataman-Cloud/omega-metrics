package cache

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/cihub/seelog"
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

func WriteStringToRedis(key string, value string) error {
	conn := Open()
	defer conn.Close()
	log.Debugf("redis Set marathon event id %s, app %s", key, value)
	_, err := conn.Do("SETEX", key, config.DefaultTimeout, value)
	return err
}

func WriteListToRedis(key, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	var err error
	conf := config.Pairs()
	log.Debugf("redis LPUSH id %s, json %s", key, value)
	if err = conn.Send("LPUSH", key, value); err != nil {
		return err
	}

	log.Debugf("redis EXPIRE id %s, json %s", key, value)

	if timeout == -1 {
		if err = conn.Send("EXPIRE", key); err != nil {
			return err
		}
	}

	if err = conn.Send("EXPIRE", key, timeout); err != nil {
		return err
	}

	_, err = conn.Do("LTRIM", key, 0, conf.Cache.Llen)
	return err
}

func WriteHashToRedis(key, field, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	var err error
	log.Debugf("redis HSET: %s, field: %s, value: %s", key, field, value)
	if err = conn.Send("HSET", key, field, value); err != nil {
		return err
	}

	if timeout == -1 {
		if _, err = conn.Do("EXPIRE", key, timeout); err != nil {
			return err
		}
	}

	_, err = conn.Do("EXPIRE", key, timeout)
	return err
}

func HashDelFromRedis(key, field string) error {
	conn := Open()
	defer conn.Close()
	var err error
	log.Debugf("redis HDEL: %s, field: %s", key, field)
	_, err = conn.Do("HDEL", key, field)
	return err
}

func ReadFromRedis(key string) (string, error) {
	conn := Open()
	defer conn.Close()
	log.Debugf("redis Get key %s", key)
	value, err := redis.String(conn.Do("GET", key))
	return value, err
}

func DeleteRedisByKey(key string) error {
	conn := Open()
	defer conn.Close()
	log.Debugf("redis delete key %s", key)
	_, err := conn.Do("DEL", key)
	return err
}
