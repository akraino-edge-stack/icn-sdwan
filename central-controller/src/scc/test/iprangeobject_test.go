package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/manager"
	"github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
	"os"
	"testing"
)

var BaseUrl string
var OverlayUrl string

func TestMain(m *testing.M) {
	servIp := flag.String("ip", "127.0.0.1", "SDEWAN Central Controller IP Address")
	flag.Parse()
	OverlayUrl = "http://" + *servIp + ":9015/scc/v1/" + manager.OverlayCollection
	BaseUrl = OverlayUrl + "/overlay1/" + manager.IPRangeCollection

	var overlay_object = module.OverlayObject{
		Metadata:      module.ObjectMetaData{"overlay1", "", "", ""},
		Specification: module.OverlayObjectSpec{}}

	var iprange_object1 = module.IPRangeObject{
		Metadata:      module.ObjectMetaData{"ipr1", "", "", ""},
		Specification: module.IPRangeObjectSpec{"192.168.0.2", 10, 12}}
	var iprange_object2 = module.IPRangeObject{
		Metadata:      module.ObjectMetaData{"ipr2", "", "", ""},
		Specification: module.IPRangeObjectSpec{"192.168.1.3", 32, 36}}

	createControllerObject(OverlayUrl, &overlay_object, &module.OverlayObject{})
	createControllerObject(BaseUrl, &iprange_object1, &module.IPRangeObject{})
	createControllerObject(BaseUrl, &iprange_object2, &module.IPRangeObject{})

	var ret = m.Run()

	deleteControllerObject(BaseUrl, "ipr1")
	deleteControllerObject(BaseUrl, "ipr2")
	deleteControllerObject(OverlayUrl, "overlay1")

	os.Exit(ret)
}

func TestGetObjects(t *testing.T) {
	url := BaseUrl
	res, err := callRest("GET", url, "")
	if err != nil {
		printError(err)
		t.Errorf("Test case GetObjects: can not get Objects")
		return
	}

	var objs []module.IPRangeObject
	err = json.Unmarshal([]byte(res), &objs)

	if len(objs) == 0 {
		fmt.Printf("Test case GetObjects: no object found")
		return
	}

	p_data, _ := json.Marshal(objs)
	fmt.Printf("%s\n", string(p_data))
}

func TestGetObject(t *testing.T) {
	tcases := []struct {
		name            string
		object_name     string
		expectedErr     bool
		expectedErrCode int
	}{
		{
			name:        "Normal",
			object_name: "ipr1",
		},
		{
			name:            "GetFoolName",
			object_name:     "foo_name",
			expectedErr:     true,
			expectedErrCode: 500,
		},
	}

	for _, tcase := range tcases {
		_, err := getControllerObject(BaseUrl, tcase.object_name, &module.IPRangeObject{})
		handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
	}
}

func TestCreateObject(t *testing.T) {
	tcases := []struct {
		name            string
		url             string
		obj             module.IPRangeObject
		expectedErr     bool
		expectedErrCode int
	}{
		{
			name: "EmptyName",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"", "object 1", "", ""},
				Specification: module.IPRangeObjectSpec{}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 422,
		},
		{
			name: "DumplicateName",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"ipr1", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.2.3", 10, 15}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 409,
		},
		{
			name: "WrongOverlayName",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.2.3", 10, 15}},
			url:             OverlayUrl + "/foooverlay/" + manager.IPRangeCollection,
			expectedErr:     true,
			expectedErrCode: 500,
		},
		{
			name: "WrongSubnet",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.2.3.0", 1, 15}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 422,
		},
		{
			name: "WrongMinIP",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.2.3", 0, 15}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 422,
		},
		{
			name: "WrongMaxIP",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.1.3", 1, 300}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 422,
		},
		{
			name: "WrongMinMaxIP",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.2.3", 20, 15}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 422,
		},
		{
			name: "ConflictRange1",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.0.3", 11, 15}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 500,
		},
		{
			name: "ConflictRange2",
			obj: module.IPRangeObject{
				Metadata:      module.ObjectMetaData{"my-ipr", "", "", ""},
				Specification: module.IPRangeObjectSpec{"192.168.1.3", 30, 40}},
			url:             BaseUrl,
			expectedErr:     true,
			expectedErrCode: 500,
		},
	}

	for _, tcase := range tcases {
		_, err := createControllerObject(tcase.url, &tcase.obj, &module.IPRangeObject{})
		handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
	}
}

func TestDeleteObject(t *testing.T) {
	tcases := []struct {
		name            string
		object_name     string
		expectedErr     bool
		expectedErrCode int
	}{
		{
			name:            "FoolName",
			object_name:     "foo_name",
			expectedErr:     true,
			expectedErrCode: 500,
		},
	}

	for _, tcase := range tcases {
		_, err := deleteControllerObject(BaseUrl, tcase.object_name)
		handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
	}
}

func TestHappyPath(t *testing.T) {
	ipr_name := "my-ipr"

	obj := module.IPRangeObject{
		Metadata:      module.ObjectMetaData{ipr_name, "", "", ""},
		Specification: module.IPRangeObjectSpec{"192.168.2.3", 10, 15}}

	ret_obj, err := createControllerObject(BaseUrl, &obj, &module.IPRangeObject{})
	if err != nil {
		printError(err)
		t.Errorf("Test Case 'Happy Path' failed: create object")
		return
	}

	if ret_obj.(*module.IPRangeObject).Specification.Subnet != "192.168.2.3" {
		t.Errorf("Test Case 'Happy Path' failed: create object")
		return
	}

	_, err = deleteControllerObject(BaseUrl, ipr_name)
	if err != nil {
		printError(err)
		t.Errorf("Test Case 'Happy Path' failed: delete object")
		return
	}
}
