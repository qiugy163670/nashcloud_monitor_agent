package config

import (
	"database/sql"
	"time"

	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql"
)

const (
	//数据库链接
	dbUrl = "nash:Nashcloud123.@tcp(172.28.101.61:3306)/crust?charset=utf8"
	//dbUrl = "nashcloud_product:Nash789product@tcp(rm-uf66a0j64jv59cbeqvo.mysql.rds.aliyuncs.com:3306)/nashcloud_product?charset=utf8"
	dbType = "mysql"
)

var (
	db *sql.DB
)

func Init() {
	SetupLogger()
}

func InitDBConnection() (*sql.DB, error) {
	Init()
	defer log.Flush()

	var err error
	db, err = sql.Open(dbType, dbUrl)
	if err != nil {
		log.Error("get db connection failed: %s at %d", err.Error(), time.Now().Unix())
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

func GetDBConnection() (*sql.DB, error) {
	if db == nil {
		return InitDBConnection()
	}
	return db, nil
}
