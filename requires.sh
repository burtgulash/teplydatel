#!/bin/bash

echo install golang
sudo apt-get install npm golang

# http://stackoverflow.com/questions/26320901/cannot-install-nodejs-usr-bin-env-node-no-such-file-or-directory
sudo ln -s /usr/bin/nodejs /usr/bin/node

echo install typescript, bower, gulp
sudo npm install -g typescript
sudo npm install -g bower
sudo npm install -g gulp


echo install gulp plugins
sudo npm install --save-dev main-bower-files
sudo npm install --save-dev del
sudo npm install --save-dev gulp
sudo npm install --save-dev gulp-util
sudo npm install --save-dev gulp-rename
sudo npm install --save-dev gulp-concat
sudo npm install --save-dev gulp-uglify
sudo npm install --save-dev gulp-sass
sudo npm install --save-dev gulp-typescript
sudo npm install --save-dev gulp-sourcemaps
sudo npm install --save-dev gulp-flatten
sudo npm install --save-dev gulp-filter
sudo npm install --save-dev gulp-minify-css

echo install bower dependencies defined in bower.json
bower install
