Your distribution may already have the Intel® QAT drivers, these instructions guide the user on how to download the kernel sources, compile kernel driver modules against those sources, and load them onto the host.

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
- Ensure SR-IOV and VT-d are enabled in BIOS

### Software

- Enable IOMMU and reboot

  Enable IOMMU support by setting the correct [kernel parameter](https://wiki.archlinux.org/index.php/Kernel_parameter) depending on the type of CPU in use:

  - For Intel® CPUs (VT-d) set `intel_iommu=on`
  - For AMD CPUs (AMD-Vi) set `amd_iommu=on`

  You should also append the `iommu=pt` parameter. This will prevent Linux from touching devices which cannot be passed through.

  > Edit the file /etc/default/grub and change the `GRUB_CMDLINE_LINUX`

  ```shell
  vim /etc/default/grub
  # Change and save
  ...
  GRUB_CMDLINE_LINUX="iommu=pt, intel_iommu=on, apparmor=0"
  ...

  sudo grub2-mkconfig -o /boot/grub2/grub.cfg
  sudo reboot

  # We can check it
  dmesg | grep -i iommu
  [    0.000000] DMAR: IOMMU enabled
  ```

- Install needed software package

  ```shell
  # For CentOS
  yum -y groupinstall "Development Tools"
  yum -y install pciutils
  yum -y install libudev-devel
  yum -y install kernel-devel-$(uname -r)
  yum -y install gcc
  yum -y install openssl-devel

  # For Ubuntu
  apt-get update
  apt-get install pciutils-dev

  # For Fedora
  dnf groupinstall "Development Tools"
  dnf install gcc-c++
  dnf install systemd-devel
  dnf install kernel-devel-`uname -r`
  dnf install elfutils-devel
  dnf install openssl-devel
  ```

---

## Quick Start

Notice that the `linux kernel` we used should be version >=4.16. Because the [bug](https://patchwork.ozlabs.org/project/netdev/patch/2f86c9d7c39cfad21fdb353a183b12651fc5efe9.1583311902.git.lucien.xin@gmail.com/) before v4.15 and fixed [here]( https://elixir.bootlin.com/linux/v4.16-rc1/source/net/xfrm/xfrm_device.c).

1. Get the Intel® QAT driver from 01.org

   ```shell
   mkdir qat_driver
   cd qat_driver
   wget https://downloadmirror.intel.com/649693/QAT.L.4.15.0-00011.tar.gz
   tar zxvf QAT.L.4.15.0-00011.tar.gz
   ```

2. Configure the Intel® QAT driver and enable LKCF

   ```shell
   # You can list the configuration by this command
   ./configure -h
   # We need enable the Intel® QAT sriov
   ./configure --enable-icp-sriov=host --enable-qat-lkcf
   ```

3. Build driver and install driver and samples

   ```shell
   make -j <number>
   make install
   make samples-install
   ```

4. Check the driver module

   ```shell
   lsmod | grep qa
     qat_c62xvf             16384  0
     qat_c62x               20480  0
     intel_qat             212992  3 qat_c62x,usdm_drv,qat_c62xvf
     uio                    20480  1 intel_qat
   ```

5. Configure the QAT device

   SDEWAN CNF will use QAT PFs/VFs bound in kernel driver, we should configure them with kernel section enabled in file like `/etc/c6xx_dev*.conf` for QAT PFs and `/etc/c6xxvf_dev*.conf` for QAT VFs. You also can replace them with the following examples.

   ```shell
   # An example for PF configuration
   ################################################################
   # This file is provided under a dual BSD/GPLv2 license.  When using or
   #   redistributing this file, you may do so under either license.
   #
   #   GPL LICENSE SUMMARY
   #
   #   Copyright(c) 2007-2021 Intel Corporation. All rights reserved.
   #
   #   This program is free software; you can redistribute it and/or modify
   #   it under the terms of version 2 of the GNU General Public License as
   #   published by the Free Software Foundation.
   #
   #   This program is distributed in the hope that it will be useful, but
   #   WITHOUT ANY WARRANTY; without even the implied warranty of
   #   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   #   General Public License for more details.
   #
   #   You should have received a copy of the GNU General Public License
   #   along with this program; if not, write to the Free Software
   #   Foundation, Inc., 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
   #   The full GNU General Public License is included in this distribution
   #   in the file called LICENSE.GPL.
   #
   #   Contact Information:
   #   Intel Corporation
   #
   #   BSD LICENSE
   #
   #   Copyright(c) 2007-2021 Intel Corporation. All rights reserved.
   #   All rights reserved.
   #
   #   Redistribution and use in source and binary forms, with or without
   #   modification, are permitted provided that the following conditions
   #   are met:
   #
   #     * Redistributions of source code must retain the above copyright
   #       notice, this list of conditions and the following disclaimer.
   #     * Redistributions in binary form must reproduce the above copyright
   #       notice, this list of conditions and the following disclaimer in
   #       the documentation and/or other materials provided with the
   #       distribution.
   #     * Neither the name of Intel Corporation nor the names of its
   #       contributors may be used to endorse or promote products derived
   #       from this software without specific prior written permission.
   #
   #   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
   #   "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
   #   LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
   #   A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
   #   OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
   #   SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
   #   LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
   #   DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
   #   THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
   #   (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
   #   OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
   #
   #
   #  version: QAT.L.4.15.0-00011
   ################################################################
   [GENERAL]
   ServicesEnabled = cy;dc

   # Set the service profile to determine available features
   # =====================================================================
   #                               DEFAULT    CRYPTO  COMPRESSION  CUSTOM1
   # Asymmetric Crypto                *         *                     *
   # Symmetric Crypto                 *         *                     *
   # Hash                             *         *          *          *
   # Cipher                           *         *                     *
   # MGF KeyGen                       *         *
   # SSL/TLS KeyGen                   *         *                     *
   # HKDF                                       *                     *
   # Compression                      *                    *          *
   # Decompression (stateless)        *                    *          *
   # Decompression (stateful)         *                    *
   # Service Chaining                                      *
   # Device Utilization                         *          *          *
   # Rate Limiting                              *          *          *
   # =====================================================================
   ServicesProfile = DEFAULT

   ConfigVersion = 2

   #Default values for number of concurrent requests*/
   CyNumConcurrentSymRequests = 512
   CyNumConcurrentAsymRequests = 64

   #Statistics, valid values: 1,0
   statsGeneral = 1
   statsDh = 1
   statsDrbg = 1
   statsDsa = 1
   statsEcc = 1
   statsKeyGen = 1
   statsDc = 1
   statsLn = 1
   statsPrime = 1
   statsRsa = 1
   statsSym = 1


   # Specify size of intermediate buffers for which to
   # allocate on-chip buffers. Legal values are 32 and
   # 64 (default is 64). Specify 32 to optimize for
   # compressing buffers <=32KB in size.
   DcIntermediateBufferSizeInKB = 64

   # This flag is to enable device auto reset on heartbeat error
   AutoResetOnError = 0

   ##############################################
   # Kernel Instances Section
   ##############################################
   [KERNEL]
   NumberCyInstances = 1
   NumberDcInstances = 0

   # Crypto - Kernel instance #0
   Cy0Name = "IPSec0"
   Cy0IsPolled = 0
   Cy0CoreAffinity = 0

   ##############################################
   # User Process Instance Section
   ##############################################
   [SSL]
   NumberCyInstances = 6
   NumberDcInstances = 2
   NumProcesses = 1
   LimitDevAccess = 0

   # Crypto - User instance #0
   Cy0Name = "SSL0"
   Cy0IsPolled = 1
   # List of core affinities
   Cy0CoreAffinity = 0

   # Crypto - User instance #1
   Cy1Name = "SSL1"
   Cy1IsPolled = 1
   # List of core affinities
   Cy1CoreAffinity = 1

   # Crypto - User instance #2
   Cy2Name = "SSL2"
   Cy2IsPolled = 1
   # List of core affinities
   Cy2CoreAffinity = 2

   # Crypto - User instance #3
   Cy3Name = "SSL3"
   Cy3IsPolled = 1
   # List of core affinities
   Cy3CoreAffinity = 3

   # Crypto - User instance #4
   Cy4Name = "SSL4"
   Cy4IsPolled = 1
   # List of core affinities
   Cy4CoreAffinity = 4

   # Crypto - User instance #5
   Cy5Name = "SSL5"
   Cy5IsPolled = 1
   # List of core affinities
   Cy5CoreAffinity = 5


   # Data Compression - User instance #0
   Dc0Name = "Dc0"
   Dc0IsPolled = 1
   # List of core affinities
   Dc0CoreAffinity = 0

   # Data Compression - User instance #1
   Dc1Name = "Dc1"
   Dc1IsPolled = 1
   # List of core affinities
   Dc1CoreAffinity = 1
   ```

   **Changes compared with default configuration**

   ```shell
   # An example for VF configuration
   ################################################################
   # This file is provided under a dual BSD/GPLv2 license.  When using or
   #   redistributing this file, you may do so under either license.
   #
   #   GPL LICENSE SUMMARY
   #
   #   Copyright(c) 2007-2021 Intel Corporation. All rights reserved.
   #
   #   This program is free software; you can redistribute it and/or modify
   #   it under the terms of version 2 of the GNU General Public License as
   #   published by the Free Software Foundation.
   #
   #   This program is distributed in the hope that it will be useful, but
   #   WITHOUT ANY WARRANTY; without even the implied warranty of
   #   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   #   General Public License for more details.
   #
   #   You should have received a copy of the GNU General Public License
   #   along with this program; if not, write to the Free Software
   #   Foundation, Inc., 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
   #   The full GNU General Public License is included in this distribution
   #   in the file called LICENSE.GPL.
   #
   #   Contact Information:
   #   Intel Corporation
   #
   #   BSD LICENSE
   #
   #   Copyright(c) 2007-2021 Intel Corporation. All rights reserved.
   #   All rights reserved.
   #
   #   Redistribution and use in source and binary forms, with or without
   #   modification, are permitted provided that the following conditions
   #   are met:
   #
   #     * Redistributions of source code must retain the above copyright
   #       notice, this list of conditions and the following disclaimer.
   #     * Redistributions in binary form must reproduce the above copyright
   #       notice, this list of conditions and the following disclaimer in
   #       the documentation and/or other materials provided with the
   #       distribution.
   #     * Neither the name of Intel Corporation nor the names of its
   #       contributors may be used to endorse or promote products derived
   #       from this software without specific prior written permission.
   #
   #   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
   #   "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
   #   LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
   #   A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
   #   OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
   #   SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
   #   LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
   #   DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
   #   THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
   #   (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
   #   OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
   #
   #
   #  version: QAT.L.4.15.0-00011
   ################################################################
   [GENERAL]
   ServicesEnabled = cy;dc

   ConfigVersion = 2

   #Default values for number of concurrent requests*/
   CyNumConcurrentSymRequests = 512
   CyNumConcurrentAsymRequests = 64

   #Statistics, valid values: 1,0
   statsGeneral = 1
   statsDh = 1
   statsDrbg = 1
   statsDsa = 1
   statsEcc = 1
   statsKeyGen = 1
   statsDc = 1
   statsLn = 1
   statsPrime = 1
   statsRsa = 1
   statsSym = 1

   ##############################################
   # Kernel Instances Section
   ##############################################
   [KERNEL]
   NumberCyInstances = 1
   NumberDcInstances = 0

   # Crypto - Kernel instance #0
   Cy0Name = "IPSec0"
   Cy0IsPolled = 0
   Cy0CoreAffinity = 0
   ```

6. Start QAT service

   ```shell
   # Start QAT and QAT VFs service
   service qat_service restart
   service qat_service_vfs restart
   # Check the devices status
   adf_ctl status
   ```

Then you can utilize the Intel® QAT VFs from your privileged docker container.

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

