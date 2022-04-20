--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.index", package.seeall)

function index()
    ver = "v1"
    entry({"sdewan", ver}, call("help")).dependent = false
    entry({"sdewan", "mwan3", ver}, call("help")).dependent = false
    entry({"sdewan", "firewall", ver}, call("help")).dependent = false
    entry({"sdewan", "networkfirewall", ver}, call("help")).dependent = false
    entry({"sdewan", "ipsec", ver}, call("help")).dependent = false
    entry({"sdewan", "service", ver}, call("help")).dependent = false
    entry({"sdewan", "application", ver}, call("help")).dependent = false
    entry({"sdewan", "route", ver}, call("help")).dependent = false
    entry({"sdewan", "rule", ver}, call("help")).dependent = false
    entry({"sdewan", "nat", ver}, call("help")).dependent = false

end

function help()
    luci.http.prepare_content("application/json")
    luci.http.write('{"message":"sdewan restful API service v1"}')
end
