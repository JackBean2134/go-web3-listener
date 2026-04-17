package main

import (
	"log"
	"net/http"

	"go-web3-listener/config"
	"go-web3-listener/ethclient"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin为发布模式（去掉调试日志）
	gin.SetMode(gin.ReleaseMode)

	// 初始化Gin路由
	r := gin.Default()
	// 去掉代理安全警告
	r.SetTrustedProxies([]string{})

	// 启动USDT转账监听（关键：调用正确的函数名）
	go ethclient.ListenUSDTTransfers(config.RpcUrl)

	// 定义接口：查询地址的USDT转账（示例）
	r.GET("/deposit/address/:addr", func(c *gin.Context) {
		addr := c.Param("addr")
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data":    addr,
		})
	})

	// 启动HTTP服务
	log.Println("HTTP服务启动成功，监听端口: 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("HTTP服务启动失败: %v", err)
	}
}
