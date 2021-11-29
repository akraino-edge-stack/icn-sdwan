--- SPDX-License-Identifier: Apache-2.0
--- Copyright (c) 2021 Intel Corporation

module("luci.controller.rest_v1.nat_rest", package.seeall)

local uci = require "luci.model.uci"

json = require "luci.jsonc"
io = require "io"
sys = require "luci.sys"
utils = require "luci.controller.rest_v1.utils"
ifutil = require "luci.controller.rest_v1.ifutil"

uci_conf = "firewall-nat"

nat_validator = {
    create_section_name=false,
    object_validator=function(value) return check_nat(value) end,
    {name="name"},
    {name="src", validator=function(value) return utils.in_array(value, {"#default", "#source"}) or ifutil.is_interface_available(value) end, message="invalid src", code="428"},
    {name="src_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_ip"},
    {name="src_dip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid src_dip"},
    {name="src_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="src_dport", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid src_port"},
    {name="proto", validator=function(value) return utils.in_array(value, {"tcp", "udp", "tcpudp", "udplite", "icmp", "esp", "ah", "sctp", "all"}) end, message="invalid proto"},
    {name="dest", validator=function(value) return utils.in_array(value, {"#default", "#source"}) or ifutil.is_interface_available(value) end, message="invalid dest", code="428"},
    {name="dest_ip", validator=function(value) return utils.is_valid_ip(value) end, message="invalid dest_ip"},
    {name="dest_port", validator=function(value) return utils.is_integer_and_in_range(value, 0) end, message="invalid dest_port"},
    {name="target", validator=function(value) return utils.in_array(value, {"DNAT", "SNAT", "MASQUERADE"}) end, message="invalid target"},
    {name="index", validator=function(value) return utils.is_integer_and_in_range(value, -1) end, message="invalid index"},
}

nat_processor = {
    nat={create="create_nat", delete="delete_nat", validator=nat_validator},
    configuration=uci_conf
}

function index()
    ver = "v1"
    configuration = "nat"
    entry({"sdewan", configuration, ver, "nats"}, call("handle_request")).leaf = true
end

function check_nat(value)
    local target = value["target"]
    if target == "SNAT" then
        if value["src_dip"] == nil then
            return false, "src_dip is required for SNAT"
        end
        if value["dest"] == nil then
            return false, "dest is required for SNAT"
        end
    end

    if target == "MASQUERADE" then
        if value["dest"] == nil then
            return false, "dest is required for SNAT MASQUERADE"
        end
    end

    if target == "DNAT" then
--      if value["src"] == nil then
--          return false, "src is required for DNAT"
--      end
        if value["dest_ip"] == nil and value["dest_port"] == nil then
            return false, "dest_ip or dest_port are required for DNAT"
        end
    end

    return true, value
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
        return utils[handler](_M, nat_processor)
    end
end

-- generate iptables command for nat
function nat_command(nat, op)
    local target = nat["target"]
    local proto = nat["proto"]
    local src = nat["src"]
    local src_ip = nat["src_ip"]
    if src == "#default" then
        src = ifutil.get_default_ifname()
    end
    local src_dip = nat["src_dip"]
    local src_port = nat["src_port"]
    local src_dport = nat["src_dport"]
    local dest = nat["dest"]
    local dest_ip = nat["dest_ip"]
    if dest == "#default" then
        dest = ifutil.get_default_ifname()
    end
    local dest_port = nat["dest_port"]
    local index = nat["index"]
    if index == nil or index == "" then
        index = "0"
    end

    local comm = "iptables -t nat"
    if op == "create" then
        if index == "0" then
            comm = comm .. " -A"
        else
            comm = comm .. " -I"
        end
    else
        comm = comm .. " -D"
    end
    if target == "SNAT" or target == "MASQUERADE" then
        comm = comm .. " POSTROUTING"
        if index ~= "0" and op == "create" then
            comm = comm .. " " .. index
        end
        if dest == "#source" then
            dest = ifutil.get_name_by_ip(src_dip)
        end
        if dest ~= nil and dest ~= "" then
            comm = comm .. " -o " .. dest
        end
    else
        comm = comm .. " PREROUTING"
        if index ~= "0" and op == "create" then
            comm = comm .. " " .. index
        end
        if src ~= nil and src ~= "" then
            comm = comm .. " -i " .. src
        end
    end

    if proto ~= nil and proto ~= "" then
        comm = comm .. " -p " .. proto
    end
    if src_ip ~= nil and src_ip ~= "" then
        comm = comm .. " -s " .. src_ip
    end
    if src_port ~= nil and src_port ~= "" then
        comm = comm .. " --sport " .. src_port
    end

    if target == "SNAT" then
        if dest_ip ~= nil and dest_ip ~= "" then
            comm = comm .. " -d " .. dest_ip
        end
        if dest_port ~= nil and dest_port ~= "" then
            comm = comm .. " --dport " .. dest_port
        end
        local new_src = src_dip
        if src_dport ~= nil and src_dport ~= "" then
            new_src = new_src .. ":" .. src_dport
        end
        comm = comm .. " -j SNAT --to-source " .. new_src
    elseif target == "DNAT" then
        if src_dip ~= nil and src_dip ~= "" then
            comm = comm .. " -d " .. src_dip
        end
        if src_dport ~= nil and src_dport ~= "" then
            comm = comm .. " --dport " .. src_dport
        end
        local new_des = dest_ip
        if new_des ~= nil and new_des ~= "" then
            if dest_port ~= nil and dest_port ~= "" then
                new_des = new_des .. ":" .. dest_port
            end
            comm = comm .. " -j DNAT --to-destination " .. new_des
        else
            if dest_port ~= nil and dest_port ~= "" then
                new_des = dest_port
                comm = comm .. " -j REDIRECT --to-port " .. new_des
            end
        end
    else
        if dest_ip ~= nil and dest_ip ~= "" then
            comm = comm .. " -d " .. dest_ip
        end
        if dest_port ~= nil and dest_port ~= "" then
            comm = comm .. " --dport " .. dest_port
        end
        comm = comm .. " -j MASQUERADE"
    end
    utils.log(comm)
    return comm
end

-- create a nat
function create_nat(nat)
    local name = nat.name
    local res, code, msg = utils.create_uci_section(uci_conf, nat_validator, "nat", nat)

    if res == false then
        uci:revert(uci_conf)
        return res, code, msg
    end

    -- create nat rule
    local comm = nat_command(nat, "create")
    os.execute(comm)

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end

-- delete a nat
function delete_nat(name)
    -- check whether nat is defined
    local nat = utils.get_object(_M, nat_processor, "nat", name)
    if nat == nil then
        return false, 404, "nat " .. name .. " is not defined"
    end

    -- delete nat rule in iptable
    local comm = nat_command(nat, "delete")
    os.execute(comm)

    utils.delete_uci_section(uci_conf, nat_validator, nat, "nat")

    -- commit change
    uci:save(uci_conf)
    uci:commit(uci_conf)

    return true
end
