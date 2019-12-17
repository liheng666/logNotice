#!/bin/bash

# 交叉编译Linux64位系统版本

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
bin="./bin/conf/"
if [ ! -e bin ]; then
  mkdir -p ./bin/conf/
fi

mv ./logNotice ./bin/logNotice
cp ./conf/conf.yaml.default ./bin/conf/conf.yaml.default
