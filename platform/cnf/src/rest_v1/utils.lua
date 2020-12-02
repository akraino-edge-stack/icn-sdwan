-- Copyright 2020 Intel Corporation, Inc
-- Licensed to the public under the Apache License 2.0.

module("luci.controller.rest_v1.utils", package.seeall)

local json = require "luci.jsonc"
local uci = require "luci.model.uci"
REQUEST_METHOD = "REQUEST_METHOD"

function index()
end

function log(obj)
    return io.stderr:write(tostring(obj) .. "\n")
end

function printTableKeys(t, prefix)
    for key, value in pairs(t) do
        if string.sub(key,1,string.len(prefix)) == prefix then
            log(key)
        end
    end
end

-- check whether s starts with prefix
function start_with(s, prefix)
    return (string.find(s, "^" .. prefix) ~= nil)
end

function end_with(s, postfix)
    return (string.find(s, postfix .. "$") ~= nil)
end

-- check ip
function is_match(s, reg)
    local ret = string.match(s, reg)
    if s == ret then
        return true
    end
    return false
end

function is_valid_ip_address(s)
    local array = {}
    local reg = string.format("([^%s]+)", ".")
    for item in string.gmatch(s, reg) do
        if is_match(item, "(25[0-5])") or is_match(item, "(2[0-4]%d)") or is_match(item, "([01]?%d?%d)") then
            table.insert(array, item)
        else
            return false
        end
    end

    if #array ~= 4 then
        return false
    end

    return true, s
end

function is_valid_ip(s)
    local array = {}
    local reg = string.format("([^%s]+)", "/")
    for item in string.gmatch(s, reg) do
        table.insert(array, item)
    end

    if #array == 1 then
        -- check ip address
        return is_valid_ip_address(array[1])
    else
        if #array == 2 then
            if is_valid_ip_address(array[1]) then
                -- check mask
                if is_integer_and_in_range(array[2], -1, 33) then
                    return true, s
                end
            end
        end
    end

    return false
end

function is_valid_mac(s)
    local array = {}
    local reg = string.format("([^%s]+)", ":")
    for item in string.gmatch(s, reg) do
        if is_match(item, "([a-fA-F0-9][a-fA-F0-9])") then
            table.insert(array, item)
        else
            return false
        end
    end

    if #array == 6 then
        -- check mac
        return true, s
    end

    return false
end

-- trim a string
function trim(s)
    return s:match("^%s*(.-)%s*$")
end

-- split a string based on sep
function split_and_trim(str, sep)
    local array = {}
    local reg = string.format("([^%s]+)", sep)
    for item in string.gmatch(str, reg) do
        item_trimed = trim(item)
        if string.len(item_trimed) > 0 then
            table.insert(array, item_trimed)
        end
    end
    return array
end

-- Check whether value1 is equal to value2
function equal(value1, value2)
    return value1 == value2
end

-- Check whether value is in values array
function in_array(value, values)
    for i=1, #values do
        if (value == values[i]) then
            return true, value
        end
    end
    return false
end

-- Check whether value is an integer and in special range
function is_integer_and_in_range(value, min_value, max_value)
    if(type(value) == "string") then
        local num_value = tonumber(value)
        if not (num_value == nil) then
            local int_value = math.floor(num_value)
            if int_value == num_value then
                if (min_value ~= nil) then
                    if (int_value <= min_value) then
                        return false
                    end
                end
                if (max_value ~= nil) then
                    if (int_value >= max_value) then
                        return false
                    end
                end

                return true
            end
        end
    end

    return false
end

-- Rest API handler function
handles_table = {
    PUT = "handle_put",
    GET = "handle_get",
    POST = "handle_post",
    DELETE = "handle_delete"
}

function get_validator_type(validator)
    config_type = validator["config_type"]
    if config_type ~= nil and type(config_type) == "function" then
        config_type = nil
    end

    return config_type
end

-- set target from src based on data validator definition
function get_uci_section(configuration, validator, object_type, name)
    local object_data = uci:get_all(configuration, name)
    if object_data == nil then
        -- check if name defined as option
        local obj = {}
        local found = false
        uci:foreach(configuration, object_type,
            function (section)
                if name == section["name"] then
                    set_data(configuration, validator, section, obj)
                    found = true
                    return false
                end
            end
        )

        if found == true then
            return obj
        else
            return nil
        end
    else
        if object_data[".type"] ~= object_type then
            return nil
        end
        -- name is defined as section name
        local obj = {}
        obj["name"] = name
        set_data(configuration, validator, object_data, obj)

        return obj
    end
end

function set_data(configuration, data_validator, src, target)
    for i,v in pairs(data_validator) do
        if(type(v) == "table") then
            local name = v["name"]
            local src_name = v["target"]
            local is_list = v["is_list"]
            if is_list == nil then
                is_list = false
            end
            if src_name == nil then
                src_name = name
            end
            local value = src[src_name]
            if v["load_func"] ~= nil and type(v["load_func"]) == "function" then
                value = v["load_func"](src)
            end
            if value ~= nil then
                if is_list and type(value) ~= "table" then
                    value = { value }
                end

                if v["item_validator"] ~= nil and type(v["item_validator"]) == "table" then
                    array_value = {}
                    local index = 1
                    for j=1, #value do
                        local item_obj = nil
                        if value[j].section ~= nil then
                            item_obj = get_uci_section(configuration, v["item_validator"], value[j].section, value[j].name)
                        else
                            item_obj = get_uci_section(configuration, v["item_validator"], get_validator_type(v["item_validator"]), value[j])
                        end
                        if item_obj ~= nil then
                            array_value[index] = item_obj
                            index = index + 1
                        end
                    end

                    target[name] = array_value
                else
                    if v["validator"] ~= nil and type(v["validator"]) == "table" then
                        if value.section ~= nil then
                            target[name] = get_uci_section(configuration, v["validator"], value.section, value.name)
                        else
                            target[name] = get_uci_section(configuration, v["validator"], get_validator_type(v["validator"]), value)
                        end
                    else
                        target[name] = value
                    end
                end
            end
        end
    end
end

-- get
function get_object_type(value)
    if end_with(value, "s") then
        return string.sub(value, 1, string.len(value)-1)
    end
    return value
end

function get_object(module_table, processors, object_type, name)
    if processors ~= nil and processors[object_type] ~= nil
       and processors[object_type]["get"] ~= nil
       and module_table[processors[object_type]["get"]] ~= nil then
            return module_table[processors[object_type]["get"]](name)
    else
        local object_conf_type = get_validator_type(processors[object_type].validator)
        if object_conf_type == nil then
            object_conf_type = object_type
        end
        return get_uci_section(processors["configuration"], processors[object_type].validator, object_conf_type, name)
    end
end

function get_objects(module_table, processors, object_types, object_type)
    if processors ~= nil and processors[object_type] ~= nil
       and processors[object_type]["gets"] ~= nil
       and module_table[processors[object_type]["gets"]] ~= nil then
            return module_table[processors[object_type]["gets"]]()
    else
        local res = {}
        res[object_types] = {}

        local index = 1
        local object_conf_type = get_validator_type(processors[object_type].validator)
        if object_conf_type == nil then
            object_conf_type = object_type
        end
        uci:foreach(processors["configuration"], object_conf_type,
            function (section)
                local obj = {}
                obj["name"] = section[".name"]
                set_data(processors["configuration"], processors[object_type].validator, section, obj)
                res[object_types][index] = obj
                index = index + 1
            end
        )

        return res
    end
end

function handle_get(module_table, processors)
    local uri_list = get_URI_list()
    if uri_list == nil then
        return
    end

    if #uri_list == 6 then
        -- get objects
        local object_types = uri_list[#uri_list]
        local object_type = get_object_type(object_types)
        if processors["get_type"] ~= nil and type(processors["get_type"]) == "function" then
            object_type = processors["get_type"](object_types)
        end
        local objs = get_objects(module_table, processors, object_types, object_type)
        if objs == nil then
            response_error(404, "Cannot find " .. object_type)
            return
        end
        response_object(objs)
    elseif #uri_list == 7 then
        -- get object
        local object_types = uri_list[#uri_list-1]
        local object_type = get_object_type(object_types)
        if processors["get_type"] ~= nil and type(processors["get_type"]) == "function" then
            object_type = processors["get_type"](object_types)
        end
        local name = uri_list[#uri_list]
        local obj = get_object(module_table, processors, object_type, name)

        if obj == nil then
            response_error(404, "Cannot find " .. object_type .. "[" .. name .. "]" )
            return
        end
        response_object(obj)
    else
        response_error(400, "Bad request URI")
        return
    end
end

-- put
function update_object(module_table, processors, object_type, obj)
    if processors ~= nil and processors[object_type] ~= nil 
       and processors[object_type]["update"] ~= nil 
       and module_table[processors[object_type]["update"]] ~= nil then
            return module_table[processors[object_type]["update"]](obj)
    else
        local name = obj.name
        res, code, msg = delete_object(module_table, processors, object_type, name)
        if res == true or code == 404 then
            return create_object(module_table, processors, object_type, obj)
        end

        return false, code, msg
    end
end

function handle_put(module_table, processors)
    local uri_list = get_URI_list(7)
    if uri_list == nil then
        return
    end

    local object_types = uri_list[#uri_list-1]
    local object_type = get_object_type(object_types)
    if processors["get_type"] ~= nil and type(processors["get_type"]) == "function" then
        object_type = processors["get_type"](object_types)
    end

    local name = uri_list[#uri_list]
    -- check content and get object
    local obj = get_and_validate_body_object(processors[object_type].validator)
    if obj == nil then
        return
    end
    obj.name = name

    res, code, msg = update_object(module_table, processors, object_type, obj)
    if res == false then
        response_error(code, msg)
    else
        response_object(get_object(module_table, processors, object_type, name))
    end
end

-- post
function create_uci_section(configuration, validator, object_type, obj)
    -- check whether object is exist
    local obj_name = obj.name
    local exist_obj = uci:get_all(configuration, obj_name)
    if exist_obj ~= nil then
        return false, 409, "Conflict"
    end

    -- create object
    local create_section_name = validator["create_section_name"]
    if create_section_name == nil then
        create_section_name = true
    end

    local obj_section = nil
    local obj_section_type = object_type
    local obj_section_config_type = validator["config_type"]
    if obj_section_config_type ~= nil then
        if type(obj_section_config_type) == "function" then
            obj_section_type = obj_section_config_type(obj)
        else
            obj_section_type = obj_section_config_type
        end
    end
    if create_section_name == true then
        obj_section = uci:section(configuration, obj_section_type, obj_name)
    else
        obj_section = uci:section(configuration, obj_section_type)
    end

    for i,v in pairs(validator) do
        if(type(v) == "table") then
            local name = v["name"]
            if create_section_name == false or name ~= "name" then
                if(obj[name] ~= nil) then
                    local target_name = v["target"]
                    local use_target_name = false
                    if(target_name == nil) then
                        target_name = name
                    else
                        use_target_name = true
                    end
                    local obj_value = obj[name]
                    if v["save_func"] ~= nil and type(v["save_func"]) == "function" then
                        res, obj_value = v["save_func"](obj)
                        if res == false then
                            return res, obj_value
                        end
                    end
                    if type(obj_value) == "table" then
                        -- Support creating multiple options/sections:
                        -- change obj_value to standard format: {{option="option_name", values={...}},}
                        if obj_value._standard_format == nil then
                            obj_value = {{option=target_name, values=obj_value},}
                        end
                        for j=1, #obj_value do
                            local option_name = obj_value[j].option
                            local section_obj = obj_value[j].values
                            if v["item_validator"] ~= nil and type(v["item_validator"]) == "table" then
                                local sub_obj_names = {}
                                for k=1, #section_obj do
                                    sub_obj_names[k] = section_obj[k].name
                                    local res, msg = create_uci_section(configuration, v["item_validator"], option_name, section_obj[k])
                                    if res == false then
                                        return res, msg
                                    end
                                end
                                uci:set_list(configuration, obj_section, option_name, sub_obj_names)
                            else
                                uci:set_list(configuration, obj_section, option_name, section_obj)
                            end
                        end
                    else
                        if v["validator"] ~= nil and type(v["validator"]) == "table" then
                            local res, msg = create_uci_section(configuration, v["validator"], target_name, obj_value)
                            if res == false then
                                return res, msg
                            end
                            uci:set(configuration, obj_section, target_name, obj_value.name)
                        else
                            if obj_value ~= "" then
                                uci:set(configuration, obj_section, target_name, obj_value)
                            end
                        end
                    end
                end
            end
        end
    end

    return true
end

function create_object(module_table, processors, object_type, obj)
    if processors ~= nil and processors[object_type] ~= nil
       and processors[object_type]["create"] ~= nil
       and module_table[processors[object_type]["create"]] ~= nil then
            return module_table[processors[object_type]["create"]](obj)
    else
        local res, msg = create_uci_section(processors["configuration"], processors[object_type].validator, object_type, obj)

        if res == false then
            uci:revert(processors["configuration"])
            return res, msg
        end

        -- commit change
        uci:save(processors["configuration"])
        uci:commit(processors["configuration"])

        return true
    end
end

function handle_post(module_table, processors)
    local uri_list = get_URI_list(6)
    if uri_list == nil then
        return
    end

    local object_types = uri_list[#uri_list]
    local object_type = get_object_type(object_types)
    if processors["get_type"] ~= nil and type(processors["get_type"]) == "function" then
        object_type = processors["get_type"](object_types)
    end

    -- check content and get policy object
    local obj = get_and_validate_body_object(processors[object_type].validator)
    if obj == nil then
        return
    end

    -- check name is not empty
    local name = obj.name
    if name ~= nil and type(name) == "string" then
        name = trim(name)
    else
        name = nil
    end
    if name == nil or name == "" then
        response_error(400, "Field[name] checked failed: required")
        return
    end

    res, code, msg = create_object(module_table, processors, object_type, obj)
    if res == false then
        response_error(code, msg)
    else
        response_object(get_object(module_table, processors, object_type, name), 201)
    end
end

-- delete
function delete_uci_section(configuration, validator, obj, object_type)
    local name = obj["name"]
    local obj_section_config_type = validator["config_type"]
    if obj_section_config_type ~= nil then
        if type(obj_section_config_type) == "function" then
            object_type = obj_section_config_type(obj)
        else
            object_type = obj_section_config_type
        end
    end

    -- delete depedency uci section
    for i,v in pairs(validator) do
        if(type(v) == "table") then
            if obj[v["name"]] ~= nil then
                if v["item_validator"] ~= nil and type(v["item_validator"]) == "table" then
                    for j=1, #obj[v["name"]] do
                        local sub_obj = obj[v["name"]][j]
                        delete_uci_section(configuration, v["item_validator"], obj[v["name"]][j], v["name"])
                    end
                else
                    if v["validator"] ~= nil and type(v["validator"]) == "table" then
                        delete_uci_section(configuration, v["validator"], obj[v["name"]], v["name"])
                    end
                end
            end
        end
    end

    -- delete uci section
    uci:foreach(configuration, object_type,
        function (section)
            if name == section[".name"] or name == section["name"] then
                for i,v in pairs(validator) do
                    if(type(v) == "table") then
                        if v["delete_func"] ~= nil and type(v["delete_func"]) == "function" then
                            v["delete_func"](section)
                        end
                    end
                end
                uci:delete(configuration, section[".name"])
            end
        end
    )
end

function delete_object(module_table, processors, object_type, name)
    if processors ~= nil and processors[object_type] ~= nil
       and processors[object_type]["delete"] ~= nil
       and module_table[processors[object_type]["delete"]] ~= nil then
            return module_table[processors[object_type]["delete"]](name)
    else
        local obj = get_object(module_table, processors, object_type, name)
        if obj == nil then
            return false, 404, object_type .. " " .. name .. " is not defined"
        end

        delete_uci_section(processors["configuration"], processors[object_type].validator, obj, object_type)

        -- commit change
        uci:save(processors["configuration"])
        uci:commit(processors["configuration"])

        return true
    end
end

function handle_delete(module_table, processors)
    local uri_list = get_URI_list(7)
    if uri_list == nil then
        return
    end

    local object_types = uri_list[#uri_list-1]
    local object_type = get_object_type(object_types)
    if processors["get_type"] ~= nil and type(processors["get_type"]) == "function" then
        object_type = processors["get_type"](object_types)
    end
    local name = uri_list[#uri_list]

    res, code, msg = delete_object(module_table, processors, object_type, name)
    if res == false then
        response_error(code, msg)
    else
        response_success()
    end
end

-- Validate src with validator then set target (array)
function validate_and_set_data_array(validator, src)
    if #src == 0 then
        return false, "Not an array"
    end

    local ret_obj = {}
    for i=1, #src do
        local res = false
        local ret_obj_item = nil
        if type(validator) == "function" then
            res, ret_obj_item = validator(src[i])
        else
            res, ret_obj_item = validate_and_set_data(validator, src[i])
        end

        if not res then
            return false, ret_obj_item
        end
        ret_obj[i] = ret_obj_item
    end

    return true, ret_obj
end

-- Validate src with validator then set target (set data field with default value if applied)
-- return:
--   res: validate result
--   ret_obj: copy of src based on validator definition(res=true) or error message(res=false)
function validate_and_set_data(validator, src)
    local target = {}
    for i,v in pairs(validator) do
        if(type(v) == "table") then
            local name = v["name"]
            local val_func = v["validator"]
            local item_val_func = v["item_validator"]
            local error_message = v["message"]
            local required = v["required"]
            local default = v["default"]
            local target_name = name

            if required == nil then
                required = false
            end

            if error_message == nil then
                error_message = ""
            end

            local value = src[name]
            if value ~= nil and type(value) == "string" then
                value = trim(value)
            end

            if value ~= nil and value ~= "" then
                local res = true
                local ret_obj = nil
                if val_func ~= nil then
                    if type(val_func) == "function" then
                        res, ret_obj = val_func(value)
                    else
                        res, ret_obj = validate_and_set_data(val_func, value)
                    end
                else
                    if item_val_func ~= nil then
                        res, ret_obj = validate_and_set_data_array(item_val_func, value)
                    end
                end

                if res == false then
                    if ret_obj ~= nil and ret_obj ~= "" then
                        return false, "Field[" .. name .. "] checked failed: " .. error_message .. " [" .. ret_obj .. "]"
                    else
                        return false, "Field[" .. name .. "] checked failed: " .. error_message
                    end
                else
                    if ret_obj ~= nil then
                        value = ret_obj
                    end
                end

                target[target_name] = value
            else
                if required then
                    return false, "Field[" .. name .. "] checked failed: required"
                else
                    -- set default value
                    if default ~= nil then
                        target[target_name] = default
                    end
                end
            end
        end
    end

    -- check with object_validator
    local object_validator = validator["object_validator"]
    if object_validator ~= nil then
        return object_validator(target)
    end

    return true, target
end

-- parse URI
function get_URI_list(required_lenth)
    local uri = luci.http.getenv("REQUEST_URI")
    local uri_list = split_and_trim(uri, "/")
    if (required_lenth ~= nil) and (not (#uri_list == required_lenth)) then
        response_error(400, "Bad request URI")
        return nil
    end

    return uri_list
end

-- get request body
function get_request_body_object()
    local content, size = luci.http.content()
    local content_obj, err = json.parse(content)
    if not (err == nil) then
        response_error(400, "Body parse error: " .. err)
    end

    return content_obj
end

-- get and validate body object
function get_and_validate_body_object(validator)
    local body_obj = get_request_body_object()
    if body_obj == nil then
        return nil
    end

    local res, res_obj = validate_and_set_data(validator, body_obj)
    if not res then
        response_error(400, res_obj)
        return nil
    end

    return res_obj
end

-- get request method
function get_req_method()
    return luci.http.getenv("REQUEST_METHOD")
end

-- check request method
function validate_req_method(method)
    local res = equal(get_req_method(), method)
    if not res then
        response_error(405, "Method Not Allowed")
    end

    return res
end

-- response with error
function response_error(code, message)
    if message == nil then
        message = ""
    end
    luci.http.status(code, message)
    luci.http.prepare_content("text/plain")
    luci.http.write('{"message":"' .. message .. '"}')
end

-- response with json object
function response_object(obj, code)
    if code ~= nil then
        luci.http.status(code, "")
    end
    luci.http.prepare_content("application/json")
    luci.http.write_json(obj)
end

-- response success message
function response_success(code)
    local message = '{"message":"success"}'
    if code ~= nil then
        luci.http.status(code, message)
    end
	response_message(message)
end

-- response with message
function response_message(message)
    luci.http.prepare_content("text/plain")
    luci.http.write(message)
end
