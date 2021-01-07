-- Copyright 2020 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.modules.wan", package.seeall)

util = require "luci.util"

function register()
    return "wan", _M["get_wan_info"]
end

function get_wan_info()
    return util.ubus("mwan3", "status", {})
end