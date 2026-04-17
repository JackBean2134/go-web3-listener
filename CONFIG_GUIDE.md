# 配置文件使用指南

## 快速开始

### 1. 创建配置文件

复制示例配置文件：

```powershell
copy config.example.yaml config.yaml
```

### 2. 修改配置

编辑 `config.yaml` 文件，根据你的实际情况修改以下配置：

#### MySQL数据库配置（必须修改）

```yaml
mysql:
  host: "127.0.0.1"      # MySQL服务器地址
  port: 3306              # MySQL端口
  user: "root"            # 用户名
  password: "your_password"  # ⚠️ 修改为你的密码
  dbname: "web3_listener" # 数据库名
```

#### RPC节点配置（可选）

默认已配置多个公共节点，如果需要添加自己的节点：

```yaml
rpc:
  type: http
  nodes:
    - name: "MyNode"
      url: "https://your-rpc-node.com"
      type: "http"
```

#### 告警配置（可选）

启用钉钉告警：

```yaml
alert:
  enabled: true
  watch_to_addrs:
    - "0xyour_address_here"  # 监控的转入地址（小写）
  threshold:
    USDT: "1000"  # USDT阈值
  dingtalk:
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
```

### 3. 运行程序

```powershell
# 使用默认配置文件 (config.yaml)
go run .

# 或指定配置文件路径
go run . -config path/to/config.yaml
```

## 配置项详解

### 服务器配置 (server)

```yaml
server:
  port: 8080        # HTTP服务端口
  mode: release     # 运行模式: debug, release, test
```

### RPC节点配置 (rpc)

```yaml
rpc:
  type: http  # RPC类型: http 或 ws
  nodes:      # 节点列表（按优先级排序）
    - name: "Node1"
      url: "https://node1.com"
      type: "http"
    - name: "Node2"
      url: "https://node2.com"
      type: "http"
```

**建议：**
- 配置至少3个节点以提高可用性
- 将稳定的节点放在前面
- 避免使用需要API密钥的节点（除非已配置）

### 合约监听配置 (contracts)

```yaml
contracts:
  - type: "USDT"
    address: "0x55d398326f99059fF775478AFB9D749AD585d14E"
    decimals: 6
  - type: "BTCB"
    address: "0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c"
    decimals: 18
```

**注意：**
- BSC上的USDT是6位小数
- BTCB和BNB是18位小数
- 不要修改 Transfer Event Topic

### MySQL配置 (mysql)

```yaml
mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "password"
  dbname: "web3_listener"
  max_idle_conns: 10          # 最大空闲连接数
  max_open_conns: 50          # 最大打开连接数
  conn_max_lifetime: 30m      # 连接最大存活时间
  conn_max_idle_time: 5m      # 空闲连接最大存活时间
```

**性能调优：**
- 高并发场景：增加 `max_open_conns` 到100-200
- 低内存环境：减少 `max_idle_conns` 到5

### 告警配置 (alert)

```yaml
alert:
  enabled: false  # 是否启用告警
  
  # 监控的转入地址（小写）
  watch_to_addrs:
    - "0xabc123..."
  
  # 告警阈值（人类可读金额）
  threshold:
    USDT: "1000"
    BTCB: "0.1"
    BNB: "10"
  
  # 钉钉机器人
  dingtalk:
    webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
  
  # SMTP邮件
  smtp:
    host: "smtp.qq.com"
    port: 587
    user: "your_email@qq.com"
    password: "authorization_code"  # 授权码，不是密码
    from: "your_email@qq.com"
    to:
      - "recipient@example.com"
```

### 监听器配置 (listener)

```yaml
listener:
  poll_interval: 10s        # 轮询间隔
  request_interval: 200ms   # 请求间隔（避免限流）
  retry_max_attempts: 3     # 重试最大次数
  connection_timeout: 15s   # 连接超时时间
```

**性能优化：**
- 降低 `poll_interval` 可以减少延迟，但会增加RPC压力
- 增加 `request_interval` 可以避免限流，但会降低处理速度
- 根据RPC节点的速率限制调整这两个参数

## 常见问题

### Q1: 如何切换到测试网？

修改RPC节点为测试网节点：

```yaml
rpc:
  nodes:
    - name: "BSC Testnet"
      url: "https://data-seed-prebsc-1-s1.binance.org:8545/"
      type: "http"
```

同时修改合约地址为测试网合约。

### Q2: 如何只监听USDT？

注释掉其他合约：

```yaml
contracts:
  - type: "USDT"
    address: "0x55d398326f99059fF775478AFB9D749AD585d14E"
    decimals: 6
  # - type: "BTCB"
  #   address: "0x7130d2a12b9bcbfae4f2634d864a1ee1ce3ead9c"
  #   decimals: 18
```

### Q3: 如何禁用数据库？

目前版本必须要有MySQL连接。如果只想测试监听功能，可以：
1. 安装MySQL
2. 创建空数据库
3. 程序会自动创建表结构

### Q4: 配置文件在哪里？

- 默认：项目根目录的 `config.yaml`
- 自定义：使用 `-config` 参数指定路径

```powershell
go run . -config D:\configs\my-config.yaml
```

### Q5: 如何验证配置文件是否正确？

运行程序时会显示配置信息：

```
========== 配置信息 ==========
服务器端口: 8080
服务器模式: release
RPC节点数量: 11
监听合约数量: 3
MySQL主机: 127.0.0.1:3306
MySQL数据库: web3_listener
告警启用: false
轮询间隔: 10s
请求间隔: 200ms
==============================
```

## 最佳实践

1. **不要提交敏感信息**
   - `config.yaml` 已在 `.gitignore` 中
   - 使用 `config.example.yaml` 作为模板

2. **多环境配置**
   ```
   config.dev.yaml      # 开发环境
   config.test.yaml     # 测试环境
   config.prod.yaml     # 生产环境
   ```

3. **定期备份配置**
   ```powershell
   copy config.yaml config.yaml.backup
   ```

4. **使用环境变量覆盖敏感信息**
   （未来版本支持）

## 下一步

配置文件系统已完成，你还可以：
- 批量写入优化
- 监控指标导出
- WebSocket订阅模式
- 多链支持

查看 README.md 了解更多功能。
