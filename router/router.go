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

func NewRouter () *Router {
  r := new(Router)
  r.routes = make(map[string] http.Handler)
  r.aliases = make(map[string] string)
  r.DefaultHandler = new(DefaultErrorHandler)
  return r
}

func (router *Router) ServeHTTP (w http.ResponseWriter, r *http.Request) {
  host := strings.Split(r.Host, ":")[0]
  host = strings.ToLower(host)

  handler := router.getHandler(host)
  handler.ServeHTTP(w, r)
}

func (router *Router) getHandler (host string) http.Handler {
  handler, ok := router.routes[host]
  if ok {
    return handler
  }

  alias, ok := router.aliases[host]
  if ok {
    return router.getHandler(alias)
  }

  return router.DefaultHandler
}


func (router *Router) Add (host string, handler http.Handler) {
  host = strings.ToLower(host)
  router.routes[host] = handler
}

func (router *Router) Alias (host, to string) {
  host = strings.ToLower(host)
  to = strings.ToLower(to)

  router.aliases[host] = to
}



