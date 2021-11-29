--- SPDX-License-Identifier: Apache-2.0
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.ifutil", package.seeall)

NX = require("nixio")
io = require "io"
json = require "luci.jsonc"
sys = require "luci.sys"
util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"

fields_table = {
    {field="ip_address", key="inet", type="array", format=function(data) return format_ip(data) end},
    {field="mac_address", key="link/ether"},
    {field="ip6_address", key="inet6", format=function(data) return format_ip(data) end},
}

function index()
end

function is_interface_available(interface)
    local f = io.open("/sys/class/net/" .. interface .. "/operstate", "r")
    if f == nil then
        return false
    end
    f:close()
    return true
end

function format_ip(data)
    local i, j = string.find(data, "/")
    if i ~= nil then
        return string.sub(data, 1, i-1)
    end
    return data
end

function get_field(data, key, field_type, format)
    if type(key) == "function" then
        return key(data)
    end

    local reg = {
        key .. " [^%s]+[%s]",
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
                if format ~= nil and type(format) == "function" then
                    value = format(value)
                end

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
    local data = util.exec("ip a show dev " .. interface)
    if data == nil then
        for j=1, 3 do
            utils.log("ip command failed, retrying ... ")
            NX.nanosleep(1)
            data = util.exec("ip a show dev " .. interface)
            if data ~= nil then
                break
            end
        end
    end
    ret["name"] = interface
    for i,v in pairs(fields_table) do
        local value = get_field(data, v["key"], v["type"], v["format"])
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

function get_name_by_ip(ip_addr)
    local ifs = get_interface_info()
    for i, interface in pairs(ifs) do
        if interface["ip_address"] ~= nil then
            for j, ipa in pairs(interface["ip_address"]) do
                if ipa == ip_addr then
                    return interface["name"]
                end
            end
        end
    end
    return nil
end

function get_default_ifname()
    local data = util.exec("ip route | grep '^default' 2>/dev/null")
    if data ~= nil then
        reg = "dev [^%s]+[%s]"
        for item in string.gmatch(data, reg) do
            local value = nil
            local i,j = string.find(item, "dev ")
            if i ~= nil then
                value = string.sub(item, j+1, string.len(item)-1)
                return value
            end
        end
    end
    return nil
end
