const LocalStrategy = require('passport-local').Strategy;
const bcrypt = require('bcrypt');
const fs = require('fs');

module.exports = (passport) => {
	passport.use(new LocalStrategy(
		function(name, password, done) {
			console.log("came in with: ", name, password)
			fs.readFile('static-users.json', 'utf8', (err, contents) => {
				if (err) {
					return done(err)
				}

				let users = JSON.parse(contents)
				let usr = users[name]
				if (!usr) {
					console.log("No user ", usr)
					return done("bad password/user")
				}

				bcrypt.compare(password, usr.password, (err, res) => {
					if (err || !res) {
						console.log("Bcrypt pwd mismatch ", err)
						return done("bad password/user")
					}

					return done(null, {
						localId: name,
						displayName: name,
					})
				})
			})
		})
	)
};
