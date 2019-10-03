#!/bin/bash

export JWT_SIGN_KEY=$(cat key.txt)
./api
