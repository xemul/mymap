# mymap
OSM-based map painter and visits keeper

To start

* git clone https://github.com/hjnilsson/country-flags
* cd static/img/ && ln -s .../country-flags flags
* cp noflag.svg flags/xx.svg
* dd if=/dev/urandom bs=128 count=1 | base64 -w0 > key.txt
* /* prepare google-config.txt in app/ dir */

* cd app/
* npm install express
* npm install passport
* npm install passport-google-oauth
* npm install body-parser
* npm install express-session
* npm install jsonwebtoken
* ./run.sh

* cd api/
* go get github.com/gorilla/mux
* go get github.com/gorilla/handlers
* go get github.com/dgrijalva/jwt-go
* go build
* ./run.sh

* $browser http://localhost:8081
