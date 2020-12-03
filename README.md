# Introduction to Akraino ICN SD-EWAN solution

SD-EWAN main functionality include

* IPsec tunnels across K8s clusters - Supporting multiple types of K8s clusters
"K8s clusters having static public IP address", "K8s clusters having dynamic public
IP address with static FQDN" and "K8s clusters with no public IP".
* Stateful inspection firewall (for inbound and outbound connections)
* Source NAT and Destination NAT for supporting K8s clusters whose POD and
ClusterIP subnets are overlapping.
* Multiple WAN link support

SD-EWAN is based on set of Linux packages

* mwan3 (for Multiple WAN link support)
* IPTables (for firewall, SNAT, DNAT)
* Strongswan (for IPsec)
* TC (Traffic Control)

## SD-EWAN in Akraino/ICN

SD-EWAN functionality is realized via CNF (Containerized Network Function)
and deployed by K8s. SD-EWAN CNF leverages Linux kernel functionality for packet
processing of above functions. Actual CNF is set of user space processes
consisting of fw3, mwan3, strongswan and others.

SD-EWAN considered as platform feature by ICN.

Refer - https://www.linkedin.com/pulse/software-defined-edge-wan-edges-srinivasa-addepalli/

### platform
SDEWAN platform features include cnf, cr definition and controller.

### traffic-hub
traffic-hub's high level design can be found at: https://www.linkedin.com/pulse/software-defined-edge-wan-central-control-traffic-hub-addepalli/

### central-controller
central-controller's high level design can be found at: https://www.linkedin.com/pulse/software-defined-edge-wan-central-control-traffic-hub-addepalli/

## Comprehensive Documentation
- [How to use](platform/crd-ctrlr#deployment-guide)
- [Development](platform/crd-ctrlr#developer-guide)
- [test](platform/test#sdewan-solution-e2e-test)

## Contact Us

For any questions about ovn4nfv k8s , feel free to ask a question in
#general in the [ICN slack](https://akraino-icn-admin.herokuapp.com/), or open up a https://jira.opnfv.org/issues/.
