package main

import (
    "io/ioutil"
    "flag"
    "encoding/base64"
    "encoding/json"
    "log"
    "bytes"
    "net/http"
    "errors"
    "os"
    "strings"
    "fmt"
)

const (
    overlayCollection="overlays"
    certificateCollection="certificates"
)

type CertData struct {
	RootCA string
	Ca     string
	Key    string
}

func main(){
    servIp := flag.String("ip", "127.0.0.1", "SDEWAN Central Controller IP Address")
    servPort := flag.String("port", "9015", "SDEWAN Central Controller Port Number")
    overlayName := flag.String("overlay", "", "Overlay the cert belongs to")
    certName := flag.String("certName", "", "Certificate name to query")
    flag.Parse()

    overlayUrl := "http://" + *servIp + ":" + *servPort + "/scc/v1/" + overlayCollection
    certUrl := overlayUrl + "/" + *overlayName + "/" + certificateCollection
    certObj := `{"metadata":{"name":"` + *certName + `","description":"object 1","userData1":"","userData2":""},"spec":{},"data":{"rootca":"","ca":"","key":""}}`

    _, _, err := getObject(overlayUrl, *overlayName)
    if err != nil {
	log.Println("Fetch overlay", *overlayName,"with error", err)
	log.Println("Please re-create the overlay")
	os.Exit(0)
    }

    log.Println("Fetch certificate with name", *certName, "within overlay", *overlayName)
    res, resp_code, err := getObject(certUrl, *certName)
    if err != nil {
	if resp_code != 500 {
	    log.Println("Fetch certificate with name", *certName, "with error", err)
            os.Exit(0)
	}

	log.Println("Certificate not found. Creating certificate with name", *certName)
	res, err = createObject(certUrl, certObj)
	if err != nil {
            log.Println("Certificate creation failed with error", err)
	    os.Exit(0)
	}
	parseCertBundle(res)
	os.Exit(0)
    }
    parseCertBundle(res)

}

func callRest(method string, url string, request string) (string, int, error) {
    client := &http.Client{}
    req_body := bytes.NewBuffer([]byte(request))
    req, _ := http.NewRequest(method, url, req_body)

    req.Header.Set("Cache-Control", "no-cache")

    resp, err := client.Do(req)
    if err != nil {
        return "", 0, err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    if resp.StatusCode >= 400 {
        return "", resp.StatusCode, errors.New(string(body))
    }

    return string(body), resp.StatusCode, nil
}

func createObject(baseUrl string, obj_str string) (string, error) {
    url := baseUrl

    res, _, err := callRest("POST", url, obj_str)
    if err != nil {
        return "", err
    }

    return res, nil
}

func getObject(baseUrl string, name string) (string, int, error) {
    url := baseUrl + "/" + name

    res, resp_code, err := callRest("GET", url, "")
    if err != nil {
         return "", resp_code, err
    }

    return res, resp_code, nil
}

func parseCertBundle(val string) (string, string, string, error) {
    var vi interface{}
    err := json.Unmarshal([]byte(val), &vi)
    if err != nil {
	log.Println("Error in formatting cert data", err)
	return "", "", "", err
    }

    data_interface := vi.(map[string]interface{})["data"]
    data, err := json.Marshal(data_interface)
    if err != nil {
	log.Println("Error in formatting cert data", err)
        return "", "", "", err
    }

    var certData CertData
    err = json.Unmarshal(data, &certData)
    if err != nil {
        log.Println("Error in formatting cert data", err)
        return "", "", "", err
    }

    ca_decoded, err := base64.StdEncoding.DecodeString(certData.RootCA)
    if err != nil {
	log.Println("Error in formatting cert rootca data", err)
	return "", "", "", err
    }

    caChain := strings.SplitAfter(string(ca_decoded), "-----END CERTIFICATE-----")

    ca_encoded := base64.StdEncoding.EncodeToString([]byte(caChain[1]))
    fmt.Println("Input for shared_ca: ")
    fmt.Println(ca_encoded)
    fmt.Println("Input for local_public_cert: ")
    fmt.Println(certData.Ca)
    fmt.Println("Input for local_private_cert: ")
    fmt.Println(certData.Key)

    return certData.Ca, certData.Key, ca_encoded, nil
}
