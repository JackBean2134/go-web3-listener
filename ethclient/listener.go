package ethclient

import (
	"context"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"go-web3-listener/config"
	"go-web3-listener/model"
)

// 强制使用IPV4（解决IPV6连接问题）
func init() {
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, "tcp4", addr)
	}
}

// ListenUSDTTransfers 监听转账（轮询模式），同时监听 USDT/BTCB/BNB 三个合约。
func ListenUSDTTransfers(rpcUrl string) {
	ctx := context.Background()

	// 1) 初始化RPC节点池（HTTP轮询模式）
	pool, err := NewRPCPool(config.RPCTypeHTTP, config.RPCNodes)
	if err != nil {
		// 兜底：如果没有配置列表，则退回单节点
		log.Printf("RPC节点池初始化失败，退回单节点: %v", err)
		config.RPCNodes = []config.RPCNode{
			{Name: "Single-Node", URL: rpcUrl, Type: config.RPCTypeHTTP},
		}
		pool, _ = NewRPCPool(config.RPCTypeHTTP, config.RPCNodes)
	}
	// 启动节点健康检查（20秒一次）
	pool.StartHealthCheck(ctx, 20*time.Second)

	// 2) 初始连接RPC节点
	client, curIdx, err := pool.DialCurrent(ctx)
	if err != nil {
		log.Fatalf("连接RPC失败: %v", err)
	}

	transferTopic := common.HexToHash(config.TransferEventTopic)
	contracts := config.Contracts
	contractAddrs := make([]common.Address, 0, len(contracts))
	contractMeta := make(map[string]config.ContractConfig, len(contracts))
	for _, c := range contracts {
		addr := strings.ToLower(c.Address)
		contractAddrs = append(contractAddrs, common.HexToAddress(addr))
		contractMeta[addr] = c
	}

	// 3) 获取初始最新区块（带重试机制）
	var latestBlock uint64
	maxInitRetries := len(config.RPCNodes)
	initSuccess := false

	for attempt := 0; attempt < maxInitRetries; attempt++ {
		latestBlock, err = client.BlockNumber(ctx)
		if err == nil {
			initSuccess = true
			break
		}

		log.Printf("获取最新区块失败（节点: %s，尝试 %d/%d）: %v", pool.nodes[curIdx].Name, attempt+1, maxInitRetries, err)

		// 标记故障并切换节点
		pool.MarkFailure(curIdx, err)
		pool.SwitchToNext(err)
		client.Close()

		// 重新连接新节点
		client, curIdx, err = pool.DialCurrent(ctx)
		if err != nil {
			log.Printf("切换RPC节点失败: %v", err)
			continue
		}
	}

	if !initSuccess {
		log.Fatalf("无法连接到任何RPC节点，请检查网络连接或更换RPC节点配置")
	}

	currentBlock := latestBlock
	node, _ := pool.CurrentNode()
	log.Printf("✅ 开始轮询转账（USDT/BTCB/BNB），起始区块: %d，当前RPC节点: %s", currentBlock, node.Name)

	// 4) 定时轮询（每10秒查一次）
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
		client.Close()
	}()

	for range ticker.C {
		// 5.1 获取最新区块
		newLatestBlock, err := client.BlockNumber(ctx)
		if err != nil {
			// 失败/限流：标记故障并切换节点
			if IsRateLimitErr(err) {
				log.Printf("RPC疑似限流（节点: %s）: %v", pool.nodes[curIdx].Name, err)
			} else {
				log.Printf("获取最新区块失败（节点: %s）: %v", pool.nodes[curIdx].Name, err)
			}

			pool.MarkFailure(curIdx, err)
			pool.SwitchToNext(err)
			client.Close()

			// 重新连接新节点
			client, curIdx, err = pool.DialCurrent(ctx)
			if err != nil {
				log.Printf("切换RPC后仍失败: %v", err)
				continue
			}
			// 重新获取最新区块
			newLatestBlock, err = client.BlockNumber(ctx)
			if err != nil {
				log.Printf("新节点获取最新区块失败: %v", err)
				continue
			}
		}

		// 5.2 遍历新区块（逐个查询）
		if newLatestBlock > currentBlock {
			node, _ := pool.CurrentNode()
			log.Printf("发现新区块，从 %d 到 %d（当前节点: %s）", currentBlock+1, newLatestBlock, node.Name)
			for blockNum := currentBlock + 1; blockNum <= newLatestBlock; blockNum++ {
				// 4.1 构建查询条件：同一个 topic + 多个合约地址
				query := ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(blockNum)),
					ToBlock:   big.NewInt(int64(blockNum)),
					Addresses: contractAddrs,
					Topics:    [][]common.Hash{{transferTopic}},
				}

				// 间隔200ms，避免限流（公共节点需要更长的间隔）
				time.Sleep(200 * time.Millisecond)

				// 5.2.2 查询日志（带重试机制）
				var logs []types.Log
				maxRetries := 3
				retrySuccess := false

				for attempt := 1; attempt <= maxRetries; attempt++ {
					logs, err = client.FilterLogs(ctx, query)
					if err == nil {
						retrySuccess = true
						break
					}

					log.Printf("查询区块 %d 日志失败（节点: %s，尝试 %d/%d）: %v", blockNum, pool.nodes[curIdx].Name, attempt, maxRetries, err)

					// 如果是限流错误，等待更长时间后重试
					if IsRateLimitErr(err) {
						waitTime := time.Duration(attempt*2) * time.Second
						log.Printf("检测到限流，等待 %v 后重试...", waitTime)
						time.Sleep(waitTime)
					}

					// 标记故障并切换节点
					pool.MarkFailure(curIdx, err)
					pool.SwitchToNext(err)
					client.Close()

					// 重新连接新节点
					client, curIdx, err = pool.DialCurrent(ctx)
					if err != nil {
						log.Printf("切换RPC后仍失败: %v", err)
						time.Sleep(1 * time.Second)
						continue
					}
				}

				if !retrySuccess {
					log.Printf("区块 %d 查询失败，已跳过", blockNum)
					continue
				}

				// 4.2 获取区块时间戳（用于落库）
				blk, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNum)))
				if err != nil {
					log.Printf("获取区块 %d 时间戳失败: %v", blockNum, err)
					continue
				}
				blockTS := blk.Time()

				// 4.3 解析转账日志
				for _, vLog := range logs {
					if len(vLog.Topics) < 3 || len(vLog.Data) == 0 {
						continue
					}

					contractAddr := strings.ToLower(vLog.Address.Hex())
					meta, ok := contractMeta[contractAddr]
					if !ok {
						// 理论上不会出现（因为 query 已限制地址），但这里防御一下
						continue
					}

					from := common.HexToAddress(vLog.Topics[1].Hex())
					to := common.HexToAddress(vLog.Topics[2].Hex())
					value := new(big.Int).SetBytes(vLog.Data)
					amountFmt := FormatUnits(value, meta.Decimals)

					// 打印转账信息：区分合约类型
					log.Printf(
						"✅ 发现转账 - 合约类型: %s, 区块: %d, 时间戳: %d, TxHash: %s, 转出: %s, 转入: %s, 金额: %s (raw=%s) (节点: %s)",
						meta.Type, blockNum, blockTS, vLog.TxHash.Hex(), from.Hex(), to.Hex(), amountFmt, value.String(), node.Name,
					)

					rec := &model.DepositRecord{
						ContractType:   string(meta.Type),
						ContractAddr:   contractAddr,
						Decimals:       meta.Decimals,
						BlockNum:       uint64(blockNum),
						BlockTimestamp: uint64(blockTS),
						TxHash:         strings.ToLower(vLog.TxHash.Hex()),
						LogIndex:       uint(vLog.Index),
						FromAddr:       strings.ToLower(from.Hex()),
						ToAddr:         strings.ToLower(to.Hex()),
						AmountRaw:      value.String(),
						Amount:         amountFmt,
					}
					if err := model.UpsertDeposit(ctx, rec); err != nil {
						log.Printf("写入MySQL失败: %v", err)
					}

					// 告警：命中转入地址 + 达到阈值
					if shouldAlert(config.DefaultAlert, meta, strings.ToLower(to.Hex()), value) {
						SendAlertWithRetry(ctx, config.DefaultAlert, AlertEvent{
							ContractType: string(meta.Type),
							ContractAddr: contractAddr,
							Amount:       amountFmt,
							FromAddr:     strings.ToLower(from.Hex()),
							ToAddr:       strings.ToLower(to.Hex()),
							TxHash:       strings.ToLower(vLog.TxHash.Hex()),
							BlockNum:     uint64(blockNum),
						})
					}
				}
			}
			// 更新当前区块
			currentBlock = newLatestBlock
		}
	}
}

func shouldAlert(cfg config.AlertConfig, meta config.ContractConfig, toAddr string, raw *big.Int) bool {
	if !cfg.Enabled {
		return false
	}

	// 1) 地址命中
	hit := false
	for _, a := range cfg.WatchToAddrs {
		if strings.ToLower(strings.TrimSpace(a)) == toAddr {
			hit = true
			break
		}
	}
	if !hit {
		return false
	}

	// 2) 阈值判断（若该合约未配置阈值，则直接触发）
	thStr, ok := cfg.Threshold[meta.Type]
	if !ok || strings.TrimSpace(thStr) == "" {
		return true
	}
	thRaw, err := ParseDecimalToInt(thStr, meta.Decimals)
	if err != nil {
		// 阈值配置错误时，不阻塞业务，仅记录并降级为不触发
		log.Printf("告警阈值解析失败 contract=%s threshold=%s err=%v", meta.Type, thStr, err)
		return false
	}
	return raw.Cmp(thRaw) >= 0
}
