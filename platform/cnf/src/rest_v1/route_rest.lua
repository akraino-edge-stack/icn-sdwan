-- Copyright 2020 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.route_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

function index()
    ver = "v1"
    configuration = "route"
    entry({"sdewan", configuration, ver, "routes"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local method = utils.get_req_method()
    if method == "PUT" then
        return update_route()
    elseif method == "POST" then
        return create_route()
    elseif method == "DELETE" then
        return delete_route()
    elseif method == "GET" then
        return get_route()
    else
        utils.response_error(405, "Method Not Allowed")
    end
end

-- Post
function create_route()
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "No Route Data")
        return
    end
    if is_duplicated(obj.name, obj.dst) then
        utils.response_error(409, "Duplicated Route Configuration")
        return
    end
    if not utils.is_valid_ip(obj.dst) then
        utils.response_error(400, "Invalid Destination IP Address")
        return
    end
    if not utils.is_valid_ip_address(obj.gw) then
        utils.response_error(400, "Invalid gateway IP Address")
        return
    end

    if obj.table == "default" then
        local comm = "ip route add "..obj.dst.." via "..obj.gw.." dev "..obj.dev
        os.execute(comm)
    elseif obj.table == "cnf" then
        local comm = "ip route add table 40 "..obj.dst.." via "..obj.gw.." dev "..obj.dev
        os.execute(comm)
    else
        utils.response_error(400, "Bad route table")
        return
    end
    local file = io.open("/etc/route_cr.info", "a+")
    file:write(obj.name, " ", obj.dst, " ", obj.gw, " ", obj.dev, " ", obj.table, "\n")
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Delete
function delete_route()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local file = io.open("/etc/route_cr.info", "r")
    content = {}
    for line in file:lines() do
        local message = split(line, ' ')
        if name ~= message[1] then
            content[#content+1] = line
        else
            if message[5] == "cnf" then
                local comm = "ip route del table 40 "..obj.dst.." via "..obj.gw.." dev "..obj.dev
                os.execute(comm)
            else
                local comm = "ip route del "..obj.dst.." via "..obj.gw.." dev "..obj.dev
            end
        end
    end
    file:close()
    local file = io.open("/etc/route_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i])
    end
    file:close()
end

-- Update
function update_route()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "Route CR not found")
        return
    end
    if obj.name ~= name then
        utils.response_error(400, "Route CR name mismatch")
        return
    end
    if not utils.is_valid_ip(obj.dst) then
        utils.response_error(400, "Invalid Destination IP Address")
        return
    end
    if not utils.is_valid_ip_address(obj.gw) then
        utils.response_error(400, "Invalid gateway IP Address")
        return
    end

    local file = io.open("/etc/route_cr.info", "r")
    content = {}
    for line in file:lines() do
        local message = split(line, ' ')
        if name ~= message[1] then
            content[#content+1] = line
        else
            if obj.dst ~= message[2] or obj.dst ~= message[5] then
                utils.response_error(400, "Route CR mismatch")
                return
            end
            if obj.table == "default" then
                local comm = "ip route replace "..obj.dst.." via "..obj.gw.." dev "..obj.dev
                os.execute(comm)
            elseif obj.table == "cnf" then
                local comm = "ip route replace table 40 "..obj.dst.." via "..obj.gw.." dev "..obj.dev
                os.execute(comm)
            else
                utils.response_error(400, "Bad route table")
                return
            end
            content[#content+1] = obj.name.." "..obj.dst.." "..obj.gw.." "..obj.dev.." "..obj.table.."\n"
        end
    end
    file:close()
    local file = io.open("/etc/route_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i])
    end
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Get
function get_route()
    local uri_list = utils.get_URI_list()
    local file = io.open("/etc/route_cr.info", "r")
    if #uri_list == 6 then
        local objs = {}
        objs["routes"] = {}
        for line in file:lines() do
            local message = split(line, ' ')
            local obj = {}
            obj["name"] = message[1]
            obj["dst"] = message[2]
            obj["gw"] = message[3]
            obj["dev"] = message[4]
            obj["table"] = message[5]
            table.insert(objs["routes"], obj)
        end
        luci.http.prepare_content("application/json")
        luci.http.write_json(objs)
    elseif #uri_list == 7 then
        local name = uri_list[#uri_list]
        local no = true
        for line in file:lines() do
            local message = split(line, ' ')
            if name == message[1] then
                no = false
                local obj = {}
                obj["name"] = message[1]
                obj["dst"] = message[2]
                obj["gw"] = message[3]
                obj["dev"] = message[4]
                obj["table"] = message[5]
                luci.http.prepare_content("application/json")
                luci.http.write_json(obj)
                break
            end
        end
        if no then
            utils.response_error(404, "Cannot find ".."Route CR ".."[".. name.."]" )
        end
    else
        utils.response_error(400, "Bad request URI")
    end
    file:close()
end

-- Sync and validate
function split(str,reps)
    local arr = {}
    string.gsub(str,'[^'..reps..']+',function(w)
        table.insert(arr, w)
    end)
    return arr
end

function is_duplicated(name, dst)
    local file = io.open("/etc/route_cr.info", "r")
    local judge = false
    for line in file:lines() do
        local message = split(line, ' ')
        if name == message[1] then
            judge = true
            break
        end
        if dst == message[2] then
            judge = true
            break
        end
    end
    file:close()
    return judge
end
