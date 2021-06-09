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

    createControllerObject(BaseUrl, &object1, &module.OverlayObject{})

    var ret = m.Run()

    deleteControllerObject(BaseUrl, "overlay1")

    os.Exit(ret)
}

func TestGetObjects(t *testing.T) {
    url := BaseUrl + "/overlay1/devices/device1/cnfs"
    res, err := callRest("GET", url, "")
    if err != nil {
        printError(err)
        t.Errorf("Test case GetObjects: can not get Objects")
        return
    }

    var objs []module.CNFObject
    err = json.Unmarshal([]byte(res), &objs)

    if len(objs) == 0 {
        fmt.Printf("Test case GetObjects: no object found")
        return
    }

    p_data, _ := json.Marshal(objs)
    fmt.Printf("%s\n", string(p_data))
}