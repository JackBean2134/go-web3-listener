package config

import (
	"fmt"
	"time"
)

// RPCType RPC类型
type RPCType string

const (
	RPCTypeHTTP RPCType = "http"
	RPCTypeWS   RPCType = "ws"
)

// ContractType 合约类型
type ContractType string

const (
	ContractUSDT ContractType = "USDT"
	ContractBTCB ContractType = "BTCB"
	ContractBNB  ContractType = "BNB"
)

// RPCNode BSC RPC 节点配置（支持 HTTP/WS）。
type RPCNode struct {
	Name string  `yaml:"name"`
	URL  string  `yaml:"url"`
	Type RPCType `yaml:"type"`
}

// ContractConfig 合约配置
type ContractConfig struct {
	Type     ContractType `yaml:"type"`
	Address  string       `yaml:"address"`
	Decimals int          `yaml:"decimals"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebHook string `yaml:"webhook"`
}

// SMTPConfig SMTP邮件配置
type SMTPConfig struct {
	Host     string   `yaml:"host"` // 如 smtp.qq.com
	Port     int      `yaml:"port"` // 如 465/587/25
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
	To       []string `yaml:"to"`
	From     string   `yaml:"from"` // 可为空，默认用 User
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled      bool                    `yaml:"enabled"`
	WatchToAddrs []string                `yaml:"watch_to_addrs"` // 触发条件：转入地址命中（全部小写对比）
	Threshold    map[ContractType]string `yaml:"threshold"`      // 阈值（人类可读金额字符串）
	DingTalk     DingTalkConfig          `yaml:"dingtalk"`
	SMTP         SMTPConfig              `yaml:"smtp"`
}

// MySQLConfig MySQL连接配置。
type MySQLConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// GetDSN 获取MySQL DSN连接字符串
func (c *MySQLConfig) GetDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}
