const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const passport = require('passport');
const auth = require('./auth');
const session = require('express-session');

auth(passport)

app.use(session({ secret: process.env.SESSION_SECRET, cookie: { maxAge: 60000 }, resave: false, saveUninitialized: false }));
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
		console.log('Logged in as ', req.user.profile)
		req.session.user = {
			id: req.user.profile.id,
			name: req.user.profile.displayName,
			emails: req.user.profile.emails,
		}
		res.redirect('/')
	}
)


app.listen(8081, function () {
  console.log('Example app listening on port 8081!');
});
