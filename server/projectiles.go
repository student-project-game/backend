package server

import (
	"encoding/json"
	"fmt"
	"math"
	"stp/utils"
	"time"
)

var PROJECTILE_MAP map[string]Troop = map[string]Troop{
	"archer": {
		MovementSpeed: 200,
		movementFunc:  ProjectileMovement,
		Damage:        10,
		Radius:        1,
		Range:         10,
		Type:          "projectile",
		Name:          "Arrow",
	},
}

func BasicRangedAttack(t *Troop) {
	var troop Troop = PROJECTILE_MAP[t.Name]
	troop.ID = fmt.Sprintf("%d", time.Now().UnixMilli())
	troop.Tile = t.Tile
	troop.NextTile = t.Tile
	troop.gameId = t.gameId
	troop.State = "moving"
	troop.Lock = t.Lock
	target := games[troop.gameId].Troops[troop.Lock]
	troop.Direction = GetDirection(troop, ShortestInRadius(troop, target))

	games[t.gameId].moveLoop[troop.ID] = t.MovementSpeed
	games[t.gameId].Troops[troop.ID] = troop

	jsonTroop, _ := json.Marshal(troop)
	var action Action = Action{Name: "troop", Body: fmt.Sprintf(`{"troop": %s}`, jsonTroop)}
	response, _ := json.Marshal(action)
	servers[t.gameId].Broadcast(response)
}

func GetDirection(t Troop, target Tile) (direction Tile) {
	direction.X = target.X - t.Tile.X
	direction.Y = target.Y - t.Tile.Y

	if direction.X > 0 {
		direction.X = 1
	} else if direction.X < 0 {
		direction.X = -1
	}

	if direction.Y > 0 {
		direction.Y = 1
	} else if direction.Y < 0 {
		direction.Y = -1
	}

	return
}

func ShortestInRadius(t Troop, target Troop) Tile {
	var tile Tile = target.Tile
	var shortest float64 = math.Inf(1)

	for y := target.Tile.Y - t.Radius; y <= target.Tile.Y+t.Radius; y++ {
		for x := target.Tile.X - t.Radius; x <= target.Tile.X+t.Radius; x++ {
			step := Tile{X: x, Y: y}
			if _, ok := games[t.gameId].Positions[step]; !ok && step != target.Tile {
				d := utils.Euclidean(x, y, t.Tile.X, t.Tile.Y)
				if d < shortest {
					shortest = d
					tile = step
				}
			}
		}
	}

	return tile
}

func ProjectileMovement(t *Troop) {
	var tile Tile = games[t.gameId].Troops[t.Lock].Tile
	d := utils.Euclidean(tile.X, tile.Y, t.NextTile.X, t.NextTile.Y)

	if math.Floor(d) == 1 {
		target := games[t.gameId].Troops[t.Lock]
		target.HP -= t.Damage
		fmt.Println(target)
		games[t.gameId].Troops[target.ID] = target
		if target.HP <= 0 {
			target.Kill(servers[target.gameId])
		} else {
			target.Broadcast()
		}
		t.Kill(servers[t.gameId])
		return
	}

	if t.Range == 0 {
		t.Kill(servers[t.gameId])
		return
	}

	t.Tile = t.NextTile
	t.NextTile.X += t.Direction.X
	t.NextTile.Y += t.Direction.Y
	t.Range -= 1
}
