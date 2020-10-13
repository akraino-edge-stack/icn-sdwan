-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.svc_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

uci_conf = "service"

function index()
    ver = "v1"
    configuration = "service"
    entry({"sdewan", configuration, ver, "services"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local method = utils.get_req_method()
    if method == "GET" then
        return get_service()
    elseif method == "POST" then
        return create_service()
    elseif method == "DELETE" then
        return delete_service()
    else
        utils.response_error(405, "Method Not Allowed")
    end
end

-- Post
function create_service()
    local obj = utils.get_request_body_object()
    if obj == nil then
        return
    end
    if is_invalid(obj.port, obj.dport) then
        utils.response_error(416, "Invalid Port Range")
    elseif is_duplicated(obj.name, obj.port) then
        utils.response_error(409, "Duplicated Service Configuration")
    else
        local count = 0
        local ns = "nslookup "..obj.fullname.." | tail -n2 | awk -F':' '{print $2}' | head -1"
        local ip
        local exec
        while count < 6 do
            exec = io.popen(ns)
            ip = string.gsub(exec:read("*a"), "^%s*(.-)%s*$", "%1")
            if ip ~= "NXDOMAIN"
             then
                break
            end
            os.execute("sleep " .. tonumber(5))
            count = count + 1
        end
        if ip ~= "NXDOMAIN" then
            exec:close()
            local file = io.open("/etc/sdewan_svc.info", "a+")
            file:write(obj.name, " ", obj.fullname, " ", obj.port, " ", obj.dport, " ", ip, " ", "0\n")
            local comm = "iptables -t nat -I PREROUTING 2 -p tcp --dport "..obj.port.." -j DNAT --to-destination "..ip..":"..obj.dport
            os.execute(comm)
            file:close()
            sync_info()
        else
            utils.response_error(408, "Timeout: waiting for service ready...")
        end
    end
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- Delete
function delete_service()
    local uri_list = utils.get_URI_list(7)
    local name = uri_list[#uri_list]
    local info_file = io.open("/etc/sdewan_svc.info", "w")
    local up_file = io.open("/etc/sdewan_svc.up", "r")
    for line in up_file:lines() do
        local message = split(line)
        if name ~= message[1] then
            info_file:write(line, "\n")
        else
            local comm = "iptables -t nat -D PREROUTING -p tcp --dport "..message[3].." -j DNAT --to-destination "..message[5]..":"..message[4]
            os.execute(comm)
        end
    end
    info_file:close()
    up_file:close()
    sync_info()
end

-- Get
function get_service()
    local uri_list = utils.get_URI_list()
    local file = io.open("/etc/sdewan_svc.info", "r")
    if #uri_list == 6 then
        local objs = {}
        objs["services"] = {}
        for line in file:lines() do
            local message = split(line)
            local obj = {}
            obj["name"] = message[1]
            obj["fullname"] = message[2]
            obj["port"] = message[3]
            obj["dport"] = message[4]
            table.insert(objs["services"], obj)
        end
        luci.http.prepare_content("application/json")
        luci.http.write_json(objs)
    elseif #uri_list == 7 then
        local name = uri_list[#uri_list]
        local no = true
        for line in file:lines() do
            local message = split(line)
            if name == message[1] then
                no = false
                local obj = {}
                obj["name"] = message[1]
                obj["fullname"] = message[2]
                obj["port"] = message[3]
                obj["dport"] = message[4]
                luci.http.prepare_content("application/json")
                luci.http.write_json(obj)
                break
            end
        end
        if no then
            utils.response_error(404, "Cannot find " .. "service" .. "[" .. name .. "]" )
        end
    else
        utils.response_error(400, "Bad request URI")
    end
    file:close()
end

-- Sync and validate
function sync_info()
    local in_file = io.open("/etc/sdewan_svc.info", "r")
    local out_file = io.open("/etc/sdewan_svc.up", "w")
    local content = in_file:read("*a")
    out_file:write(content)
    in_file:close()
    out_file:close()
end

function split(str)
    local arr = {}
    for w in string.gmatch(str, "%S+") do
        table.insert(arr, w)
    end
    return arr
end

function is_duplicated(name, port)
    local file = io.open("/etc/sdewan_svc.info", "r")
    local judge = false
    for line in file:lines() do
        local message = split(line)
        if name == message[1] then
            judge = true
            break
        end
        if port == message[2] then
            judge = true
            break
        end
    end
    file:close()
    return judge
end

function is_invalid(port, dport)
    local judge = false
    if not utils.is_integer_and_in_range(port, 0, 65535) then
        judge = true
    end
    if not utils.is_integer_and_in_range(dport, 0, 65535) then
        judge = true
    end
    return judge
end

