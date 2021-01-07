-- Copyright 2020 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.status_rest", package.seeall)

utils = require "luci.controller.rest_v1.utils"
local fs = require "nixio.fs"
local deb = require "debug"

local module_list = {}
local module_list_updated_time = 0
__file__ = deb.getinfo(1, 'S').source:sub(2)

function index()
    ver = "v1"
    entry({"sdewan", ver, "status"}, call("handle_request")).leaf = true
end

-- base path
function basepath()
    return fs.dirname(__file__)
end

-- create module list
function create_module_list(cur)
    module_list_updated_time = cur

    local base = basepath() .. "/modules/"
    for mfile in (fs.glob(base .. "*.lua")) do
        local mname = "luci.controller.rest_v1.modules." .. mfile:sub(#base+1, #mfile-4):gsub("/", ".")
        local s, mod = pcall(require, mname)
        if s then
            local reg_func = mod.register
            if type(reg_func) == "function" then
                local name, qfunc = reg_func()
                module_list[#module_list + 1] = {name = name, func = qfunc}
            else
                utils.log("Function register is not found in " .. mname)
            end
        else
            utils.log("Error to load module " .. mname)
        end
    end
end

-- refresh module list
function refresh_module_list()
    local cur = os.time()
    if #module_list == 0 or (cur - module_list_updated_time) >= 10 then
        module_list = {}
        create_module_list(cur)
    end
end

function get_status()
    utils.log("Query all module status")
    refresh_module_list()

    local objs = {}
    for i=1, #module_list do
        local obj = {}
        obj["name"] = module_list[i].name
        obj["status"] = module_list[i].func()
        objs[#objs + 1] = obj
    end

    return objs
end

function get_module_status(module_name)
	utils.log("Query module status: %s" % module_name)
    refresh_module_list()

    local obj = {}
    for i=1, #module_list do
        if module_name == module_list[i].name then
            obj = module_list[i].func()
            return obj
        end
    end

    return nil
end

-- Request Handler
function handle_request()
    if not (utils.validate_req_method("GET")) then
        return
    end

    local uri_list = utils.get_URI_list()
    if uri_list == nil then
        return
    end

    if #uri_list == 5 then
        -- get all status
        local objs = get_status()
        utils.response_object(objs)
    else
        if #uri_list == 6 then
            -- get module status
            local module_name = uri_list[#uri_list]
            local obj = get_module_status(module_name)
            if obj == nil then
                utils.response_error(400, "Bad request URI: " .. module_name .. " is not supported")
            else
                utils.response_object(obj)
            end
        else
            utils.response_error(400, "Bad request URI")
        end
    end
end
