#SD-eWAN test scenario
[Overview]
In this test scenario, three clusters are created for edge-a, edge-b and sdewan-hub.
Two tunnels are established between the edge and the hub, and also two applications
are installed in edga-a and edge-b. Tunnels are verified thru the connection test between 
the two applications.

[Test guide]
Run the test.sh under sdwan/platform/test/e2e-test/ to invoke the vm creation and configurations.
  $ ./test.sh

Scripts description:
1. The Vagrantfile will be used to setup the base environment.
2. The installer.sh script contains the minimal Ubuntu instructions required for bringing up ICN.
3. The setup-cnf.sh script creates ovn networks, sdewan cnfs and application pods if needed.
4. The sdwan_verifier.sh script inserts configs into the sdewan cnf, including firewall and ipsec.
5. The test-connection.sh script under edge-a tests the connection between the applications
reside in edge-a and edge-b.


[License]

Apache-2.0

[1]: https://gerrit.akraino.org/r/icn/sdwan

[2]: https://git.onap.org/multicloud/k8s

[3]: https://www.vagrantup.com/

