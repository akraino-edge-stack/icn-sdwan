--- SPDX-License-Identifier: Apache-2.0 
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.fw_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
ifutil = require "luci.controller.rest_v1.ifutil"

uci_conf = "firewall-rule"

fw_validator = {
    create_section_name=false,
    object_validator=function(value) return check_rule(value) end,
    {name="name"},
    {name="src", required=true},
    {name="src_val"},
    {name="src_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_ip"},
    {name="src_mac", validator=function(value) return utils.is_valid_mac(value) end, message="invalid src_mac"},
    {name="src_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="proto", validator=function(value) return utils.in_array(value, {"tcp", "udp", "tcpudp", "udplite", "icmp", "esp", "ah", "sctp", "all"}) end, message="invalid proto"},
    {name="icmp_type", is_list=true, item_validator=function(value) return check_icmp_type(value) end, message="invalid icmp_type"},
    {name="dest"},
    {name="dest_val"},
    {name="dest_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid dest_ip"},
    {name="dest_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid dest_port"},
    {name="mark"},
    {name="target", required=true, validator=function(value) return utils.in_array(value, {"ACCEPT", "REJECT", "DROP", "MARK", "NOTRACK"}) end, message="invalid target"},
    {name="set_mark"},
    {name="set_xmark"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "any"}) end, message="invalid family"},
    {name="extra"},
}

fw_processor = {
    rule={create="create_rule", delete="delete_rule", validator=fw_validator},
    configuration=uci_conf
}

function index()
    ver = "v1"
    configuration = "networkfirewall"
    entry({"sdewan", configuration, ver, "rules"}, call("handle_request")).leaf = true
end

function check_rule(value)
    local target = value["target"]
    if target == "MARK" then
        if value["set_mark"] == nil and value["set_xmark"] == nil then
            return false, "set_mark or set_xmark is required for MARK"
        end
    end

    -- src
    local src = value["src"]
    local src_val = src
    if utils.start_with(src, "#") then
        local src_name = string.sub(src, 2, string.len(src))
        if src_name == "default" then
            src_val = ifutil.get_default_ifname()
        else
            src_val = ifutil.get_name_by_ip(src_name)
        end
    end

    if src_val == nil or (not ifutil.is_interface_available(src_val)) then
        return false, "428:Field[src] checked failed: Invalid interface"
    end

    value["src_val"] = src_val

    -- dest
    local dest = value["dest"]
    if dest ~= nil then
        local dest_val = dest
        if utils.start_with(dest, "#") then
            local dest_name = string.sub(dest, 2, string.len(dest))
            if dest_name == "default" then
                dest_val = ifutil.get_default_ifname()
            else
                dest_val = ifutil.get_name_by_ip(dest_name)
            end
        end

        if dest_val == nil or (not ifutil.is_interface_available(dest_val)) then
            return false, "428:Field[dest] checked failed: Invalid interface"
        end

        value["dest_val"] = dest_val
    end

    return true, value
end

function check_icmp_type(value)
    return utils.in_array(value, {"address-mask-reply", "address-mask-request", "any", "communication-prohibited",
        "destination-unreachable", "echo-reply", "echo-request", "fragmentation-needed", "host-precedence-violation",
        "host-prohibited", "host-redirect", "host-unknown", "host-unreachable", "ip-header-bad", "network-prohibited",
        "network-redirect", "network-unknown", "network-unreachable", "parameter-problem", "ping", "pong",
        "port-unreachable", "precedence-cutoff", "protocol-unreachable", "redirect", "required-option-missing",
        "router-advertisement", "router-solicitation", "source-quench", "source-route-failed", "time-exceeded",
        "timestamp-reply", "timestamp-request", "TOS-host-redirect", "TOS-host-unreachable", "TOS-network-redirect",
        "TOS-network-unreachable", "ttl-exceeded", "ttl-zero-during-reassembly", "ttl-zero-during-transit"})
end

-- Request Handler
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
        return utils[handler](_M, fw_processor)
    end
end

-- generate iptables command for rule
function rule_commands(rule, op)
    local target = rule["target"]
    local proto = rule["proto"]
    if proto == nil then
        proto = "all"
    end
    local family = rule["family"]
    local extra = rule["extra"]

    local src_val = rule["src_val"]
    local src_ip = rule["src_ip"]
    local src_mac = rule["src_mac"]
    local src_port = rule["src_port"]
    
    local icmp_type = rule["icmp_type"]

    local dest_val = rule["dest_val"]
    local dest_ip = rule["dest_ip"]
    local dest_port = rule["dest_port"]
    
    local mark = rule["mark"]
    local set_mark = rule["set_mark"]
    local set_xmark = rule["set_xmark"]

    local comms = {}

    local comm = "iptables"
    local chain = "INPUT"
    if dest_val ~= nil then
        chain = "FORWARD"
    end
    if target == "MARK" then
        comm = comm .. " -t mangle"
        chain = "PREROUTING"
    end

    if op == "create" then
        comm = comm .. " -A " .. chain
    else
        comm = comm .. " -D " .. chain
    end

    comm = comm .. " -i " .. src_val
    if dest_val ~= nil and dest_val ~= "" and target ~= "MARK" then
        comm = comm .. " -o " .. dest_val
    end
    if src_ip ~= nil and src_ip ~= "" then
        comm = comm .. " -s " .. src_ip
    end
    if dest_ip ~= nil and dest_ip ~= "" then
        comm = comm .. " -d " .. dest_ip
    end

    local post_comm = ""
    if src_mac ~= nil and src_mac ~= "" then
        post_comm = post_comm .. " -m mac --mac-source " .. src_mac
    end
    if mark ~= nil and mark ~= "" then
        post_comm = post_comm .. " -m mark --mark " .. mark
    end
    post_comm = post_comm .. " -j " .. target
    if target == "MARK" then
        if set_xmark ~= nil and set_xmark ~= "" then
            post_comm = post_comm .. " --set-xmark " .. set_xmark
        else
            if set_mark ~= nil and set_mark ~= "" then
                post_comm = post_comm .. " --set-mark " .. set_mark
            end
        end
    end

    local port_val = ""
    if src_port ~= nil and src_port ~= "" then
        port_val = port_val .. " --sport " .. src_port
    end
    if dest_port ~= nil and dest_port ~= "" then
        port_val = port_val .. " --dport " .. dest_port
    end
    local proto_vals = {}
    if proto == "all" then
        proto_vals[1] = ""
    end
    if proto == "tcp" then
        proto_vals[1] = " -p tcp -m tcp" .. port_val
    end
    if proto == "udp" then
        proto_vals[1] = " -p udp -m udp" .. port_val
    end
    if proto == "tcpudp" then
        proto_vals[1] = " -p tcp -m tcp" .. port_val
        proto_vals[2] = " -p udp -m udp" .. port_val
    end
    if proto == "icmp" then
        for j=1, #icmp_type do
            proto_vals[j] = " -p icmp -m icmp --icmp-type " .. icmp_type[j]
        end
    end
    if proto == "udplite" or proto == "esp" or proto == "ah" or proto == "sctp" then
        proto_vals[1] = " -p " .. proto
    end
    
    for j=1, #proto_vals do
        comms[j] = comm .. proto_vals[j] .. post_comm
    end

    return comms
end

-- create a rule
function create_rule(rule)
    local name = rule.name
    local res, code, msg = utils.create_uci_section(uci_conf, fw_validator, "rule", rule)

    if res == false then
        uci:revert(uci_conf)
        return res, code, msg
    end

    -- create firewall rule
    local comms = rule_commands(rule, "create")
    for j=1, #comms do
        utils.log(comms[j])
        os.execute(comms[j])
    end

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end

-- delete a rule
function delete_rule(name)
    -- check whether rule is defined
    local rule = utils.get_object(_M, fw_processor, "rule", name)
    if rule == nil then
        return false, 404, "rule " .. name .. " is not defined"
    end

    -- delete rule rule in iptable
    local comms = rule_commands(rule, "delete")
    for j=1, #comms do
        utils.log(comms[j])
        os.execute(comms[j])
    end

    utils.delete_uci_section(uci_conf, fw_validator, rule, "rule")

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end