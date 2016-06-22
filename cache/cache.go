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
	if pool != nil {
		return pool.Get()
	}

	mutex := &sync.Mutex{}
	mutex.Lock()
	defer mutex.Unlock()
	InitCache()

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
		log.Errorf("initcache has error: %s, can't connect cache server: %+v", err.Error(), conf.Cache)
		panic(-1)
	}
	log.Debugf("reach cache server: %s initialized cache: %+v ", pong, conf.Cache)
}

func DestroyCache() {
	log.Info("destroying Cache")
	if pool != nil {
		pool.Close()
		log.Info("cache was closed")
	}
}

func WriteStringToRedis(key string, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	if timeout != -1 {
		_, err := conn.Do("SETEX", key, timeout, value)
		return err
	}
	_, err := conn.Do("SET", key, value)
	return err
}

func WriteSetToRedis(key, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	var err error
	if _, err = conn.Do("SADD", key, value); err != nil {
		return err
	}

	if timeout != -1 {
		_, err = conn.Do("EXPIRE", key, timeout)
		return err
	}
	return nil
}

func WriteListToRedis(key, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	var err error
	conf := config.Pairs()
	if err = conn.Send("LPUSH", key, value); err != nil {
		return err
	}

	if timeout != -1 {
		if err = conn.Send("EXPIRE", key, timeout); err != nil {
			return err
		}
	}

	_, err = conn.Do("LTRIM", key, 0, conf.Cache.Llen)
	return err
}

func WriteHashToRedis(key, field, value string, timeout int) error {
	conn := Open()
	defer conn.Close()
	var err error
	if _, err = conn.Do("HSET", key, field, value); err != nil {
		return err
	}

	if timeout != -1 {
		_, err = conn.Do("EXPIRE", key, timeout)
		return err
	}
	return nil
}

func HashDelFromRedis(key, field string) error {
	conn := Open()
	defer conn.Close()
	var err error
	_, err = conn.Do("HDEL", key, field)
	return err
}

func ReadFromRedis(key string) (string, error) {
	conn := Open()
	defer conn.Close()
	value, err := redis.String(conn.Do("GET", key))
	return value, err
}

func DeleteRedisByKey(key string) error {
	conn := Open()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func ReadSetMembers(key string) ([]string, error) {
	conn := Open()
	defer conn.Close()
	return redis.Strings(conn.Do("SMEMBERS", key))
}
