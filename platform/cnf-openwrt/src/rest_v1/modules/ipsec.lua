-- Copyright 2021 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.modules.ipsec", package.seeall)

util = require "luci.util"
utils = require "luci.controller.rest_v1.utils"
uci = require "luci.model.uci"
json = require "luci.jsonc"

uci_conf = "ipsec"
fields_table = {
        {field="connecting", key="connecting"},
        {field="up", key="up"},
        {field="connection", key="==="}
}

function register()
    return "ipsec", _M["get_ipsec_info"]
end

function get_field(data, key, field_type)
        if type(key) == "function" then
                return key(data)
        end

        local reg = {
                "%d+%s" .. key,
                "%w+%p%w+%p%w+%p%w+%p%w+%s" .. key .. "%s%w+%p%w+%p%w+%p%w+%p%w+"
        }

        local ret = nil

        if (key == "===") then
                index = 2
        else
                index = 1
        end

        for item in string.gmatch(data, reg[index]) do
                local value = nil
                local i,j = string.find(item, " " .. key)
                if i ~= nil then
                        if (key == "===") then
                                value = item
                        else
                                value = string.sub(item, 1, i-1)
                        end
                end
                if value ~= nil
                then
                        ret = value
                        break
                end
        end
        return ret
end

function get_ipsec_detail(stat)
        local ret = nil
        for i,v in pairs(fields_table) do
                local value = get_field(stat, v["key"], v["type"])
                if value ~= nil then
                        if ret == nil then
                            ret = {}
                        end
                        ret[v["field"]] = value
                end
        end
        return ret
end

function getTunnelCounts(configuration)
    local c = 0
    uci:foreach(configuration, "remote",
        function(session)
           print(json.stringify(session))
           local t = session["tunnel"]
           if t ~= nil and #t > 0 then
               c = c + #t
           end
           t = session["transport"]
           if t ~= nil and #t > 0 then
              c = c + #t
           end
        end
    )
    return c
end

function get_ipsec_info()
    local ret = {}
    local index = 1
    local stats = "InitConnection"
    local upi
    local connecti
    local total = getTunnelCounts(uci_conf)
    ret[stats] = "success"
    ret["details"] = {}
    for stat in util.execi("ipsec status") do
        local res = get_ipsec_detail(stat)
        if res ~= nil then
                for k, v in pairs(res) do
                        if (k  == "up") then
                                upi = tonumber(v)
                        elseif (k == "connecting") then
                                connecti = tonumber(v)
                        else
                                break
                        end
                end
                if (upi + connecti < total) then
			util.perror("Finding connection failed ...")
                        ret[stats] = "fail"
                end
                ret["details"][index] = res
                index = index + 1
        end
    end
    return ret
end
