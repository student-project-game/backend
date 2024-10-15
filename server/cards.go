package server

import (
	"math"
	"stp/utils"
)

var CARD_MAP map[string]Troop = map[string]Troop{
  "tower": {
    MaxHp: 10000,
    HP: 10000,
    AttackSpeed: 500, MovementSpeed: 100, 
    attackFunc: BasicAttack, movementFunc: BuildingDetection,
    Vision: 5,
    Radius: 5,
    Damage: 10,
    Type: "building",
  },
  "water": {},
  "wizard": {
    AttackSpeed: 300, MovementSpeed: 500, DeploySpeed: 1000, 
    attackFunc: BasicAttack, movementFunc: BasicMovement,
    Vision: 5,
    Radius: 1,
    MaxHp: 50,
    HP: 50,
    Damage: 10,
  }, 
  "hog_rider": {
    AttackSpeed: 300, MovementSpeed: 100, DeploySpeed: 1000, 
    attackFunc: BasicAttack, movementFunc: BasicMovement,
    Radius: 1,
  }, 
}

func BasicAttack(t *Troop) {
  target := games[t.gameId].Troops[t.Lock] 
  target.HP -= t.Damage
  games[t.gameId].Troops[t.Lock] = target
  target.Broadcast()

  if target.HP <= 0 {
    target.Kill(servers[t.gameId])
    t.State = "moving" 
    games[t.gameId].moveLoop[t.ID] = t.MovementSpeed
    t.movementFunc(t)
  }
  games[t.gameId].Positions[t.Tile] = t.ID 
}

func (t *Troop) ShortestInRadius(center Tile) (Tile, float64) {
  var tile Tile = center;
  var shortest float64 = math.Inf(1)

  for y := center.Y - t.Radius; y <= center.Y + t.Radius; y++ {
    for x := center.X - t.Radius; x <= center.X + t.Radius; x++ {
      step := Tile{X: x, Y: y}
      if _, ok := games[t.gameId].Positions[step]; !ok && step != center {
	d := utils.Euclidean(x, y, t.Tile.X, t.Tile.Y)
	if (d < shortest) {
	  shortest = d
	  tile = step
	}
      }
    } 
  } 

  return tile, shortest
}

func (t *Troop) ShortestPath(target Tile) Tile {
  var tile Tile;
  var shortest float64 = math.Inf(1)

  for y := t.Tile.Y - 1; y <= t.Tile.Y + 1; y++ {
    for x := t.Tile.X - 1; x <= t.Tile.X + 1; x++ {
      step := Tile{X: x, Y: y}
      if _, ok := games[t.gameId].Positions[step]; !ok {
	d := utils.Euclidean(x, y, target.X, target.Y)
	
	if (d < shortest && !(d != 0 && step == t.Tile)) {
	  shortest = d
	  tile = step
	}
      }
    } 
  } 

  return tile
}

func (t *Troop) ClosestTroop() Tile {
  if (t.Type != "building") {
    t.NearestTower()
  }

  var tile Tile = games[t.gameId].Troops[t.Lock].Tile;
  var shortest float64 = utils.Euclidean(t.Tile.X, t.Tile.Y, tile.X, tile.Y) 

  for y := t.Tile.Y - t.Vision; y <= t.Tile.Y + t.Vision; y++ {
    for x := t.Tile.X - t.Vision; x <= t.Tile.X + t.Vision; x++ {
      step := Tile{X: x, Y: y}
      if troop, ok := games[t.gameId].Positions[step]; ok {
	target := games[t.gameId].Troops[troop]
	if target.Player != t.Player && target.Player != "" {
	  d := utils.Euclidean(t.Tile.X, t.Tile.Y, target.NextTile.X, target.NextTile.Y)

	  if (d < shortest) {
	    shortest = d
	    tile = target.NextTile
	    t.Lock = troop
	  }
	} 
      }
    } 
  } 

  return tile
}

func (t *Troop) Pathfinding() Tile {
  target, d := t.ShortestInRadius(t.ClosestTroop())

  if d == 0 {
    t.State = "attacking" 
    games[t.gameId].attackLoop[t.ID] = t.AttackSpeed
  }

  step := t.ShortestPath(target)

  return step
}

func (t *Troop) NearestTower() {
  tower_key := "tower_down"

  if (t.Player == games[t.gameId].Players[0]) {
    tower_key = "tower_up"
  }

  t_l := games[t.gameId].Troops[tower_key + "_left"]
  t_r := games[t.gameId].Troops[tower_key + "_right"]
  left := utils.Euclidean(t.Tile.X, t.Tile.Y, t_l.Tile.X, t_l.Tile.Y)
  right := utils.Euclidean(t.Tile.X, t.Tile.Y, t_r.Tile.X, t_r.Tile.Y)

  if (left < right) {
    t.Lock = tower_key + "_left"
  } else {
    t.Lock = tower_key + "_right" 
  }
}

func BasicMovement(t *Troop) {
  delete(games[t.gameId].Positions, t.Tile)
  t.Tile = t.NextTile 
  t.NextTile = t.Pathfinding()

  games[t.gameId].Positions[t.Tile] = t.ID 
}


func (t *Troop) InRadius(target Tile) bool {
  for y := t.Tile.Y - t.Radius; y <= t.Tile.Y + t.Radius; y++ {
    for x := t.Tile.X - t.Radius; x <= t.Tile.X + t.Radius; x++ {
      tile := Tile{X: x, Y: y}
      if target == tile {
	return true
      }
    } 
  } 

  return false 
}

func BuildingDetection(t *Troop) {
  if !t.InRadius(t.ClosestTroop()) {
    return
  }

  t.State = "attacking" 
  games[t.gameId].attackLoop[t.ID] = t.AttackSpeed
}
