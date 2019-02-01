var foods = [];
var boxWidth = 20;  // 横向方块的个数
var boxHeight = 20; // 纵向方块的个数
var boxSize = 24;   // 方块的尺寸（像素）

var box = document.getElementById('can').getContext('2d');
function draw(seat, color) {
	box.fillStyle = color;
	box.fillRect(seat % boxWidth * boxSize + 1, ~~(seat / boxHeight) * boxSize + 1, boxSize-2, boxSize-2);
}

function Snake(body) {
	this.snake = body;  // snake 队列表示蛇身
	this.direction = 1; // -1表示向左，-20表示向上，1表示向右，20表示向下
	this.gameOver = false;
	this.snake.forEach(function(n) {
		draw(n, "lime");
	});
}
Snake.prototype.frameForward = function(keyCode) {
	if (this.gameOver) return;

	// 计算出蛇头的位置
	var newDir = [-1, -boxWidth, 1, boxWidth][keyCode - 37] || this.direction;
	this.direction = (this.snake[1] - this.snake[0] == newDir) ? this.direction : newDir;
	var head = this.snake[0] + this.direction;

	// 判断是否撞到自己或者墙壁
	if (this.snake.indexOf(head) > 0 || head < 0 || head > 399 || this.direction == 1 && head % boxWidth == 0 || this.direction == -1 && head % boxWidth == boxWidth-1) {
		// 结束！
		this.gameOver = true;

		// 画残骸
		this.snake.forEach(function(n) {
			draw(n, 'darkgreen');
		});
		return;
	}

	// 画出蛇头
	this.snake.unshift(head);
	draw(head, "lime");

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

	var cid = 0; // my client id
	var snakes = {};

	astilectron.onMessage(function(message) {
		console.log("onMessage:", JSON.stringify(message));
		switch (message.name) {
		case "about":
			return;

		case "kick-off":
			// 开局
			box.fillStyle = 'black';
			box.fillRect(0, 0, 400, 400);
			message.payload.foods.forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			cid = message.payload.cid || 0;
			snakes[cid] = new Snake(message.payload.snakes[cid].body);
			return;

		case "frame":
			// 帧驱动
			var kc = message.payload.keycodes[cid];
			snakes[cid].frameForward(kc);
			(message.payload.foods||[]).forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			return;

		case "finish":
			// 结束一局
			$('#btn-1p').removeAttr('disabled')

			box.font = "40px Arial";
			box.textAlign = "center";
			box.strokeStyle = "red";
			box.strokeText("GAME OVER", boxWidth*boxSize/2, boxHeight*boxSize/2);
			return;
		}
	});
});
