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

// ---------------- 合约监听配置（USDT / BTCB / BNB） ----------------

type ContractType string

const (
	ContractUSDT ContractType = "USDT"
	ContractBTCB ContractType = "BTCB"
	ContractBNB  ContractType = "BNB"
)

// TransferEventTopic ERC20/BEP20 Transfer(address,address,uint256)
// 三个合约共用同一个 topic（只要事件签名一致）。
const TransferEventTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// BTCB 合约地址（BSC）- 常见为 BTCB BEP20
const BtcbContract = "0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c"

// BNB 合约地址（BSC）- 这里使用 WBNB（因为原生 BNB 没有合约事件）
const BnbContract = "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"

type ContractConfig struct {
	Type     ContractType
	Address  string
	Decimals int
}

// Contracts 默认监听合约列表。
// 注意：你的需求指定 USDT 6 位小数、BNB 18 位小数；这里按需求填写。
var Contracts = []ContractConfig{
	{Type: ContractUSDT, Address: UsdtContract, Decimals: 6},
	{Type: ContractBTCB, Address: BtcbContract, Decimals: 18},
	{Type: ContractBNB, Address: BnbContract, Decimals: 18},
}

// ---------------- 告警配置（钉钉 / 邮件） ----------------

type DingTalkConfig struct {
	WebHook string
}

type SMTPConfig struct {
	Host     string // 如 smtp.qq.com
	Port     int    // 如 465/587/25
	User     string
	Password string
	To       []string
	From     string // 可为空，默认用 User
}

type AlertConfig struct {
	Enabled bool

	// 触发条件：转入地址命中（全部小写对比）
	WatchToAddrs []string

	// 阈值（人类可读金额字符串，比如 "1000"、"0.5"）
	// key 为合约类型（USDT/BTCB/BNB）
	Threshold map[ContractType]string

	DingTalk DingTalkConfig
	SMTP     SMTPConfig
}

// DefaultAlert 默认告警配置（按需修改）。
var DefaultAlert = AlertConfig{
	Enabled:      false,
	WatchToAddrs: []string{}, // 例如: []string{"0xabc..."}（请用小写）
	Threshold: map[ContractType]string{
		ContractUSDT: "1000",
	},
	DingTalk: DingTalkConfig{
		WebHook: "",
	},
	SMTP: SMTPConfig{
		Host:     "",
		Port:     587,
		User:     "",
		Password: "",
		To:       []string{},
		From:     "",
	},
}

type RPCType string

const (
	RPCTypeHTTP RPCType = "http"
	RPCTypeWS   RPCType = "ws"
)

// RPCNode BSC RPC 节点配置（支持 HTTP/WS）。
type RPCNode struct {
	Name string
	URL  string
	Type RPCType
}

// RPCNodes 默认节点列表（按优先级排序）。
//
// 说明：公共节点可能存在限流/波动，因此建议配置多个；监听层会自动做健康检查与故障切换。
var RPCNodes = []RPCNode{
	// HTTP - BSC官方节点（最稳定）
	{Name: "BinanceSeed-1", URL: "https://bsc-dataseed1.binance.org/", Type: RPCTypeHTTP},
	{Name: "BinanceSeed-2", URL: "https://bsc-dataseed2.binance.org/", Type: RPCTypeHTTP},
	{Name: "BinanceSeed-3", URL: "https://bsc-dataseed3.binance.org/", Type: RPCTypeHTTP},
	{Name: "BinanceSeed-4", URL: "https://bsc-dataseed4.binance.org/", Type: RPCTypeHTTP},
	// HTTP - Defibit节点
	{Name: "Defibit-1", URL: "https://bsc-dataseed1.defibit.io/", Type: RPCTypeHTTP},
	{Name: "Defibit-2", URL: "https://bsc-dataseed2.defibit.io/", Type: RPCTypeHTTP},
	{Name: "Defibit-3", URL: "https://bsc-dataseed3.defibit.io/", Type: RPCTypeHTTP},
	{Name: "Defibit-4", URL: "https://bsc-dataseed4.defibit.io/", Type: RPCTypeHTTP},
	// HTTP - 其他公共节点
	{Name: "Nodies-1", URL: "https://rpc.nodereal.io/v1/64a9df0874fb4a93b9d0a3849de012d3", Type: RPCTypeHTTP},
	{Name: "PublicNode", URL: "https://bsc.publicnode.com", Type: RPCTypeHTTP},
	{Name: "OnFinality", URL: "https://bsc.api.onfinality.io/public", Type: RPCTypeHTTP},
}

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
