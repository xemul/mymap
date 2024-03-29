const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const passport = require('passport');
const gauth = require('./auth-google');
const lauth = require('./auth-local');
const session = require('express-session');
const jwt = require('jsonwebtoken');
const parseURL = require('parseurl');

const tokenKey = new Buffer(process.env.JWT_SIGN_KEY, 'base64')

gauth(passport)
lauth(passport)

app.use(session({ secret: process.env.SESSION_SECRET, cookie: { }, resave: false, saveUninitialized: false }));
app.use(passport.initialize())
app.use('/static', express.static('static'))
app.use(bodyParser.json())

app.get('/', (req, res) => {
	res.redirect('/map')
})

app.get('/map', (req, res) => {
	let map = req.query.viewmap

	if (!map || map == 'my') {
		req.session.viewmap = ""
	} else {
		req.session.viewmap = map
	}

	res.sendFile(__dirname + '/static/map.html')
})

app.get('/config', (req, res) => {
	res.json({
		viewmap: req.session.viewmap,
		backend: 'http://localhost:8082',
	})
})

app.get('/login', (req, res) => {
	res.redirect('/static/login')
})

app.get('/logout', (req, res) => {
	req.logout();
	req.session.user = null;
	res.redirect('/');
})

app.get('/creds', (req, res) => {
	if (req.session.user) {
		res.json(req.session.user)
	} else {
		res.status(401).send('not authenticated')
	}
})

app.get('/auth/local', passport.authenticate('local'),
	(req, res) => {
		let user = {
			id: 'local.' + req.user.localId,
			name: req.user.displayName,
		}

		console.log('Logged in ', user)

		user.token = jwt.sign( {
				id: user.id,
				exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60),
			}, tokenKey)

		req.session.user = user
		res.redirect('/')
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
		let user = {
			id: 'google.' + req.user.profile.id,
			name: req.user.profile.displayName,
			emails: req.user.profile.emails,
		}

		console.log('Logged in ', user)

		user.token = jwt.sign( {
				id: user.id,
				exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60),
			}, tokenKey)

		req.session.user = user
		res.redirect('/')
	}
)


app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
