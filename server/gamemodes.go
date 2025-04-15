package server

type Gamemode struct {
  PlayerCount int `json:"playerCount"`
  MaxElixir int
  InitialElixir int
  ElixirRate int
  Timer int
}

var GAMEMODE_MAP map[string]Gamemode = map[string]Gamemode {
  "1v1": {
    PlayerCount: 2,
    MaxElixir: 10,
    InitialElixir: 8,
    ElixirRate: 1000,
    Timer: 2000000000000000,
  },
  "2v1": {
    PlayerCount: 3,
    MaxElixir: 10,
    InitialElixir: 8,
    ElixirRate: 1000,
    Timer: 2000000000000000,
  },
}

