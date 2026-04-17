package main

import (
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
	// 设置Gin为发布模式（去掉调试日志）
	gin.SetMode(gin.ReleaseMode)

	// 初始化MySQL（若启动时不可用，监听写库/查询会自动触发重连）
	if err := model.InitDB(config.MySQLDsn); err != nil {
		log.Printf("MySQL初始化失败（将继续运行）: %v", err)
	}

	// 初始化Gin路由
	r := gin.Default()
	// 去掉代理安全警告
	r.SetTrustedProxies([]string{})

	// 启动USDT转账监听（关键：调用正确的函数名）
	go ethclient.ListenUSDTTransfers(config.RpcUrl)

	// 原有接口示例：按地址查询（保留）
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
	// GET /deposit/contract/:contractAddr/list?page=1&size=10
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
	log.Println("HTTP服务启动成功，监听端口: 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("HTTP服务启动失败: %v", err)
	}
}
