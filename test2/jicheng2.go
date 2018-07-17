package main

import (
	"regexp"
	"BitCoin/utils"
	"fmt"
	"strings"
)

func main() {
	s := "ok_sub_spot_eth_btc_ticker"
	var myExp = utils.MyRegexp{regexp.MustCompile(`^ok_sub_spot_(?P<coin>(\w+)*)_ticker`)}
	m := myExp.FindStringSubmatchMap(s)
	if s, ok := m["coin"]; ok {
		fmt.Println(strings.Replace(s, "_", "", -1))
	}
}
