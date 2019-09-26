var express = require('express');
var app = express();
const bodyParser = require('body-parser');

app.use(express.static('static'))
app.use(bodyParser.json())

app.post('/visited', (req, resp) => {
	console.log("Visited: ", req.body);
	resp.sendStatus(200)
})

app.get('/visited', (req, resp) => {
	resp.json([])
})

app.delete('/visited', (req, resp) => {
	console.log("Unvisited: ", req.query.id)
	resp.sendStatus(204)
})

app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
