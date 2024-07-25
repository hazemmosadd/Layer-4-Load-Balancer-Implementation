package main

import (
    "flag"
    "fmt"
    "math/rand"
    "net"
    "os"
    "sync"
)

var (
    servers     []string
    nextServer  int
    lock        sync.Mutex
    selectServer func() string // this is like a pointer to a function 
    algoFlag    string
)

// the connection that will be in the queue to be handled 
type Job struct {
    conn net.Conn
}

func init() {
    flag.StringVar(&algoFlag, "algo", "roundrobin", "Load balancing algorithm ('roundrobin' or 'random')")
    initializeServers()
}

func main() {
    jobQueue := make(chan Job, 1000)
    flag.Parse()

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

    fmt.Println("initializing the workers pool ")

    for i := 0 ; i < 3 ; i++ {
        go worker(i , jobQueue) ; 
    }


    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        fmt.Println("New connection from", conn.RemoteAddr().String())

        // this was the old setup where a new go routine is created for every connection 
        // go handleClient(conn)
        // now we will inser the connection in the channel (channel is a thread safe place for all the the concurrent go routines to access )
        jobQueue <- Job{conn: conn} // Send job to channel

    }
}

func initializeServers() {
    // Example server addresses
    //TODO: make the servers dynamic in run time 
    servers = []string{"127.0.0.1:4001", "127.0.0.1:4000", "127.0.0.1:4002"}
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

// this is a worker go routine 
// it handles clients by reading from the thread safe channel 
// so we are sure that the number of go routines we have dong go really large 
func worker(id int , jobQueue chan Job){
    for job := range jobQueue{
        handleClient(job.conn)
    }
}




// ALGORITHMS 
// next server is a shared variable so we need to handle it correctly to prevent race conditions 
func getNextServerRoundRobin() string {
    lock.Lock()
    server := servers[nextServer]
    nextServer = (nextServer + 1) % len(servers)
    lock.Unlock()
    return server
}


// The lock here is not nessecary for now (servers are static ) 
func getNextServerRandom() string {
    lock.Lock()
    server := servers[rand.Intn(len(servers))]
    lock.Unlock()
    return server
}
