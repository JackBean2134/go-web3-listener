# go-web3-listener
Go Web3 Listener 是一个基于Go语言开发的BSC（币安智能链）区块链转账监听服务，能够实时监控USDT、BTCB和BNB等代币的转账事件，并将交易数据持久化存储到MySQL数据库中。该项目提供了RESTful API接口用于查询转账记录，并支持钉钉和邮件告警功能。

主要特性 
 • 多合约监听：同时监听USDT、BTCB、BNB三种代币的转账事件
 • RPC节点池管理：支持多个RPC节点配置，具备健康检查和自动故障切换能力 
 • 实时数据处理：每10秒轮询新区块，实时捕获转账事件 • 数据持久化：将转账记录存储到MySQL数据库，支持幂等去重
 • RESTful API：提供按地址和按合约查询转账记录的API接口 • 智能告警系统：支持钉钉机器人和SMTP邮件告警，可设置金额阈值和监控地址 • 高可用性设计：内置重试机制、连接恢复和限流处理  

技术栈 
 • 后端框架：Gin Web Framework 
 • 区块链交互：go-ethereum客户端库
 • 数据库：MySQL + GORM ORM 
 • 部署环境：Windows/Linux  核心功能 区块链监听 
 • 使用轮询模式监听BSC区块链上的Transfer事件 
 • 支持HTTP和WebSocket类型的RPC节点 • 自动处理RPC限流和节点故障  数据存储
 • 转账记录包含完整的交易信息（区块号、时间戳、交易哈希、发送/接收地址、金额等）
 • 使用(tx_hash, log_index)唯一索引确保数据不重复  

API接口 
 • GET /deposit/address/:addr - 按地址查询转账记录 
 • GET /deposit/contract/:contractAddr/list - 按合约地址查询转账记录  

告警系统 
 • 钉钉Webhook消息推送 
 • SMTP邮件通知 
 • 可配置的监控地址和金额阈值  

应用场景
 • 加密货币交易所充值监控
 • DeFi应用资金流动追踪 
 • 钱包应用交易提醒
 • 区块链数据分析平台  

这个项目为需要实时监控BSC链上转账的应用提供了一个稳定可靠的解决方案，特别适合需要跟踪特定地址或合约资金变动的场景。
