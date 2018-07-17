package server

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"BitCoin/model"
	"strconv"
)

type D map[string]interface{}

func HandleIndex(c *gin.Context) {

	c.JSON(http.StatusOK, D{
		"exchanges": model.GetExchangesJson(),
		"users":     model.GetUsersJson(),
	})

	//c.JSON(http.StatusOK, gin.H{
	//	"exchanges": model.ExchangeList,
	//	"users":     model.UserList,
	//})
}

func HandleBuy(c *gin.Context) {
	symbol := c.Query("symbol")
	num := c.Query("num")
	username := c.Query("username")
	f, _ := strconv.ParseFloat(num, 64)

	var CurrentUser model.User
	for _, v := range model.UserList {
		if v.Name == username {
			CurrentUser = v
		}
	}
	CurrentUser.Buy(symbol, f, "limit")
}

func HandleUserDeposit(c *gin.Context) {
	username := c.Query("username")
	symbol := c.Query("symbol")
	numQ := c.Query("num")
	num, _ := strconv.ParseFloat(numQ, 64)

	var CurrentUser model.User
	for _, v := range model.UserList {
		if v.Name == username {
			CurrentUser = v
		}
	}

	CurrentUser.Deposit(symbol, num)
}

func HandleUserWithDraw(c *gin.Context) {
	username := c.Query("username")
	symbol := c.Query("symbol")
	numQ := c.Query("num")
	num, _ := strconv.ParseFloat(numQ, 64)

	var CurrentUser model.User
	for _, v := range model.UserList {
		if v.Name == username {
			CurrentUser = v
		}
	}

	CurrentUser.WithDraw(symbol, num)
}

func HandleUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"users": model.UserList,
	})
}
