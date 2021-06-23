--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.rule_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

function index()
    ver = "v1"
    configuration = "rule"
    entry({"sdewan", configuration, ver, "rules"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local method = utils.get_req_method()
    if method == "PUT" then
        return update_rule()
    elseif method == "POST" then
        return create_rule()
    elseif method == "DELETE" then
        return delete_rule()
    elseif method == "GET" then
        return get_rule()
    else
        utils.response_error(405, "Method Not Allowed")
    end
end

-- Post
function create_rule()
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "No Rule Data")
        return
    end
    if is_duplicated(obj.name, obj.src, obj.dst) then
        utils.response_error(409, "Duplicated Rule Configuration")
        return
    end
    if not is_valid_format(obj.src, obj.dst, obj.prio, obj.table, obj.fwmark) then
        utils.response_error(400, "Invalid rule format")
        return
    end

    local comm = "ip rule add "
    comm = rule_gen(comm, obj.src, obj.dst, obj.prio, obj.table, obj.fwmark, obj.flag)
    os.execute(comm)

    local file = io.open("/etc/rule_cr.info", "a+")
    local rule_str = input_format(obj.name, obj.src, obj.dst, obj.prio, obj.table, obj.fwmark, obj.flag)
    file:write(rule_str, "\n")
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Delete
function delete_rule()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local file = io.open("/etc/rule_cr.info", "r")
    local content = {}
    for line in file:lines() do
        local message = split(line, ',')
        if name ~= message[1] then
            content[#content+1] = line
        else
            local comm = "ip rule del "
            comm = rule_gen(comm, message[2], message[3], message[4], message[5], message[6], message[7])
            os.execute(comm)
        end
    end
    file:close()
    local file = io.open("/etc/rule_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i], "\n")
    end
    file:close()
end

-- Update
function update_rule()
    local uri_list = utils.get_URI_list(7)
    if uri_list == nil then
        return
    end
    local name = uri_list[#uri_list]
    local obj = utils.get_request_body_object()
    if obj == nil then
        utils.response_error(400, "Rule CR not found")
        return
    end
    if obj.name ~= name then
        utils.response_error(400, "Rule CR name mismatch")
        return
    end
    if not is_valid_format(obj.src, obj.dst, obj.prio, obj.table, obj.fwmark) then
        utils.response_error(400, "Invalid rule format")
        return
    end

    local file = io.open("/etc/rule_cr.info", "r")
    local content = {}
    local is_found = false
    for line in file:lines() do
        local message = split(line, ',')
        if name ~= message[1] then
            content[#content+1] = line
        else
            is_found = true
            local pre_comm = "ip rule del "
            pre_comm = rule_gen(pre_comm, message[2], message[3], message[4], message[5], message[6], message[7])
            os.execute(pre_comm)
            local post_comm = "ip rule add "
            post_comm = rule_gen(post_comm, obj.src, obj.dst, obj.prio, obj.table, obj.fwmark, obj.flag)
            os.execute(post_comm)
            content[#content+1] = input_format(obj.name, obj.src, obj.dst, obj.prio, obj.table, obj.fwmark, obj.flag)
        end
    end
    file:close()

    if not is_found then
        utils.response_error(404, "Cannot find ".."Rule ".."[".. name.."]".." to update." )
        return
    end

    local file = io.open("/etc/rule_cr.info", "w+")
    for i = 1, #content do
        file:write(content[i], "\n")
    end
    file:close()
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Get
function get_rule()
    local uri_list = utils.get_URI_list()
    local file = io.open("/etc/rule_cr.info", "r")
    if #uri_list == 6 then
        local objs = {}
        objs["rules"] = {}
        for line in file:lines() do
            local message = split(line, ',')
            local obj = {}
            obj["name"] = message[1]
            obj["src"] = message[2]
            obj["dst"] = message[3]
            obj["prio"] = message[4]
            obj["table"] = message[5]
            obj["fwmark"] = message[6]
            if message[7] == "false" then
                obj["flag"] = false
            else
                obj["flag"] = true
            end
            table.insert(objs["rules"], obj)
        end
        luci.http.prepare_content("application/json")
        luci.http.write_json(objs)
    elseif #uri_list == 7 then
        local name = uri_list[#uri_list]
        local no = true
        for line in file:lines() do
            local message = split(line, ',')
            if name == message[1] then
                no = false
                local obj = {}
                obj["name"] = message[1]
                obj["src"] = message[2]
                obj["dst"] = message[3]
                obj["prio"] = message[4]
                obj["table"] = message[5]
                obj["fwmark"] = message[6]
                if message[7] == "false" then
                    obj["flag"] = false
                else
                    obj["flag"] = true
                end
                luci.http.prepare_content("application/json")
                luci.http.write_json(obj)
                break
            end
        end
        if no then
            utils.response_error(404, "Cannot find ".."Rule CR ".."[".. name.."]" )
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

function is_duplicated(name, src, dst)
    local file = io.open("/etc/rule_cr.info", "r")
    local judge = false
    for line in file:lines() do
        local message = split(line, ',')
        if name == message[1] then
            judge = true
            break
        end
        if src == "" then
            src = "NULL"
        end
        if dst == "" then
            dst = "NULL"
        end
        if src == message[2] and dst == message[3] then
            judge = true
            break
        end
    end
    file:close()
    return judge
end

function is_valid_format(src, dst, prio, table, fwmark)
    local judge = true
    if src == "" and dst == "" then
        judge = false
    elseif src == "" then
        judge = utils.is_valid_ip(dst)
    elseif dst == "" then
        judge = utils.is_valid_ip(src)
    else
        judge = utils.is_valid_ip(dst) and utils.is_valid_ip(src)
    end

    if prio ~= "" then
        judge = judge and utils.is_integer_and_in_range(prio, 0)
    end

    if fwmark ~= "" then
        local num = tonumber(fwmark, 16)
        if not num then
            judge = false
        elseif string.len(fwmark) > 10 then
            judge = false
        end
    end

    if table == "main" or table == "local" or table == "default" or table == "" then
        return judge
    else
        table_id = get_table_id(table)
        judge = judge and utils.is_integer_and_in_range(table_id, 0)
        return judge
    end
end

function rule_gen(comm, src, dst, prio, table, fwmark, flag)
    if tostring(flag) == "true" then
        comm = comm.."not "
    end
    if prio ~= "" and prio ~= "NULL" then
        comm = comm.."prio "..prio.." "
    end
    if src == "" or src == "NULL" then
        comm = comm.."to "..dst.." "
    elseif dst == "" or dst == "NULL" then
        comm = comm.."from "..src.." "
    else
        comm = comm.."from "..src.." to "..dst.." "
    end
    local table_id = get_table_id(table)
    comm = comm.."lookup "..table_id
    if fwmark ~= "" and fwmark ~= "NULL" then
        comm = comm.." fwmark "..fwmark
    end
    return comm
end

function get_table_id(table)
    --TODO
    local table_id = table
    if table == "" then
        table_id = "main"
    end
    return table_id
end

function input_format(name, src, dst, prio, table, fwmark, flag)
    local str = name
    if src == "" then
        str = str..",".."NULL"
    else
        str = str..","..src
    end
    if dst == "" then
        str = str..",".."NULL"
    else
        str = str..","..dst
    end
    if prio == "" then
        str = str..",".."NULL"
    else
        str = str..","..prio
    end
    str = str..","..get_table_id(table)
    if fwmark  == "" then
        str = str..",".."NULL"
    else
        str = str..","..fwmark
    end
    str = str..","..tostring(flag)
    return str
end
