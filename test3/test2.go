package main

import (
	"fmt"
	"strconv"
	"time"
)

type Pool struct {
	Queue          chan func(id int) string
	RuntimeNumber  int
	Total          int
	Result         chan string
	FinishCallback func()
}

//初始化
func (p *Pool) Init(runtimeNumber int, total int) {
	p.RuntimeNumber = runtimeNumber
	p.Total = total
	p.Queue = make(chan func(id int) string, total)
	p.Result = make(chan string, total)
}

func (p *Pool) InitWithCallBack(runtimeNumber int, total int, finishcallback func()) {
	p.RuntimeNumber = runtimeNumber
	p.Total = total
	p.Queue = make(chan func(id int) string, total)
	p.Result = make(chan string, total)
	p.FinishCallback = finishcallback
}

//这个是工作线程，处理具体的业务逻辑，将jobs中的任务取出，处理后将处理结果放置在results中。
func (p Pool) worker(id int, jobs <-chan func(id int) string, results chan<- string) {
	for j := range jobs {
		//fmt.Println("worker", id)
		time.Sleep(time.Second)
		results <- j(id)
	}
}

func (p Pool) Start() {
	for w := 1; w <= p.RuntimeNumber; w++ {
		go p.worker(w, p.Queue, p.Result)
	}

	for j := 1; j <= p.Total; j++ {
		s := "http://" + strconv.Itoa(j) + "-------.com"
		p.Queue <- func(id int) string {
			fmt.Println("worker ", id, " downloading ", s)
			return s + "下载完成"
		}
	}

	close(p.Queue)

	for a := 1; a <= p.Total; a++ {
		r := <-p.Result
		fmt.Println("result: ", r)
	}

	if p.FinishCallback != nil {
		p.FinishCallback()
	}
}

func main() {
	p := Pool{}
	p.InitWithCallBack(3, 9, func() {
		fmt.Println("全部任务完成")
	})
	start := time.Now()
	p.Start()
	end := time.Now()

	fmt.Printf("一共所用时间: %.3f 秒\n", end.Sub(start).Seconds())
	fmt.Printf("%v\n", p)
	fmt.Printf("%+v\n", p)

	fmt.Printf("%#v\n", p)
}
