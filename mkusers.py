import bcrypt
import json

users = {}

with open('static-users.txt') as suf:
	for user in suf:
		(name, pwd) = user.strip().split(' ', 2)
		users[name] = { 'password': bcrypt.hashpw(pwd, bcrypt.gensalt()) }

print(json.dumps(users, indent=4, sort_keys=True))
