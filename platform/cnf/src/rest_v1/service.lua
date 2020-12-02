-- Copyright 2020 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.service", package.seeall)

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

available_services = {"mwan3", "firewall", "ipsec"}
action_tables = {
    mwan3 = {command="/etc/init.d/mwan3", reload="restart"},
    firewall = {command="/etc/init.d/firewall", reload_command="fw3"},
    ipsec = {command="/etc/init.d/ipsec", reload="restart"}
}

executeservice_validator = {
    {name="action", required=true, validator=function(value) return utils.in_array(value, {"start", "stop", "restart", "reload"}) end, message="wrong action"}
}

function index()
    ver = "v1"
    entry({"sdewan", ver, "services"}, call("getServices"))
    entry({"sdewan", ver, "service"}, call("executeService")).leaf = true
end

function getServices()
    if not (utils.validate_req_method("GET")) then
        return
    end

    local res = {}
    res["services"] = available_services

    utils.response_object(res)
end

function getservicename()
    local uri_list = utils.get_URI_list(6)
    if uri_list == nil then
        return nil
    end

    local service = uri_list[#uri_list]
    if not utils.in_array(service, available_services) then
        utils.response_error(400, "Bad request URI")
        return nil
    end

    return service
end

function executeService()
    -- check request method
    if not (utils.validate_req_method("PUT")) then
        return
    end

    -- get service name
    local service = getservicename()
    if service == nil then
        return
    end

    -- check content
    local body_obj = utils.get_and_validate_body_object(executeservice_validator)
    if body_obj == nil then
        return
    end

    local action = body_obj["action"]

    local exec_command = action_tables[service][action .. "_command"]
    if exec_command == nil then
        exec_command = action_tables[service]["command"]
    end
    local exec_action = action_tables[service][action]
    if exec_action == nil then
        exec_action  = action
    end

    exec_command = exec_command .. " " .. exec_action

    -- execute command
    utils.log("Execute Command: %s" % exec_command)
    sys.exec(exec_command)

    utils.response_success()
end
