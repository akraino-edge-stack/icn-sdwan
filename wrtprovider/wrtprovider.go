package wrtprovider

import (
	"fmt"
	"reflect"
	sdewanv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_sdewan")

func NetworkInterfaceMap(instance *sdewanv1alpha1.Sdewan) map[string]string {
	ifMap := make(map[string]string)
	for i, network := range instance.Spec.Networks {
		prefix := "lan_"
		if network.IsProvider {
			prefix = "wan_"
		}
		if network.Interface == "" {
			network.Interface = fmt.Sprintf("net%d", i)
		}
		ifMap[network.Name] = prefix + fmt.Sprintf("net%d", i)
	}
	return ifMap
}

func Mwan3ReplacePolicies(policies []openwrt.SdewanPolicy, existOnes []openwrt.SdewanPolicy, client *openwrt.Mwan3Client) error {
	// create/update new policies
	for _, policy := range policies {
		found := false
		for _, p := range existOnes {
			if p.Name == policy.Name {
				if !reflect.DeepEqual(policy, p) {
					_, err := client.UpdatePolicy(policy)
					if err != nil {
						return err
					}
				}
				found = true
				break
			}
		}
		if found == false {
			_, err := client.CreatePolicy(policy)
			if err != nil {
				return err
			}
		}
	}

	// remove old policies
	for _, p := range existOnes {
		found := false
		for _, policy := range policies {
			if p.Name == policy.Name {
				found = true
				break
			}
		}
		if found == false {
			err := client.DeletePolicy(p.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Mwan3ReplaceRules(rules []openwrt.SdewanRule, existOnes []openwrt.SdewanRule, client *openwrt.Mwan3Client) error {
	// create/update new rules
	for _, rule := range rules {
		found := false
		for _, r := range existOnes {
			if r.Name == rule.Name {
				if !reflect.DeepEqual(rule, r) {
					_, err := client.UpdateRule(rule)
					if err != nil {
						return err
					}
				}
				found = true
				break
			}
		}
		if found == false {
			_, err := client.CreateRule(rule)
			if err != nil {
				return err
			}
		}
	}

	// remove old rules
	for _, r := range existOnes {
		found := false
		for _, rule := range rules {
			if r.Name == rule.Name {
				found = true
				break
			}
		}
		if found == false {
			err := client.DeleteRule(r.Name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// apply policy and rules
func Mwan3Apply(mwan3Conf *sdewanv1alpha1.Mwan3Conf, sdewan *sdewanv1alpha1.Sdewan) error {
	reqLogger := log.WithValues("Mwan3Provider", mwan3Conf.Name, "Sdewan", sdewan.Name)
	openwrtClient := openwrt.NewOpenwrtClient(sdewan.Name, "root", "")
	mwan3 := openwrt.Mwan3Client{OpenwrtClient: openwrtClient}
	service := openwrt.ServiceClient{OpenwrtClient: openwrtClient}
	netMap := NetworkInterfaceMap(sdewan)
	var policies []openwrt.SdewanPolicy
	for policyName, members := range mwan3Conf.Spec.Policies {
		openwrtMembers := make([]openwrt.SdewanMember, len(members.Members))
		for i, member := range members.Members {
			openwrtMembers[i] = openwrt.SdewanMember{
				Interface: netMap[member.Network],
				Metric:    fmt.Sprintf("%d", member.Metric),
				Weight:    fmt.Sprintf("%d", member.Weight),
			}
		}
		policies = append(policies, openwrt.SdewanPolicy{
			Name:    policyName,
			Members: openwrtMembers})
	}
	existPolicies, err := mwan3.GetPolicies()
	if err != nil {
		reqLogger.Error(err, "Failed to fetch existing policies")
		return err
	}
	err = Mwan3ReplacePolicies(policies, existPolicies.Policies, &mwan3)
	if err != nil {
		reqLogger.Error(err, "Failed to apply Policies")
		return err
	}
	var rules []openwrt.SdewanRule
	for ruleName, rule := range mwan3Conf.Spec.Rules {
		openwrtRule := openwrt.SdewanRule{
			Name:     ruleName,
			Policy:   rule.UsePolicy,
			SrcIp:    rule.SrcIP,
			SrcPort:  rule.SrcPort,
			DestIp:   rule.DestIP,
			DestPort: rule.DestPort,
			Proto:    rule.Proto,
			Family:   rule.Family,
			Sticky:   rule.Sticky,
			Timeout:  rule.Timeout,
		}
		rules = append(rules, openwrtRule)
	}
	existRules, err := mwan3.GetRules()
	if err != nil {
		reqLogger.Error(err, "Failed to fetch existing rules")
		return err
	}
	err = Mwan3ReplaceRules(rules, existRules.Rules, &mwan3)
	if err != nil {
		reqLogger.Error(err, "Failed to apply rules")
		return err
	}
	_, err = service.ExecuteService("mwan3", "restart")
	if err != nil {
		reqLogger.Error(err, "Failed to restart mwan3 service")
		return err
	}
	return nil
}
