package game

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID        string     `json:"id"`
	Body      []Position `json:"body"`      // array of positions representing the snake's body segments, with the head at index 0
	Direction Position   `json:"direction"` // e.g. (1,0) for right, (-1,0) for left, (0,1) for down, (0,-1) for up
	Score     int        `json:"score"`
	IsDead    bool       `json:"isDead"`
}

type GameState struct {
	Snakes map[string]*Snake `json:"snakes"` // mapping from snake ID to Snake struct
	Food   []Position        `json:"food"`   // array of positions representing food items on the board
	Width  int               `json:"width"`
	Height int               `json:"height"`
}
