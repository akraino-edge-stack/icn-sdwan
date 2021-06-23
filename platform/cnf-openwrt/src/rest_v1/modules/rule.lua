--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.modules.rule", package.seeall)

NX = require("nixio")
sys = require "luci.sys"
util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"

fields_table = {
    {field="src", key="from"},
    {field="dst", key="to"},
    {field="prio", key=function(data) return split(data, ':')[1] end},
    {field="fwmark", key="fwmark"},
    {field="table", key="lookup"},
    {field="not", key=function(data) if string.match(data, "[%s]not[%s]") ~= nil then return "true" else return "false" end end},
}

function register()
    return "rule", _M["get_rule_info"]
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

function get_rule(rule)
    local ret = {}
    for i,v in pairs(fields_table) do
        local value = get_field(rule, v["key"], v["type"])
        if value ~= nil then
            ret[v["field"]] = value
        end
    end
    return ret
end

function get_rule_info()
    local ret = {}
    local index = 1
    for rule in util.execi("ip rule") do
        ret[index] = get_rule(rule)
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
