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

func TestMain(m *testing.M) {
    servIp := flag.String("ip", "127.0.0.1", "SDEWAN Central Controller IP Address")
    flag.Parse()
    BaseUrl = "http://" + *servIp + ":9015/scc/v1/" + manager.OverlayCollection

    var object1 = module.OverlayObject{
        Metadata: module.ObjectMetaData{"overlay1", "", "", ""}, 
        Specification: module.OverlayObjectSpec{}}
    var object2 = module.OverlayObject{
        Metadata: module.ObjectMetaData{"overlay2", "", "", ""}, 
        Specification: module.OverlayObjectSpec{}}

    createControllerObject(BaseUrl, &object1, &module.OverlayObject{})
    createControllerObject(BaseUrl, &object2, &module.OverlayObject{})

    var ret = m.Run()

    deleteControllerObject(BaseUrl, "overlay1")
    deleteControllerObject(BaseUrl, "overlay2")

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

    var objs []module.OverlayObject
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
            object_name: "overlay1",
        },
        {
            name: "GetFoolName",
            object_name: "foo_name",
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        _, err := getControllerObject(BaseUrl, tcase.object_name, &module.OverlayObject{})
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateObject(t *testing.T) {
    tcases := []struct {
        name string
        obj module.OverlayObject
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            obj: module.OverlayObject{
                Metadata: module.ObjectMetaData{"", "object 1", "", ""}, 
                Specification: module.OverlayObjectSpec{}},
            expectedErr: true,
            expectedErrCode: 422,
        },
        {
            name: "DumplicateName",
            obj: module.OverlayObject{
                Metadata: module.ObjectMetaData{"overlay1", "", "", ""}, 
                Specification: module.OverlayObjectSpec{}},
            expectedErr: true,
            expectedErrCode: 409,
        },
    }

    for _, tcase := range tcases {
        _, err := createControllerObject(BaseUrl, &tcase.obj, &module.OverlayObject{})
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestUpdateObject(t *testing.T) {
    tcases := []struct {
        name string
        object_name string
        obj module.OverlayObject
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            object_name: "overlay1",
            obj: module.OverlayObject{
                Metadata: module.ObjectMetaData{"", "object 1", "", ""}, 
                Specification: module.OverlayObjectSpec{}},
            expectedErr: true,
            expectedErrCode: 422,
        },
        {
            name: "MisMatchName",
            object_name: "overlay2",
            obj: module.OverlayObject{
                Metadata: module.ObjectMetaData{"overlay1", "", "", ""}, 
                Specification: module.OverlayObjectSpec{}},
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        _, err := updateControllerObject(BaseUrl, tcase.object_name, &tcase.obj, &module.OverlayObject{})
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
    overlay_name := "my-overlay"

    obj := module.OverlayObject{
        Metadata: module.ObjectMetaData{overlay_name, "object 1", "", ""},
        Specification: module.OverlayObjectSpec{}}

    obj_update := module.OverlayObject{
        Metadata: module.ObjectMetaData{overlay_name, "object 2", "", ""},
        Specification: module.OverlayObjectSpec{}}

    ret_obj, err := createControllerObject(BaseUrl, &obj, &module.OverlayObject{})
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: create object")
        return
    }

    if ret_obj.(*module.OverlayObject).Metadata.Description != "object 1" {
        t.Errorf("Test Case 'Happy Path' failed: create object")
        return
    }

    ret_obj, err = updateControllerObject(BaseUrl, overlay_name, &obj_update, &module.OverlayObject{})
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: update object")
        return
    }

    if ret_obj.(*module.OverlayObject).Metadata.Description != "object 2" {
        t.Errorf("Test Case 'Happy Path' failed: update object")
        return
    }

    _, err = deleteControllerObject(BaseUrl, overlay_name)
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: delete object")
        return
    }
}
