-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.firewall_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"

uci_conf = "firewall"

zone_validator = {
    create_section_name=false,
    {name="name"},
    {name="network", item_validator=function(value) return is_network_interface_available(value) end, message="invalid network"},
    {name="masq", validator=function(value) return utils.in_array(value, {"0", "1"}) end, message="invalid masq"},
    {name="masq_src", item_validator=function(value) return is_valid_masq_subset(value) end, message="invalid masq_src"},
    {name="masq_dest", item_validator=function(value) return is_valid_masq_subset(value) end, message="invalid masq_dest"},
    {name="masq_allow_invalid", validator=function(value) return utils.in_array(value, {"0", "1"}) end, message="invalid masq_allow_invalid"},
    {name="mtu_fix", validator=function(value) return utils.in_array(value, {"0", "1"}) end, message="invalid mtu_fix"},
    {name="input", validator=function(value) return utils.in_array(value, {"ACCEPT", "REJECT", "DROP"}) end, message="invalid input"},
    {name="forward", validator=function(value) return utils.in_array(value, {"ACCEPT", "REJECT", "DROP"}) end, message="invalid forward"},
    {name="output", validator=function(value) return utils.in_array(value, {"ACCEPT", "REJECT", "DROP"}) end, message="invalid output"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "any"}) end, message="invalid family"},
    {name="subnet", item_validator=function(value) return utils.is_valid_ip(value) end, message="invalid subnet"},
    {name="extra_src"},
    {name="etra_dest"},
}

redirect_validator = {
    create_section_name=false,
    object_validator=function(value) return check_redirect(value) end,
    {name="name"},
    {name="src", validator=function(value) return is_zone_available(value) end, message="invalid src"},
    {name="src_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_ip"},
    {name="src_dip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_dip"},
    {name="src_mac", validator=function(value) return utils.is_valid_mac(value) end, message="invalid src_mac"},
    {name="src_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="src_dport", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="proto", validator=function(value) return utils.in_array(value, {"tcp", "udp", "tcpudp", "udplite", "icmp", "esp", "ah", "sctp", "all"}) end, message="invalid proto"},
    {name="dest", validator=function(value) return is_zone_available(value) end, message="invalid dest"},
    {name="dest_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid dest_ip"},
    {name="dest_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid dest_port"},
    {name="mark"},
    {name="target", validator=function(value) return utils.in_array(value, {"DNAT", "SNAT"}) end, message="invalid target"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "any"}) end, message="invalid family"},
}

rule_validator = {
    create_section_name=false,
    object_validator=function(value) return check_rule(value) end,
    {name="name"},
    {name="src", validator=function(value) return is_zone_available(value) end, message="invalid src"},
    {name="src_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_ip"},
    {name="src_mac", validator=function(value) return utils.is_valid_mac(value) end, message="invalid src_mac"},
    {name="src_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="proto", validator=function(value) return utils.in_array(value, {"tcp", "udp", "tcpudp", "udplite", "icmp", "esp", "ah", "sctp", "all"}) end, message="invalid proto"},
    {name="icmp_type", is_list=true, item_validator=function(value) return check_icmp_type(value) end, message="invalid icmp_type"},
    {name="dest", validator=function(value) return is_zone_available(value) end, message="invalid dest"},
    {name="dest_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid dest_ip"},
    {name="dest_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid dest_port"},
    {name="mark"},
    {name="target", validator=function(value) return utils.in_array(value, {"ACCEPT", "REJECT", "DROP", "MARK", "NOTRACK"}) end, message="invalid target"},
    {name="set_mark"},
    {name="set_xmark"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "any"}) end, message="invalid family"},
    {name="extra"},
}

forwarding_validator = {
    create_section_name=false,
    {name="name"},
    {name="src", required=true, validator=function(value) return is_zone_available(value) end, message="invalid src"},
    {name="dest", required=true, validator=function(value) return is_zone_available(value) end, message="invalid dest"},
    {name="family", validator=function(value) return utils.in_array(value, {"ipv4", "ipv6", "any"}) end, message="invalid family"},
}

firewall_processor = {
    zone={update="update_zone", delete="delete_zone", validator=zone_validator},
    redirect={validator=redirect_validator},
    rule={validator=rule_validator},
    forwarding={validator=forwarding_validator},
    configuration=uci_conf
}

zone_checker = {
    {name="rule", checker={"src", "dest"}},
    {name="forwarding", checker={"src", "dest"}},
    {name="redirect", checker={"src", "dest"}},
}

function index()
    ver = "v1"
    configuration = "firewall"
    entry({"sdewan", configuration, ver, "zones"}, call("handle_request")).leaf = true
    entry({"sdewan", configuration, ver, "redirects"}, call("handle_request")).leaf = true
    entry({"sdewan", configuration, ver, "rules"}, call("handle_request")).leaf = true
    entry({"sdewan", configuration, ver, "forwardings"}, call("handle_request")).leaf = true
end

function is_network_interface_available(interface)
    local interfaces = uci:get_all("network", interface)
    if interfaces == nil then
        return false, "Interface[" .. interface .. "] is not definied"
    end

    return true, interface
end

function is_valid_masq_subset(s)
    local ip = s
    if utils.start_with(ip, "!") then
        ip = string.sub(ip, 2, string.len(ip))
    end

    local res, _ = utils.is_valid_ip(ip)
    if res then
        return true, s
    else
        return false
    end
end

function check_redirect(value)
    local target = value["target"]
    if target == "SNAT" then
        if value["src_dip"] == nil then
            return false, "src_dip is required for SNAT"
        end
        if value["dest"] == nil then
            return false, "dest is required for SNAT"
        end
    end

    if target == "DNAT" then
        if value["src"] == nil then
            return false, "src is required for DNAT"
        end
    end

    return true, value
end

function check_rule(value)
    local target = value["target"]
    if target == "MARK" then
        if value["set_mark"] == nil and value["set_xmark"] == nil then
            return false, "set_mark or set_xmark is required for MARK"
        end
    end

    return true, value
end

function check_icmp_type(value)
    return utils.in_array(value, {"address-mask-reply", "address-mask-request ", "any", "communication-prohibited",
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
    local handler = utils.handles_table[utils.get_req_method()]
    if handler == nil then
        utils.response_error(405, "Method Not Allowed")
    else
        return utils[handler](_M, firewall_processor)
    end
end

-- Zone APIs
-- check if zone is used by rule, forwarding or redirect
function is_zone_used(name)
    local is_used = false

    for i,v in pairs(zone_checker) do
        local section_name = v["name"]
        local checker = v["checker"]
        uci:foreach(uci_conf, section_name,
            function(section)
                for j=1, #checker do
                    if name == section[checker[j]] then
                        is_used = true
                        return false
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

function is_zone_available(name)
    local zone = utils.get_object(_M, firewall_processor, "zone", name)
    if zone == nil then
        return false, "Zone[" .. name .. "] is not definied"
    end

    return true, name
end

-- delete a zone
function delete_zone(name, check_used)
    -- check whether zone is defined
    local zone = utils.get_object(_M, firewall_processor, "zone", name)
    if zone == nil then
        return false, 404, "zone " .. name .. " is not defined"
    end

    if check_used == nil then
        check_used = true
    else
        check_used = false
    end

    -- Todo: check whether the zone is used by a rule
    if check_used == true and is_zone_used(name) then
        return false, 400, "zone " .. name .. " is used"
    end

    -- delete zone
    uci:foreach(uci_conf, "zone",
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

-- update a zone
function update_zone(zone)
    local name = zone.name
    res, code, msg = delete_zone(name, false)
    if res == true or code == 404 then
        return utils.create_object(_M, firewall_processor, "zone", zone)
    end

    return false, code, msg
end