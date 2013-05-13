package main

import (
    "fmt"
    "time"
)

type Pool struct {
    available int
    free      chan *GenConn
}

type GenConn struct {
    // Network Connection pool
    // Database Connection pool
    // Workflow pool
    // ...
}

func NewGenConn() (conn *GenConn, err error) {
    c := new(GenConn)
    return c, nil
}

func (c *GenConn) Call(i int) (ret int) {
    return i
}

func (this *Pool) push(c *GenConn) {
    this.free <- c
}

func (this *Pool) pull() (c *GenConn, err error) {
    return <-this.free, nil
}

func main() {

    pl := new(Pool)
    pl.available = 200
    pl.free = make(chan *GenConn, pl.available)
    for i := 0; i < pl.available; i++ {
        c, _ := NewGenConn()
        pl.free <- c
    }

    maxrequest := 5000000
    status := make(chan int, 2)
    start := time.Now()

    for i := 1; i <= maxrequest; i++ {

        conn, _ := pl.pull()

        go func(i int, conn *GenConn, pl *Pool) {

            defer pl.push(conn)

            if ret := conn.Call(i); ret == maxrequest {
                status <- 1
            }

        }(i, conn, pl)
    }

    select {
    case <-status:
        fmt.Printf("Executed %v in %v\n", maxrequest, time.Since(start))
    case <-time.After(60e9):
        fmt.Printf("Timeout %v\n", int(time.Since(start)))
    }
}
