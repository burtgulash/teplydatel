#!/bin/bash

echo 1. go build
go build

echo 2. typescript compile
tsc ts/logic.ts --outFile js/logic.js
