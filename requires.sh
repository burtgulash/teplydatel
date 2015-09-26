#!/bin/bash

# Run under root!

apt-get install npm golang ruby

# http://stackoverflow.com/questions/26320901/cannot-install-nodejs-usr-bin-env-node-no-such-file-or-directory
ln -s /usr/bin/nodejs /usr/bin/node

npm install -g typescript
npm install -g gulp

npm install --save-dev gulp
npm install --save-dev gulp-util
npm install --save-dev gulp-sass
npm install --save-dev gulp-typescript
npm install --save-dev gulp-sourcemaps

gem install sass
