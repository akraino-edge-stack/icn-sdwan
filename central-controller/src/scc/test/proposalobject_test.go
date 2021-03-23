package test

import (
    "testing"
    "flag"
    "encoding/json"
    "fmt"
    "os"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/manager"
)

var BaseUrl string
var OverlayUrl string

func TestMain(m *testing.M) {
    servIp := flag.String("ip", "127.0.0.1", "SDEWAN Central Controller IP Address")
    flag.Parse()
    OverlayUrl = "http://" + *servIp + ":9015/scc/v1/" + manager.OverlayCollection
    IPBaseUrl := OverlayUrl + "/overlay1/" + manager.IPRangeCollection
    BaseUrl = OverlayUrl + "/overlay1/" + manager.ProposalCollection

    var overlay_object = module.OverlayObject{
        Metadata: module.ObjectMetaData{"overlay1", "", "", ""}, 
        Specification: module.OverlayObjectSpec{}}
    
    var iprange_object1 = module.IPRangeObject{
        Metadata: module.ObjectMetaData{"ipr1", "", "", ""}, 
        Specification: module.IPRangeObjectSpec{"192.168.0.2", 10, 20}}
    var iprange_object2 = module.IPRangeObject{
        Metadata: module.ObjectMetaData{"ipr2", "", "", ""}, 
        Specification: module.IPRangeObjectSpec{"192.168.2.2", 18, 20}}

    var proposal_object1 = module.ProposalObject{
        Metadata: module.ObjectMetaData{"proposal1", "", "", ""}, 
        Specification: module.ProposalObjectSpec{"aes256", "sha256", "modp4096"}}
    var proposal_object2 = module.ProposalObject{
        Metadata: module.ObjectMetaData{"proposal2", "", "", ""}, 
        Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}}
    
    createControllerObject(OverlayUrl, &overlay_object, &module.OverlayObject{})
    createControllerObject(IPBaseUrl, &iprange_object1, &module.IPRangeObject{})
    createControllerObject(IPBaseUrl, &iprange_object2, &module.IPRangeObject{})
    createControllerObject(BaseUrl, &proposal_object1, &module.ProposalObject{})
    createControllerObject(BaseUrl, &proposal_object2, &module.ProposalObject{})

    var ret = m.Run()

    deleteControllerObject(BaseUrl, "proposal1")
    deleteControllerObject(BaseUrl, "proposal2")
    deleteControllerObject(IPBaseUrl, "ipr1")
    deleteControllerObject(IPBaseUrl, "ipr2")
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

    var objs []module.ProposalObject
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
        name string
        object_name string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "Normal",
            object_name: "proposal1",
        },
        {
            name: "GetFoolName",
            object_name: "foo_name",
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        _, err := getControllerObject(BaseUrl, tcase.object_name, &module.ProposalObject{})
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateObject(t *testing.T) {
    tcases := []struct {
        name string
        url string
        obj module.ProposalObject
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            obj: module.ProposalObject{
                Metadata: module.ObjectMetaData{"", "object 1", "", ""}, 
                Specification: module.ProposalObjectSpec{}},
            url: BaseUrl,
            expectedErr: true,
            expectedErrCode: 422,
        },
        {
            name: "WrongOverlayName",
            obj: module.ProposalObject{
                Metadata: module.ObjectMetaData{"proposal1", "", "", ""}, 
                Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}},
            url: OverlayUrl + "/foooverlay/" + manager.ProposalCollection,
            expectedErr: true,
            expectedErrCode: 500,
        },
        {
            name: "DumplicateName",
            obj: module.ProposalObject{
                Metadata: module.ObjectMetaData{"proposal1", "", "", ""}, 
                Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}},
            url: BaseUrl,
            expectedErr: true,
            expectedErrCode: 409,
        },
    }

    for _, tcase := range tcases {
        _, err := createControllerObject(tcase.url, &tcase.obj, &module.ProposalObject{})
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateObject(t *testing.T) {
    tcases := []struct {
        name string
        object_name string
        obj module.ProposalObject
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            object_name: "proposal1",
            obj: module.ProposalObject{
                Metadata: module.ObjectMetaData{"", "object 1", "", ""}, 
                Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}},
            expectedErr: true,
            expectedErrCode: 422,
        },
        {
            name: "MisMatchName",
            object_name: "proposal2",
            obj: module.ProposalObject{
                Metadata: module.ObjectMetaData{"proposal1", "", "", ""}, 
                Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}},
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        _, err := updateControllerObject(BaseUrl, tcase.object_name, &tcase.obj, &module.ProposalObject{})
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestDeleteObject(t *testing.T) {
    tcases := []struct {
        name string
        object_name string
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "FoolName",
            object_name: "foo_name",
        },
    }

    for _, tcase := range tcases {
        _, err := deleteControllerObject(BaseUrl, tcase.object_name)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestHappyPath(t *testing.T) {
    proposal_name := "my-proposal"

    obj := module.ProposalObject{
        Metadata: module.ObjectMetaData{proposal_name, "object 1", "", ""},
        Specification: module.ProposalObjectSpec{"aes256", "sha256", "modp4096"}}

    obj_update := module.ProposalObject{
        Metadata: module.ObjectMetaData{proposal_name, "object 1", "", ""},
        Specification: module.ProposalObjectSpec{"aes512", "sha512", "modp4096"}}

    ret_obj, err := createControllerObject(BaseUrl, &obj, &module.ProposalObject{})
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: create object")
        return
    }

    if ret_obj.(*module.ProposalObject).Specification.Encryption != "aes256" {
        t.Errorf("Test Case 'Happy Path' failed: create object")
        return
    }

    ret_obj, err = updateControllerObject(BaseUrl, proposal_name, &obj_update, &module.ProposalObject{})
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: update object")
        return
    }

    if ret_obj.(*module.ProposalObject).Specification.Encryption != "aes512" {
        t.Errorf("Test Case 'Happy Path' failed: update object")
        return
    }

    _, err = deleteControllerObject(BaseUrl, proposal_name)
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: delete object")
        return
    }
}