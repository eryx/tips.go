package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
	//"io"
	//"io/ioutil"
	"net"
	"net/rpc"
)

type Command int
type CommandQuery string
type CommandReply string

const (
	httpPort string = "9600"
	rpcPort  string = "9601"
)

func ApiQuery(w http.ResponseWriter, r *http.Request) {

	defer func() {
		//r.Body.Close()
	}()

	//_, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//    return
	//}

	var err error

	status := make(chan int, 2)

	go func() {

		var sock *rpc.Client
		if sock, err = rpc.DialHTTP("tcp", "localhost:9601"); err != nil {
			return
		}
		defer sock.Close()

		//r  := new(CommandReply)
		var rsp CommandReply
		rs := sock.Go("Command.Query", "nil", &rsp, nil)

		select {
		case <-rs.Done:
			status <- 1
		case <-time.After(3e9):
			status <- 9
		}

		//runtime.Goexit()
		return
	}()

	for {
		select {
		case <-status:
			goto L
		case <-time.After(3e9):
			goto L
		}
	}

L:
	//io.WriteString(w, "{\"status\": \"OK\"}")
	close(status)

	return
}

func (c *Command) Query(q *CommandQuery, r *CommandReply) error {

	*r = "OK"

	return nil
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	// HTTP Service
	go func() {

		http.HandleFunc("/api/query", ApiQuery)

		s := &http.Server{
			Addr:           ":" + httpPort,
			Handler:        nil,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		s.ListenAndServe()

		return
	}()

	// RPC Service
	ln, err := net.Listen("tcp", ":"+rpcPort)
	if err != nil {
		panic(err)
	}
	// RPC Func Register
	cmd := new(Command)
	rpc.Register(cmd)
	rpc.HandleHTTP()
	go http.Serve(ln, nil)

	// Monitor
	var m runtime.MemStats
	for {
		runtime.ReadMemStats(&m)

		fmt.Printf("NumGoroutine %d; Alloc %d; TotalAlloc %d; Sys %v; NumGC %v; MemUse %v\n", runtime.NumGoroutine(), m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC, m.Mallocs-m.Frees)

		runtime.GC()
		time.Sleep(3e9)
	}
}
