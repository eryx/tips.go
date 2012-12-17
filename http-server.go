package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
	//"io/ioutil"
	//"net/rpc"
	//"net"
)

const (
	httpPort string = "9600"
)

func ApiQuery(w http.ResponseWriter, r *http.Request) {

	//defer func() {
	//r.Body.Close()
	//}()

	//_, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//    return
	//}

	io.WriteString(w, "{\"status\": \"OK\"}")

	return
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

	// Monitor
	var m runtime.MemStats
	for {
		runtime.ReadMemStats(&m)

		fmt.Printf("NumGoroutine %d; Alloc %d; TotalAlloc %d; Sys %v; NumGC %v; MemUse %v\n", runtime.NumGoroutine(), m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC, m.Mallocs-m.Frees)

		//runtime.GC()
		time.Sleep(3e9)
	}
}
