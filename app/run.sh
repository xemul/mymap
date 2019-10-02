#!/bin/bash

CONFIG="google-config.txt"
export GOOGLE_CLIENT_ID=$(awk '/client/{print $2}' $CONFIG)
export GOOGLE_CLIENT_SECRET=$(awk '/secret/{print $2}' $CONFIG)
export GOOGLE_REDIRECT_URL=$(awk '/redirect/{print $2}' $CONFIG)

node app.js
