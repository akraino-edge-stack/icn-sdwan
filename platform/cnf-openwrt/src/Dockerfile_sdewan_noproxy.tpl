# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

FROM openwrtorg/rootfs:x86-64-openwrt-19.07

#EXPOSE 80

RUN mkdir /var/lock && \
    opkg update && \
    opkg install shadow-chpasswd sudo && \
    opkg install luci-ssl && \
    opkg install uhttpd-mod-lua && \
    uci set uhttpd.main.interpreter='.lua=/usr/bin/lua' && \
    uci commit uhttpd && \
    opkg install shadow-useradd shadow-groupadd shadow-usermod && \
    opkg install mwan3 jq bash conntrack && \
    opkg install strongswan-default luasocket strongswan-mod-af-alg && \
    opkg install luci-app-mwan3; exit 0

COPY strongswan.conf /etc/strongswan.conf
COPY system /etc/config/system
COPY ipsec /etc/config/ipsec
COPY ipsec_exec /etc/init.d/ipsec
COPY updown /etc/updown
COPY updown_oip /etc/updown_oip
COPY sdewan.user /etc/sdewan.user
COPY sdewan_svc.info /etc/sdewan_svc.info
COPY app_cr.info /etc/app_cr.info
COPY route_cr.info /etc/route_cr.info
COPY rule_cr.info /etc/rule_cr.info
COPY default_firewall /etc/config/firewall
COPY rest_v1 /usr/lib/lua/luci/controller/rest_v1
COPY 10-default.conf /etc/sysctl.d/10-default.conf

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN groupadd --system sudo && useradd wrt
RUN usermod -a -G sudo wrt

USER wrt

# using exec format so that /sbin/init is proc 1 (see procd docs)
CMD ["/sbin/init"]
