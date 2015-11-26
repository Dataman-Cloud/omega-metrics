package util

import (
	"fmt"
	"sync"

	"github.com/Dataman-Cloud/omega-metrics/config"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func DB() *sqlx.DB {
	if db != nil {
		return db
	}

	mutex := sync.Mutex{}
	mutex.Lock()
	InitDB()
	defer mutex.Unlock()

	return db
}

func InitDB() {
	conf := config.Pairs()
	url := DBUrl()
	var err error
	db, err = sqlx.Open("mysql", url)
	if err != nil {
		log.Error("can't open db: ", url, " err: ", err)
		panic(-1)
	}

	err = db.Ping()
	if err != nil {
		log.Error("can't ping db: ", url, " err: ", err)
		panic(-1)
	}
	db.SetMaxIdleConns(conf.Db.MaxIdleConns)
	db.SetMaxOpenConns(conf.Db.MaxOpenConns)

	log.Debug("initialized db: ", url)
}

func DestroyDB() {
	log.Info("destroying DB")
	if db != nil {
		db.Close()
		log.Info("db closed")
	}
}

func DBUrl() string {
	conf := config.Pairs()
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		conf.Db.User, conf.Db.Password, conf.Db.Host, conf.Db.Port, conf.Db.Name)
}
