package fadada

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// Client of Fadada
type Client struct {
	Endpoint   string
	AppID      string
	AppSecret  string
	httpClient *http.Client
}

type Response struct {
	Result  string
	Code    string
	Message string `json:"msg"`
	Data    string
}

func (res Response) IsSuccess() bool {
	return res.Code == "1000"
}

func NewClient(endpoint, appID, appSecret string) Client {
	return Client{
		Endpoint:   endpoint,
		AppID:      appID,
		AppSecret:  appSecret,
		httpClient: http.DefaultClient,
	}
}

func (c Client) URL(path string) string {
	return c.Endpoint + path
}

// AccountRegister 注册账号，用业务系统用户编号换取法大大的 customerID 客户编号
func (c Client) AccountRegister(openID string) (customerID string, err error) {
	params := url.Values{}
	params.Add("open_id", openID)
	params.Add("account_type", "1")
	json, err := c.sendRequest("POST", "/account_register.api", params)
	if err != nil {
		return
	}

	if gjson.Get(json, "code").String() != "1" {
		err = fmt.Errorf("Response not success: %s", json)
	}

	return gjson.Get(json, "data").String(), nil
}

// findPersonCertInfo 查询个人实名认证信息
func (c Client) findPersonCertInfo() {

}

// GetPersonVerifyURL 获取个人实名认证地址
// 调用这个，给一个 returnURL，让用户跳转到认证地址，完成认证后，将会带上认证信息并跳转回来
// callback?personName=李华顺&transactionNo=ab4feb43763e4a31bb5378d33b199f05&authenticationType=1&status=2&sign=NUU4MEVEQUREN0RFMDY5RTdDRDFDNkY5RDU1M0ZBNkZFNzYwQTIzNw==
func (c Client) GetPersonVerifyURL(customerID, returnURL string) (verifyURL string, transactionNo string, err error) {
	params := url.Values{}
	params.Add("customer_id", customerID)
	params.Add("verified_way", "0")
	params.Add("page_modify", "1")
	params.Add("notify_url", returnURL)
	params.Add("return_url", returnURL)

	json, err := c.sendRequest("POST", "/get_person_verify_url.api", params)
	if err != nil {
		return
	}

	if gjson.Get(json, "code").String() != "1" {
		err = fmt.Errorf("Response not success: %s", json)
	}

	verifyURL1, err := base64.StdEncoding.DecodeString(gjson.Get(json, "data.url").String())
	if err != nil {
		return
	}

	transactionNo = gjson.Get(json, "data.transactionNo").String()

	return string(verifyURL1), transactionNo, nil
}

// ApplyCert 实名证书申请, 调用此接口可以给相关账号颁发证书。
func (c Client) ApplyCert(customerID string, transactionNo string) (err error) {
	params := url.Values{}
	params.Add("customer_id", customerID)
	params.Add("verified_serialno", transactionNo)

	json, err := c.sendRequest("POST", "/apply_cert.api", params)

	if gjson.Get(json, "code").String() != "1" {
		return fmt.Errorf("Response not success, %s", json)
	}

	return nil
}

func (c Client) UploadDocs(contactID, docTitle, docURL, docType string) (err error) {
	params := url.Values{}
	params.Add("contract_id", contactID)
	params.Add("doc_title", docTitle)
	params.Add("doc_url", docURL)
	params.Add("doc_type", docType)

	json, err := c.sendRequest("POST", "/uploaddocs.api", params)
	if err != nil {
		return
	}

	if gjson.Get(json, "code").String() != "1000" {
		err = fmt.Errorf("Response not success: %s", json)
	}

	return nil
}

func (c Client) GenerateSignURL(transactionID, contractID, customerID, docTitle, returnURL string) (signURL string) {
	params := url.Values{}
	params.Add("transaction_id", transactionID)
	params.Add("contract_id", contractID)
	params.Add("customer_id", customerID)
	params.Add("doc_title", docTitle)
	params.Add("return_url", returnURL)
	params.Add("read_time", "10")
	// open_environment 打开环境 1:客户微信小程序
	params.Add("open_environment", "1")
	_, rawURL := c.newRequest("GET", "/extsign.api", params)
	return rawURL
}

func (c *Client) newRequest(method, path string, params url.Values) (req *http.Request, rawURL string) {
	t := time.Now().Format("20060102150405")

	sign := c.sign(t, params)

	params.Add("app_id", c.AppID)
	params.Add("timestamp", t)
	params.Add("v", "2.0")
	params.Add("msg_digest", sign)

	rawURL = c.URL(path)
	if method == "GET" {
		rawURL = rawURL + "?" + params.Encode()
	}

	req, _ = http.NewRequest(method, rawURL, strings.NewReader(params.Encode()))
	if req.Method != "GET" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, rawURL
}

func (c *Client) sendRequest(method, path string, params url.Values) (json string, err error) {
	req, _ := c.newRequest(method, path, params)
	fmt.Println("method", req.Method, " ", req.URL.RequestURI(), req.URL.RawQuery)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Request %s with status: %d", req.URL.String(), resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	json = string(body)
	return json, nil

	return
}

func (c *Client) sign(t string, params url.Values) string {
	sortedStr := ""

	if len(params.Get("transaction_id")) > 0 {
		t = params.Get("transaction_id") + t
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		sortedStr += params.Get(key)
	}
	signStr := c.AppSecret + sortedStr

	sha1Str := sha1Digest(signStr)

	h1 := md5.New()
	h1.Write([]byte(t))
	tDigest := fmt.Sprintf("%X", h1.Sum(nil))

	allStr := sha1Digest(c.AppID + tDigest + sha1Str)
	return base64.StdEncoding.EncodeToString([]byte(allStr))
}

func sha1Digest(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%X", h.Sum(nil))
}
