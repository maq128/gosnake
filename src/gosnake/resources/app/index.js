var boxWidth = 20;  // 横向的方块个数
var boxHeight = 20; // 纵向的方块个数
var boxSize = 24;   // 方块的尺寸（像素）
var canvasWidth = 480;
var canvasHeight = 480;

var snakes = [];
var foods = [];

var canvas = $('#canvas').get(0).getContext('2d');
function draw(seat, color) {
	canvas.fillStyle = color;
	canvas.fillRect(seat % boxWidth * boxSize + 1, ~~(seat / boxHeight) * boxSize + 1, boxSize-2, boxSize-2);
}

function Snake(body, isMe) {
	this.body = body;  // 蛇身
	this.colorHead = isMe ? 'lime' : 'red'; // 蛇身颜色
	this.colorDead = isMe ? 'darkgreen' : 'darkred'; // 残骸颜色

	this.body.forEach((n) => {
		draw(n, this.colorHead);
	});
}

function playFrame(keycodes) {
	snakes.forEach(function(snake, cid) {
		if (snake == null) return; // 此蛇已死
		var keyCode = keycodes[cid];

		// 根据按键操作计算出行进方向
		var oldDir = snake.body[0] - snake.body[1];
		var newDir = [-1, -boxWidth, 1, boxWidth][keyCode - 37] || oldDir;
		// 选择有效的行进方向（不能逆向）
		newDir = (newDir + oldDir == 0) ? oldDir : newDir;
		// 蛇头的位置
		var head = snake.body[0] + newDir;

		// 判断是否撞到墙壁
		var gameOver = false;
		if (head < 0 // 上边界
			|| head >= boxWidth*boxHeight // 下边界
			|| newDir == 1 && head % boxWidth == 0 // 右边界
			|| newDir == -1 && head % boxWidth == boxWidth-1 // 左边界
			) {
			gameOver = true;
		}

		// 判断是否撞到自己或其它蛇身
		snakes.forEach((snake) => {
			if (snake == null) return; // 此蛇已死
			if (snake.body.indexOf(head) >= 0) {
				gameOver = true;
			}
		});

		if (gameOver) {
			// 画残骸
			snake.body.forEach((n) => {
				draw(n, snake.colorDead);
			});
			snakes[cid] = null;
			return;
		}

		// 蛇头并入身体
		snake.body.unshift(head);

		// 画出蛇头
		draw(head, snake.colorHead);

		// 判断是否吃到食物
		var pos = foods.indexOf(head);
		if (pos >= 0) {
			// 吃到，清除食物
			foods.splice(pos, 1);
		} else {
			// 没有吃到，清除蛇尾
			draw(snake.body.pop(), "black");
		}
	});
}

document.addEventListener('astilectron-ready', function() {
	$('button')
		.removeAttr('disabled')
		.on('click', function(evt) {
			$('button').attr('disabled', true);
			// 请求后端开始一局
			astilectron.sendMessage({
				name: 'start',
				payload: $(evt.target).attr('data-mode') * 1,
			});

			canvas.fillStyle = 'black';
			canvas.fillRect(0, 0, canvasWidth, canvasHeight);
			canvas.font = "40px Arial";
			canvas.textAlign = "center";
			canvas.strokeStyle = "yellow";
			canvas.strokeText("Waiting ...", canvasWidth/2, canvasHeight/2);
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

	astilectron.onMessage(function(message) {
		// console.log("onMessage:", JSON.stringify(message));
		switch (message.name) {
		case "about":
			canvas.font = "15px Arial";
			canvas.textAlign = "center";
			canvas.fillStyle = "gray";
			canvas.fillText("GoSnake, powered by Golang/Electron/Astilectron", canvasWidth/2, 20);
			return;

		case "kick-off":
			// 开局
			boxWidth = message.payload.width;
			boxHeight = message.payload.height;
			boxSize = canvasWidth / boxWidth;
			canvasHeight = boxSize * boxHeight;

			// 清场
			canvas.fillStyle = 'black';
			canvas.fillRect(0, 0, canvasWidth, canvasHeight);

			// 初始化每条蛇
			var myid = message.payload.cid || 0;
			snakes = [];
			message.payload.snakes.forEach(function(snake, cid) {
				snakes.push(new Snake(snake.body, myid == cid));
			});

			// 初始化所有食物
			foods = [];
			message.payload.foods.forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			return;

		case "frame":
			// 帧驱动
			playFrame(message.payload.keycodes);
			(message.payload.foods||[]).forEach(function(food) {
				foods.push(food);
				draw(food, "yellow");
			});
			return;

		case "finish":
			// 结束一局
			$('button').removeAttr('disabled')

			canvas.font = "40px Arial";
			canvas.textAlign = "center";
			canvas.strokeStyle = "red";
			canvas.strokeText("GAME OVER", canvasWidth/2, canvasHeight/2);
			return;
		}
	});
});
