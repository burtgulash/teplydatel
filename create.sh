#!/bin/bash

echo 1. gulp clean
gulp clean

echo 2. gulp build-css
gulp build-css

echo 3. gulp build-javascript
gulp build-js

echo 4. gulp publish-bower-components
gulp publish-bower-components

echo 4. go get
go get

echo 5. go build
go build
