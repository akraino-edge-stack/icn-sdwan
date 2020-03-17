package openwrt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
)

type OpenwrtError struct {
	Code    int
	Message string
}

func (e *OpenwrtError) Error() string {
	return fmt.Sprintf("Error Code: %d, Error Message: %s", e.Code, e.Message)
}

type openwrtClient struct {
	ip       string
	user     string
	password string
	token    string
}

func CloseClient(o *openwrtClient) {
	o.logout()
	runtime.SetFinalizer(o, nil)
}

func NewOpenwrtClient(ip string, user string, password string) *openwrtClient {
	client := &openwrtClient{
		ip:       ip,
		user:     user,
		password: password,
		token:    "",
	}

	runtime.SetFinalizer(client, CloseClient)
	return client
}

// openwrt base URL
func (o *openwrtClient) getBaseURL() string {
	return "http://" + o.ip + "/cgi-bin/luci/"
}

// login to openwrt http server
func (o *openwrtClient) login() error {
	client := &http.Client{
		// block redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// login
	login_info := "luci_username=" + o.user + "&luci_password=" + o.password
	var req_body = []byte(login_info)
	req, _ := http.NewRequest("POST", o.getBaseURL(), bytes.NewBuffer(req_body))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	} else if resp.StatusCode != 302 {
		// fail to auth
		return &OpenwrtError{Code: resp.StatusCode, Message: "Unauthorized"}
	} else {
		// get token
		res_cookie := resp.Header["Set-Cookie"][0]
		res_cookies := strings.Split(res_cookie, ";")
		for _, cookie := range res_cookies {
			cookie := strings.TrimSpace(cookie)
			index := strings.Index(cookie, "=")
			var key = cookie
			var value = ""
			if index != -1 {
				key = cookie[:index]
				value = cookie[index+1:]
			}

			if key == "sysauth" {
				o.token = value
				break
			}
		}
	}

	return nil
}

// logout to openwrt http server
func (o *openwrtClient) logout() error {
	if o.token != "" {
		_, err:= o.Get("admin/logout")
		o.token = ""
		return err
	}

	return nil
}

// call openwrt restful API
func (o *openwrtClient) call(method string, url string, request string) (string, error) {
	for i := 0; i < 2; i++ {
		if o.token == "" {
			err := o.login()
			if err != nil {
				return "", err
			}
		}

		client := &http.Client{}
		req_body := bytes.NewBuffer([]byte(request))
		req, _ := http.NewRequest(method, o.getBaseURL()+url, req_body)
		req.Header.Add("Cookie", "sysauth="+o.token)
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			if resp.StatusCode == 403 {
				// token expired, retry
				o.token = ""
				continue
			} else {
				// error request
				return "", &OpenwrtError{Code: resp.StatusCode, Message: string(body)}
			}
		}

		return string(body), nil
	}

	return "", nil
}

// call openwrt Get restful API
func (o *openwrtClient) Get(url string) (string, error) {
	return o.call("GET", url, "")
}

// call openwrt restful API
func (o *openwrtClient) Post(url string, request string) (string, error) {
	return o.call("POST", url, request)
}

// call openwrt restful API
func (o *openwrtClient) Put(url string, request string) (string, error) {
	return o.call("PUT", url, request)
}

// call openwrt restful API
func (o *openwrtClient) Delete(url string) (string, error) {
	return o.call("DELETE", url, "")
}
