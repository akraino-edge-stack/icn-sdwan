# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

#################################################################
# EWOCTL - CLI for EWO
#################################################################

Ewoctl is command line tool for interacting with EWO.
All commands take input a file. An input file can contain one or more resources.


### Syntax for describing a resource

```
version: <domain-name>/<api-version>
resourceContext:
  anchor: <URI>
metadata:
   Name: <name>
   Description: <text>
   userData1: <text>
   userData2: <text>
spec:
  <key>: <value>
```

### Example resource file

```
version: ewo/v1
resourceContext:
  anchor: overlays
metadata:
   Name: overlay1
   Description: test
   userData1: test1
   userData2: test2

---
version: ewo/v1
resourceContext:
  anchor: overlays/overlay1/ipranges
metadata:
  name: iprange1
  description: test
  userData1: test1
  userData2: test2
spec:
  subnet: 10.10.10.10
  minIp: 0
  maxIp: 255
```

### EWO CLI Commands

1. Create Ewo Resources

This command will apply the resources in the file. The user is responsible to ensuring the hierarchy of the resources.

`$ ewoctl apply -f filename.yaml`

2. Get Ewo Resources

Get the resources in the input file. This command will use the metadata name in each of the resources in the file to get information about the resource.

`$ ewoctl get -f filename.yaml`

For getting information for one resource anchor can be provided as an arguement

`$ ewoctl get <anchor>`

`$ ewoctl get overlays/overlay1`

3. Delete Ewo Resources

Delete resources in the file. The ewoctl will start deleting resources in the reverse order than given in the file to maintain hierarchy. This command will use the metadata name in each of the resources in the file to delete the resource..

`$ ewoctl delete -f filename.yaml`

For deleting one resource anchor can be provided as an arguement

`$ ewoctl delete <anchor>`

`$ ewoctl delete overlays/overlay1`

4. Update Ewo Resources

This command will call update (PUT) for the resources in the file.

`$ ewoctl update -f filename.yaml`


### Running the ewoctl

```
* Make sure that the ewoctl is build. You can build it by issuing the 'make' command.
Dir : $EWO_HOME/src/tools/ewoctl
```
* Then run the ewoctl by command:
```
./ewoctl --config ./examples/ewo-cfg.yaml apply -f ./examples/test.yaml

```

Here, ewo-cfg.yaml contains the config/port details of each of the microservices you are using.
A sample configuration is :

```
  ewo:
    host: localhost
    port: 9015
```

### Running the ewoctl with template file

```
* Ewoctl supports template values in the input file. The example input file with this feature is
examples/test_template.yaml. This file can be used with examples/values.yaml like below.
```
* Then run the ewoctl with values file:
```
ewoctl --config ./examples/ewo-cfg.yaml apply -f ./examples/test_template.yaml -v ./examples/values.yaml

```

### Running the ewoctl with token

```
* Ewoctl supports JWT tokens for interacting with EWO when EWO services are running with Istio Ingress and OAuth2 server.
```
* Then run the ewoctl with values file:
```
ewoctl --config ./examples/ewo-cfg.yaml apply -f ./examples/test.yaml -t "<token>"

```