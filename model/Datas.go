package model

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net/http"
	"bytes"
	"time"
	"BitCoin/utils"
)

//var SymbolList = []string{
//	"bcnbtc", "dashbtc", "dogebtc",
//	"dshbtc", "emcbtc", "ethbtc",
//	"fcnbtc", "lskbtc", "ltcbtc",
//	"nxtbtc", "qcnbtc", "sbdbtc",
//	"scbtc", "steembtc", "xdnbtc",
//	"xembtc", "xmrbtc", "ardrbtc",
//	"zecbtc", "wavesbtc", "maidbtc",
//	"ampbtc", "busbtc", "dgdbtc",
//	"icnbtc", "snglsbtc", "1stbtc",
//	"trstbtc", "timebtc", "gnobtc",
//	"repbtc", "zrcbtc", "bosbtc",
//	"dctbtc", "antbtc", "aeonbtc",
//	"gupbtc", "plubtc", "lunbtc",
//	"taasbtc", "nxcbtc", "edgbtc",
//	"rlcbtc", "swtbtc", "tknbtc",
//	"wingsbtc", "xaurbtc", "aebtc",
//	"ptoybtc", "etcbtc", "cfibtc",
//	"plbtbtc", "bntbtc", "xdncobtc",
//	"snmbtc", "xtzbtc", "dicebtc",
//	"xrpbtc", "stratbtc", "eosbtc",
//	"mnebtc", "dltbtc", "qaubtc",
//	"dntbtc", "fypbtc", "optbtc",
//	"iftbtc", "stxbtc", "tntbtc",
//	"catbtc", "bchbtc", "sncbtc",
//	"oaxbtc", "zrxbtc", "rvtbtc",
//	"icosbtc", "ppcbtc", "veribtc",
//	"prgbtc", "bmcbtc", "cndbtc",
//	"skinbtc", "emgobtc", "funbtc",
//	"hvnbtc", "fuelbtc", "poebtc",
//	"mcapbtc", "airbtc", "ambbtc",
//	"ntobtc", "icobtc", "pingbtc",
//	"gamebtc", "hpcbtc", "mthbtc",
//	"wmgobtc", "lrcbtc", "icxbtc",
//	"neobtc", "csnobtc", "ormebtc",
//	"pixbtc", "kickbtc", "yoyowbtc",
//	"mipsbtc", "cdtbtc", "xvgbtc",
//	"dgbbtc", "dcnbtc", "latbtc",
//	"vibebtc", "voisebtc", "enjbtc",
//	"zscbtc", "etbsbtc", "trxbtc",
//	"venbtc", "artbtc", "evxbtc",
//	"ebtcoldbtc", "bkbbtc", "exnbtc",
//	"tgtbtc", "ctrbtc", "bmtbtc",
//	"subbtc", "wtcbtc", "cnxbtc",
//	"atbbtc", "odnbtc", "btmbtc",
//	"b2xbtc", "atmbtc", "lifebtc",
//	"vibbtc", "omgbtc", "paybtc",
//	"cossbtc", "pptbtc", "sntbtc",
//	"btgbtc", "smartbtc", "xucbtc",
//	"cldbtc", "elmbtc", "edobtc",
//	"pollbtc", "ixtbtc", "atsbtc",
//	"sclbtc", "atlbtc", "ebtcnewbtc",
//	"etpbtc", "otxbtc", "drpubtc",
//	"neblbtc", "hacbtc", "ctxbtc",
//	"elebtc", "arnbtc", "sisabtc",
//	"stubtc", "indibtc", "btxbtc",
//	"plrbtc", "surbtc", "bqxbtc",
//	"itsbtc", "ammbtc", "dbixbtc",
//	"prebtc", "kbrbtc", "tbtbtc",
//	"erobtc", "smsbtc", "zapbtc",
//	"dovbtc", "frdbtc", "otnbtc",
//	"hsrbtc", "lendbtc", "sbtcbtc",
//	"btcabtc", "wrcbtc", "locbtc",
//	"swftcbtc", "stormbtc", "dimbtc",
//	"ngcbtc", "mcobtc", "manabtc",
//	"echbtc", "databtc", "uttbtc",
//	"kmdbtc", "qtumbtc", "ekobtc",
//	"adxbtc", "tiobtc", "waxbtc",
//	"eetbtc", "c20btc", "idhbtc",
//	"iplbtc", "covbtc", "sentbtc",
//	"smtbtc", "casbtc", "chatbtc",
//	"grmdbtc", "avhbtc", "pclbtc",
//	"cloutbtc", "utkbtc", "chsbbtc",
//	"neubtc", "taubtc", "mekbtc",
//	"barbtc", "flpbtc", "rbtc",
//	"pktbtc", "wlkbtc", "evnbtc",
//	"cpgbtc", "bptnbtc", "betrbtc",
//	"arctbtc", "dbetbtc", "bezbtc",
//	"ctebtc", "utnpbtc", "cpybtc",
//	"bcptbtc", "actbtc", "adabtc",
//	"sigbtc", "rpmbtc", "mtxbtc",
//	"bggbtc", "wizbtc", "dadibtc",
//	"datxbtc", "truebtc", "drgbtc",
//	"bancabtc", "autobtc", "noahbtc",
//	"socbtc", "wildbtc", "insurbtc",
//	"ocnbtc", "stqbtc", "xlmbtc",
//	"iotabtc", "drtbtc", "mldbtc",
//	"ertbtc", "crptbtc", "meshbtc",
//	"ihtbtc", "sccbtc", "yccbtc",
//	"danbtc", "telbtc", "bubobtc",
//	"vitbtc", "clrbtc", "nctbtc",
//	"axpbtc", "bmhbtc", "hqxbtc",
//	"ldcbtc", "xmobtc", "berrybtc",
//	"bstnbtc", "shipbtc", "lncbtc",
//	"uncbtc", "rpxbtc", "clbtc",
//	"daybtc", "daxtbtc", "fotabtc",
//	"sethbtc",
//}
var SymbolList = []string{
	"dashbtc", "etcbtc",
	"eosbtc", "omgbtc", "ethbtc",
	"xrpbtc", "zecbtc", "ltcbtc",
	"vitbtc", "clrbtc", "nctbtc",
	"axpbtc", "bmhbtc", "hqxbtc",
	"ldcbtc", "xmobtc", "berrybtc",
	"bstnbtc", "shipbtc", "lncbtc",
	"uncbtc", "rpxbtc", "clbtc",
	"daybtc", "daxtbtc", "fotabtc",
	"sethbtc",
	"nxtbtc", "qcnbtc", "sbdbtc",
	"scbtc", "steembtc", "xdnbtc",
	"xembtc", "xmrbtc", "ardrbtc",
}

var ExchangeList []BigE
var UserList []User

func init() {
	//result := getSymbols()

	//SymbolList = []string{}

	//result.ForEach(func(key, value gjson.Result) bool {
	//	SymbolList = append(SymbolList, strings.ToUpper(value.String()))
	//	fmt.Println("追加:", value.String())
	//	return true
	//})

	//huobi := NewHuoBiExchange()
	//binance := NewBinanceExchange()
	//okex := NewOkexExchange()
	//bitfinex := NewBitfinexExchange()
	//bittrex := NewBittrexExchange()
	//hitbtc := NewHitbtcExchange()
	//zb := NewZbExchange()
	//gate := NewGateIOExchange()
	//otcbtc := NewOtcBtcExchange()
	//exx := NewExxExchange()
	//bibox := NewBiboxExchange()
	//hadax := NewHadaxExchange()
	//bithumb := NewBithumbExchange()
	//coinbene := NewCoinbeneExchange()
	//bcex := NewBcexExchange()
	bitz := NewBitzExchange()

	//lbank := NewLBankExchange()
	//kucoin := NewKuCoinExchange()
	//bittrex := NewBittrexExchange()
	//poloniex := NewPoloniexExchange()
	//upbit := NewUpbitExchange()
	//fatbtc := NewFatbtcExchange()
	//allcoin := NewAllcoinExchange()
	//quoine := NewQuoineExchange()

	//ExchangeList = append(ExchangeList, huobi)
	//ExchangeList = append(ExchangeList, binance)
	//ExchangeList = append(ExchangeList, okex)
	//ExchangeList = append(ExchangeList, bitfinex)
	//ExchangeList = append(ExchangeList, bittrex)
	//ExchangeList = append(ExchangeList, hitbtc)
	//ExchangeList = append(ExchangeList, zb)
	//ExchangeList = append(ExchangeList, gate)
	//ExchangeList = append(ExchangeList, otcbtc)
	//ExchangeList = append(ExchangeList, exx)
	//ExchangeList = append(ExchangeList, bibox)
	//ExchangeList = append(ExchangeList, hadax)
	//ExchangeList = append(ExchangeList, bithumb)
	//ExchangeList = append(ExchangeList, coinbene)
	//ExchangeList = append(ExchangeList, bcex)
	ExchangeList = append(ExchangeList, bitz)

	//ExchangeList = append(ExchangeList, kucoin)
	//ExchangeList = append(ExchangeList, bittrex)
	//ExchangeList = append(ExchangeList, poloniex)
	//ExchangeList = append(ExchangeList, upbit)
	//ExchangeList = append(ExchangeList, fatbtc)
	//ExchangeList = append(ExchangeList, allcoin)
	//ExchangeList = append(ExchangeList, quoine)

	for i := 0; i < 10; i++ {
		user := User{
			Name:       fmt.Sprintf("Test%d", i),
			AmountDict: make(map[string]float64),
		}
		UserList = append(UserList, user)
	}

	for _, e := range ExchangeList {
		e.CreateRun("")
		//e.CreateRun(v)
	}
	//for _, v := range SymbolList {
	//	for _, u := range UserList {
	//		newSymbol := strings.Replace(v, "btc", "", -1)
	//		//newSymbol = strings.Replace(v, "eth", "", 1)
	//		u.AmountDict[newSymbol] = 0
	//	}
	//}
}

type D map[string]interface{}

func GetExchangesJson() []D {
	var exchangesList []D

	for _, e := range ExchangeList {
		var symbols []D
		for _, s := range SymbolList {
			var symbol = D{}
			symbol["s"] = s
			symbol["p"] = e.GetLastPrice(s)
			symbol["n"] = e.GetAmount(utils.GetCoinBySymbol(s))
			symbol["tf"] = e.GetTradeFee(s, "taker")
			symbol["mf"] = e.GetTradeFee(s, "maker")
			symbols = append(symbols, symbol)
		}
		var coins []D
		for _, s := range SymbolList {
			var coin = D{}
			coinName := utils.GetCoinBySymbol(s)
			coin["name"] = coinName
			fee := e.GetTransferFee(coinName)
			coin["fee"] = fee
			coins = append(coins, coin)
		}
		var exchange = D{
			"exchange": D{
				"name":    e.GetName(),
				"symbols": symbols,
				"coins":   coins,
			},
		}
		exchangesList = append(exchangesList, exchange)
	}
	return exchangesList
}

func GetUsersJson() []D {
	var usersList []D

	for _, u := range UserList {
		var users []D
		for _, s := range SymbolList {
			var symbol = D{}
			symbol["s"] = s[:3]
			symbol["n"] = u.AmountDict[s[:3]]
			users = append(users, symbol)
		}
		var exchange = D{
			"user": D{
				"name":    u.Name,
				"symbols": users,
			},
		}
		usersList = append(usersList, exchange)
	}
	return usersList
}

func getSymbols() gjson.Result {
	//uProxy, _ := url.Parse("http://183.51.191.251:9797")
	client := &http.Client{
		//Transport: &http.Transport{
		//	Proxy: http.ProxyURL(uProxy),
		//},
	}
	client.Timeout = time.Second * 20

	u := "https://api.hitbtc.com/api/2/public/symbol"
	resp := &http.Response{}
	err := error(nil)
	for {
		resp, err = client.Get(u)

		if err != nil {
			fmt.Println(err)
		}
	}
	buf := bytes.NewBuffer(make([]byte, 0, 512))

	buf.ReadFrom(resp.Body)
	results := gjson.GetBytes(buf.Bytes(), "#[quoteCurrency==\"BTC\"]#.id")

	return results
}
