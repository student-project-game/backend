package server

import (
	"math"
	"stp/utils"
)

var CARD_MAP map[string]Troop = map[string]Troop{
	"tower": {
		MaxHP:       10000,
		HP:          10000,
		AttackSpeed: 500, MovementSpeed: 100,
		attackFunc: BasicMeleeAttack, movementFunc: BuildingDetection,
		Vision: 5,
		Radius: 5,
		Damage: 5,
		Type:   "building",
	},
	"water": {},
	"wizard": {
		AttackSpeed: 300, MovementSpeed: 1000, DeploySpeed: 1000,
		attackFunc: BasicSplashAttack, movementFunc: BasicMovement,
		Vision: 5,
		Radius: 1,
		Splash: 2,
		MaxHP:  50,
		HP:     50,
		Cost:   2,
		Damage: 10,
		Name:   "Wizard",
	},
	"archer": {
		Name:        "archer",
		AttackSpeed: 2000, MovementSpeed: 500, DeploySpeed: 1000,
		attackFunc: BasicRangedAttack, movementFunc: BasicMovement,
		Vision: 5,
		Radius: 4,
		Splash: 2,
		MaxHP:  1000,
		HP:     1000,
		Cost:   3,
		Damage: 10,
	},
	"hog_rider": {
		AttackSpeed: 300, MovementSpeed: 100, DeploySpeed: 1000,
		attackFunc: BasicMeleeAttack, movementFunc: BasicMovement,
		Radius: 1,
		Vision: 5,
		MaxHP:  100,
		HP:     100,
		Cost:   5,
		Damage: 50,
	},
}

func Attack(t *Troop, damage func(t *Troop, target Troop) bool) {
	target := games[t.gameId].Troops[t.Lock]
	if target.ID == "" {
		t.InitiateMovement()
		games[t.gameId].Positions[t.Tile] = t.ID
		return
	}

	isKilled := damage(t, target)

	if isKilled {
		t.InitiateMovement()
	}

	games[t.gameId].Positions[t.Tile] = t.ID
}

func BasicMeleeAttack(t *Troop) {
	Attack(t, DirectAttack)
}

func BasicSplashAttack(t *Troop) {
	Attack(t, SplashAttack)
}

func DirectAttack(t *Troop, target Troop) bool {
	target.HP -= t.Damage
	games[t.gameId].Troops[target.ID] = target
	if target.HP <= 0 {
		target.Kill(servers[target.gameId])
		return true
	} else {
		target.Broadcast()
	}
	return false
}

func SplashAttack(t *Troop, target Troop) bool {
	isLockKilled := false
	for y := t.Tile.Y - t.Splash; y <= t.Tile.Y+t.Splash; y++ {
		for x := t.Tile.X - t.Splash; x <= t.Tile.X+t.Splash; x++ {
			tile := Tile{X: x, Y: y}
			if troop, ok := games[t.gameId].Positions[tile]; ok {
				enemy := games[t.gameId].Troops[troop]
				if enemy.Team != t.Team && enemy.Team != "" {
					isKilled := DirectAttack(t, enemy)
					if t.Lock == troop {
						isLockKilled = isKilled
					}
				}
			}
		}
	}
	return isLockKilled
}

func (t *Troop) InitiateMovement() {
	t.State = "moving"
	games[t.gameId].moveLoop[t.ID] = t.MovementSpeed
	t.movementFunc(t)
}

func (t *Troop) ShortestInRadius(center Tile) (Tile, float64) {
	var tile Tile = center
	var shortest float64 = math.Inf(1)

	for y := center.Y - t.Radius; y <= center.Y+t.Radius; y++ {
		for x := center.X - t.Radius; x <= center.X+t.Radius; x++ {
			step := Tile{X: x, Y: y}
			if _, ok := games[t.gameId].Positions[step]; !ok && step != center {
				d := utils.Euclidean(x, y, t.Tile.X, t.Tile.Y)
				if d < shortest {
					shortest = d
					tile = step
				}
			}
		}
	}

	return tile, shortest
}

func (t *Troop) ShortestPath(target Tile) Tile {
	var tile Tile
	var shortest float64 = math.Inf(1)

	for y := t.Tile.Y - 1; y <= t.Tile.Y+1; y++ {
		for x := t.Tile.X - 1; x <= t.Tile.X+1; x++ {
			step := Tile{X: x, Y: y}
			if _, ok := games[t.gameId].Positions[step]; !ok {
				d := utils.Euclidean(x, y, target.X, target.Y)

				if d < shortest && !(d != 0 && step == t.Tile) {
					shortest = d
					tile = step
				}
			}
		}
	}

	return tile
}

func (t *Troop) ClosestTroop() Tile {
	if t.Type != "building" {
		t.NearestTower()
	}

	var tile Tile = games[t.gameId].Troops[t.Lock].Tile
	var shortest float64 = utils.Euclidean(t.Tile.X, t.Tile.Y, tile.X, tile.Y)

	for y := t.Tile.Y - t.Vision; y <= t.Tile.Y+t.Vision; y++ {
		for x := t.Tile.X - t.Vision; x <= t.Tile.X+t.Vision; x++ {
			step := Tile{X: x, Y: y}
			if troop, ok := games[t.gameId].Positions[step]; ok {
				target := games[t.gameId].Troops[troop]
				if target.Team != t.Team && target.Team != "" {
					d := utils.Euclidean(t.Tile.X, t.Tile.Y, target.NextTile.X, target.NextTile.Y)

					if d < shortest {
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

	if t.Team == "up" {
		tower_key = "tower_up"
	}

	t_l := games[t.gameId].Troops[tower_key+"_left"]
	t_r := games[t.gameId].Troops[tower_key+"_right"]
	left := utils.Euclidean(t.Tile.X, t.Tile.Y, t_l.Tile.X, t_l.Tile.Y)
	right := utils.Euclidean(t.Tile.X, t.Tile.Y, t_r.Tile.X, t_r.Tile.Y)

	if left < right {
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
	xMin := t.Tile.X - t.Radius
	xMax := t.Tile.X + t.Radius
	yMin := t.Tile.Y - t.Radius
	yMax := t.Tile.Y + t.Radius

	if target.X >= xMin && target.X <= xMax {
		if target.Y >= yMin && target.Y <= yMax {
			return true
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
