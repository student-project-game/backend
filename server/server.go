package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var servers map[string]*Server = make(map[string]*Server)

func HomeActions(a Action, s *Server, connections ...*websocket.Conn) {
  switch a.Name {
    case "join":
      JoinMatch(a.Body, s, connections[0])
  } 
}

func MatchActions(a Action, s *Server, connections ...*websocket.Conn) {
  switch a.Name {
  case "place":
    games[s.id].Place(a.Body, s, connections[0])
  }
}

func JoinMatch(body string, s *Server, ws *websocket.Conn) {
  var gamemode Gamemode;
  json.Unmarshal([]byte(body), &gamemode)

  log.Println(gamemode, body)

  playerID := ws.Config().Protocol[0]
  for _, server := range servers {
    maxPlayers := games[server.id].Gamemode.PlayerCount
    
    if maxPlayers == gamemode.PlayerCount && len(server.conns) >= 0 && len(server.conns) < maxPlayers {
      direction := "up"
      if len(server.conns) % 2 == 1 {
	direction = "down"
      }
      var player Player = Player{ID: playerID, Team: direction}
      games[server.id].Players = append(games[server.id].Players, player)
      var action Action = Action{Name: "game_id", Body: fmt.Sprintf(`{"id": "%s", "direction": "%s"}`, server.id, direction)}
      response, _ := json.Marshal(action) 
      ws.Write(response)
      return
    }	
  } 

  serverId := fmt.Sprintf("%d", len(servers))
  games[serverId] = MakeGame(serverId)
  games[serverId].Gamemode = gamemode
  var player Player = Player{ID: playerID, Team: "up"}
  games[serverId].Players = append(games[serverId].Players, player)
  var action Action = Action{Name: "game_id", Body: fmt.Sprintf(`{"id": "%s", "direction": "up"}`, serverId)}
  response, _ := json.Marshal(action) 
  ws.Write(response)
}

func Serve() {
  s := MakeServer(HomeActions, "home")
  http.HandleFunc("/home",
    func (w http.ResponseWriter, req *http.Request) {
      s := websocket.Server{Handler: s.Handler}
      s.ServeHTTP(w, req)
    },
  );

  http.HandleFunc("/games/{id}",
    func (w http.ResponseWriter, req *http.Request) {
      var game *Server;
      id := req.PathValue("id")
      if (servers[id] != nil) {
	game = servers[id]
      } else {
	game = MakeServer(MatchActions, req.PathValue("id"))
	servers[id] = game
      }

      if len(game.conns) == games[id].Gamemode.PlayerCount {
	w.WriteHeader(423)
	return
      }

      s := websocket.Server{Handler: game.Handler}
      s.ServeHTTP(w, req)
    },
  );
  err := http.ListenAndServe(":12345", nil) 
  if err != nil {
    log.Fatal(err)
  } 
}
