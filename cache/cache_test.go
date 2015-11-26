package cache

import (
	"fmt"
	redis "github.com/garyburd/redigo/redis"
	"testing"
)

func TestCache(t *testing.T) {
	conn := Open()
	defer conn.Close()
	token, err := redis.String(conn.Do("GET", "AutoScaleToken"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(token)
}
