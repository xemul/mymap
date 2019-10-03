const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const passport = require('passport');
const auth = require('./auth');
const session = require('express-session');
const jwt = require('jsonwebtoken');
const parseURL = require('parseurl');

const tokenKey = new Buffer(process.env.JWT_SIGN_KEY, 'base64')

auth(passport)

app.use(session({ secret: process.env.SESSION_SECRET, cookie: { }, resave: false, saveUninitialized: false }));
app.use(passport.initialize())
app.use(express.static('static'))
app.use(bodyParser.json())

app.get('/login', (req, res) => {
	res.redirect('/auth/google')
})

app.get('/logout', (req, res) => {
	req.logout();
	req.session.user = null;
	res.redirect('/');
})

app.get('/config', (req, res) => {
	res.json({
		viewmap: req.session.viewmap,
		backend: 'http://localhost:8082',
	})
})

app.get('/map/*', (req, res) => {
	let map = parseURL(req).pathname.substring(5)
	if (map == 'my') {
		req.session.viewmap = ""
	} else {
		req.session.viewmap = map
	}
	res.redirect('/')
})

app.get('/creds', (req, res) => {
	if (req.session.user) {
		res.json(req.session.user)
	} else {
		res.status(401).send('not authenticated')
	}
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
			id: req.user.profile.id,
			name: req.user.profile.displayName,
			emails: req.user.profile.emails,
		}

		console.log('Logged in ', user)

		user.token = jwt.sign( {
				id: "google." + user.id,
				exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60),
			}, tokenKey)

		req.session.user = user
		res.redirect('/')
	}
)


app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
