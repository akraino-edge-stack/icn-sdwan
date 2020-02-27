-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.mwan3_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

policy_member_validator = {
    {name="interface", required=true, validator=function(value) return is_interface_available(value) end, message="invalid interface"},
    {name="metric", required=true, validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="metric should be greater than 0"},
    {name="weight", required=true, validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="weight should be greater than 0"}
}

policy_validator = {
    {name="name"},
    {name="members", required=true, item_validator=policy_member_validator, message="Incorrect policy member"} 
}

rule_validator = {
    {name="name"},
    {name="policy", required=true, target="use_policy", validator=function(value) return is_policy_available(value) end, message="invalid policy"},
    {name="src_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid ip address"},
    {name="src_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="dest_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid ip address"},
    {name="dest_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid dest_port"},
    {name="proto", validator=function(value) return utils.in_array(value, {"tcp", "udp", "icmp", "all"}) end, message="wrong proto"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "all"}) end, message="wrong family"},
    {name="sticky", validator=function(value) return utils.in_array(value, {"0", "1"}) end, message="wrong sticky"},
    {name="timeout", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid timeout"}
}

mwan3_processor = {
    policy={get="get_policy", update="update_policy", create="create_policy", delete="delete_policy", validator=policy_validator},
    rule={validator=rule_validator},
    configuration="mwan3"
}

function index()
    ver = "v1"
    entry({"sdewan", "mwan3", ver, "policies"}, call("get_policies"))
    entry({"sdewan", "mwan3", ver, "rules"}, call("get_rules"))
    entry({"sdewan", "mwan3", ver, "policy"}, call("handle_request")).leaf = true
    entry({"sdewan", "mwan3", ver, "rule"}, call("handle_request")).leaf = true
end

-- Request Handler
function handle_request()
    local handler = utils.handles_table[utils.get_req_method()]
    if handler == nil then
        utils.response_error(405, "Method Not Allowed")
    else
        return utils[handler](_M, mwan3_processor)
    end
end

-- check interface
function is_interface_available(name)
    local interfaces = uci:get_all("mwan3", name)
    if interfaces == nil then
        return false, "Interface[" .. name .. "] is not definied"
    end
    
    return true
end

-- check policy
function is_policy_available(name)
    local policy = uci:get_all("mwan3", name)
    if policy == nil then
        return false, "Policy[" .. name .. "] is not definied"
    end
    
    return true
end

-- policy APIs
-- get a policy
function get_policy(name)
    local members = uci:get_all("mwan3", name)
    if members == nil then
        return nil
    end
    members = members["use_member"]

    local policy = {}
    policy["name"] = name
    policy["members"] = {}

    for i=1, #members do
        local member_name = members[i]
        local member_data = uci:get_all("mwan3", member_name)
        if not (member_data == nil) then
            local member = {}
            member["name"] = member_name
            utils.set_data("mwan3", policy_member_validator, member_data, member)
            policy["members"][i] = member
        end
    end

    return policy
end

-- get /policies
function get_policies()
    if not (utils.validate_req_method("GET")) then
        return
    end

    local res = {}
    res["policies"] = {}

    local index = 1
    uci:foreach("mwan3", "policy",
        function (section)
            policy = get_policy(section[".name"])
            if not (policy == nil) then
                res["policies"][index] = policy
                index = index + 1
            end
        end
    )

    utils.response_object(res)
end

-- create a policy
function create_policy(policy)
    -- check whether policy is exist
    local policy_name = policy.name
    local exist_policy_obj = uci:get_all("mwan3", policy_name)
    if exist_policy_obj ~= nil then
        return false, 409, "Conflict"
    end

    -- check member interfaces
    local member_interface_names = {}
    local members = policy.members
    for i=1, #members do
        if utils.in_array(policy_name .. "_" .. members[i].interface, member_interface_names) then
            return false, 400, "Same interface is used in different members"
        end
        member_interface_names[i] = policy_name .. "_" .. members[i].interface
    end

    -- create member
    for i=1, #members do
        uci:section("mwan3", "member", member_interface_names[i], {
            interface = members[i].interface,
            metric = members[i].metric,
            weight = members[i].weight})
    end

    -- create policy
    uci:section("mwan3", "policy", policy_name)
    uci:set_list("mwan3", policy_name, "use_member", member_interface_names)

    -- commit change
    uci:save("mwan3")
    uci:commit("mwan3")

    return true
end

-- check if policy is used by rule
function is_policy_used(name)
    local is_used = false

    uci:foreach("mwan3", "rule",
        function (section)
            rule = utils.get_object(_M, mwan3_processor, "rule", section[".name"])
            if not (rule == nil) then
                if rule.policy == name then
                    is_used = true
                    return false
                end
            end
        end
    )

    return is_used
end

-- delete a policy
function delete_policy(name, check_used)
    -- check whether policy is defined
    local policy = get_policy(name)
    if policy == nil then
        return false, 404, "policy " .. name .. " is not defined"
    end

    if check_used == nil then
        check_used = true
    else
        check_used = false
    end

    -- Todo: check whether the policy is used by a rule
    if check_used == true and is_policy_used(name) then
        return false, 400, "policy " .. name .. " is used"
    end

    -- delete policy
    uci:delete("mwan3", name)

    -- delete members
    local members = policy.members
    for i=1, #members do
        uci:delete("mwan3", members[i].name)
    end

    -- commit change
    uci:save("mwan3")
    uci:commit("mwan3")

    return true
end

-- update a policy
function update_policy(policy)
    local name = policy.name
    res, code, msg = delete_policy(name, false)
    if res == true then
        return create_policy(policy)
    end

    return false, code, msg
end

-- Rule APIs
-- get /rules
function get_rules()
    utils.handle_get_objects("rules", "mwan3", "rule", rule_validator)
end
