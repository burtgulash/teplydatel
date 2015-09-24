#!/bin/bash

clean() {
    [ -d "./$1" ] && rm -r "./$1"
    mkdir "./$1"
}

echo 1. sass compile
clean "css"
sass --scss ./sass/*.scss > ./css/style.css

echo 2. typescript compile
clean "js"
tsc ./ts/race.ts --outFile js/race.js

echo 3. go build
go build

