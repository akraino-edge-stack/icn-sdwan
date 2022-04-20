--- SPDX-License-Identifier: Apache-2.0
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.app_rest", package.seeall)

local uci = require "luci.model.uci"
json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
uci_conf = "app-cnf"

application_validator = {
    config_type="rule",
    {name="name"},
    {name="iplist",validator=function(value) return is_valid_ip(value) end,target="src", message="Invalid Source IP Address"},
 }


 application_processor = {
    application={create="create_application", delete="delete_application", validator=application_validator},
    configuration=uci_conf,
}


function index()
    ver = "v1"
    configuration = "application"
    entry({"sdewan", configuration, ver, "applications"}, call("handle_request")).leaf = true
end



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
        return utils[handler](_M,application_processor)
    end
end


function application_command(rule, op)
    local comm_list={}
    local src = rule["iplist"]

    if src == "" or src == nil then
        return comm_list
    end
    local src_ips = split(src, ',')
    if op == "create" then
        for i, ip in ipairs(src_ips) do
            comm_list[i]="ip rule add from "..ip.." lookup 40"
        end
    else
        for i, ip in ipairs(src_ips) do
            comm_list[i]="ip rule del from "..ip.." lookup 40"
        end
    end
    return comm_list
end


function create_application(application)

    local application_name =  application.name
    local res, code, msg = utils.create_uci_section(uci_conf, application_validator, "rule", application)
    if res == false then
        uci:revert(uci_conf)
        return res, code, msg
    end
    local comm_list = application_command(application, "create")
    for _,comm in ipairs(comm_list) do
        os.execute(comm)
        utils.log(comm)
    end
    uci:save(uci_conf)
    uci:commit(uci_conf)
    return true
end

function delete_application(name)

    -- check whether rule is defined
    local application = utils.get_object(_M, application_processor, "application", name)
    if application == nil then
        return false, 404, "application " .. name .. " is not defined"
    end

    -- delete  rule
    local comm_list = application_command(application, "delete")
    for _,comm in ipairs(comm_list) do
        os.execute(comm)
        utils.log(comm)
    end

    utils.delete_uci_section(uci_conf, application_validator, application, "application")
    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true

end

function is_valid_ip(iplist)
    if iplist ~= nil then
        local iplist = utils.split_and_trim(iplist, ',')
        local judge = true
        for _, ip in ipairs(iplist) do
            judge, _ = utils.is_valid_ip_address(ip)
            if not judge then
                return false
            end
        end
        return true
    else
        return false
    end
end


function split(str,reps)
    local arr = {}
    string.gsub(str,'[^'..reps..']+',function(w)
        table.insert(arr, w)
    end)
    return arr
end


