-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.ipsec_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
mime = require "mime"

uci_conf = "ipsec"

encrypto_alg={"3des", "cast128", "blowfish128", "blowfish", "blowfish192", "blowfish256", "null", "aes", "aes128", "aes192", "aes256", "aes128ctr", "aes192ctr", "aes256ctr", "aes128ccm8", "aes192ccm8", "aes256ccm8", "aes128ccm64", "aes192ccm64", "aes256ccm64", "aes128ccm12", "aes192ccm12", "aes256ccm12", "aes128ccm96", "aes192ccm96", "aes256ccm96", "aes128ccm16", "aes192ccm16", "aes256ccm16", "aes128ccm128", "aes192ccm128", "aes256ccm128", "aes128gcm8", "aes192gcm8", "aes256gcm8", "aes128gcm64", "aes192gcm64", "aes256gcm64", "aes128gcm12", "aes192gcm12", "aes256gcm12", "aes128gcm96", "aes192gcm96", "aes256gcm96", "aes128gcm16", "aes192gcm16", "aes256gcm16", "aes128gcm128", "aes192gcm128", "aes256gcm128", "camellia128", "camellia192", "camellia256", "camellia", "camellia128ctr", "camellia192ctr", "camellia256ctr", "camellia128ccm8", "camellia192ccm8", "camellia256ccm8", "camellia128ccm64", "camellia192ccm64", "camellia256ccm64", "camellia128ccm12", "camellia192ccm12", "camellia256ccm12", "camellia128ccm96", "camellia192ccm96", "camellia256ccm96", "camellia128ccm16", "camellia192ccm16", "camellia256ccm16", "camellia128ccm128", "camellia192ccm128", "camellia256ccm128", "chacha20poly1305"}

hash_alg={"md5", "sha", "sha1", "aesxcbc", "sha256", "sha2_256", "sha384", "sha2_384", "sha512", "sha2_512", "sha256_96", "sha2_256_96"}

dh_group={"modp768", "modp1024", "modp1536", "modp2048", "modp3072", "modp4096", "modp6144", "modp8192", "modp1024s160", "modp2048s224", "modp2048s256", "ecp192", "ecp224", "ecp256", "ecp384", "ecp521", "ecp224bp", "ecp256bp", "ecp384bp", "ecp512bp", "curve25519", "x25519", "curve448", "x448"}

proposal_validator = {
    {name="name"},
    {name="encryption_algorithm", validator=function(value) return utils.in_array(value, encrypto_alg) end, message="invalid encryption_algorithm"},
    {name="hash_algorithm", validator=function(value) return utils.in_array(value, hash_alg) end, message="invalid hash_algorithm"},
    {name="dh_group", validator=function(value) return utils.in_array(value, dh_group) end, message="invalid dh_group"},
}

connection_validator = {
config_type=function(value) return value["conn_type"] end,
    {name="name", required=true},
    {name="conn_type", required=true, validator=function(value) return utils.in_array(value, {"tunnel", "transport"}) end, load_func=function(value) return value[".type"] end, save_func=function(value) return true, "" end, message="invalid type"},
    {name="mode", required=true, validator=function(value) return utils.in_array(value, {"start", "add", "route"}) end, message="invalid connection mode"},
    {name="local_subnet"},
    {name="local_nat"},
    {name="local_sourceip"},
    {name="local_updown"},
    {name="local_firewall", validator=function(value) return utils.in_array(value, {"yes", "no"}) end},
    {name="remote_subnet"},
    {name="remote_sourceip"},
    {name="remote_updown"},
    {name="remote_firewall", validator=function(value) return utils.in_array(value, {"yes", "no"}) end},
    {name="crypto_proposal", is_list=true, item_validator=function(value) return is_proposal_available(value) end, message="invalid crypto_proposal"},
    {name="mark"},

}

remote_validator = {
    config_type="remote",
    object_validator=function(value) return check_auth_method(value) end,
    {name="name"},
    {name="type"},
    {name="gateway", required=true},
    {name="enabled", default="1"},
    {name="authentication_method", required=true, validator=function(value) return utils.in_array(value, {"psk", "pubkey"}) end},
    {name="pre_shared_key"},
    {name="local_identifier"},
    {name="remote_identifier"},
    {name="crypto_proposal", is_list=true, item_validator=function(value) return is_proposal_available(value) end, message="invalid crypto_proposal"},
    {name="force_crypto_proposal", validator=function(value) return utils.in_array(value, {"0", "1"}) end, message="invalid input for ForceCryptoProposal"},
    {name="local_public_cert",
        load_func=function(value) return load_cert(value["local_public_cert"]) end,
        save_func=function(value) return save_cert(value["local_public_cert"], "/etc/ipsec.d/certs/" .. value["name"] .. "_public.pem") end,
        delete_func=function(value) return delete_cert(value["local_public_cert"]) end},
    {name="local_private_cert",
        load_func=function(value) return load_cert(value["local_private_cert"]) end,
        save_func=function(value) return save_cert(value["local_private_cert"], "/etc/ipsec.d/private/" .. value["name"] .. "_private.pem") end,
        delete_func=function(value) return delete_cert(value["local_private_cert"]) end},
    {name="shared_ca",
        load_func=function(value) return load_cert(value["shared_ca"]) end,
        save_func=function(value) return save_cert(value["shared_ca"], "/etc/ipsec.d/cacerts/" .. value["name"] .. "_ca.pem") end,
        delete_func=function(value) return delete_cert(value["shared_ca"]) end},
    {name="connections", item_validator=connection_validator, message="invalid connection",
        load_func=function(value) return load_connection(value) end,
        save_func=function(value) return save_connection(value) end,},
}

ipsec_processor = {
    proposal={update="update_proposal", delete="delete_proposal", validator=proposal_validator},
    remote={validator=remote_validator},
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
    entry({"sdewan", configuration, ver, "proposals"}, call("handle_request")).leaf = true
    entry({"sdewan", configuration, ver, "remotes"}, call("handle_request")).leaf = true
end

-- Validate authentication method and secrets
function check_auth_method(value)
    local method = value["authentication_method"]
    local psk = value["pre_shared_key"]
    local pubkey = value["local_public_cert"]
    local privatekey = value["local_private_cert"]
    local sharedca = value["shared_ca"]
    if method == "psk" then
        if psk ~= nil then
            return true, value
        else
            return false, "Secret does not exists for the authentication method: " .. method
        end
    elseif method == "pubkey" then
        if pubkey ~= nil and privatekey ~= nil
            and sharedca ~= nil then
            return true, value
        else
            return false, "Secret does not exists for the authentication method: " .. method
        end
    else
        return false, "Invalid authentication method"
    end
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

    mime.decode("base64")
    local cert = mime.unb64(content)
    file:write(cert)
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
    mime.decode("base64")
    return mime.unb64(content)
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
        add_to_key_list(ret_value, connections[i]["conn_type"], connections[i])
    end

    return true, ret_value
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
                    if section[checker[j]] ~= nil then
                        for k=1, #section[checker[j]] do
                            if name == section[checker[j]][k] then
                                is_used = true
                                return false
                            end
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
    if res == true or code == 404 then
        return utils.create_object(_M, ipsec_processor, "proposal", proposal)
    end

    return false, code, msg
end
