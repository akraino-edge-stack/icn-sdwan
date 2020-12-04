# sdewan solution e2e test

This folder contains the e2e test scripts based on RESTful API RESTful API exported by SDEWAN CNF, as well as scripts based on kubernetes customresourcedefiniton(CRD) exported by SDEWAN CRD Controller.

* Test scenario
The test scripts will setup 3 all-in-one node k8s clusters within virtual machines which act as edge-a, edge-b and the hub. The three clusters are linked through provider network based on OVN. The two edges only obtain a private IP address, hence they cannot be reached by the hub. However, the hub obtains a public IP address which is accessible by others. IPSEC tunnel are then setup up between edge and hub to enable security connection.

The test script then do the following steps for the edges:
- Setup SD-EWAN CNF as well as CRD controller(this only appears in the CRD case)
- Setup test application(Httpbin)
- Apply IPSEC CRs
- Apply Firewall CRs

For the hub, it will do the following ones:
- Setup SD-EWAN CNF as well as CRD controller(this only appears in the CRD case)
- Apply IPSEC CRs

After all configurations are applied, it will launch a connection test between the
two edges and see if the connection is working fine.


* Test scripts based on RESTful API

The e2e-test folder contains the test scripts for our end-to-end scenario based on RESTful API.

cd e2e-test
./test.sh

This will automatically bring up our test scenario, as well as connection check.

* Test scripts based on CRD

The e2e-test-crd folder contains the test scripts for our end-to-end scenario based on CRD.

cd e2e-test-crd
./test.sh

This will automatically bring up our test scenario, as well as connection check.
