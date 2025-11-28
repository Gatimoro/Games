// Tank.io Client
// ===============

class TankGame {
    constructor() {
        // DOM elements
        this.menuEl = document.getElementById('menu');
        this.lobbyEl = document.getElementById('lobby');
        this.gameEl = document.getElementById('game');
        this.canvas = document.getElementById('gameCanvas');
        this.ctx = this.canvas.getContext('2d');

        // Game state
        this.ws = null;
        this.playerId = null;
        this.lobbyCode = null;
        this.gameState = null;
        this.inputState = {
            up: false,
            down: false,
            left: false,
            right: false,
            mouseX: 0,
            mouseY: 0,
            firing: false
        };
        this.activeWeapon = 'cannon';
        this.lastInputSent = 0;
        this.inputSendInterval = 1000 / 30; // 30 times per second

        this.setupEventListeners();
    }

    setupEventListeners() {
        // Menu buttons
        document.getElementById('createLobbyBtn').addEventListener('click', () => this.createLobby());
        document.getElementById('joinLobbyBtn').addEventListener('click', () => this.showJoinForm());
        document.getElementById('submitJoinBtn').addEventListener('click', () => this.joinLobby());
        document.getElementById('lobbyCodeInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.joinLobby();
        });
        document.getElementById('playAgainBtn').addEventListener('click', () => this.returnToMenu());

        // Game input
        document.addEventListener('keydown', (e) => this.handleKeyDown(e));
        document.addEventListener('keyup', (e) => this.handleKeyUp(e));
        this.canvas.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        this.canvas.addEventListener('mousedown', (e) => this.handleMouseDown(e));
        this.canvas.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        this.canvas.addEventListener('contextmenu', (e) => e.preventDefault());
    }

    // ==================
    // Menu & Lobby Logic
    // ==================

    showError(message) {
        document.getElementById('errorMessage').textContent = message;
    }

    showJoinForm() {
        document.getElementById('joinForm').classList.toggle('active');
    }

    async createLobby() {
        try {
            const response = await fetch('/api/create-lobby', { method: 'POST' });
            if (!response.ok) {
                const text = await response.text();
                this.showError(text);
                return;
            }
            const data = await response.json();
            this.connectToLobby(data.code);
        } catch (err) {
            this.showError('Failed to create lobby: ' + err.message);
        }
    }

    async joinLobby() {
        const code = document.getElementById('lobbyCodeInput').value.toUpperCase();
        if (code.length !== 4) {
            this.showError('Please enter a 4-character code');
            return;
        }
        this.connectToLobby(code);
    }

    connectToLobby(code) {
        this.lobbyCode = code;
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?lobby=${code}`;

        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('Connected to lobby:', code);
        };

        this.ws.onmessage = (event) => {
            // Handle multiple messages (newline separated)
            const messages = event.data.split('\n');
            for (const msg of messages) {
                if (msg.trim()) {
                    this.handleServerMessage(JSON.parse(msg));
                }
            }
        };

        this.ws.onerror = (err) => {
            console.error('WebSocket error:', err);
            this.showError('Connection error');
        };

        this.ws.onclose = () => {
            console.log('Disconnected from server');
            if (this.gameState && this.gameState.state !== 'gameover') {
                this.showError('Connection lost');
                this.returnToMenu();
            }
        };
    }

    handleServerMessage(msg) {
        console.log('Server message:', msg.type, msg);

        switch (msg.type) {
            case 'connected':
                this.playerId = msg.payload.playerId;
                this.showLobby();
                break;

            case 'lobby_info':
                this.updateLobbyInfo(msg.payload);
                break;

            case 'game_state':
                this.gameState = msg.payload;
                if (msg.payload.state === 'playing' && this.gameEl.style.display !== 'block') {
                    this.startGame();
                }
                if (msg.payload.state === 'gameover') {
                    this.showGameOver(msg.payload.winnerId);
                }
                break;

            case 'error':
                this.showError(msg.payload.message);
                break;
        }
    }

    showLobby() {
        this.menuEl.style.display = 'none';
        this.lobbyEl.style.display = 'block';
        document.getElementById('lobbyCode').textContent = this.lobbyCode;
    }

    updateLobbyInfo(info) {
        const slot1 = document.getElementById('player1Slot');
        const slot2 = document.getElementById('player2Slot');
        const waitingText = document.getElementById('waitingText');

        const players = info.players || [];

        slot1.textContent = players[0] ? 'Player 1' : 'Waiting...';
        slot1.className = 'player-slot' + (players[0] ? ' filled' : '');
        if (players[0] === this.playerId) {
            slot1.textContent = 'You';
            slot1.classList.add('you');
        }

        slot2.textContent = players[1] ? 'Player 2' : 'Waiting...';
        slot2.className = 'player-slot' + (players[1] ? ' filled' : '');
        if (players[1] === this.playerId) {
            slot2.textContent = 'You';
            slot2.classList.add('you');
        }

        if (players.length === 2) {
            waitingText.textContent = 'Game starting...';
        } else {
            waitingText.textContent = 'Waiting for opponent...';
        }
    }

    // ==================
    // Game Logic
    // ==================

    startGame() {
        this.lobbyEl.style.display = 'none';
        this.gameEl.style.display = 'block';
        document.getElementById('gameOver').classList.add('hidden');

        // Set canvas size to match map
        if (this.gameState && this.gameState.map) {
            this.canvas.width = this.gameState.map.width;
            this.canvas.height = this.gameState.map.height;
        } else {
            this.canvas.width = 1200;
            this.canvas.height = 800;
        }

        // Start game loop
        this.gameLoop();
    }

    gameLoop() {
        if (!this.gameState || this.gameState.state === 'gameover') {
            return;
        }

        this.update();
        this.render();

        requestAnimationFrame(() => this.gameLoop());
    }

    update() {
        // Send input to server at fixed rate
        const now = Date.now();
        if (now - this.lastInputSent > this.inputSendInterval) {
            this.sendInput();
            this.lastInputSent = now;
        }
    }

    sendInput() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'input',
                payload: this.inputState
            }));
        }
    }

    // ==================
    // Input Handling
    // ==================

    handleKeyDown(e) {
        if (this.gameEl.style.display !== 'block') return;

        switch (e.key.toLowerCase()) {
            case 'w':
            case 'arrowup':
                this.inputState.up = true;
                break;
            case 's':
            case 'arrowdown':
                this.inputState.down = true;
                break;
            case 'a':
            case 'arrowleft':
                this.inputState.left = true;
                break;
            case 'd':
            case 'arrowright':
                this.inputState.right = true;
                break;
        }
    }

    handleKeyUp(e) {
        switch (e.key.toLowerCase()) {
            case 'w':
            case 'arrowup':
                this.inputState.up = false;
                break;
            case 's':
            case 'arrowdown':
                this.inputState.down = false;
                break;
            case 'a':
            case 'arrowleft':
                this.inputState.left = false;
                break;
            case 'd':
            case 'arrowright':
                this.inputState.right = false;
                break;
        }
    }

    handleMouseMove(e) {
        const rect = this.canvas.getBoundingClientRect();
        this.inputState.mouseX = e.clientX - rect.left;
        this.inputState.mouseY = e.clientY - rect.top;
    }

    handleMouseDown(e) {
        if (this.gameEl.style.display !== 'block') return;
        e.preventDefault();

        if (e.button === 0) { // Left click - cannon
            this.activeWeapon = 'cannon';
            this.switchWeapon('cannon');
            this.inputState.firing = true;
        } else if (e.button === 2) { // Right click - mortar
            this.activeWeapon = 'mortar';
            this.switchWeapon('mortar');
            this.fire();
        }
        this.updateWeaponUI();
    }

    handleMouseUp(e) {
        if (e.button === 0) {
            this.inputState.firing = false;
        }
    }

    switchWeapon(weapon) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'switch_weapon',
                payload: { weapon }
            }));
        }
    }

    fire() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'fire'
            }));
        }
    }

    updateWeaponUI() {
        const cannonInd = document.getElementById('cannonIndicator');
        const mortarInd = document.getElementById('mortarIndicator');

        cannonInd.classList.toggle('active', this.activeWeapon === 'cannon');
        mortarInd.classList.toggle('active', this.activeWeapon === 'mortar');

        // Update mortar ammo if we have state
        if (this.gameState && this.gameState.tanks && this.gameState.tanks[this.playerId]) {
            const tank = this.gameState.tanks[this.playerId];
            document.getElementById('mortarAmmo').textContent =
                `${tank.mortarAmmo}/${tank.mortarMaxAmmo}`;
        }
    }

    // ==================
    // Rendering
    // ==================

    render() {
        const ctx = this.ctx;
        const state = this.gameState;

        // Clear canvas
        ctx.fillStyle = '#0f0f23';
        ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw grid
        this.drawGrid();

        // Draw bullets
        if (state.bullets) {
            for (const bullet of state.bullets) {
                this.drawBullet(bullet);
            }
        }

        // Draw tanks
        if (state.tanks) {
            for (const [id, tank] of Object.entries(state.tanks)) {
                this.drawTank(tank, id === this.playerId);
            }
        }

        // Update UI
        this.updateWeaponUI();
    }

    drawGrid() {
        const ctx = this.ctx;
        ctx.strokeStyle = '#1a1a3a';
        ctx.lineWidth = 1;

        const gridSize = 50;
        for (let x = 0; x <= this.canvas.width; x += gridSize) {
            ctx.beginPath();
            ctx.moveTo(x, 0);
            ctx.lineTo(x, this.canvas.height);
            ctx.stroke();
        }
        for (let y = 0; y <= this.canvas.height; y += gridSize) {
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(this.canvas.width, y);
            ctx.stroke();
        }
    }

    drawTank(tank, isPlayer) {
        const ctx = this.ctx;
        const x = tank.position.x;
        const y = tank.position.y;
        const bodyRotation = tank.rotation;
        const turretRotation = tank.turretAngle;

        ctx.save();
        ctx.translate(x, y);

        // Draw body
        ctx.save();
        ctx.rotate(bodyRotation);

        // Tank body
        ctx.fillStyle = isPlayer ? '#4ecca3' : '#e94560';
        ctx.fillRect(-20, -15, 40, 30);

        // Tank tracks
        ctx.fillStyle = '#333';
        ctx.fillRect(-22, -17, 6, 34);
        ctx.fillRect(16, -17, 6, 34);

        ctx.restore();

        // Draw turret (rotates independently)
        ctx.save();
        ctx.rotate(turretRotation);

        // Turret base
        ctx.fillStyle = isPlayer ? '#3ba885' : '#c73a52';
        ctx.beginPath();
        ctx.arc(0, 0, 12, 0, Math.PI * 2);
        ctx.fill();

        // Turret barrel
        ctx.fillStyle = isPlayer ? '#2d8a6a' : '#a12e42';
        ctx.fillRect(0, -4, 30, 8);

        ctx.restore();

        // Draw health bar
        if (tank.health < tank.maxHealth) {
            const barWidth = 40;
            const barHeight = 6;
            const healthPercent = tank.health / tank.maxHealth;

            ctx.fillStyle = '#333';
            ctx.fillRect(-barWidth/2, -35, barWidth, barHeight);

            ctx.fillStyle = healthPercent > 0.5 ? '#4ecca3' : '#e94560';
            ctx.fillRect(-barWidth/2, -35, barWidth * healthPercent, barHeight);
        }

        ctx.restore();
    }

    drawBullet(bullet) {
        const ctx = this.ctx;

        if (bullet.type === 'normal') {
            // Normal bullet - simple circle
            ctx.fillStyle = '#ffcc00';
            ctx.beginPath();
            ctx.arc(bullet.position.x, bullet.position.y, 5, 0, Math.PI * 2);
            ctx.fill();

            // Glow effect
            ctx.shadowColor = '#ffcc00';
            ctx.shadowBlur = 10;
            ctx.fill();
            ctx.shadowBlur = 0;
        } else if (bullet.type === 'mortar') {
            // Mortar - draw impact zone and projectile arc
            const progress = bullet.flightProgress || 0;

            // Draw impact zone (always visible)
            ctx.strokeStyle = 'rgba(255, 100, 100, 0.5)';
            ctx.lineWidth = 2;
            ctx.setLineDash([5, 5]);
            ctx.beginPath();
            ctx.arc(bullet.impactPos.x, bullet.impactPos.y, bullet.impactRadius || 50, 0, Math.PI * 2);
            ctx.stroke();
            ctx.setLineDash([]);

            // Draw projectile if in flight
            if (progress < 1) {
                // Calculate arc position (parabolic)
                const startX = bullet.position.x;
                const startY = bullet.position.y;
                const endX = bullet.impactPos.x;
                const endY = bullet.impactPos.y;

                // Lerp position
                const currentX = startX + (endX - startX) * progress;
                const currentY = startY + (endY - startY) * progress;

                // Add arc height
                const arcHeight = 100 * Math.sin(progress * Math.PI);

                // Draw shadow on ground
                ctx.fillStyle = 'rgba(0, 0, 0, 0.3)';
                ctx.beginPath();
                ctx.ellipse(currentX, currentY, 10, 5, 0, 0, Math.PI * 2);
                ctx.fill();

                // Draw mortar shell
                ctx.fillStyle = '#ff6600';
                ctx.beginPath();
                ctx.arc(currentX, currentY - arcHeight, 8, 0, Math.PI * 2);
                ctx.fill();
            } else {
                // Explosion effect
                ctx.fillStyle = 'rgba(255, 100, 50, 0.6)';
                ctx.beginPath();
                ctx.arc(bullet.impactPos.x, bullet.impactPos.y, bullet.impactRadius || 50, 0, Math.PI * 2);
                ctx.fill();
            }
        }
    }

    // ==================
    // Game Over
    // ==================

    showGameOver(winnerId) {
        const overlay = document.getElementById('gameOver');
        const resultText = document.getElementById('resultText');

        overlay.classList.remove('hidden');

        if (winnerId === this.playerId) {
            resultText.textContent = 'Victory!';
            resultText.className = 'victory';
        } else {
            resultText.textContent = 'Defeat';
            resultText.className = 'defeat';
        }
    }

    returnToMenu() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        this.gameState = null;
        this.playerId = null;
        this.lobbyCode = null;

        this.gameEl.style.display = 'none';
        this.lobbyEl.style.display = 'none';
        this.menuEl.style.display = 'block';

        document.getElementById('joinForm').classList.remove('active');
        document.getElementById('errorMessage').textContent = '';
        document.getElementById('lobbyCodeInput').value = '';
    }
}

// Initialize game when page loads
window.addEventListener('load', () => {
    window.game = new TankGame();
});
