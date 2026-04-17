package model

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

var (
	dbMu     sync.Mutex
	lastDSN  string
	sqlDBRef *sql.DB
)

// InitDB 初始化MySQL连接，并自动迁移表结构。
// 若启动时连接失败，可在后续写入/查询时通过 EnsureDB() 触发重连。
func InitDB(dsn string) error {
	lastDSN = dsn

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDBRef = sqlDB

	// 连接池配置：降低“空闲连接被回收/断开后复用”导致的错误概率
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := EnsureDB(); err != nil {
		return err
	}

	return DB.AutoMigrate(&DepositRecord{})
}

// EnsureDB 检查连接可用性；不可用时尝试重连。
func EnsureDB() error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if DB == nil || sqlDBRef == nil {
		if lastDSN == "" {
			return errors.New("mysql not initialized")
		}
		return reconnectLocked()
	}

	if err := sqlDBRef.Ping(); err != nil {
		log.Printf("mysql ping failed, try reconnect: %v", err)
		return reconnectLocked()
	}
	return nil
}

func reconnectLocked() error {
	if lastDSN == "" {
		return errors.New("mysql dsn is empty")
	}

	db, err := gorm.Open(mysql.Open(lastDSN), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	DB = db
	sqlDBRef = sqlDB
	return DB.AutoMigrate(&DepositRecord{})
}

func IsConnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid connection") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "server has gone away") ||
		strings.Contains(msg, "lost connection")

}
