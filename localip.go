package main

import (
    "fmt"
    "net"
    "bytes"
    "os/exec"
    "strings"
)

func main() {
    
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        panic(err)
    }
    for _, addr := range addrs {
        fmt.Println("Method 1:", addr.String())
    }
    
    var out bytes.Buffer
    cmd := "ip addr|grep inet|grep -v inet6|grep -v 127.0.|head -n1" +
        "|awk ' {print $2}'|awk -F \"/\" '{print $1}'"
    ec := exec.Command("sh", "-c", cmd)
    ec.Stdout = &out
    if err := ec.Run(); err == nil {
        fmt.Println("Method 2:", strings.TrimSpace(out.String()))
    }
}
