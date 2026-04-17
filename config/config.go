package config

import "fmt"

const (
	// BSC公共RPC（轮询模式可用）
	// RpcUrl = "https://bsc-dataseed1.binance.org:443"
	RpcUrl     = "https://bsc-dataseed4.defibit.io/"
	ServerAddr = ":8080"

	// USDT合约地址（BSC）
	UsdtContract = "0x55d398326f99059fF775478AFB9D749AD585d14E"
)

// MySQLConfig MySQL连接配置。
//
// 说明：当前项目未引入配置文件加载逻辑，因此这里以常量形式提供默认值；
// 若后续接入 yml/json/env，只需在启动时覆盖这些字段即可。
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

var DefaultMySQL = MySQLConfig{
	Host:     "127.0.0.1",
	Port:     3306,
	User:     "root",
	Password: "root",
	DBName:   "web3_listener",
}

// MySQLDsn 兼容旧代码（如仍直接引用该常量/变量）。
// 推荐使用 DefaultMySQL.DSN()，方便以后扩展。
var MySQLDsn = DefaultMySQL.DSN()

func (c MySQLConfig) DSN() string {
	// parseTime=True 对 time.Time 字段非常关键（否则扫描会失败）。
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}
