package ethclient

import (
	"context"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 强制使用IPV4（解决IPV6连接问题）
func init() {
	// 正确代码（显式调用net包的DialContext）
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, "tcp4", addr) // 改用Dialer对象调用，避免未定义
	}
}

// USDT合约地址（BSC链）
const usdtContractAddr = "0x55d398326f99059ff775478afb9d749ad585d14e"

// 监听USDT转账（适配公共RPC限流的轮询模式）
func ListenUSDTTransfers(rpcUrl string) {
	// 连接以太坊客户端
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatalf("连接RPC失败: %v", err)
	}
	defer client.Close()

	// USDT合约地址转成common.Address
	contractAddr := common.HexToAddress(usdtContractAddr)
	// Transfer事件的Topic（固定值，USDT的Transfer事件签名）
	transferTopic := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	// 从最新区块开始轮询
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("获取最新区块失败: %v", err)
	}
	currentBlock := latestBlock
	log.Printf("开始轮询USDT转账，起始区块: %d", currentBlock)

	// 定时轮询（每10秒查一次，适配公共RPC限流）
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取当前最新区块
		newLatestBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("获取最新区块失败: %v", err)
			continue
		}

		// 如果有新区块，逐个查询（每次只查1个，避免限流）
		if newLatestBlock > currentBlock {
			log.Printf("发现新区块，从 %d 到 %d", currentBlock+1, newLatestBlock)
			// 遍历每个新区块（逐个查，加小间隔）
			for blockNum := currentBlock + 1; blockNum <= newLatestBlock; blockNum++ {
				// 构建日志查询条件（单次只查1个区块）
				query := ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(blockNum)),
					ToBlock:   big.NewInt(int64(blockNum)),
					Addresses: []common.Address{contractAddr},
					Topics:    [][]common.Hash{{transferTopic}},
				}

				// 查询前加100ms间隔，避免触发RPC限流
				time.Sleep(100 * time.Millisecond)

				// 查询日志
				logs, err := client.FilterLogs(context.Background(), query)
				if err != nil {
					log.Printf("查询区块 %d 日志失败: %v", blockNum, err)
					continue
				}

				// 解析每一条Transfer日志
				for _, vLog := range logs {
					// 解析转账人（from）、接收人（to）、金额（value）
					from := common.HexToAddress(vLog.Topics[1].Hex())
					to := common.HexToAddress(vLog.Topics[2].Hex())
					value := new(big.Int).SetBytes(vLog.Data)

					// 打印转账信息（这里可以改成写入MySQL）
					log.Printf(
						"✅ 发现USDT转账 - 区块: %d, 交易哈希: %s, 转出: %s, 转入: %s, 金额: %s",
						blockNum,
						vLog.TxHash.Hex(),
						from.Hex(),
						to.Hex(),
						value.String(),
					)
				}
			}
			// 更新当前区块
			currentBlock = newLatestBlock
		}
	}
}
