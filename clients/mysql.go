package clients

import (
	"errors"
	"fmt"
	"log-detect/global"
	"time"
	"log-detect/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func LoadDatabase() {

	// 取得config參數
	host := global.EnvConfig.Database.Host
	port := global.EnvConfig.Database.Port
	user := global.EnvConfig.Database.User
	password := global.EnvConfig.Database.Password
	dbname := global.EnvConfig.Database.Db
	parameter := global.EnvConfig.Database.Params

	fmt.Println("host", host)
	fmt.Println("port", port)
	fmt.Println("user", user)
	fmt.Println("password", password)
	fmt.Println("dbname", dbname)
	fmt.Println("parameter", parameter)

	// var err error
	err := errors.New("mock error")
	for err != nil {
		global.Mysql, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, host, port, dbname, parameter)), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		time.Sleep(1 * time.Second)
	}
	// if err != nil {
	// 	panic(err)
	// }

	sqlDB, err := global.Mysql.DB()
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(int(global.EnvConfig.Database.MaxIdle))

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(int(global.EnvConfig.Database.MaxOpenConn))

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	lifeTime, _ := time.ParseDuration(global.EnvConfig.Database.MaxLifeTime)
	sqlDB.SetConnMaxLifetime(lifeTime)
	fmt.Println("SQL Database 連線成功")
	log.Logrecord("SQL ", "SQL Database 連線成功")
}
