var express = require('express');
var app = express();
const bodyParser = require('body-parser');

app.use(express.static('static'))
app.use(bodyParser.json())

app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
