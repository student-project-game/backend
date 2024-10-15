package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"stp/utils"

	"golang.org/x/net/websocket"
)

type Server struct {
  conns map[*websocket.Conn]bool
  actions func (Action, *Server, ...*websocket.Conn)
  id string
}

type Action struct {
  Name string `json:"name"`
  Body string `json:"body"`
}

func MakeServer(actionsFunc func (Action, *Server, ...*websocket.Conn), id string) *Server {
  return &Server{
    conns: make(map[*websocket.Conn]bool),
    actions: actionsFunc,
    id: id,
  }
}

func (s *Server) Log(message string) {
  log.Println(fmt.Sprintf("%s -> %s", s.id, message))
}

func (s *Server) Handler(ws *websocket.Conn) {
  s.Log(fmt.Sprintf("Client %s %s", ws.Request().RemoteAddr, utils.Foreground("0;255;0", "CONNECTED")))
  s.conns[ws] = true

  if (s.id != "home" && len(s.conns) == 2) {
    s.StartMatch();
  } 

  go func() {
   s.Ping(ws)
  }()
  s.Read(ws)
}

func (s *Server) Read(ws *websocket.Conn) {
  buf := make([]byte, 1024)
  for {
    n, err := ws.Read(buf)
    if err != nil {
      if err == io.EOF {
	break
      }
      log.Println("error:", err)
      continue
    }
    msg := buf[:n]
    log.Println(ws.Request().RemoteAddr, "sent:", string(msg)) 
    var action Action
    json.Unmarshal(msg, &action)

    s.actions(action, s, ws)

    s.Broadcast(msg)
  }
}

func (s *Server) Ping(conn *websocket.Conn) {
  for {
    active := s.conns[conn]
    if (active) {
      var m Action = Action{Name: "ping", Body: "ping"}
      body, _ := json.Marshal(m)
      _, err := conn.Write([]byte(body)) 
      if err != nil && errors.Is(err, net.ErrClosed) {
	delete(s.conns, conn)
	s.Log(fmt.Sprintf("Client %s %s", conn.Request().RemoteAddr, utils.Foreground("255;0;0", "DISCONNECTED")))

	if (s.id != "home" && len(s.conns) == 0) {
	  delete(games, s.id)
	  delete(servers, s.id)
	}
      }
    }
    time.Sleep(time.Second)
  }
}

func (s *Server) Broadcast(msg []byte) {
  for conn, active  := range s.conns {
    go func(conn *websocket.Conn, active bool) {
      if (active) {
	conn.Write(msg) 
      }
    }(conn, active)
  }
}
