var food = 43,   // 食物的位置
	box = document.getElementById('can').getContext('2d');

// 从0到399表示box里[0~19]*[0~19]的所有节点，每20px一个节点
function draw(seat, color) {
	box.fillStyle = color;
	box.fillRect(seat % 20 * 20 + 1, ~~(seat / 20) * 20 + 1, 18, 18);
	// 用color填充一个矩形，以前两个参数为x，y坐标，后两个参数为宽和高。
}

function Snake(s) {
	this.snake = s;     // snake队列表示蛇身，初始节点存在但不显示
	this.direction = 1; // 1表示向右，-1表示向左，20表示向下，-20表示向上
	this.n = 0;         // 与下次移动的位置有关
	this.alive = true;
}
Snake.prototype.frameForward = function(keyCode) {
	if (!this.alive) return;
	this.direction = this.snake[1] - this.snake[0] == (this.n = [-1, -20, 1, 20][keyCode - 37] || this.direction) ? this.direction : this.n;
	this.snake.unshift(this.n = this.snake[0] + this.direction);
	// 此时的n为下次蛇头出现的位置，n进入队列
	if (this.snake.indexOf(this.n, 1) > 0 || this.n < 0 || this.n > 399 || this.direction == 1 && this.n % 20 == 0 || this.direction == -1 && this.n % 20 == 19) {
		// if语句判断贪吃蛇是否撞到自己或者墙壁，碰到时返回，结束程序
		this.alive = false;
		return alert("GAME OVER!");
	}
	draw(this.n, "lime");   // 画出蛇头下次出现的位置
	if (this.n == food) {   // 如果吃到食物时，产生一个蛇身以外的随机的点，不会去掉蛇尾
		while (this.snake.indexOf(food = ~~(Math.random() * 400)) >= 0);
		draw(food, "yellow");
	} else {                //没有吃到食物时正常移动，蛇尾出队列
		draw(this.snake.pop(), "black");
	}
}

document.addEventListener('astilectron-ready', function() {
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

	var s;

	astilectron.onMessage(function(message) {
		console.log("onMessage:", JSON.stringify(message));
		switch (message.name) {
		case "about":
			return;

		case "kick-off":
			food = message.payload.food;
			s = new Snake(message.payload.snakes[0].body);
			return;

		case "frame":
			var kc = message.payload.keycode;
			s.frameForward(kc);
			return;
		}
	});
});
