// 从0到399表示box里[0~19]*[0~19]的所有节点，每20px一个节点
var box = document.getElementById('can').getContext('2d');
function draw(seat, color) {
	box.fillStyle = color;
	box.fillRect(seat % 20 * 20 + 1, ~~(seat / 20) * 20 + 1, 18, 18);
	// 用color填充一个矩形，以前两个参数为x，y坐标，后两个参数为宽和高。
}

var foods = [];

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
	var newDir = [-1, -20, 1, 20][keyCode - 37] || this.direction;
	this.direction = (this.snake[1] - this.snake[0] == newDir) ? this.direction : newDir;
	var head = this.snake[0] + this.direction;
	this.snake.unshift(head);

	// 判断是否撞到自己或者墙壁
	if (this.snake.indexOf(head, 1) > 0 || head < 0 || head > 399 || this.direction == 1 && head % 20 == 0 || this.direction == -1 && head % 20 == 19) {
		// 结束！
		this.gameOver = true;
		$('#btn-1p').removeAttr('disabled')
		box.font = "40px Arial";
		box.strokeStyle = "red";
		box.textAlign = "center";
		box.strokeText("GAME OVER", 200, 200);
		return;
	}

	// 画出蛇头
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

	var myid = 0;
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
			myid = message.payload.myid;
			snakes[myid] = new Snake(message.payload.snakes[myid].body);
			return;

		case "frame":
			// 帧驱动
			var kc = message.payload.keycodes[myid];
			snakes[myid].frameForward(kc);
			return;
		}
	});
});
