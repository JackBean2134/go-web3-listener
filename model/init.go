package model

import (
	"database/sql"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strings"
	"sync"
	"time"
)

var DB *gorm.DB
var (
	dbMu     sync.Mutex
	lastDSN  string
	sqlDBRef *sql.DB
)

func InitDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db
	lastDSN = dsn

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDBRef = sqlDB
	// 连接池配置：避免长连接被 MySQL/NAT 静默断开后继续复用
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := EnsureDB(); err != nil {
		return err
	}

	if err := db.AutoMigrate(&DepositRecord{}); err != nil {
		return err
	}
	return nil
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

	// 快速探活
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

	// 迁移失败不影响重连，但这里返回错误便于上层感知
	if err := db.AutoMigrate(&DepositRecord{}); err != nil {
		return err
	}
	return nil
}

func IsConnErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	// 覆盖常见场景：driver bad connection / server has gone away / reset by peer 等
	if strings.Contains(msg, "invalid connection") ||
		strings.Contains(msg, "bad connection") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "server has gone away") ||
		strings.Contains(msg, "lost connection") {
		return true
	}
	return false
}
