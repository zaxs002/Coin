package main

import (
	"BitCoin/utils"
	"fmt"
	"strings"
)

func main() {
	s := "5,000,000BTS"

	s = strings.Replace(s, ",", "", -1)
	m := utils.GetByZb(s)
	fmt.Println(m)
}
