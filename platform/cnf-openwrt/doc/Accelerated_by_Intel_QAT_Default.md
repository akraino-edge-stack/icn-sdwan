---
* [Introduction](#introduction)
* [Requirements](#requirements)
    * [Hardware](#hardware)
    * [Software](#software)
* [Quick Start](#quick-start)
* [SDEWAN CNF setup](#sdewan-cnf-setup)

## Introduction

Intel® QuickAssist Technology (Intel® QAT) is developed by Intel® and runs on Intel® architecture providing security and compression acceleration capabilities to improve performance and efficiency. It offloads workloads from the CPU to hardware. Server, networking, big data, and storage applications use Intel® QAT to offload compute-intensive operations, such as:

- Symmetric cryptography functions, including cipher operations and authentication operations.
- Public key functions, including RSA, Diffie-Hellman, and elliptic curve cryptography.
- Compression and decompression functions, including DEFLATE.

Intel® QAT has improved the function of in many areas, such as Hadoop* acceleration, OpenSSL integration, SDN and NFV Solutions Boost and so on. On a more detailed level, we can get the benefits from the following aspects:

- 4G LTE and 5G encryption algorithm offload for mobile gateways and infrastructure.
- VPN traffic acceleration, with up to 50 Gbps crypto throughput and support for IPsec and SSL acceleration.
- Compression/decompression, with up to 24 Gbps throughput.
- I/O virtualization using PCI-SIG Single-Root I/O Virtualization (SR-IOV).

---

## Requirements

### Hardware

- Intel® QAT device on your machine with the following forms:
  - Chipset: Intel® C6xx series chipset
  - PCIE: Intel® QuickAssist Adapter 89xx
  - SoC: Intel Atom® C3000 processor series (Denverton NS) / Rangeley

### Software

- Linux Kernel Version >= 4.16

Notice that the `linux kernel` we used should be version >=4.16. Because the [bug](https://patchwork.ozlabs.org/project/netdev/patch/2f86c9d7c39cfad21fdb353a183b12651fc5efe9.1583311902.git.lucien.xin@gmail.com/) before v4.15 and fixed [here]( https://elixir.bootlin.com/linux/v4.16-rc1/source/net/xfrm/xfrm_device.c).

---

## Quick Start

SDEWAN CNF is configurated to use LKCF (Linux* Kernel Crypto Framework) as the encryption engine to setup IPSec tunnel between clusters, and LKCF will use Intel® QAT automatically to accelerate data traffic when the QAT device available. To guarantee the performance of traffic processing, when deploy SDEWAN CNF in a k8s cluster with Intel QuickAssist Technology (QAT) device plugin enabled, we recommend deploying the Intel® QAT device plugin and SDEWAN CNF in separated nodes. 
This can be done by configuring podAntiAffinity in intel qat plugin’s deployment file. e.g. if SDEWAN CNF is deployed on a node with label “sdewan-cnf:enabled”, then add the following podAntiAffinity rule for QAT device plugin:

```yaml
...
spec:
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
        matchExpressions:
        - key: sdewan-cnf
          operator: In
          values:
          - enabled
...

```

---

## SDEWAN CNF setup

We recommend you to deploy the SDEWAN CNF using the `sdewan_cnf` helm chart part in this [guide](https://github.com/akraino-edge-stack/icn-sdwan/blob/master/platform/deployment/README.md). And suppose you have successfully deployed two CNFs and the network between them is accessible.

Note: We have two pre-build images `sdewan-cnf:qat` and `sdewan-cnf:qat-test`, you can use them to deploy you strongswan ipsec tunnel accelerated by your host Intel® QAT devices, what you should do is only replace the image field in Helm chart value. `qat-test` tag is used for performance testing. Of course, you can configure your custemer images based on `openwrt` to enbale QAT accelerating workload as the following guide.

```shell
# We use `kubectl exec` to login to the CNF container and execute the following command
# under root to enable QAT and load testing tools

opkg update
opkg install vim
opkg install iperf3
opkg install pciutils
opkg install tcpdump
opkg install strongswan-mod-af-alg

# Remove the alogrithm plugin not in linux kernel crypto framework
cd /etc/strongswan.d/charon
rm aes.conf des.conf sha1.conf sha2.conf fips-prf.conf md5.conf
```

Then we can configure the ipsec configuration and secrets for the tunnel established.

```yaml
# Add the following section to each container under `/etc/ipsec.conf`, note that they
# would have an opposite value in left* and right*
conn con
       authby=secret
       auto=add
       type=tunnel
       leftid=l
       left=<left_container_ip>
       rightid=r
       right=<right_container_ip>
       ike=aes256-sha1-modp2048
       keyexchange=ikev2
       ikelifetime=28800s
       esp=aes256-sha1-modp1536
```

and add the following simple key to `/etc/ipsec.secrets`

```yaml
l r : PSK 'AAABBCCDD'
```

Then we can start `ipsec` and load the connection to test the performance

```shell
ipsec restart
ipsec up con

# On one CNF, we start a `iperf3` server to receive the packages
iperfs -s
# On another CNF, we use `iperf3` as a client to send packages to server
# TCP
iperf3 -c <iperf3_server_ip> -b 10g
# UDP
iperf3 -c <iperf3_server_ip> -l 64k -b 10g -u

```

