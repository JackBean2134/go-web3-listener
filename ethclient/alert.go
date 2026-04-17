package ethclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"go-web3-listener/config"
)

type AlertEvent struct {
	ContractType string
	ContractAddr string
	Amount       string // 格式化后的金额
	FromAddr     string
	ToAddr       string
	TxHash       string
	BlockNum     uint64
}

func (e AlertEvent) Text() string {
	return fmt.Sprintf(
		"[BSC Transfer告警]\n合约类型: %s\n合约地址: %s\n金额: %s\n转出: %s\n转入: %s\nTxHash: %s\n区块号: %d",
		e.ContractType,
		e.ContractAddr,
		e.Amount,
		e.FromAddr,
		e.ToAddr,
		e.TxHash,
		e.BlockNum,
	)
}

func SendAlertWithRetry(ctx context.Context, cfg config.AlertConfig, ev AlertEvent) {
	if !cfg.Enabled {
		return
	}

	var senders []func(context.Context) error
	if strings.TrimSpace(cfg.DingTalk.WebHook) != "" {
		webhook := cfg.DingTalk.WebHook
		senders = append(senders, func(ctx context.Context) error {
			return sendDingTalk(ctx, webhook, ev.Text())
		})
	}
	if strings.TrimSpace(cfg.SMTP.Host) != "" && strings.TrimSpace(cfg.SMTP.User) != "" && len(cfg.SMTP.To) > 0 {
		smtpCfg := cfg.SMTP
		senders = append(senders, func(ctx context.Context) error {
			return sendEmail(ctx, smtpCfg, "BSC Transfer告警", ev.Text())
		})
	}

	if len(senders) == 0 {
		log.Printf("告警启用但未配置发送通道（钉钉/邮件）")
		return
	}

	for _, send := range senders {
		retry(ctx, 3, func() error { return send(ctx) })
	}
}

func retry(ctx context.Context, max int, fn func() error) {
	var lastErr error
	for i := 1; i <= max; i++ {
		if err := fn(); err == nil {
			return
		} else {
			lastErr = err
			log.Printf("告警发送失败（第%d/%d次）: %v", i, max, err)
			sleep := time.Duration(1<<uint(i-1)) * time.Second // 1s,2s,4s
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleep):
			}
		}
	}
	if lastErr != nil {
		log.Printf("告警发送最终失败: %v", lastErr)
	}
}

func sendDingTalk(ctx context.Context, webhook, content string) error {
	type dingReq struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}
	var req dingReq
	req.MsgType = "text"
	req.Text.Content = content

	b, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dingtalk http status: %s", resp.Status)
	}
	return nil
}

func sendEmail(ctx context.Context, cfg config.SMTPConfig, subject, body string) error {
	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = cfg.User
	}
	if cfg.Port == 0 {
		return errors.New("smtp port is empty")
	}

	// net/smtp 不支持显式 ctx，这里用 select 提前取消只做“软超时”
	done := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
		msg := buildMail(from, cfg.To, subject, body)
		done <- smtp.SendMail(addr, auth, from, cfg.To, []byte(msg))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		return errors.New("smtp send timeout")
	}
}

func buildMail(from string, to []string, subject, body string) string {
	var buf strings.Builder
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	return buf.String()
}

