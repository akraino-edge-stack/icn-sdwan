package test

import (
    "fmt"
    "io/ioutil"
    
        "net/http"
    
        "bytes"
    "reflect"
    "testing"
    "encoding/json"
    "github.com/akraino-edge-stack/icn-sdwan/central-controller/src/scc/pkg/module"
)

type TestError struct {
    Code    int
    Message string
}

func (e *TestError) Error() string {
    return fmt.Sprintf("Error Code: %d, Error Message: %s", e.Code, e.Message)
}

// Error handler
func handleError(t *testing.T, err error, name string, expectedErr bool, errorCode int) {
    if (err != nil) {
        if (expectedErr) {
            switch err.(type) {
            case *TestError:
                if(errorCode != err.(*TestError).Code) {
                    t.Errorf("Test case '%s': expected '%d', but got '%d'", name, errorCode, err.(*TestError).Code)
                } else {
                    fmt.Printf("%s\n", err.(*TestError).Message)
                }
            default:
                t.Errorf("Test case '%s': expected openwrt.OpenwrtError, but got '%s'", name, reflect.TypeOf(err).String())
            }
        } else {
            t.Errorf("Test case '%s': expected success, but got '%s'", name, reflect.TypeOf(err).String())
        }
    } else {
        if (expectedErr) {
            t.Errorf("Test case '%s': expected error code '%d', but success", name, errorCode)
        }
    }
}

func printError(err error) {
    switch err.(type) {
    case *TestError:
        fmt.Printf("%s\n", err.(*TestError).Message)
    default:
        fmt.Printf("%s\n", reflect.TypeOf(err).String())
    }
}

func callRest(method string, url string, request string) (string, error) {
    client := &http.Client{}
    req_body := bytes.NewBuffer([]byte(request))
    req, _ := http.NewRequest(method, url, req_body)

    req.Header.Set("Cache-Control", "no-cache")
    
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode >= 400 {
        return "", &TestError{Code: resp.StatusCode, Message: string(body)}
    }

    return string(body), nil
}

func createControllerObject(baseUrl string, obj module.ControllerObject, retObj module.ControllerObject) (module.ControllerObject, error) {
    url := baseUrl
    obj_str, _ := json.Marshal(obj)

    res, err := callRest("POST", url, string(obj_str))
    if err != nil {
        return retObj, err
    }

    err = json.Unmarshal([]byte(res), retObj)
    if err != nil {
        return retObj, err
    }

    return retObj, nil
}

func getControllerObject(baseUrl string, name string, retObj module.ControllerObject) (module.ControllerObject, error) {
    url := baseUrl + "/" + name

    res, err := callRest("GET", url, "")
    if err != nil {
         return retObj, err
    }

    err = json.Unmarshal([]byte(res), retObj)
    if err != nil {
        return retObj, err
    }

    return retObj, nil
}

func updateControllerObject(baseUrl string, name string, obj module.ControllerObject, retObj module.ControllerObject) (module.ControllerObject, error) {
    url := baseUrl + "/" + name
    obj_str, _ := json.Marshal(obj)

    res, err := callRest("PUT", url, string(obj_str))
    if err != nil {
        return retObj, err
    }

    err = json.Unmarshal([]byte(res), retObj)
    if err != nil {
        return retObj, err
    }

    return retObj, nil
}

func deleteControllerObject(baseUrl string, name string) (bool, error) {
    url := baseUrl + "/" + name

    _, err := callRest("DELETE", url, "")
    if err != nil {
        printError(err)
        return false, err
    }

    _, err = callRest("GET", url, "")
    if err == nil {
         return false, &TestError{Code: 500, Message: "Filed to delete object"}
    }

    return true, nil
}