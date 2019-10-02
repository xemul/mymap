const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const passport = require('passport');
const auth = require('./auth');

auth(passport)

app.use(passport.initialize())
app.use(express.static('static'))
app.use(bodyParser.json())

app.get('/login', (req, res) => {
	res.redirect('/auth/google')
})

app.get('/auth/google', passport.authenticate('google', {
	scope: [
		'https://www.googleapis.com/auth/userinfo.email',
		'https://www.googleapis.com/auth/userinfo.profile',
	]
}))

app.get('/auth/google/callback', passport.authenticate('google', {
		failureRedirect: '/',
	}), (req, res) => {
		console.log('Logged in as ', req.user)
		res.redirect('/')
	}
)


app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
