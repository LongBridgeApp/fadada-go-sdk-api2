package fadada

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	client = NewClient("http://test.api.fabigbig.com:8888/api", "404068", "PqDC96A66qRN6fBQwvJaD4Yu")
)

func Test_sign(t *testing.T) {
	client1 := NewClient("http://test.api.fabigbig.com:8888/api", "000000", "fdd20141119")
	params := url.Values{}
	params.Add("account_id", "1")
	params.Add("contract_id", "testopenid")
	assert.Equal(t, "QjQ5MUQ2OUM1RTEyOTFBQkZDNTc1MkQ2Mjc4M0I1QjUwMjJGQUI1RA==", client1.sign("20191012114711", params))
}

func TestAccountRegister(t *testing.T) {
	customerID, err := client.AccountRegister("104")
	assert.NoError(t, err)
	assert.Equal(t, "59669721A0BC651ADF68CE491014345F", customerID)
}

func TestGetPersonVerifyURL(t *testing.T) {
	verifyURL, _, err := client.GetPersonVerifyURL("59669721A0BC651ADF68CE491014345F", "https://mp.longbridge.global/foo/bar")
	assert.NoError(t, err)
	// assert.Equal(t, "041b33c5014c458ba0e9aa41b982f0b2", transactionNo)
	assert.Equal(t, "", verifyURL)
}

func TestApplyCert(t *testing.T) {
	err := client.ApplyCert("59669721A0BC651ADF68CE491014345F", "73341ddab387406a87d9a79cc0dee3bc")
	assert.NoError(t, err)
}

func TestUploadDocs(t *testing.T) {
	err := client.UploadDocs("1001", "hello world", "https://cdn-support.lbkrs.com/files/202005/v5TpW6MH8rLqwvsW/Disclosure-Statement-and-Disclaimer.pdf", ".pdf")
	assert.NoError(t, err)
}

func TestGenerateSignURL(t *testing.T) {
	rawURL := client.GenerateSignURL("A1000001", "1001", "59669721A0BC651ADF68CE491014345F", "Hello world.pdf", "https://mp.longbridge.global/foo/bar")
	uri, err := url.Parse(rawURL)
	assert.NoError(t, err)
	assert.Equal(t, "test.api.fabigbig.com:8888", uri.Host)
	params := uri.Query()
	assert.Equal(t, "A1000001", params.Get("transaction_id"))
	assert.Equal(t, "1001", params.Get("contract_id"))
	assert.Equal(t, "59669721A0BC651ADF68CE491014345F", params.Get("customer_id"))
	assert.Equal(t, "https://mp.longbridge.global/foo/bar", params.Get("return_url"))
	assert.Equal(t, "Hello world.pdf", params.Get("doc_title"))
	assert.Equal(t, "", rawURL)
}
