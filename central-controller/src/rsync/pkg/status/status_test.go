package status_test

import (
	"io/ioutil"
	"testing"

	rb "gitlab.com/project-emco/core/emco-base/src/monitor/pkg/apis/k8splugin/v1alpha1"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/appcontext"
	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/contextdb"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/context"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/internal/utils"
	"gitlab.com/project-emco/core/emco-base/src/rsync/pkg/status"
	. "gitlab.com/project-emco/core/emco-base/src/rsync/pkg/types"

	"k8s.io/api/core/v1"
)

func init() {
	edb := new(contextdb.MockConDb)
	edb.Err = nil
	contextdb.Db = edb
}

var rbfile, _ = ioutil.ReadFile("test/test.yaml")

var TestCA CompositeApp = CompositeApp{
	CompMetadata: appcontext.CompositeAppMeta{Project: "proj1", CompositeApp: "ca1", Version: "v1", Release: "r1",
		DeploymentIntentGroup: "dig1", Namespace: "default", Level: "0"},
	AppOrder: []string{"collectd"},
	Apps: map[string]*App{"collectd": {
		Name: "collectd",
		Clusters: map[string]*Cluster{"provider1+cluster1": {
			Name: "provider1+cluster1",
			Resources: map[string]*AppResource{"r1": {Name: "r1", Data: "a1c1r1"},
				"r2": {Name: "r2", Data: "a1c1r2"},
			},
			ResOrder: []string{"r1", "r2"}}},
	},
	},
}

func TestAppReadyOnAllClusters(t *testing.T) {
	// Read in test data

	data := &rb.ResourceBundleState{}
	_, err := utils.DecodeYAMLData(string(rbfile), data)
	if err != nil {
		return
	}
	cid, _ := context.CreateCompApp(TestCA)

	testCases := []struct {
		label                  string
		expectedValue          bool
		podReady               v1.ConditionStatus
		updatedNumberScheduled int32
	}{
		{
			label:                  "AppReadyOnAllClusters Success Case",
			expectedValue:          true,
			updatedNumberScheduled: 1,
			podReady:               v1.ConditionTrue,
		},
		{
			label:                  "AppReadyOnAllClusters Daemonset not ready",
			expectedValue:          false,
			updatedNumberScheduled: 0,
			podReady:               v1.ConditionTrue,
		},
		{
			label:                  "AppReadyOnAllClusters Pod not ready",
			expectedValue:          false,
			updatedNumberScheduled: 1,
			podReady:               v1.ConditionFalse,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			data.Status.DaemonSetStatuses[0].Status.UpdatedNumberScheduled = testCase.updatedNumberScheduled
			data.Status.PodStatuses[0].Status.Conditions[1].Status = testCase.podReady
			val := status.UpdateAppReadyStatus(cid, "collectd", "provider1+cluster1", data)
			if val != testCase.expectedValue {
				t.Fatalf("TestAppReadyOnAllClusters Failed")
			}
		})
	}
}
