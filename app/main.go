package main

import (
	"github.com/gin-gonic/gin"
	"BitCoin/server"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/", server.HandleIndex)
	r.GET("/buy", server.HandleBuy)
	r.GET("/deposit", server.HandleUserDeposit)
	r.GET("/withdraw", server.HandleUserWithDraw)
	r.GET("/users", server.HandleUsers)

	r.Run(":8080")
}
