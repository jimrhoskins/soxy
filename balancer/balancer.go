package balancer

import (
  "net/http"
)


type task struct {
  w http.ResponseWriter
  r *http.Request
  complete chan bool
}


type LoadBalancer struct {
  handlers []http.Handler
  tasks chan *task
}

func New(h ...http.Handler) *LoadBalancer {
  r := &LoadBalancer{
    handlers: h,
    tasks: make(chan *task),
  }

  go func(){
    for {
      for _, handler := range r.handlers {
        task := <-r.tasks
        go func (h http.Handler){
          h.ServeHTTP(task.w, task.r)
          task.complete <-true
        }(handler)
      }
    }
  }()

  return r
}

func (self *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  t := &task{w, r, make(chan bool)}
  self.tasks <- t
  <-t.complete
}
