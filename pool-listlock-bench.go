package main

import (
    "container/list"
    "sync"
    "time"
    "fmt"
)

type Pool struct {
    available int
    free      list.List
    lock      sync.Mutex
    emptyCond *sync.Cond
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
    this.lock.Lock()
    defer this.lock.Unlock()

    this.free.PushFront(c)

    this.available++
    this.emptyCond.Signal()
}

func (this *Pool) pull() (c *GenConn, err error) {
    
    this.lock.Lock()
    defer this.lock.Unlock()

    for this.available == 0 {
        this.emptyCond.Wait()
    }

    if this.free.Len() > 0 {
        c, _ = this.free.Remove(this.free.Back()).(*GenConn)
    } else {
        if c, err = NewGenConn(); err != nil {
            this.emptyCond.Signal()
            return nil, err
        }
    }

    this.available--
    return c, nil
}

func main() {
    
    pl := new(Pool)
    pl.available = 200
    pl.emptyCond = sync.NewCond(&pl.lock)

    maxrequest  := 5000000
    status      := make(chan int, 2)
    start       := time.Now()

    for i := 1; i <= maxrequest; i++ {
        
        conn, _ := pl.pull()
        
        go func(i int, conn *GenConn, pl *Pool) {
            
            defer pl.push(conn)
            
            if ret := conn.Call(i); ret == maxrequest {
                status <-1
            }
                   
        }(i, conn, pl)
    }
    
    select {
    case <-status:
        fmt.Printf("Executed %v in %v\n", maxrequest, time.Since(start))
    case <-time.After(60e9):
        fmt.Println("Timeout", time.Since(start))
    }
}
