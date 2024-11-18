package server

import (
	"encoding/json"
	"fmt"
)

type Tile struct {
  X int `json:"x"`
  Y int `json:"y"`
}

type Troop struct {
  ID string
  Name string
  gameId string
  AttackSpeed int
  MovementSpeed int
  DeploySpeed int
  Tile Tile
  NextTile Tile
  Radius int
  Vision int
  MaxHP int
  HP int
  Cost int
  Damage int
  Splash int
  Range int
  Direction Tile 
  State string
  Lock string
  attackFunc func (t *Troop)
  movementFunc func (t *Troop)
  Team string
  Type string
}

func (t *Troop) Broadcast() {
  if t.gameId == "" {
    return
  }
  jsonTroop, _ := json.Marshal(t)
  var action Action = Action{Name: "troop", Body: fmt.Sprintf(`{"troop": %s}`, jsonTroop)}
  response, _ := json.Marshal(action)

  servers[t.gameId].Broadcast(response)
}

func (t *Troop) Kill(s *Server) {
  var action Action = Action{Name: "death", Body: fmt.Sprintf(`{"troop": "%s"}`, t.ID)}
  fmt.Println(action)
  response, _ := json.Marshal(action)
  delete(games[s.id].Positions, t.Tile)
  delete(games[s.id].attackLoop, t.ID)
  delete(games[s.id].moveLoop, t.ID)
  delete(games[s.id].Troops, t.ID)

  s.Broadcast(response)
}
