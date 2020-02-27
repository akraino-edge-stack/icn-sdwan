-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.ipsec_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

uci_conf = "ipsec"

proposal_validator = {
    {name="name"},
    {name="encryption_algorithm", validator=function(value) return true, value end, message="invalid encryption_algorithm"},
    {name="hash_algorithm", validator=function(value) return true, value end, message="invalid hash_algorithm"},
    {name="dh_group", validator=function(value) return true, value end, message="invalid dh_group"},
}

connection_validator = {
    config_type=function(value) return value["type"] end,
    {name="name"},
    {name="type", required=true, validator=function(value) return utils.in_array(value, {"tunnel", "transport"}) end, load_func=function(value) return value[".type"] end, save_func=function(value) return true, "" end, message="invalid type"},
    {name="mode"},
    {name="local_subnet"},
    {name="local_nat"},
    {name="local_sourceip"},
    {name="local_updown"},
    {name="local_firewall"},
    {name="remote_subnet"},
    {name="remote_sourceip"},
    {name="remote_updown"},
    {name="remote_firewall"},
    {name="crypto_proposal", is_list=true, item_validator=function(value) return is_proposal_available(value) end, message="invalid crypto_proposal"},
}

site_validator = {
    config_type="remote",
    {name="name"},
    {name="gateway"},
    {name="pre_shared_key"},
    {name="authentication_method"},
    {name="local_identifier"},
    {name="remote_identifier"},
    {name="crypto_proposal", is_list=true, item_validator=function(value) return is_proposal_available(value) end, message="invalid crypto_proposal"},
    {name="force_crypto_proposal"},
    {name="local_public_cert",
        load_func=function(value) return load_cert(value["local_public_cert"]) end,
        save_func=function(value) return save_cert(value["local_public_cert"], "/tmp/" .. value["name"] .. "_public.cert") end,
        delete_func=function(value) return delete_cert(value["local_public_cert"]) end},
    {name="local_private_cert",
        load_func=function(value) return load_cert(value["local_private_cert"]) end,
        save_func=function(value) return save_cert(value["local_private_cert"], "/tmp/" .. value["name"] .. "_private.cert") end,
        delete_func=function(value) return delete_cert(value["local_private_cert"]) end},
    {name="shared_ca"},
    {name="connections", item_validator=connection_validator, message="invalid connection",
        load_func=function(value) return load_connection(value) end,
        save_func=function(value) return save_connection(value) end,},
}

ipsec_processor = {
    proposal={update="update_proposal", delete="delete_proposal", validator=proposal_validator},
    site={validator=site_validator},
    configuration=uci_conf
}

proposal_checker = {
    {name="remote", checker={"crypto_proposal"}},
    {name="tunnel", checker={"crypto_proposal"}},
    {name="transport", checker={"crypto_proposal"}},
}

function index()
    ver = "v1"
    configuration = "ipsec"
    entry({"sdewan", configuration, ver, "proposals"}, call("get_proposals"))
    entry({"sdewan", configuration, ver, "sites"}, call("get_sites"))
    entry({"sdewan", configuration, ver, "proposal"}, call("handle_request")).leaf = true
    entry({"sdewan", configuration, ver, "site"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local handler = utils.handles_table[utils.get_req_method()]
    if handler == nil then
        utils.response_error(405, "Method Not Allowed")
    else
        return utils[handler](_M, ipsec_processor)
    end
end

function save_cert(content, path)
    local file = io.open(path, "w")
    if file == nil then
        return false, "Can not generate cert at: " .. path
    end

    file:write(content)
    file:close()

    return true, path
end

function load_cert(path)
    if path == nil then
        return nil
    end
    content = path
    local file = io.open(path, "rb")
    if file ~= nil then
        content = file:read "*a"
        file:close()
    end
    return content
end

function delete_cert(path)
    if path ~= nil then
        os.remove(path)
    end
end

function add_to_key_list(arr, key, value)
    local sub_arr = nil
    for i=1, #arr do
        if key == arr[i].option then
            sub_arr = arr[i].values
        end
    end

    if sub_arr == nil then
        sub_arr = {}
        arr[#arr+1] = {option=key, values=sub_arr}
    end
    sub_arr[#sub_arr+1] = value
end

-- load uci configuration as format: {{section="tunnel/transport", name="section_name"},}
function load_connection(value)
    local ret_value={}
    local tunnels = value["tunnel"]
    if tunnels ~= nil and #tunnels > 0 then
        for i=1, #tunnels do
            ret_value[#ret_value+1]={section="tunnel", name=tunnels[i]}
        end
    end
    local transports = value["transport"]
    if transports ~= nil and #transports > 0 then
        for i=1, #transports do
            ret_value[#ret_value+1]={section="transport", name=transports[i]}
        end
    end
    if #ret_value == 0 then
        return nil
    end
    return ret_value
end

-- save connections as standard format:
-- {_standard_format=true, {option="tunnel", values={...}}, {option="transport", values={...}}}
function save_connection(value)
    local connections = value["connections"]
    local ret_value = {_standard_format=true}
    for i=1, #connections do
        add_to_key_list(ret_value, connections[i]["type"], connections[i])
    end

    return true, ret_value
end

-- Site APIs
-- get /sites
function get_sites()
    utils.handle_get_objects("sites", uci_conf, "remote", site_validator)
end

-- Proposal APIs
-- check if proposal is used by connection, site
function is_proposal_used(name)
    local is_used = false

    for i,v in pairs(proposal_checker) do
        local section_name = v["name"]
        local checker = v["checker"]
        uci:foreach(uci_conf, section_name,
            function(section)
                for j=1, #checker do
                    for k=1, #section[checker[j]] do
                        if name == section[checker[j]][k] then
                            is_used = true
                            return false
                        end
                    end
                end
            end
        )

        if is_used then
            break
        end
    end

    return is_used
end

function is_proposal_available(name)
    local proposal = utils.get_object(_M, ipsec_processor, "proposal", name)
    if proposal == nil then
        return false, "Proposal[" .. name .. "] is not definied"
    end

    return true, name
end

-- get /proposals
function get_proposals()
    utils.handle_get_objects("proposals", uci_conf, "proposal", proposal_validator)
end

-- delete a proposal
function delete_proposal(name, check_used)
    -- check whether proposal is defined
    local proposal = utils.get_object(_M, ipsec_processor, "proposal", name)
    if proposal == nil then
        return false, 404, "proposal " .. name .. " is not defined"
    end

    if check_used == nil then
        check_used = true
    else
        check_used = false
    end

    -- Todo: check whether the proposal is used
    if check_used == true and is_proposal_used(name) then
        return false, 400, "proposal " .. name .. " is used"
    end

    -- delete proposal
    uci:foreach(uci_conf, "proposal",
        function (section)
            if name == section[".name"] or name == section["name"] then
                uci:delete(uci_conf, section[".name"])
            end
        end
    )

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end

-- update a proposal
function update_proposal(proposal)
    local name = proposal.name
    res, code, msg = delete_proposal(name, false)
    if res == true then
        return utils.create_object(_M, ipsec_processor, "proposal", proposal)
    end

    return false, code, msg
end
