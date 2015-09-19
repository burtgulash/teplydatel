#!/bin/bash

echo 1. clean
[ -d ./js ] && rm -r ./js
mkdir ./js

echo 2. typescript compile
tsc ts/race.ts --outFile js/race.js

echo 3. go build
go build

