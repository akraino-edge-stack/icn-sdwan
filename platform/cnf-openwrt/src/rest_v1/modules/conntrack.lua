--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.modules.conntrack", package.seeall)

NX = require("nixio")
sys = require "luci.sys"
util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"

tcp_table = {
    {field="protocol", key=function(data) return split(data, ' ')[1] end},
    {field="request", key=function(data) return get_info(data, 5) end},
    {field="response", key=function(data) return get_info(data, 9) end},
    {field="mark", key="mark"},
    {field="state", key=function(data) return split(data, ' ')[4] end},
}

udp_table = {
    {field="protocol", key=function(data) return split(data, ' ')[1] end},
    {field="request", key=function(data) return get_info(data, 4) end},
    {field="response", key=function(data) return get_info(data, 8) end},
    {field="mark", key="mark"},
}

function register()
    return "conntrack", _M["get_conn_info"]
end

function get_info(data, index)
    local ret = {}
    local message = split(data, ' ')
    if string.find(message[index], 'src') == nil then
        index = index + 1
    end
    local src = split(message[index], '=')[2]
    local dst = split(message[index+1], '=')[2]
    local sport = split(message[index+2], '=')[2]
    local dport = split(message[index+3], '=')[2]
    if src == nil then
        util.perror("Invalid request or response source")
        return nil
    end
    ret["src"] = src..':'..sport
    ret["dst"] = dst..':'..dport
    return ret
end

function get_field(data, key)
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

function get_conn(conn)
    local ret = {}
    local protocol = split(conn, ' ')[1]
    if protocol  == "tcp" then
        fields_table = tcp_table
    elseif protocol == "udp" then
        fields_table = udp_table
    else
        return ret
    end
    for i,v in pairs(fields_table) do
        local value = get_field(conn, v["key"])
        if value ~= nil then
            ret[v["field"]] = value
        end
    end
    return ret
end

function get_conn_info()
    local ret = {}
    local index = 1
    for conn in util.execi("conntrack -L") do
        ret[index] = get_conn(conn)
        index = index + 1
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
