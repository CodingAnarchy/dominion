package main

import(
  "log"
  "fmt"
  "net"
  "bufio"
  "strings"
)

var domains map[string]net.IP
var client int

func handleConnection(conn net.Conn) {
  loc_client := client
  client++
  reader := bufio.NewReader(conn)
  writer := bufio.NewWriter(conn)
  var ip string
  fmt.Println("Reading from connection...")
  for {
    msg, err := reader.ReadString('\n')
    if err != nil {
      log.Println("Error receiving from client", loc_client, ": ", err)
      return
    }
    msg = strings.TrimSuffix(msg, "\n")
    if addr, ok := domains[msg]; ok {
      ip = addr.String()
    } else {
      ip = "Domain not found."
    }
    writer.WriteString(fmt.Sprint(ip, "\n"))
    writer.Flush()
  }
}

func main() {
  fmt.Println("Server starting...")
  client = 1
  fmt.Println("Creating map of domains to IP addresses...")
  domains = make(map[string]net.IP)
  domains["www.google.com"] = net.ParseIP("74.125.224.72")
  domains["www.facebook.com"] = net.ParseIP("69.63.176.13")
  domains["example.com"] = net.ParseIP("93.184.216.119")
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
    fmt.Println("Connection accepted to client", client, "...")
    go handleConnection(conn)
  }
}
