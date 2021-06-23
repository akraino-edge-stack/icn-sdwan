--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.modules.wan", package.seeall)

util = require "luci.util"

function register()
    return "wan", _M["get_wan_info"]
end

function get_wan_info()
    return util.ubus("mwan3", "status", {})
end