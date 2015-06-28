package main

import (
  "os"
  "fmt"
  "log"
  "net"
  "bufio"
)

func main() {
  fmt.Println("Starting client...")
  conn, err := net.Dial("tcp", "localhost:8080")
  if err != nil {
    log.Fatal("Connection error: ", err)
  }
  server_reply := bufio.NewReader(conn)
  server_writer := bufio.NewWriter(conn)
  input := bufio.NewReader(os.Stdin)
  for {
    fmt.Print("Enter message: ")
    in, _ := input.ReadString('\n')
    _, err = server_writer.WriteString(in)
    if err != nil {
      log.Fatal("Error writing to connection: ", err)
    }
    server_writer.Flush()

    reply, err := server_reply.ReadString('\n')
    if err != nil {
      conn.Close()
      log.Fatal("Error getting server response: ", err)
    }
    fmt.Print(reply)

  }
}
