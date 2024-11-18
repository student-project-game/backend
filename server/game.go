package server

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)


type Gamemode struct {
  PlayerCount int `json:"playerCount"`
}

type Game struct {
  ID string
  Troops map[string]Troop
  Positions map[Tile]string
  attackLoop map[string]int
  moveLoop map[string]int
  started bool
  Players []Player
  Gamemode Gamemode
  l sync.Mutex
}

type Placement struct {
  Name string `json:"name"`
  Tile Tile `json:"tile"`
}

func MakeGame(id string) *Game {
  game := Game{
    ID: id,
    attackLoop: make(map[string]int),
    moveLoop: make(map[string]int),
    Troops: make(map[string]Troop),
    Positions: make(map[Tile]string),
    Players: make([]Player, 0),
    started: false,
  }
  return &game
}

func (g *Game) GenerateTroop(card string, id string, tile Tile, team string) Troop {
  var troop Troop = CARD_MAP[card]
  troop.ID = id
  troop.Tile = tile 
  troop.NextTile = troop.Tile 
  troop.Team = team 
  troop.gameId = g.ID
  g.Positions[troop.Tile] = troop.ID

  if (card == "tower") {
    troop.State = "moving"
    g.moveLoop[troop.ID] = troop.MovementSpeed
  }

  return troop
}

func (g *Game) StaticTroops(s *Server) {
  g.Troops["tower_down_left"] = g.GenerateTroop("tower", "tower_down_left", Tile{X: 4, Y: 5}, g.Players[0].Team)
  g.Troops["tower_down_right"] = g.GenerateTroop("tower", "tower_down_right", Tile{X: 13, Y: 5}, g.Players[0].Team)
  g.Troops["tower_main_down"] = g.GenerateTroop("tower", "tower_main_down", Tile{X: 8, Y: 2}, g.Players[0].Team)

  g.Troops["tower_up_left"] = g.GenerateTroop("tower", "tower_up_left", Tile{X: 4, Y: 26}, g.Players[1].Team)
  g.Troops["tower_up_right"] = g.GenerateTroop("tower", "tower_up_right", Tile{X: 13, Y: 26}, g.Players[1].Team)
  g.Troops["tower_main_up"] = g.GenerateTroop("tower", "tower_main_up", Tile{X: 8, Y: 29}, g.Players[1].Team)

  for i := 0; i < 18; i++ {
    if i != 4 && i != 13 {
      id := fmt.Sprintf("water%d", i)
      g.Troops[id] = g.GenerateTroop("water", id, Tile{X: i, Y: 15}, "")
    }
  }
}

var games map[string]*Game = make(map[string]*Game) 

func (g *Game) Place(body string, s *Server, ws *websocket.Conn) {
  var placement Placement;
  json.Unmarshal([]byte(body), &placement)

  var troop Troop = CARD_MAP[placement.Name]

  troop.ID = fmt.Sprintf("%d", time.Now().UnixMilli())
  troop.Tile = placement.Tile
  troop.NextTile = placement.Tile
  troop.Team = ws.Config().Protocol[0]
  troop.gameId = s.id
  troop.State = "moving"

  troop.NearestTower()

  games[s.id].moveLoop[troop.ID] = troop.DeploySpeed
  games[s.id].Troops[troop.ID] = troop
  games[s.id].Positions[placement.Tile] = troop.ID
  
  jsonTroop, _ := json.Marshal(troop)
  var action Action = Action{Name: "troop", Body: fmt.Sprintf(`{"troop": %s}`, jsonTroop)}
  response, _ := json.Marshal(action)

  s.Broadcast(response)
}

func (s *Server) StartMatch() {
  games[s.id].StaticTroops(s)

  jsonTroops, _ := json.Marshal(games[s.id].Troops)
  var match_start Action = Action{Name: "match_started", Body: fmt.Sprintf(`{"troops": %s}`, string(jsonTroops))}
  broadcast, _ := json.Marshal(match_start)
  s.Broadcast(broadcast)

  if (!games[s.id].started) {
    games[s.id].started = true
    go func() {
      games[s.id].Loop(s)
    }()
  }
}

func (g *Game) Loop(s *Server) {
  for {
    // start := time.Now().UnixMilli()
    if games[s.id] == nil {
      break
    }
      
    games[s.id].l.Lock()
    for key, troop := range g.Troops {
      if troop.State == "attacking" {
	games[s.id].attackLoop[key] -= 100	
	if (games[s.id].attackLoop[key] == 0) {
	  troop.attackFunc(&troop)
	  troop.Broadcast()
	  games[s.id].attackLoop[key] = troop.AttackSpeed
	}
      } else if troop.State == "moving" {
	games[s.id].moveLoop[key] -= 100
	if (games[s.id].moveLoop[key] == 0) {
	  troop.movementFunc(&troop)
	  troop.Broadcast()
	  games[s.id].moveLoop[key] = troop.MovementSpeed
	}
      }
      games[s.id].Troops[key] = troop
    }
    games[s.id].l.Unlock()
    // s.Log(fmt.Sprintf("Loop completed in: %d miliseconds with %d troops", time.Now().UnixMilli() - start, len(games[s.id].Troops)))
    time.Sleep(time.Millisecond*100)
  }
} 
