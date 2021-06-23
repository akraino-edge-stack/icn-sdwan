--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.app_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

function index()
    ver = "v1"
    configuration = "application"
    entry({"sdewan", configuration, ver, "applications"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local method = utils.get_req_method()
    if method == "PUT" then
        return update_application()
    elseif method == "POST" then
        return create_application()
    elseif method == "DELETE" then
        return delete_application()
    elseif method == "GET" then
        return get_application()
    else
        utils.response_error(405, "Method Not Allowed")
    end
end

-- Post
function create_application()
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "No Application Data")
        return
    end
    if is_duplicated(obj.name) then
        utils.response_error(409, "Duplicated Application Configuration")
        return
    end

    local file = io.open("/etc/app_cr.info", "a+")
    if obj.iplist ~= nil then
        local iplist = utils.split_and_trim(obj.iplist, ',')
        if not is_valid_ip(iplist) then
            utils.response_error(400, "Invalid IP Address")
            return
        end
        for _, ip in ipairs(iplist) do
            local comm = "ip rule add from "..ip.." lookup 40"
            os.execute(comm)
        end

        iplist = array_to_string(iplist)
        file:write(obj.name, " ", iplist, "\n")
    else
        file:write(obj.name, "\n")
    end
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Delete
function delete_application()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local file = io.open("/etc/app_cr.info", "r")
    content = {}
    for line in file:lines() do
        local message = split(line, ' ')
        if name ~= message[1] then
            content[#content+1] = line
        else
            if #message ~= 1 then
                local iplist = split(message[2], ',')
                for _, ip in ipairs(iplist) do
                    local comm = "ip rule del from "..ip.." lookup 40"
                    os.execute(comm)
                end
            end
        end
    end
    file:close()
    local file = io.open("/etc/app_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i])
    end
    file:close()
end

-- Update
function update_application()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "Application CR not found")
        return
    end
    if obj.name ~= name then
        utils.response_error(400, "Application CR name mismatch")
        return
    end

    local file = io.open("/etc/app_cr.info", "r")
    if obj.iplist ~= nil then
        local input = utils.split_and_trim(obj.iplist, ",")
        if not is_valid_ip(input) then
            utils.response_error(400, "Invalid IP Address")
            return
        end

        content = {}
        for line in file:lines() do
            local message = split(line, ' ')
            if name ~= message[1] then
                content[#content+1] = line
            else
                if #message ~= 1 then
                    local iplist = split(message[2], ',')
                    for _, ip in ipairs(iplist) do
                        local comm = "ip rule del from "..ip.." lookup 40"
                        os.execute(comm)
                    end
                end
                for _, ip in ipairs(input) do
                    local comm = "ip rule add from "..ip.." lookup 40"
                    os.execute(comm)
                end
                local str = array_to_string(input)
                content[#content+1] = obj.name.." "..str.."\n"
            end
        end
    else
        for line in file:lines() do
            local message = split(line, ' ')
            if name ~= message[1] then
                content[#content+1] = line
            else
                content[#content+1] = obj.name.."\n"
            end
        end
    end
    file:close()
    local file = io.open("/etc/app_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i])
    end
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Get
function get_application()
    local uri_list = utils.get_URI_list()
    local file = io.open("/etc/app_cr.info", "r")
    if #uri_list == 6 then
        local objs = {}
        objs["applications"] = {}
        for line in file:lines() do
            local message = split(line, ' ')
            local obj = {}
            obj["name"] = message[1]
            if #message == 1 then
                obj["iplist"] = ""
            else
                obj["iplist"] = message[2]
            end
            table.insert(objs["applications"], obj)
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
                if #message == 1 then
                    obj["iplist"] = ""
                else
                    obj["iplist"] = message[2]
                end
                luci.http.prepare_content("application/json")
                luci.http.write_json(obj)
                break
            end
        end
        if no then
            utils.response_error(404, "Cannot find ".."application CR ".."[".. name.."]" )
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

function is_duplicated(name)
    local file = io.open("/etc/app_cr.info", "r")
    local judge = false
    for line in file:lines() do
        local message = split(line, ' ')
        if name == message[1] then
            judge = true
            break
        end
    end
    file:close()
    return judge
end

function is_valid_ip(iplist)
    local judge = true
    for _, ip in ipairs(iplist) do
        judge, _ = utils.is_valid_ip_address(ip)
        if not judge then
            return false
        end
    end
    return true
end

function array_to_string(arr)
    local str = ""
    for _, s in ipairs(arr) do
        if str == "" then
            str = s
        else
            str = str..","..s
        end
    end
    return str
end
