package main

import(
  "log"
  "fmt"
  "net"
  "bufio"
)

func handleConnection(conn net.Conn) (bool) {
  reader := bufio.NewReader(conn)
  writer := bufio.NewWriter(conn)
  fmt.Println("Reading from connection...")
  for {
    msg, err := reader.ReadString('\n')
    if err != nil {
      log.Fatal("Error receieving from client: ", err)
    }
    fmt.Println(msg)
    writer.WriteString(msg)
  }
}

func main() {
  fmt.Println("Server starting...")
  listener, err := net.Listen("tcp", ":8080")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println("Listening on TCP port 8080...")
  for {
    conn, err := listener.Accept()
    if err != nil {
      log.Fatal("Error accepting connection: ", err)
    }
    fmt.Println("Connection accepted...")
    go handleConnection(conn)
  }
}
