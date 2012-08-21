package router

import (
  "net/http"
  "strings"
)

type Router struct {
  routes map[string] http.Handler
  aliases map[string] string
  
  // DefaultHandler is used when no route matches host
  DefaultHandler http.Handler

}

type DefaultErrorHandler struct{}

func (_ *DefaultErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
  http.Error(w, "Could not find route!", http.StatusInternalServerError)
}

func New () *Router {
  r := new(Router)
  r.routes = make(map[string] http.Handler)
  r.aliases = make(map[string] string)
  r.DefaultHandler = new(DefaultErrorHandler)
  return r
}

func (self *Router) ServeHTTP (w http.ResponseWriter, r *http.Request) {
  host := strings.Split(r.Host, ":")[0]
  host = strings.ToLower(host)

  handler := self.getHandler(host)
  handler.ServeHTTP(w, r)
}

func (self *Router) getHandler (host string) http.Handler {
  handler, ok := self.routes[host]
  if ok {
    return handler
  }

  alias, ok := self.aliases[host]
  if ok {
    return self.getHandler(alias)
  }

  return self.DefaultHandler
}


func (self *Router) Add (host string, handler http.Handler) {
  host = strings.ToLower(host)
  self.routes[host] = handler
}

func (self *Router) Alias (host, to string) {
  host = strings.ToLower(host)
  to = strings.ToLower(to)

  self.aliases[host] = to
}



