--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.modules.route", package.seeall)

NX = require("nixio")
sys = require "luci.sys"
util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"

fields_table = {
    {field="gateway", key="via"},
    {field="device", key="dev"},
    {field="destination", key=function(data) return split(data, ' ')[1] end},
    {field="scope", key="scope"},
    {field="proto", key="proto"},
    {field="metric", key="metric"},
    {field="src", key="src"},
}

function register()
    return "route", _M["get_route_info"]
end

function get_field(data, key, field_type)
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
            local i,j = string.find(item, key .. " ")
            if i ~= nil then
                value = string.sub(item, j+1, string.len(item)-1)
            end
            if value ~= nil then
                ret = value
                break
            end
        end
    end
    return ret
end

function get_route(route)
    local ret = {}
    for i,v in pairs(fields_table) do
        local value = get_field(route, v["key"], v["type"])
        if value ~= nil then
            ret[v["field"]] = value
        end
    end
    return ret
end

function get_route_info()
    local ret = {}
    for table in util.execi("ip rule | awk '{print $NF}' | sort | uniq") do
        if table == "main" or table == "local" or table == "default" or utils.is_integer_and_in_range(table, 0) then
            local cont = {}
            local index = 1
            local data = {}
            for route in util.execi("ip route show table " .. table) do
                data[index] = get_route(route)
                index = index + 1
            end
            cont["item"] = data
            ret[table] = cont
        end
    end
    return ret
end

function split(str,reps)
    local arr = {}
    string.gsub(str,'[^'..reps..']+',function(w)
        table.insert(arr, w)
    end)
    return arr
end
