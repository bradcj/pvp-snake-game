package game

import (
	"fmt"
	"math/rand"
)

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
	Food   []Position        `json:"food"`   // slice of positions where food is located
	Width  int               `json:"width"`
	Height int               `json:"height"`
}

const DEFAULT_BOARD_WIDTH = 40
const DEFAULT_BOARD_HEIGHT = 30

func NewGameState() *GameState {
	return &GameState{
		Snakes: make(map[string]*Snake),
		// hardcode some initial food for now
		Food:   []Position{{X: 10, Y: 10}, {X: 20, Y: 20}, {X: 30, Y: 30}},
		Width:  DEFAULT_BOARD_WIDTH,
		Height: DEFAULT_BOARD_HEIGHT,
	}
}

func (gs *GameState) UpdateSnakeState(snake *Snake) bool {
	// update position based on direction
	oldHead := snake.Body[0]
	newHead := Position{
		X: oldHead.X + snake.Direction.X,
		Y: oldHead.Y + snake.Direction.Y,
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

// update the snake's direction in the game state
// validate that the new direction is not directly opposite to the current direction
func (gs *GameState) UpdateSnakeDirection(clientID string, newDirection Position) error {
	if snake, exists := gs.Snakes[clientID]; exists {
		// validate that new direction is not directly opposite to current direction
		if (newDirection.X == -snake.Direction.X && newDirection.Y == 0) ||
			(newDirection.Y == -snake.Direction.Y && newDirection.X == 0) {
			return fmt.Errorf("invalid direction change for client %s: cannot reverse direction", clientID)
		}
		snake.Direction = newDirection
	} else {
		return fmt.Errorf("snake for client %s not found", clientID)
	}
	return nil
}

// spawn food, in proportion with number of snakes, in random locations
func (gs *GameState) SpawnRandomFood() {
	foodToSpawn := len(gs.Snakes)
	for i := 0; i < foodToSpawn; i++ {
		foodSpawned := false
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
		ID:        id,
		Body:      []Position{{X: gs.Width / 2, Y: gs.Height / 2}},
		Direction: Position{X: 1, Y: 0}, // default direction to the right
		Score:     0,
		IsDead:    false,
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
