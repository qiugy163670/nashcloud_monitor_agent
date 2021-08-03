package config

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	//数据库链接
	dbUrl = "nashcloud_test:Nashtest345@tcp(rm-uf66a0j64jv59cbeqvo.mysql.rds.aliyuncs.com:3306)/nashcloud_test?charset=utf8"
	//dbUrl = "nashcloud_product:Nash789product@tcp(rm-uf66a0j64jv59cbeqvo.mysql.rds.aliyuncs.com:3306)/nashcloud_product?charset=utf8"
	dbType = "mysql"
)

func InitDBConnection()  {
	db, err := sql.Open(dbType, dbUrl)
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}
