-- Licensed to the public under the GNU General Public License v2.

module("luci.controller.rest_v1.index", package.seeall)

function index()
    ver = "v1"
    entry({"sdewan", ver}, call("help")).dependent = false
    entry({"sdewan", "mwan3", ver}, call("help")).dependent = false
    entry({"sdewan", "firewall", ver}, call("help")).dependent = false
    entry({"sdewan", "ipsec", ver}, call("help")).dependent = false
end

function help()
    luci.http.prepare_content("application/json")
    luci.http.write('{"message":"sdewan restful API service v1"}')
end
