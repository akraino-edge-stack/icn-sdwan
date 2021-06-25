```
SPDX-License-Identifier: Apache-2.0
Copyright (c) 2021 Intel Corporation
```

# cnf-openwrt

sdewan cnf docker image for OpenNESS SDEWAN solution

# folder structure

* src: includes all file to generate sdewan docker image
* examples: sample yaml file to create CNF
* doc: documents
* README.md: this file

# Build

Requirements:
* docker

Steps:

* Set proxy:
If proxy is required in build environment, edit `build/set_proxy` file:
docker_proxy={proxy}

* Build docker image
cd src
./build_image.sh

Note: After build, the docker image will be imported as openwrt-1806-mwan3
