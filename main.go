package main

import (
    "flag"
    "fmt"
    "math/rand"
    "net"
    "os"
    "sync"
    "time"
)

var (
    servers     []string
    nextServer  int
    lock        sync.Mutex
    selectServer func() string // this is like a pointer to a function 
    algoFlag    string
)

func init() {
    flag.StringVar(&algoFlag, "algo", "roundrobin", "Load balancing algorithm ('roundrobin' or 'random')")
    rand.Seed(time.Now().UnixNano())
}

func main() {
    flag.Parse()
    initializeServers()

    switch algoFlag {
    case "random":
        selectServer = getNextServerRandom
    default:
        selectServer = getNextServerRoundRobin
    }

    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Server is listening on port 8080...")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        fmt.Println("New connection from", conn.RemoteAddr().String())
        go handleClient(conn)
    }
}

func initializeServers() {
    // Example server addresses
    servers = []string{"127.0.0.1:1236", "127.0.0.1:1237", "127.0.0.1:1238"}
    nextServer = 0
}

func handleClient(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            fmt.Println("Error reading:", err.Error())
            return
        }
		// client part 
        fmt.Printf("Received from client: %s says: %s\n", conn.RemoteAddr().String(), string(buffer[:n]))

        serverAddr := selectServer()

		// here we are sending the bytes we recieved to one of the servers we handling 
        serverConn, err := net.Dial("tcp", serverAddr)
        if err != nil {
            fmt.Println("Error connecting to server:", err.Error())
            return
        }
        defer serverConn.Close()

        _, err = serverConn.Write(buffer[:n])
        if err != nil {
            fmt.Println("Error sending to server:", err.Error())
            return
        }

        reply := make([]byte, 1024)
        n, err = serverConn.Read(reply)
        if err != nil {
            fmt.Println("Error receiving from server:", err.Error())
            return
        }

		// client part 
        _, err = conn.Write(reply[:n])
        if err != nil {
            fmt.Println("Error sending to client:", err.Error())
        }
    }
}


// next server is a shared variable so we need to handle it correctly to prevent race conditions 

func getNextServerRoundRobin() string {
    lock.Lock()
    server := servers[nextServer]
    nextServer = (nextServer + 1) % len(servers)
    lock.Unlock()
    return server
}


// The lock here is not nessecary for now (servers are static)
func getNextServerRandom() string {
    lock.Lock()
    server := servers[rand.Intn(len(servers))]
    lock.Unlock()
    return server
}
