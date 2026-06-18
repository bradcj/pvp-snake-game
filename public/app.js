const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

// connect to websocket server - same host as client
const ws = new WebSocket(`ws://${window.location.host}/ws`);
ws.onopen = () => console.log("Connected to Snake Server! :p");

let cellSize;
let gameWidth;
let gameHeight;

// handle incoming tick updates from server
ws.onmessage = (event) => {
    const gameState = JSON.parse(event.data);
    console.log("Received game state update:", gameState);

    // calculate cell size based on board dimensions and canvas size
    if (gameState.width != gameWidth || gameState.height !== gameHeight) {
        gameWidth = gameState.width;
        gameHeight = gameState.height;
        cellSize = Math.min(canvas.width / gameWidth, canvas.height / gameHeight);
        console.log(`Updated board dimensions: ${gameWidth}x${gameHeight}, cell size: ${cellSize}`);
    }
    // clear canvas every frame before drawing new state
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // draw food
    gameState.food.forEach(food => {
        ctx.fillStyle = 'red';
        ctx.fillRect(food.x * cellSize, food.y * cellSize, cellSize, cellSize);
    });

    // draw snakes
    Object.values(gameState.snakes).forEach(snake => {
        snake.body.forEach((segment, index) => {
            if (index === 0) {
                ctx.fillStyle = snake.isDead ? 'lightgray' : 'lightgreen'; // head is darker
            } else if (index === snake.body.length - 1) {
                ctx.fillStyle = snake.isDead ? 'darkgray' : 'darkgreen'; // tail is lighter
            } else {
                ctx.fillStyle = snake.isDead ? 'gray' : 'green';
            }
            ctx.fillRect(segment.x * cellSize, segment.y * cellSize, cellSize, cellSize);
        });
    });
};

// handle keyboard input and send to server
window.addEventListener('keydown', (e) => {
    let direction = null;
    if (e.key === 'ArrowUp' || e.key === 'w') direction = 'UP';
    else if (e.key === 'ArrowDown' || e.key === 's') direction = 'DOWN';
    else if (e.key === 'ArrowLeft' || e.key === 'a') direction = 'LEFT';
    else if (e.key === 'ArrowRight' || e.key === 'd') direction = 'RIGHT';

    if (direction) {
        ws.send(JSON.stringify({ type: 'move', payload: { direction: direction } }));
    }
})
