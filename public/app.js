const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

// connect to websocket server - same host as client
const ws = new WebSocket(`ws://${window.location.host}/ws`);
ws.onopen = () => console.log("Connected to Snake Server! :p");

// handle incoming tick updates from server
ws.onmessage = (event) => {
  const gameState = JSON.parse(event.data);
  console.log("Received game state update:", gameState);
  // clear canvas every frame before drawing new state
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  // TODO: draw game state (e.g. snakes, food, etc.)
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
