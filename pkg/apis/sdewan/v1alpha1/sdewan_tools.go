package v1alpha1

import (
    "fmt"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

func NetworkInterfaceMap(instance Sdewan) map[string]string {
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

