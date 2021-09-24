#!/bin/bash

#SPDX-License-Identifier: Apache-2.0
#Copyright (c) 2021 Intel Corporation

# usage: build_images.sh

set -ex
docker_file=Dockerfile_sdewan
image_tag=openwrt-sdewan:v0.1

# generate Dockerfile
test -f ./set_proxy && . set_proxy
docker_proxy=${docker_proxy-""}
if [ -z "$docker_proxy" ]; then
    cp ${docker_file}_noproxy.tpl $docker_file
else
    cp $docker_file.tpl $docker_file
    sed -i "s,{docker_proxy},$docker_proxy,g" $docker_file
fi

# build docker images for openwrt with wman3
docker build --network=host -f $docker_file -t $image_tag .

# clear
rm -rf $docker_file
