package ethclient

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"

	"go-web3-listener/config"
)

type nodeStatus struct {
	healthy      bool
	lastError    string
	lastChecked  time.Time
	failCount    int
	lastFailTime time.Time
}

// RPCPool 管理多个 RPC 节点，提供健康检查与故障自动切换。
// 仅负责“连接层”，不包含业务订阅/解析逻辑。
type RPCPool struct {
	typ   config.RPCType
	nodes []config.RPCNode

	mu      sync.RWMutex
	status  []nodeStatus
	current int
}

func NewRPCPool(typ config.RPCType, nodes []config.RPCNode) (*RPCPool, error) {
	var filtered []config.RPCNode
	for _, n := range nodes {
		if n.Type == typ && strings.TrimSpace(n.URL) != "" {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("rpc nodes is empty")
	}

	p := &RPCPool{
		typ:     typ,
		nodes:   filtered,
		status:  make([]nodeStatus, len(filtered)),
		current: 0,
	}
	// 默认认为“未知=健康”，让业务先跑起来；健康检查会持续校正
	for i := range p.status {
		p.status[i].healthy = true
	}
	return p, nil
}

func (p *RPCPool) CurrentNode() (config.RPCNode, int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.nodes[p.current], p.current
}

func (p *RPCPool) MarkFailure(idx int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if idx < 0 || idx >= len(p.status) {
		return
	}
	p.status[idx].healthy = false
	p.status[idx].failCount++
	p.status[idx].lastFailTime = time.Now()
	if err != nil {
		p.status[idx].lastError = err.Error()
	}
}

func (p *RPCPool) MarkHealthy(idx int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if idx < 0 || idx >= len(p.status) {
		return
	}
	p.status[idx].healthy = true
	p.status[idx].lastError = ""
	p.status[idx].lastChecked = time.Now()
	p.status[idx].failCount = 0
}

func (p *RPCPool) pickNextHealthyLocked() int {
	n := len(p.nodes)
	if n == 0 {
		return -1
	}
	for step := 1; step <= n; step++ {
		i := (p.current + step) % n
		if p.status[i].healthy {
			return i
		}
	}
	return (p.current + 1) % n
}

// SwitchToNext 切换到下一个可用节点，并记录切换日志。
func (p *RPCPool) SwitchToNext(reason error) config.RPCNode {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldIdx := p.current
	newIdx := p.pickNextHealthyLocked()
	if newIdx < 0 || newIdx == oldIdx {
		return p.nodes[oldIdx]
	}
	p.current = newIdx

	oldNode := p.nodes[oldIdx]
	newNode := p.nodes[newIdx]
	if reason != nil {
		log.Printf("RPC切换：%s(%s) -> %s(%s)，原因：%v", oldNode.Name, oldNode.URL, newNode.Name, newNode.URL, reason)
	} else {
		log.Printf("RPC切换：%s(%s) -> %s(%s)", oldNode.Name, oldNode.URL, newNode.Name, newNode.URL)
	}
	return newNode
}

func (p *RPCPool) DialCurrent(ctx context.Context) (*ethclient.Client, int, error) {
	node, idx := p.CurrentNode()
	c, err := dialWithTimeout(ctx, node.URL)
	if err == nil {
		return c, idx, nil
	}
	p.MarkFailure(idx, err)
	p.SwitchToNext(err)

	node2, idx2 := p.CurrentNode()
	c2, err2 := dialWithTimeout(ctx, node2.URL)
	if err2 != nil {
		p.MarkFailure(idx2, err2)
		return nil, idx2, err2
	}
	return c2, idx2, nil
}

func dialWithTimeout(ctx context.Context, url string) (*ethclient.Client, error) {
	dctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	return ethclient.DialContext(dctx, url)
}

// StartHealthCheck 定时健康检查（BlockNumber 探活）。
func (p *RPCPool) StartHealthCheck(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.checkOnce(ctx)
			}
		}
	}()
}

func (p *RPCPool) checkOnce(ctx context.Context) {
	for i, node := range p.nodes {
		dctx, cancel := context.WithTimeout(ctx, 6*time.Second)
		c, err := ethclient.DialContext(dctx, node.URL)
		cancel()
		if err != nil {
			p.MarkFailure(i, err)
			continue
		}

		bctx, cancel2 := context.WithTimeout(ctx, 6*time.Second)
		_, err = c.BlockNumber(bctx)
		cancel2()
		c.Close()

		if err != nil {
			p.MarkFailure(i, err)
			continue
		}
		p.MarkHealthy(i)
	}
}

// IsRateLimitErr 判断是否是RPC限流错误。
func IsRateLimitErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "429") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "too many requests") ||
		strings.Contains(s, "limit exceeded") ||
		strings.Contains(s, "request rate exceeded")
}