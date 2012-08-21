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

func NewProxy (target string) *Proxy {
  p := new(Proxy)
  p.Target = target
  return p
}


func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  if r.Header.Get("Upgrade") != "websocket" {
    p.proxyHTTP(w, r)
  } else {
    p.proxyWebSocket(w, r)
  }
}


// proxyHTTP will proxy the http.Request r to the new hos
// This does modify r in the process of proxying the request
func (p *Proxy) proxyHTTP(w http.ResponseWriter, r *http.Request){

  r.Header.Add("X-Forwarded-Host", r.Host)
  r.Header.Add("X-Forwarded-For", r.RemoteAddr)

  r.Host = p.Target
  r.URL.Host = p.Target

  r.URL.Scheme = "http"
  r.RequestURI = ""


  resp, err := p.Client.Do(r)
  if err != nil {
    fmt.Println(err)
    return
  }


  for key, _ := range resp.Header {
    w.Header().Set(key, resp.Header.Get(key))
  }
  w.WriteHeader(resp.StatusCode)

  io.Copy(w, resp.Body)
}

// Couple creates a full-duplex connection betwee 2 io.ReadWriters
// It returns after both ends read an EOF or other error
func Couple(a, b io.ReadWriter) {
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

  // Write recieved header to server
  fmt.Fprintf(server, "%s %s %s\r\n",r.Method, r.URL.RequestURI(), r.Proto)
  r.Header.Write(server)
  fmt.Fprintf(server, "\r\n\r\n")

  Couple(client, server)

}

type req struct{
  W http.ResponseWriter
  R *http.Request
  done chan bool
}

func request(w http.ResponseWriter, r *http.Request) *req {
  return &req{w, r, make(chan bool)}
}

type LoadBalancer struct {
  handlers []http.Handler
  requests chan *req
}

func NewLoadBalancer(h ...http.Handler) *LoadBalancer {
  r := &LoadBalancer{
    handlers: h,
    requests: make(chan *req),
  }

  go func(){
    for {
      for _, handler := range r.handlers {
        request := <-r.requests
        go func (h http.Handler){
          h.ServeHTTP(request.W, request.R)
          request.done <-true
        }(handler)
      }
    }
  }()

  return r
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  req := request(w, r)
  lb.requests <- req
  <-req.done
}


