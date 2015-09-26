#!/bin/bash

echo 1. gulp build-css
gulp build-css

echo 2. gulp build-javascript
gulp build-javascript

echo 3. go get
go get

echo 4. go build
go build
