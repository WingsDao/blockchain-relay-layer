package clitester

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkRest "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

type RestRequest struct {
	t             *testing.T
	cdc           *codec.Codec
	httpMethod    string
	baseUrl       string
	endPointUrl   string
	urlValues     url.Values
	requestValue  interface{}
	responseValue interface{}
	timeout       time.Duration
	gas           uint64
}

// REST endpoint error object
type RestError struct {
	Error string `json:"error"`
}

// ABCI error object helper, used to unmarshal RestError.Error string
type ABCIError struct {
	Codespace string `json:"codespace"`
	Code      uint32 `json:"code"`
	Message   string `json:"message"`
}

func (r *RestRequest) SetQuery(httpMethod, subPath string, urlValues url.Values, requestValue interface{}, responseValue interface{}) *RestRequest {
	r.httpMethod = httpMethod
	r.endPointUrl = subPath
	r.urlValues = urlValues
	r.requestValue = requestValue
	r.responseValue = responseValue

	return r
}

func (r *RestRequest) ModifySubPath(targetSubStr, replaceSubStr string) *RestRequest {
	r.endPointUrl = strings.Replace(r.endPointUrl, targetSubStr, replaceSubStr, 1)

	return r
}

func (r *RestRequest) ModifyUrlValues(targetKey, newValue string) *RestRequest {
	r.urlValues.Set(targetKey, newValue)

	return r
}

func (r *RestRequest) SetTimeout(dur time.Duration) {
	r.timeout = dur
}

func (r *RestRequest) Request() (retCode int, retBody []byte) {
	u, _ := url.Parse(r.baseUrl)
	u.Path = path.Join(u.Path, r.endPointUrl)
	if r.urlValues != nil {
		u.RawQuery = r.urlValues.Encode()
	}

	_, err := url.Parse(u.String())
	require.NoError(r.t, err, "%s: ParseRequestURI: %s", r.String(), u.String())

	var reqBodyBytes []byte
	if r.requestValue != nil {
		var err error
		reqBodyBytes, err = r.cdc.MarshalJSON(r.requestValue)
		require.NoError(r.t, err, "%s: marshal requestValue", r.String())
	}

	req, err := http.NewRequest(r.httpMethod, u.String(), bytes.NewBuffer(reqBodyBytes))
	require.NoError(r.t, err, "%s: NewRequest", r.String())
	req.Header.Set("Content-type", "application/json")

	client := http.Client{}
	if r.timeout > 0 {
		client.Timeout = r.timeout
	}

	resp, err := client.Do(req)
	require.NoError(r.t, err, "%s: HTTP request", r.String())
	require.NotNil(r.t, resp, "%s: HTTP response", r.String())
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(r.t, err, "%s: HTTP response body read", r.String())

	retCode, retBody = resp.StatusCode, bodyBytes

	return
}

func (r *RestRequest) Execute() error {
	respCode, respBody := r.Request()
	if respCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP code %d: %s", r.String(), respCode, string(respBody))
	}

	// parse Tx response or Query response
	if r.responseValue != nil {
		if _, ok := r.responseValue.(*sdk.TxResponse); ok {
			if err := r.cdc.UnmarshalJSON(respBody, r.responseValue); err != nil {
				return fmt.Errorf("%s: unmarshal txResponseValue: %s", r.String(), string(respBody))
			}

			txResp := r.responseValue.(*sdk.TxResponse)
			if txResp.Code != 0 {
				return fmt.Errorf("%s: tx code %d: %s", r.String(), txResp.Code, txResp.RawLog)
			}

			return nil
		}

		if _, ok := r.responseValue.(*tmCoreTypes.ResultBlock); ok {
			if err := r.cdc.UnmarshalJSON(respBody, r.responseValue); err != nil {
				return fmt.Errorf("%s: unmarshal tmCoreTypes.ResultBlock: %s", r.String(), string(respBody))
			}

			return nil
		}

		if _, ok := r.responseValue.(*auth.StdTx); ok {
			respMsg := struct {
				Type  string     `json:"type"`
				Value auth.StdTx `json:"value"`
			}{}

			if err := r.cdc.UnmarshalJSON(respBody, &respMsg); err != nil {
				return fmt.Errorf("%s: unmarshal coreTypes.ResultBlock: %s", r.String(), string(respBody))
			}

			*r.responseValue.(*auth.StdTx) = respMsg.Value
			return nil
		}

		respMsg := sdkRest.ResponseWithHeight{}
		if err := r.cdc.UnmarshalJSON(respBody, &respMsg); err != nil {
			return fmt.Errorf("%s: unmarshal ResponseWithHeight: %s", r.String(), string(respBody))
		}

		if respMsg.Result != nil {
			if err := r.cdc.UnmarshalJSON(respMsg.Result, r.responseValue); err != nil {
				return fmt.Errorf("%s: unmarshal responseValue: %s", r.String(), string(respBody))
			}
		}
	}

	return nil
}

func (r *RestRequest) CheckSucceeded() {
	require.NoError(r.t, r.Execute())
}

func (r *RestRequest) CheckFailed(expectedCode int, expectedErr error) {
	respCode, respBody := r.Request()
	bodyStr := string(respBody)
	require.Equal(r.t, expectedCode, respCode, "%s: HTTP code", r.String())

	if expectedErr != nil {
		expectedSdkErr, ok := expectedErr.(*sdkErrors.Error)
		require.True(r.t, ok, "not a SDK error")

		require.NotNil(r.t, respBody, "%s: respBody", r.String())

		restErr, abciErr := &RestError{}, &ABCIError{}
		require.NoError(r.t, r.cdc.UnmarshalJSON(respBody, restErr), "%s: unmarshal RestError: %s", r.String(), string(respBody))

		if err := r.cdc.UnmarshalJSON([]byte(restErr.Error), abciErr); err == nil {
			require.Equal(r.t, expectedSdkErr.Codespace(), abciErr.Codespace, "%s: err codespace: %s", r.String(), string(respBody))
			require.Equal(r.t, expectedSdkErr.ABCICode(), abciErr.Code, "%s: err code: %s", r.String(), string(respBody))
			return
		} else if strings.Contains(bodyStr, expectedErr.Error()) {
			return
		}

		r.t.Fatalf("%s: ABCIError unmarshal failed and %q error not found: %s", r.String(), expectedErr.Error(), bodyStr)
	}
}

func (r *RestRequest) String() string {
	return fmt.Sprintf("%s %s", r.httpMethod, r.endPointUrl)
}
