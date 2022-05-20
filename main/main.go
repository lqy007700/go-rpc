package main

import (
	"fmt"
	gorpc "go-rpc"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}

	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	gorpc.Accept(l)
}

func main() {
	quit := make(chan struct{})
	go checkGoroutine(quit)

	time.Sleep(time.Second * 5)

	quit <- struct {}{}

	//log.SetFlags(0)
	//addr := make(chan string)
	//go startServer(addr)
	//client, _ := gorpc.Dial("tcp", <-addr)
	//defer func() { _ = client.Close() }()
	//
	//time.Sleep(time.Second)
	//// send request & receive response
	//var wg sync.WaitGroup
	//for i := 0; i < 5; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		args := fmt.Sprintf("geerpc req %d", i)
	//		var reply string
	//		if err := client.Call("Foo.Sum", args, &reply); err != nil {
	//			log.Fatal("call Foo.Sum error:", err)
	//		}
	//		log.Printf("reply:%+v", reply)
	//	}(i)
	//}
	//wg.Wait()
}

func checkGoroutine(done chan struct{}) {

	for {
		select {
		case <-done:
			return
		default:
		}
		time.Sleep(time.Second)
		fmt.Println(1)
	}
}
