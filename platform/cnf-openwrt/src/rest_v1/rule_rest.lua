--- SPDX-License-Identifier: Apache-2.0
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.rule_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
ifutil = require "luci.controller.rest_v1.ifutil"

uci_conf = "rule-cnf"

rule_validator = {
    create_section_name=false,
    object_validator=function(value) return check_rule(value) end,
    {name="name"},
    {name="src", validator=function(value) return utils.is_valid_ip(value) end, message="Invalid Source IP Address"},
    {name="dst", validator=function(value) return utils.is_valid_ip(value) end, message="Invalid Destination IP Address"},
    {name="prio", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="Invalid Prioroty"},
    {name="table", required=true, validator=function(value) return utils.in_array(value, {"main", "local", "default"}) or utils.is_integer_and_in_range(value, 0) end, message="Invalid Table"},
    {name="fwmark", validator=function(value) return check_fwmark(value) end, message="Invalid fwmark"},
    {name="flag",
        load_func=function(value) if value["flag"] == "true" then return true else return false end end,
        save_func=function(value) if value["flag"] == true then return true, "true" else return true, "false" end end},
}

rule_processor = {
    rule={create="create_rule", delete="delete_rule", validator=rule_validator},
    configuration=uci_conf
}

function index()
    ver = "v1"
    configuration = "rule"
    entry({"sdewan", configuration, ver, "rules"}, call("handle_request")).leaf = true
end

function check_rule(value)
    local src = value["src"]
    local dst = value["dst"]
    if src == "" and dst == "" then
        return false, "src or dst are required for rule"
    end

    return true, value
end

-- Request Handler
function handle_request()
    local conf = io.open("/etc/config/" .. uci_conf, "r")
    if conf == nil then
        conf = io.open("/etc/config/" .. uci_conf, "w")
    end
    conf:close()

    local handler = utils.handles_table[utils.get_req_method()]
    if handler == nil then
        utils.response_error(405, "Method Not Allowed")
    else
        return utils[handler](_M, rule_processor)
    end
end

function check_fwmark(value)
    local num = tonumber(value, 16)
    if not num then
        return false, "not a number"
    elseif string.len(value) > 10 then
        return false, "too large"
    end

    return true, value
end

-- generate command for rule
function rule_command(rule, op)
    local src = rule["src"]
    local dst = rule["dst"]
    local prio = rule["prio"]
    local t = rule["table"]
    local fwmark = rule["fwmark"]
    local flag = rule["flag"]

    local comm = "ip rule"
    if op == "create" then
        comm = comm .. " add"
    else
        comm = comm .. " del"
    end

    if tostring(flag) == "true" then
        comm = comm .. " not"
    end
    if prio ~= nil and prio ~= "" then
        comm = comm .." prio " .. prio
    end
    if src ~= nil and src ~= "" then
        comm = comm .." from " .. src
    end
    if dst ~= nil and dst ~= "" then
        comm = comm .. " to " .. dst
    end

    if t == nil or t == "" then
        t = "main"
    end
    comm = comm .. " lookup " .. t

    if fwmark ~= nil and fwmark ~= "" then
        comm = comm .. " fwmark " .. fwmark
    end

    utils.log(comm)
    return comm
end

-- create a rule
function create_rule(rule)
    local name = rule.name
    local res, code, msg = utils.create_uci_section(uci_conf, rule_validator, "rule", rule)

    if res == false then
        uci:revert(uci_conf)
        return res, code, msg
    end

    -- create rule
    local comm = rule_command(rule, "create")
    os.execute(comm)

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end

-- delete a rule
function delete_rule(name)
    -- check whether rule is defined
    local rule = utils.get_object(_M, rule_processor, "rule", name)
    if rule == nil then
        return false, 404, "rule " .. name .. " is not defined"
    end

    -- delete rule
    local comm = rule_command(rule, "delete")
    os.execute(comm)

    utils.delete_uci_section(uci_conf, rule_validator, rule, "rule")

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end
