package proxy

import (
  "io"
  "fmt"
  "net"
  "net/http"
)


type Proxy struct {
  Target string
  Client http.Client
}

func New (target string) *Proxy {
  p := new(Proxy)
  p.Target = target
  return p
}


func (self *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if r.Header.Get("Upgrade") == "websocket" {
    self.proxyWebSocket(w, r)
  } else {
    self.proxyHTTP(w, r)
  }
}


// proxyHTTP will proxy the http.Request r to the new hos
// This does modify r in the process of proxying the request
func (self *Proxy) proxyHTTP(w http.ResponseWriter, r *http.Request){

  r.Header.Add("X-Forwarded-Host", r.Host)
  r.Header.Add("X-Forwarded-For", r.RemoteAddr)

  r.Host = self.Target
  r.URL.Host = self.Target

  // Reset Request properteis for Client
  r.URL.Scheme = "http"
  r.RequestURI = ""


  resp, err := self.Client.Do(r)
  if err != nil {
    http.Error(w, BAD_GATEWAY, http.StatusBadGateway)
    return
  }


  // Copy response header
  for key, _ := range resp.Header {
    w.Header().Set(key, resp.Header.Get(key))
  }
  w.WriteHeader(resp.StatusCode)

  io.Copy(w, resp.Body)
}

// proxyWebSocket will proxy a websocket connection from r to the host
func (p *Proxy) proxyWebSocket(w http.ResponseWriter, r *http.Request) {
  hj, ok := w.(http.Hijacker)
  if !ok {
    http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
    return
  }

  client, _, err := hj.Hijack()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer client.Close()

  server, err := net.Dial("tcp", p.Target)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer server.Close()

  writeHeader(r, server)
  duplex(client, server)

}

func writeHeader(r *http.Request, target io.Writer){
  fmt.Fprintf(target, "%s %s %s\r\n",r.Method, r.URL.RequestURI(), r.Proto)
  r.Header.Write(target)
  fmt.Fprintf(target, "\r\n\r\n")
}

func duplex(a, b io.ReadWriter) {
  copydone := make(chan int, 2)

  go func () {
    io.Copy(a, b)
    copydone <- 1
  }()

  go func () {
    io.Copy(b,a)
    copydone <- 1
  }()

  <-copydone
  <-copydone
}

const BAD_GATEWAY string = `Error 502: Bad Gateway`

