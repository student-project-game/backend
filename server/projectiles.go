package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var PROJECTILE_MAP map[string]Troop = map[string]Troop{
  "archer": {
    MovementSpeed: 500, 
    movementFunc: ProjectileMovement,
    Damage: 10,
    Radius: 1,
    Range: 5,
    Type: "projectile",
  },
} 


func BasicRangedAttack(t *Troop) {
  var troop Troop = PROJECTILE_MAP[t.Name]
  troop.ID = fmt.Sprintf("%d", time.Now().UnixMilli()) 
  troop.Tile = t.Tile
  troop.NextTile = t.Tile
  troop.NextTile.Y += 1 
  troop.gameId = t.gameId
  troop.State = "moving"
  troop.Lock = t.Lock
  target := games[troop.gameId].Troops[troop.Lock];
  troop.Direction = GetDirection(troop, target)

  games[t.gameId].moveLoop[troop.ID] = t.MovementSpeed
  games[t.gameId].Troops[troop.ID] = troop
  games[t.gameId].Positions[t.Tile] = troop.ID

  jsonTroop, _ := json.Marshal(troop)
  var action Action = Action{Name: "troop", Body: fmt.Sprintf(`{"troop": %s}`, jsonTroop)}
  response, _ := json.Marshal(action)
  servers[t.gameId].Broadcast(response)
}

func GetDirection(t Troop, target Troop) (direction Tile) {
  direction.X = target.Tile.X - t.Tile.X 
  direction.Y = target.Tile.Y - t.Tile.Y

  if (direction.X > 0) {
    direction.X = 1
  } else if (direction.X < 0) {
    direction.X = -1
  }

  if (direction.Y > 0) {
    direction.Y = 1
  } else if (direction.Y < 0) {
    direction.Y = -1
  }

  log.Println(direction)
  return
}

func ProjectileMovement(t *Troop) {
  var tile Tile = games[t.gameId].Troops[t.Lock].Tile;

  if (tile == t.NextTile) {
    t.Kill(servers[t.gameId]) 
    return
  }

  if (t.Range == 0) {
    t.Kill(servers[t.gameId])
    return
  }

  delete(games[t.gameId].Positions, t.Tile)
  t.Tile = t.NextTile 
  t.NextTile.X += t.Direction.X
  t.NextTile.Y += t.Direction.Y
  games[t.gameId].Positions[t.Tile] = t.ID 
  t.Range -= 1
}
