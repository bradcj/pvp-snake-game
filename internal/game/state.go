package game

import (
	"fmt"
	"math/rand"
	"time"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

const MOVE_INPUT_QUEUE_SIZE = 2
const DEFAULT_SNAKE_SPEED = 100 * time.Millisecond

type Snake struct {
	ID               string        `json:"id"`
	Body             []Position    `json:"body"`           // array of positions representing the snake's body segments, with the head at index 0
	CurrentDirection Position      `json:"direction"`      // e.g. (1,0) for right, (-1,0) for left, (0,1) for down, (0,-1) for up
	MoveInputQueue   []Position    `json:"moveInputQueue"` // queue of move inputs for the snake
	Score            int           `json:"score"`
	IsDead           bool          `json:"isDead"`
	LastMoveTime     time.Time     `json:"lastMoveTime"` // last time the snake moved
	Speed            time.Duration `json:"speed"`        // in milliseconds
}

const DEFAULT_BOARD_WIDTH = 40
const DEFAULT_BOARD_HEIGHT = 30

type GameState struct {
	Snakes map[string]*Snake `json:"snakes"` // mapping from snake ID to Snake struct
	Food   []Position        `json:"food"`   // slice of positions where food is located
	Width  int               `json:"width"`
	Height int               `json:"height"`
}

func NewGameState() *GameState {
	return &GameState{
		Snakes: make(map[string]*Snake),
		// hardcode some initial food for now
		Food:   []Position{{X: 10, Y: 10}, {X: 20, Y: 20}, {X: 30, Y: 30}},
		Width:  DEFAULT_BOARD_WIDTH,
		Height: DEFAULT_BOARD_HEIGHT,
	}
}

func (gs *GameState) QueueSnakeMove(clientID string, nextMove Position) error {
	snake, exists := gs.Snakes[clientID]
	if !exists {
		return fmt.Errorf("snake for client %s not found", clientID)
	}
	// validate that new direction is not same or opposite as current direction
	prevMove := snake.CurrentDirection
	if len(snake.MoveInputQueue) > 0 {
		prevMove = snake.MoveInputQueue[len(snake.MoveInputQueue)-1]
	}
	sameMove := prevMove == nextMove
	oppositeMove := prevMove.X+nextMove.X == 0 && prevMove.Y+nextMove.Y == 0
	if !sameMove && !oppositeMove {
		if len(snake.MoveInputQueue) >= MOVE_INPUT_QUEUE_SIZE {
			snake.MoveInputQueue = snake.MoveInputQueue[1:] // pop first move
		}
		snake.MoveInputQueue = append(snake.MoveInputQueue, nextMove)
	}
	return nil
}

// Updates the snake's direction based on the move input queue
func UpdateSnakeDirection(snake *Snake) {
	if len(snake.MoveInputQueue) > 0 {
		nextMove := snake.MoveInputQueue[0]
		snake.MoveInputQueue = snake.MoveInputQueue[1:] // pop first move
		snake.CurrentDirection = nextMove
	}
}

func (gs *GameState) UpdateSnakeState(snake *Snake) bool {
	if time.Since(snake.LastMoveTime) < snake.Speed {
		return false
	}
	snake.LastMoveTime = time.Now()

	UpdateSnakeDirection(snake)

	// update position based on direction
	oldHead := snake.Body[0]
	newHead := Position{
		X: oldHead.X + snake.CurrentDirection.X,
		Y: oldHead.Y + snake.CurrentDirection.Y,
	}

	if len(snake.Body) > 1 && newHead.X == snake.Body[1].X && newHead.Y == snake.Body[1].Y {
		// direction is opposite of current movement, ignore movement
		newHead = oldHead
	} else {
		snake.Body = append([]Position{newHead}, snake.Body...) // add the new head at the front of the body
		snake.Body = snake.Body[:len(snake.Body)-1]             // remove the last segment
	}

	// check for collisions with walls
	if newHead.X < 0 || newHead.X >= gs.Width || newHead.Y < 0 || newHead.Y >= gs.Height {
		snake.IsDead = true
		return true
	}

	// check for collisions with self
	for i := 1; i < len(snake.Body); i++ {
		if snake.Body[i] == newHead {
			snake.IsDead = true
			return true
		}
	}
	if snake.IsDead {
		return true
	}

	// check for collisions with other snakes
	for _, otherSnake := range gs.Snakes {
		if otherSnake.ID == snake.ID {
			continue
		}
		if otherSnake.IsDead {
			continue
		}
		for _, segment := range otherSnake.Body {
			if segment == newHead {
				snake.IsDead = true
				return true
			}
		}
		if snake.IsDead {
			return true
		}
	}
	if snake.IsDead {
		return true
	}

	// check for food consumption
	for i, food := range gs.Food {
		if food == newHead {
			snake.Score++
			snake.Body = append(snake.Body, snake.Body[len(snake.Body)-1]) // grow the snake by adding a new segment at the tail
			gs.Food = append(gs.Food[:i], gs.Food[i+1:]...)                // remove the food from the game state
			break
		}
	}

	return false
}

func (gs *GameState) UpdateGameState() {
	// loop through snakes, update position based on direction, check for collisions, check for food consumption, etc.
	for _, snake := range gs.Snakes {
		snakeDied := gs.UpdateSnakeState(snake)
		if snakeDied {
			fmt.Printf("Snake %s has died\n", snake.ID)
			gs.RemoveSnake(snake.ID) // remove the snake from the game state if it has died
		}
	}
}

// Spawns food in random locations in proportion to number of snakes
func (gs *GameState) SpawnRandomFood() {
	aliveSnakes := 0
	for _, snake := range gs.Snakes {
		if !snake.IsDead {
			aliveSnakes++
		}
	}
	foodToSpawn := aliveSnakes
	for i := 0; i < foodToSpawn; i++ {
		foodSpawned := false
		// spawn food on unoccupied cells
		for !foodSpawned {
			foodPos := &Position{
				X: rand.Intn(gs.Width),
				Y: rand.Intn(gs.Height),
			}
			for _, snake := range gs.Snakes {
				for _, segment := range snake.Body {
					if segment == *foodPos {
						continue
					}
				}
			}
			for _, food := range gs.Food {
				if food == *foodPos {
					continue
				}
			}
			gs.AddFood(*foodPos)
			foodSpawned = true
		}
	}
}

func (gs *GameState) AddSnake(id string) {
	gs.Snakes[id] = &Snake{
		ID:               id,
		Body:             []Position{{X: gs.Width / 2, Y: gs.Height / 2}},
		CurrentDirection: Position{X: 1, Y: 0}, // default direction to the right
		Score:            0,
		IsDead:           false,
		Speed:            DEFAULT_SNAKE_SPEED,
		LastMoveTime:     time.Now(),
	}
}

func (gs *GameState) RemoveSnake(id string) {
	delete(gs.Snakes, id)
}

func (gs *GameState) AddFood(position Position) {
	gs.Food = append(gs.Food, position)
}

func (gs *GameState) RemoveFood(position Position) {
	for i, food := range gs.Food {
		if food == position {
			gs.Food = append(gs.Food[:i], gs.Food[i+1:]...)
			break
		}
	}
}
