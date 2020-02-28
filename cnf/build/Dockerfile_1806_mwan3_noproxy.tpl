FROM openwrt-1806-4-base

#EXPOSE 80

RUN mkdir /var/lock && \
    opkg update && \
    opkg install uhttpd-mod-lua && \
    uci set uhttpd.main.interpreter='.lua=/usr/bin/lua' && \
    uci commit uhttpd && \
    opkg install mwan3 && \
    opkg install luci-app-mwan3; exit 0

COPY system /etc/config/system
COPY rest_v1 /usr/lib/lua/luci/controller/rest_v1

USER root

# using exec format so that /sbin/init is proc 1 (see procd docs)
CMD ["/sbin/init"]
