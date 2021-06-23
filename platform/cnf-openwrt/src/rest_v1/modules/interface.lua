--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.modules.interface", package.seeall)

NX = require("nixio")
sys = require "luci.sys"
util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"

fields_table = {
    {field="ip_address", key="inet addr", type="array"},
    {field="ip6_address", key="inet6 addr", type="array"},
    {field="mac_address", key="HWaddr"},
    {field="received_packets", key="RX packets"},
    {field="send_packets", key="TX packets"},
    {field="status", key=function(data) if string.match(data, "[%s]UP[%s]") ~= nil then return "UP" else return "DOWN" end end},
}

function register()
    return "interface", _M["get_interface_info"]
end

function get_field(data, key, field_type)
    if type(key) == "function" then
        return key(data)
    end

    local reg = {
        key .. ":[^%s]+[%s]",
        key .. " [^%s]+[%s]",
        key .. ": [^%s]+[%s]",
    }

    local ret = nil
    for index=1, #reg do
        for item in string.gmatch(data, reg[index]) do
            local value = nil
            local i,j = string.find(item, key .. ": ")
            if i ~= nil then
                value = string.sub(item, j+1, string.len(item)-1)
            else
                i,j = string.find(item, key .. ":")
                if i ~= nil then
                    value = string.sub(item, j+1, string.len(item)-1)
                else
                    i,j = string.find(item, key .. " ")
                    if i ~= nil then
                        value = string.sub(item, j+1, string.len(item)-1)
                    end
                end
            end
            if value ~= nil then
                if field_type == "array" then
                    if ret == nil then
                        ret = {value}
                    else
                        ret[#ret+1] = value
                    end
                else
                    ret = value
                    break
                end
            end
        end
    end
    return ret
end

function get_interface(interface)
    local ret = {}
    local data = util.exec("ifconfig " .. interface)
    if data == nil then
        for j=1, 3 do
            utils.log("ifconfig failed, retrying ... ")
            NX.nanosleep(1)
            data = util.exec("ifconfig " .. interface)
            if data ~= nil then
                break
            end
        end
    end
    ret["name"] = interface
    for i,v in pairs(fields_table) do
        local value = get_field(data, v["key"], v["type"])
        if value ~= nil then
            ret[v["field"]] = value
        end
    end
    return ret
end

function get_interface_info()
    local ret = {}
    local index = 1
    for interface in util.execi("ifconfig | awk '/^[^ \t]+/{print $1}'") do
        if interface ~= "lo" then
           ret[index] = get_interface(interface)
           index = index + 1
        end
    end
    return ret
end
