#!/bin/bash

# Run under root!

apt-get install npm golang ruby

npm install -g typescript
gem install sass

# http://stackoverflow.com/questions/26320901/cannot-install-nodejs-usr-bin-env-node-no-such-file-or-directory
ln -s /usr/bin/nodejs /usr/bin/node
