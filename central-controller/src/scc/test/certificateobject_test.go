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
    BaseUrl = OverlayUrl + "/overlay1/" + manager.CertCollection

    var overlay_object = module.OverlayObject{
        Metadata: module.ObjectMetaData{"overlay1", "", "", ""}, 
        Specification: module.OverlayObjectSpec{}}
    
    var cert_object1 = module.CertificateObject{
        Metadata: module.ObjectMetaData{"device1", "", "", ""}}
    var cert_object2 = module.CertificateObject{
        Metadata: module.ObjectMetaData{"device2", "", "", ""}}
    
    createControllerObject(OverlayUrl, &overlay_object, &module.OverlayObject{})
    createControllerObject(BaseUrl, &cert_object1, &module.CertificateObject{})
    createControllerObject(BaseUrl, &cert_object2, &module.CertificateObject{})

    var ret = m.Run()

    deleteControllerObject(BaseUrl, "device1")
    deleteControllerObject(BaseUrl, "device2")
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

    var objs []module.CertificateObject
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
            object_name: "device1",
        },
        {
            name: "GetFoolName",
            object_name: "foo_name",
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        obj, err := getControllerObject(BaseUrl, tcase.object_name, &module.CertificateObject{})
        if err == nil {
            p_data, _ := json.Marshal(obj)
            fmt.Printf("%s\n", string(p_data))
        }
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestCreateObject(t *testing.T) {
    tcases := []struct {
        name string
        url string
        obj module.CertificateObject
        expectedErr bool
        expectedErrCode int
    }{
        {
            name: "EmptyName",
            obj: module.CertificateObject{
                Metadata: module.ObjectMetaData{"", "object 1", "", ""}},
            url: BaseUrl,
            expectedErr: true,
            expectedErrCode: 422,
        },
        {
            name: "WrongOverlayName",
            obj: module.CertificateObject{
                Metadata: module.ObjectMetaData{"device3", "", "", ""}},
            url: OverlayUrl + "/foooverlay/" + manager.CertCollection,
            expectedErr: true,
            expectedErrCode: 500,
        },
        {
            name: "DumplicateName",
            obj: module.CertificateObject{
                Metadata: module.ObjectMetaData{"device1", "", "", ""}},
            url: BaseUrl,
            expectedErr: true,
            expectedErrCode: 409,
        },
    }

    for _, tcase := range tcases {
        _, err := createControllerObject(tcase.url, &tcase.obj, &module.CertificateObject{})
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
            expectedErr: true,
            expectedErrCode: 500,
        },
    }

    for _, tcase := range tcases {
        _, err := deleteControllerObject(BaseUrl, tcase.object_name)
        handleError(t, err, tcase.name, tcase.expectedErr, tcase.expectedErrCode)
    }
}

func TestHappyPath(t *testing.T) {
    cert_name := "my-device"

    obj := module.CertificateObject{
        Metadata: module.ObjectMetaData{cert_name, "", "", ""}}

    _, err := createControllerObject(BaseUrl, &obj, &module.CertificateObject{})
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: create object")
        return
    }

    _, err = deleteControllerObject(BaseUrl, cert_name)
    if err != nil {
        printError(err)
        t.Errorf("Test Case 'Happy Path' failed: delete object")
        return
    }
}