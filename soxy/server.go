package main

import (
  "flag"
  "fmt"
  "net/http"
  "github.com/jimrhoskins/soxy/router"
  "github.com/jimrhoskins/soxy/proxy"
  "github.com/jimrhoskins/procfile"
  "log"
  "strings"
  "bufio"
  "os"
)


var (
  ip = flag.String("ip", "", "ip to bind to")
  port = flag.Int("port", 8080, "port to listen on")
  configFile = flag.String("config", "/etc/soxy", "config file location")
)

type statementFunc func (line []string)

func eachStatement (filename string, handler statementFunc) {
  f, err := os.Open(filename)
  if err != nil {
    log.Fatal(err)
  }
  defer f.Close()

  fr := bufio.NewReader(f)

  for {
    line, err := fr.ReadString('\n')
    if err != nil {
      break
    }
    line = strings.Split(line, "#")[0]
    words := strings.Fields(line)

    if len(words) > 0 {
      handler(words)
    }
  }
}

func load(path string, routes *router.Router) {
  eachStatement(path, func(tokens []string) {
    switch strings.ToUpper(tokens[0]){
    case "PROXY":
      fmt.Println("prx...", tokens[1:])
      if len(tokens) != 3 {
        log.Fatal("Proxy requires 2 arguments: host and target")
      }
      routes.Add(tokens[1], proxy.NewProxy(tokens[2]))
    case "PROCFILE":
      fmt.Println("procfile...", tokens[1:])
      routes.Add(tokens[1], procfile.NewHandler(tokens[2]))
    case "ALIAS":
      fmt.Println("ala...", tokens[1:])
    }
  })
}


func main() {
  flag.Parse()

  routes := router.NewRouter()
  load(*configFile, routes)

  http.Handle("/", routes)
  addr := fmt.Sprintf("%s:%d", *ip, *port)

  fmt.Printf("Serving on %s\n", addr)
  log.Fatal(http.ListenAndServe(addr, nil))
}
