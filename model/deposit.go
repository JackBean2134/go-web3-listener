package model

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DepositRecord USDT Transfer 转账记录（监听写入）。
type DepositRecord struct {
	ID             uint           `gorm:"primaryKey"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	BlockNum       uint64         `gorm:"not null;index"`
	BlockTimestamp uint64         `gorm:"not null;index"` // 区块时间戳（秒）
	TxHash         string         `gorm:"not null;size:80;index;uniqueIndex:uniq_tx_log"`
	LogIndex       uint           `gorm:"not null;uniqueIndex:uniq_tx_log"` // 同一Tx可能包含多条Transfer，用于去重
	FromAddr       string         `gorm:"not null;size:60;index"`
	ToAddr         string         `gorm:"not null;size:60;index"`
	Amount         string         `gorm:"not null;size:80"` // 字符串存大数（USDT 18位）
}

func (DepositRecord) TableName() string { return "deposit_records" }

func (r *DepositRecord) Validate() error {
	if r == nil {
		return errors.New("nil record")
	}
	if r.BlockNum == 0 || r.BlockTimestamp == 0 {
		return errors.New("empty block number/timestamp")
	}
	if strings.TrimSpace(r.TxHash) == "" ||
		strings.TrimSpace(r.FromAddr) == "" ||
		strings.TrimSpace(r.ToAddr) == "" ||
		strings.TrimSpace(r.Amount) == "" {
		return errors.New("empty tx/from/to/amount")
	}
	return nil
}

// UpsertDeposit 写入转账记录；使用 (tx_hash, log_index) 做幂等去重。
// 若遇到断连等连接类错误，会尝试重连并重试一次。
func UpsertDeposit(ctx context.Context, r *DepositRecord) error {
	if err := r.Validate(); err != nil {
		return err
	}

	if err := EnsureDB(); err != nil {
		return err
	}

	err := DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tx_hash"}, {Name: "log_index"}},
			DoNothing: true,
		}).
		Create(r).Error
	if err == nil {
		return nil
	}

	if IsConnErr(err) {
		if e := EnsureDB(); e != nil {
			return e
		}
		return DB.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "tx_hash"}, {Name: "log_index"}},
				DoNothing: true,
			}).
			Create(r).Error
	}
	return err
}

type DepositListResult struct {
	Total int64           `json:"total"`
	List  []DepositRecord `json:"list"`
}

// ListDepositsByAddr 查询指定地址相关转账（from/to 任一匹配），支持分页。
func ListDepositsByAddr(ctx context.Context, addr string, page, size int) (DepositListResult, error) {
	var res DepositListResult
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return res, errors.New("addr is empty")
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	if err := EnsureDB(); err != nil {
		return res, err
	}

	q := DB.WithContext(ctx).Model(&DepositRecord{}).
		Where("from_addr = ? OR to_addr = ?", addr, addr)

	if err := q.Count(&res.Total).Error; err != nil {
		if IsConnErr(err) {
			if e := EnsureDB(); e != nil {
				return res, e
			}
			if err2 := q.Count(&res.Total).Error; err2 != nil {
				return res, err2
			}
		} else {
			return res, err
		}
	}

	offset := (page - 1) * size
	if err := q.Order("block_num desc, id desc").Offset(offset).Limit(size).Find(&res.List).Error; err != nil {
		if IsConnErr(err) {
			if e := EnsureDB(); e != nil {
				return res, e
			}
			return res, q.Order("block_num desc, id desc").Offset(offset).Limit(size).Find(&res.List).Error
		}
		return res, err
	}
	return res, nil
}