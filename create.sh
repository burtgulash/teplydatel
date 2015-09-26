#!/bin/bash

clean() {
    [ -d "./$1" ] && rm -r "./$1"
    mkdir "./$1"
}

echo 1. sass compile
clean "css"
sass --scss ./frontend/sass/race.scss > css/style.css

echo 2. typescript compile
clean "js"
tsc ./frontend/typescript/race.ts --outFile js/race.js

echo 3. go get
go get

echo 4. go build
go build

