package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go-web3-listener/config"
	"go-web3-listener/ethclient"
	"go-web3-listener/model"

	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v\n提示：请确保 config.yaml 文件存在且格式正确", err)
	}

	// 打印配置信息
	cfg.PrintConfig()

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化MySQL
	mysqlDSN := cfg.MySQL.GetDSN()
	if err := model.InitDB(mysqlDSN, cfg.MySQL.MaxIdleConns, cfg.MySQL.MaxOpenConns,
		cfg.MySQL.ConnMaxLifetime, cfg.MySQL.ConnMaxIdleTime); err != nil {
		log.Printf("MySQL初始化失败（将继续运行）: %v", err)
	}

	// 初始化Gin路由
	r := gin.Default()
	r.SetTrustedProxies([]string{})

	// 启动USDT转账监听
	go ethclient.ListenUSDTTransfers(cfg)

	// 原有接口示例：按地址查询
	r.GET("/deposit/address/:addr", func(c *gin.Context) {
		addr := strings.ToLower(strings.TrimSpace(c.Param("addr")))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		res, err := model.ListDepositsByAddr(c.Request.Context(), addr, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": res})
	})

	// 新增：按合约查询
	r.GET("/deposit/contract/:contractAddr/list", func(c *gin.Context) {
		contractAddr := strings.ToLower(strings.TrimSpace(c.Param("contractAddr")))
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		res, err := model.ListDepositsByContract(c.Request.Context(), contractAddr, page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": res})
	})

	// 启动HTTP服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("HTTP服务启动成功，监听端口: %d", cfg.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("HTTP服务启动失败: %v", err)
	}
}
