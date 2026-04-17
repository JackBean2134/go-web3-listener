# Go Web3 Listener - BSC区块链转账监听器

[![Go](https://img.shields.io/badge/Go-1.25.6-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

## 项目简介

Go Web3 Listener 是一个基于Go语言开发的BSC（币安智能链）区块链转账监听服务，能够实时监控USDT、BTCB和BNB等代币的转账事件，并将交易数据持久化存储到MySQL数据库中。该项目提供了RESTful API接口用于查询转账记录，并支持钉钉和邮件告警功能。

## 主要特性

- ✅ **多合约监听**：同时监听USDT、BTCB、BNB三种代币的转账事件
- ✅ **RPC节点池管理**：支持多个RPC节点配置，具备健康检查和自动故障切换能力
- ✅ **实时数据处理**：每10秒轮询新区块，实时捕获转账事件
- ✅ **数据持久化**：将转账记录存储到MySQL数据库，支持幂等去重
- ✅ **RESTful API**：提供按地址和按合约查询转账记录的API接口
- ✅ **智能告警系统**：支持钉钉机器人和SMTP邮件告警，可设置金额阈值和监控地址
- ✅ **高可用性设计**：内置重试机制、连接恢复和限流处理

## 技术栈

- **后端框架**: [Gin](https://github.com/gin-gonic/gin) Web Framework
- **区块链交互**: [go-ethereum](https://github.com/ethereum/go-ethereum) 客户端库
- **数据库**: MySQL + [GORM](https://gorm.io/) ORM
- **部署环境**: Windows/Linux

## 核心功能

### 区块链监听
- 使用轮询模式监听BSC区块链上的Transfer事件
- 支持HTTP和WebSocket类型的RPC节点
- 自动处理RPC限流和节点故障
- 智能节点健康检查和自动切换

### 数据存储
- 转账记录包含完整的交易信息（区块号、时间戳、交易哈希、发送/接收地址、金额等）
- 使用(tx_hash, log_index)唯一索引确保数据不重复
- 支持数据库断线自动重连

### API接口
- `GET /deposit/address/:addr` - 按地址查询转账记录
- `GET /deposit/contract/:contractAddr/list` - 按合约地址查询转账记录

### 告警系统
- 钉钉Webhook消息推送
- SMTP邮件通知
- 可配置的监控地址和金额阈值
- 支持重试机制确保告警送达

## 快速开始

### 前置要求

- Go 1.25.6+
- MySQL 5.7+
- BSC RPC节点访问权限

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/JackBean2134/go-web3-listener.git
cd go-web3-listener
```

2. 安装依赖
```bash
go mod download
```

3. 配置数据库
在 `config/config.go` 中修改MySQL连接配置：
```go
var DefaultMySQL = MySQLConfig{
    Host:     "127.0.0.1",
    Port:     3306,
    User:     "root",
    Password: "your_password",
    DBName:   "web3_listener",
}
```

4. 创建数据库
```sql
CREATE DATABASE web3_listener CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

5. 运行程序
```bash
go run main.go
```

### 配置说明

#### RPC节点配置
在 `config/config.go` 中配置多个RPC节点以提高可用性：
```go
var RPCNodes = []RPCNode{
    {Name: "Defibit-4", URL: "https://bsc-dataseed4.defibit.io/", Type: RPCTypeHTTP},
    {Name: "Ankr-HTTP", URL: "https://rpc.ankr.com/bsc", Type: RPCTypeHTTP},
    // ... 更多节点
}
```

#### 告警配置
启用告警功能：
```go
var DefaultAlert = AlertConfig{
    Enabled:      true,
    WatchToAddrs: []string{"0xyour_address_here"}, // 监控的转入地址
    Threshold: map[ContractType]string{
        ContractUSDT: "1000", // USDT阈值
    },
    DingTalk: DingTalkConfig{
        WebHook: "your_dingtalk_webhook_url",
    },
}
```

## API使用示例

### 按地址查询转账记录
```bash
curl http://localhost:8080/deposit/address/0x55d398326f99059ff775478afb9d749ad585d14e?page=1&size=10
```

响应示例：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "ID": 1,
        "ContractType": "USDT",
        "ContractAddr": "0x55d398326f99059ff775478afb9d749ad585d14e",
        "BlockNum": 12345678,
        "TxHash": "0x...",
        "FromAddr": "0x...",
        "ToAddr": "0x...",
        "Amount": "100.50",
        "AmountRaw": "100500000"
      }
    ]
  }
}
```

### 按合约查询转账记录
```bash
curl http://localhost:8080/deposit/contract/0x55d398326f99059ff775478afb9d749ad585d14e/list?page=1&size=10
```

## 项目结构

```
go-web3-listener/
├── config/              # 配置文件
│   └── config.go       # 全局配置
├── contract/           # 合约ABI文件
│   ├── ERC20.go
│   └── erc20.abi
├── ethclient/          # 以太坊客户端相关
│   ├── alert.go        # 告警功能
│   ├── amount.go       # 金额格式化
│   ├── listener.go     # 转账监听
│   └── rpc_pool.go     # RPC节点池
├── model/              # 数据模型
│   ├── deposit.go      # 存款记录模型
│   └── init.go         # 数据库初始化
├── main.go             # 程序入口
├── go.mod              # Go模块依赖
└── go.sum              # 依赖校验文件
```

## 应用场景

- 🏦 加密货币交易所充值监控
- 💰 DeFi应用资金流动追踪
- 👛 钱包应用交易提醒
- 📊 区块链数据分析平台

## 注意事项

1. **RPC节点选择**: 建议使用多个可靠的RPC节点以避免限流
2. **数据库性能**: 大量转账记录时建议对数据库进行优化和分表
3. **告警频率**: 合理设置告警阈值，避免告警风暴
4. **安全性**: 不要将敏感配置（如数据库密码、API密钥）提交到版本控制

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题或建议，请通过 GitHub Issues 联系。

---

⭐ 如果这个项目对你有帮助，请给个Star支持一下！
