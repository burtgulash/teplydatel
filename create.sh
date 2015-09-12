#!/bin/bash

echo 1. typescript compile
tsc ts/logic.ts --outFile js/logic.js
tsc ts/race.ts --outFile js/race.js

echo 2. go build
go build

