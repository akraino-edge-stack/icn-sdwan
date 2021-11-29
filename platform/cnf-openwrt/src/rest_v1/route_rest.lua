--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.route_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
ifutil = require "luci.controller.rest_v1.ifutil"

uci_conf = "route-cnf"

route_validator = {
    create_section_name=false,
    {name="name"},
    {name="dst", required=true, validator=function(value) return (value == "default") or utils.is_valid_ip(value) end, message="Invalid Destination IP Address"},
    {name="src", validator=function(value) return utils.is_valid_ip_address(value) end, message="Invalid Source IP Address"},
    {name="gw", validator=function(value) return utils.is_valid_ip_address(value) end, message="Invalid Gateway IP Address"},
    {name="dev", required=true, validator=function(value) return (value == "#default") or ifutil.is_interface_available(value) end, message="Invalid interface", code="428"},
    {name="table", validator=function(value) return utils.in_array(value, {"default", "cnf"}) end, message="Bad route table"},
}

route_processor = {
    route={create="create_route", delete="delete_route", validator=route_validator},
    configuration=uci_conf
}

function index()
    ver = "v1"
    configuration = "route"
    entry({"sdewan", configuration, ver, "routes"}, call("handle_request")).leaf = true
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
        return utils[handler](_M, route_processor)
    end
end

-- generate command for route
function route_command(route, op)
    local dst = route["dst"]
    local src = route["src"]
    local gw = route["gw"]
    local dev = route["dev"]
    local t = route["table"]
    if dev == "#default" then
        dev = ifutil.get_default_ifname()
    end

    local comm = "ip route"
    if op == "create" then
        comm = comm .. " add"
    else
        comm = comm .. " del"
    end

    if t == "cnf" then
        comm = comm .. " table 40"
    end
    comm = comm .. " " .. dst
    if gw ~= nil and gw ~= "" then
        comm = comm .. " via " .. gw
    end
    comm = comm .. " dev " .. dev
    if src ~= nil and src ~= "" then
        comm = comm .. " src " .. src
    end

    utils.log(comm)
    return comm
end

-- create a route
function create_route(route)
    local name = route.name
    local res, code, msg = utils.create_uci_section(uci_conf, route_validator, "route", route)

    if res == false then
        uci:revert(uci_conf)
        return res, code, msg
    end

    -- create route rule
    local comm = route_command(route, "create")
    os.execute(comm)

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end

-- delete a route
function delete_route(name)
    -- check whether route is defined
    local route = utils.get_object(_M, route_processor, "route", name)
    if route == nil then
        return false, 404, "route " .. name .. " is not defined"
    end

    -- delete route rule
    local comm = route_command(route, "delete")
    os.execute(comm)

    utils.delete_uci_section(uci_conf, route_validator, route, "route")

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end
