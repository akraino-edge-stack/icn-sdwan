# sdewan-cnf

sdewan cnf docker image for Akraino ICN SDEWAN solution

# folder structure

* build: includes all file to generate sdewan docker image
* helm: helm chart to create sdewan CNF
* README.md: this file

# Build

Requirements:
* docker

Steps:

* Set proxy:
If proxy is required in build environment, edit `build/set_proxy` file:
docker_proxy={proxy}

* Build docker image
cd build
./build_image.sh

Note: After build, the docker image will be imported as openwrt-1806-mwan3 
  
