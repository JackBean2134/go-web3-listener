package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用总配置
type Config struct {
	Server             ServerConfig     `yaml:"server"`
	RPC                RPCConfig        `yaml:"rpc"`
	Contracts          []ContractConfig `yaml:"contracts"`
	TransferEventTopic string           `yaml:"transfer_event_topic"`
	MySQL              MySQLConfig      `yaml:"mysql"`
	Alert              AlertConfig      `yaml:"alert"`
	Listener           ListenerConfig   `yaml:"listener"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// RPCConfig RPC配置
type RPCConfig struct {
	Type  RPCType   `yaml:"type"`
	Nodes []RPCNode `yaml:"nodes"`
}

// ListenerConfig 监听器配置
type ListenerConfig struct {
	PollInterval      time.Duration `yaml:"poll_interval"`
	RequestInterval   time.Duration `yaml:"request_interval"`
	RetryMaxAttempts  int           `yaml:"retry_max_attempts"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

var AppConfig *Config

// LoadConfig 从YAML文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	AppConfig = &config
	return &config, nil
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", config.Server.Port)
	}

	// 验证RPC配置
	if len(config.RPC.Nodes) == 0 {
		return fmt.Errorf("至少需要配置一个RPC节点")
	}

	// 验证合约配置
	if len(config.Contracts) == 0 {
		return fmt.Errorf("至少需要配置一个合约")
	}

	for i, contract := range config.Contracts {
		if contract.Address == "" {
			return fmt.Errorf("合约 %d 地址不能为空", i)
		}
		if contract.Decimals < 0 {
			return fmt.Errorf("合约 %d 小数位数不能为负数", i)
		}
	}

	// 验证MySQL配置
	if config.MySQL.Host == "" {
		return fmt.Errorf("MySQL主机不能为空")
	}
	if config.MySQL.Port <= 0 {
		return fmt.Errorf("无效的MySQL端口: %d", config.MySQL.Port)
	}
	if config.MySQL.DBName == "" {
		return fmt.Errorf("MySQL数据库名不能为空")
	}

	// 设置默认值
	if config.Server.Mode == "" {
		config.Server.Mode = "release"
	}
	if config.Listener.PollInterval == 0 {
		config.Listener.PollInterval = 10 * time.Second
	}
	if config.Listener.RequestInterval == 0 {
		config.Listener.RequestInterval = 200 * time.Millisecond
	}
	if config.Listener.RetryMaxAttempts == 0 {
		config.Listener.RetryMaxAttempts = 3
	}
	if config.Listener.ConnectionTimeout == 0 {
		config.Listener.ConnectionTimeout = 15 * time.Second
	}

	return nil
}

// PrintConfig 打印配置信息（隐藏敏感信息）
func (c *Config) PrintConfig() {
	log.Println("========== 配置信息 ==========")
	log.Printf("服务器端口: %d", c.Server.Port)
	log.Printf("服务器模式: %s", c.Server.Mode)
	log.Printf("RPC节点数量: %d", len(c.RPC.Nodes))
	log.Printf("监听合约数量: %d", len(c.Contracts))
	log.Printf("MySQL主机: %s:%d", c.MySQL.Host, c.MySQL.Port)
	log.Printf("MySQL数据库: %s", c.MySQL.DBName)
	log.Printf("告警启用: %v", c.Alert.Enabled)
	log.Printf("轮询间隔: %v", c.Listener.PollInterval)
	log.Printf("请求间隔: %v", c.Listener.RequestInterval)
	log.Println("==============================")
}
