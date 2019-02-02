var foods = [];
var boxWidth = 20;  // 横向方块的个数
var boxHeight = 20; // 纵向方块的个数
var boxSize = 24;   // 方块的尺寸（像素）
var canvasWidth = 480;
var canvasHeight = 480;

var canvas = $('#canvas').get(0).getContext('2d');
function draw(seat, color) {
	canvas.fillStyle = color;
	canvas.fillRect(seat % boxWidth * boxSize + 1, ~~(seat / boxHeight) * boxSize + 1, boxSize-2, boxSize-2);
}

function Snake(body, isMe) {
	this.snake = body;  // snake 队列表示蛇身
	this.direction = 1; // -1表示向左，-20表示向上，1表示向右，20表示向下
	this.gameOver = false;
	this.colorHead = isMe ? 'lime' : 'lightblue';
	this.colorDead = isMe ? 'darkgreen' : 'blue';
	this.snake.forEach((n) => {
		draw(n, this.colorHead);
	});
}
Snake.prototype.frameForward = function(keyCode) {
	if (this.gameOver) return;

	// 根据按键操作计算出行进方向
	var newDir = [-1, -boxWidth, 1, boxWidth][keyCode - 37] || this.direction;
	// 选择有效的行进方向（不能逆向）
	this.direction = (this.snake[1] - this.snake[0] == newDir) ? this.direction : newDir;
	// 蛇头的位置
	var head = this.snake[0] + this.direction;

	// 判断是否撞到自己或者墙壁
	if (this.snake.indexOf(head) > 0 // 自身
		|| head < 0 // 上边界
		|| head >= boxWidth*boxHeight // 下边界
		|| this.direction == 1 && head % boxWidth == 0 // 右边界
		|| this.direction == -1 && head % boxWidth == boxWidth-1 // 左边界
		) {
		// 结束！
		this.gameOver = true;

		// 画残骸
		this.snake.forEach((n) => {
			draw(n, this.colorDead);
		});
		return;
	}

	// 画出蛇头
	this.snake.unshift(head);
	draw(head, this.colorHead);

	// 判断是否吃到食物
	var pos = foods.indexOf(head);
	if (pos >= 0) {
		// 吃到，清除食物
		foods.splice(pos, 1);
	} else {
		// 没有吃到，清除蛇尾
		draw(this.snake.pop(), "black");
	}
}

document.addEventListener('astilectron-ready', function() {
	$('#btn-1p')
		.removeAttr('disabled')
		.on('click', function() {
			$('#btn-1p').attr('disabled', true);
			// 开始 1P 游戏
			astilectron.sendMessage({
				name: 'start',
				payload: 1,
			});
		});


	// 把操作按键传给后端
	document.addEventListener('keydown', (evt) => {
		var kc = (evt || event).keyCode;
		if (kc >= 37 && kc <= 40) {
			astilectron.sendMessage({
				name: 'keydown',
				payload: kc,
			});
		}
	});

	var snakes = [];

	astilectron.onMessage(function(message) {
		// console.log("onMessage:", JSON.stringify(message));
		switch (message.name) {
		case "about":
			return;

		case "kick-off":
			// 开局
			boxWidth = message.payload.width;
			boxHeight = message.payload.height;
			boxSize = canvasWidth / boxWidth;
			canvasHeight = boxSize * boxHeight;

			canvas.fillStyle = 'black';
			canvas.fillRect(0, 0, canvasWidth, canvasHeight);

			var myid = message.payload.cid || 0;
			snakes = [];
			message.payload.snakes.forEach(function(snake, cid) {
				snakes.push(new Snake(snake.body, myid == cid));
			});

			foods = [];
			message.payload.foods.forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			return;

		case "frame":
			// 帧驱动
			message.payload.keycodes.forEach(function(kc, cid) {
				snakes[cid].frameForward(kc);
			});
			(message.payload.foods||[]).forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			return;

		case "finish":
			// 结束一局
			$('#btn-1p').removeAttr('disabled')

			canvas.font = "40px Arial";
			canvas.textAlign = "center";
			canvas.strokeStyle = "red";
			canvas.strokeText("GAME OVER", canvasWidth/2, canvasHeight/2);
			return;
		}
	});
});
