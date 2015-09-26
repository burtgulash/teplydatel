#!/bin/bash

# Run under root!

echo install golang
apt-get install npm golang

# http://stackoverflow.com/questions/26320901/cannot-install-nodejs-usr-bin-env-node-no-such-file-or-directory
ln -s /usr/bin/nodejs /usr/bin/node

echo install typescript, bower, gulp
npm install -g typescript
npm install -g bower
npm install -g gulp


echo install gulp plugins
npm install --save-dev main-bower-files
npm install --save-dev del
npm install --save-dev gulp
npm install --save-dev gulp-util
npm install --save-dev gulp-rename
npm install --save-dev gulp-concat
npm install --save-dev gulp-uglify
npm install --save-dev gulp-sass
npm install --save-dev gulp-typescript
npm install --save-dev gulp-sourcemaps
npm install --save-dev gulp-flatten
npm install --save-dev gulp-filter

echo install bower dependencies defined in bower.json
bower install
