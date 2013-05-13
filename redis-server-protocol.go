package main

import (
    "crypto/rand"
    "fmt"
    "io"
    "net"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "time"
)

const AGENT_IOBUF_LEN = 32
const AGENT_INLINE_MAX_SIZE = 1024 * 64 // Max size of inline reads
const AGENT_TIMEOUT = 3e9
const AGENT_QUIT = 10
const AGENT_NET_PORT = "9051"

type Agent struct {
    clients map[string]*AgentClient

    Lock sync.Mutex

    ln  net.Listener
}

type AgentClient struct {
    Sig chan int
    //Rep       *Reply
    WatchPath string
    Querybuf  []byte
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    _ = NewAgent(AGENT_NET_PORT)

    for {
        time.Sleep(10e9)
    }
}

func NewAgent(port string) *Agent {

    this := new(Agent)
    this.clients = map[string]*AgentClient{}
    var err error

    go func() {

        if this.ln, err = net.Listen("tcp", ":"+port); err != nil {
            // TODO
        }

        for {
            conn, err := this.ln.Accept()
            if err != nil {
                // handle error
                continue
            }
            go this.Handler(conn)
        }

    }()

    return this
}

func (this *Agent) Handler(conn net.Conn) {

    sid := NewRandString(16)

    c := new(AgentClient)
    c.Sig = make(chan int, 4)
    //c.Rep = new(Reply)
    c.Querybuf = []byte{}

    this.clients[sid] = c

    defer func() {
        conn.Close()
        this.Lock.Lock()
        delete(this.clients, sid)
        this.Lock.Unlock()
    }()

    multiBulkLen := 0
    bulkLen := -1
    pos := 0

    argc := 0
    argv := map[int][]byte{}

    for {

        var buf [AGENT_IOBUF_LEN]byte
        n, err := conn.Read(buf[0:])

        if err != nil {
            return
        }
        if n > 0 {
            c.Querybuf = append(c.Querybuf, buf[0:n]...)
        }
        n = len(c.Querybuf)
        //fmt.Println("Query Buffer", len(c.Querybuf), "[", string(c.Querybuf), "]")

        // Process Multibulk Buffer
        if multiBulkLen == 0 {
            // Multi bulk length cannot be read without a \r\n
            li := strings.SplitN(string(c.Querybuf[0:n]), "\r", 2)
            if len(li) == 1 {
                // TODO "Protocol error: too big mbulk count string"
                if len(li[0]) > AGENT_INLINE_MAX_SIZE {
                    _, _ = conn.Write([]byte("-ERR\r\n"))
                }
                return // TODO
            }

            // Buffer should also contain \n
            if len(li[1]) < 1 || li[1][0] != 10 {
                return // TODO
            }

            // We know for sure there is a whole line since newline != NULL,
            // so go ahead and find out the multi bulk length.
            if c.Querybuf[0] != []byte("*")[0] {
                return // TODO
            }
            // multi bulk length can not be empty
            if len(li[0]) < 2 {
                return // TODO
            }
            //
            mblen, err := strconv.Atoi(li[0][1:])
            if err != nil || mblen > 1024*1024 {
                return // TODO "Protocol error: invalid multibulk length"
            }

            multiBulkLen = mblen
            pos = len(li[0]) + 2

            // Reset all
            argc = 0
            argv = map[int][]byte{}
            c.WatchPath = ""
        }

        for {
            // Read bulk length if unknown
            if bulkLen == -1 {

                li := strings.SplitN(string(c.Querybuf[pos:]), "\r", 2)
                if len(li) == 1 {
                    if len(li[0]) > AGENT_INLINE_MAX_SIZE {
                        // "Protocol error: too big bulk count string"
                        _, _ = conn.Write([]byte("-ERR\r\n"))
                    }
                    break // TODO
                }

                // Buffer should also contain \n
                if len(li[1]) < 1 || li[1][0] != 10 {
                    break // TODO
                }

                if c.Querybuf[pos] != []byte("$")[0] {
                    return // TODO
                }

                lis, err := strconv.Atoi(li[0][1:])
                if err != nil || lis < 0 || lis > 512*1024*1024 {
                    return // TODO "Protocol error: invalid bulk length"
                }

                pos += len(li[0]) + 2
                bulkLen = lis
            }

            /* Read bulk argument */
            if n-pos < bulkLen+2 {
                // Not enough data (+2 == trailing \r\n)
                break
            } else {

                argv[argc] = c.Querybuf[pos : pos+bulkLen]
                argc++

                pos += bulkLen + 2
                bulkLen = -1
                multiBulkLen--
            }

            if multiBulkLen <= 0 {
                //fmt.Println("multi bulk len END", len(cmd.Argv))
                break
            }
        }

        // RPC: Process Command
        if multiBulkLen == 0 && argc > 0 {

            c.Querybuf = c.Querybuf[pos:n]

            // fmt.Println("Agent DONE Buffer", sid, pos, len(c.Querybuf), string(c.Querybuf[0:pos]), string(c.Querybuf[pos:]))

            fmt.Println("req", argv)

            rsp := "+OK\r\n"
            _, _ = conn.Write([]byte(rsp))

        }

    }

    return
}

func NewRandString(len int) string {

    u := make([]byte, len/2)

    // Reader is a global, shared instance of a cryptographically strong pseudo-random generator.
    // On Unix-like systems, Reader reads from /dev/urandom.
    // On Windows systems, Reader uses the CryptGenRandom API.
    _, err := io.ReadFull(rand.Reader, u)
    if err != nil {
        panic(err)
    }

    return fmt.Sprintf("%x", u)
}
