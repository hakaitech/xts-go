package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

/*
	COMMON STRUCTS AND METHODS
*/

var XTS_baseURL string

type XTSClient struct {
	secretKey        string
	appKey           string
	sessionToken     string
	ClientID         string
	UserID           string
	isInvestorClient bool
}

type XTSOrder struct {
	ClientID             string
	OrderUID             string
	ExchangeSegment      string
	ProductType          string
	OrderType            string
	OrderSide            string
	TimeInForce          string
	AppOrderID           float64
	LimitPrice           float64
	StopPRice            float64
	DisclosedQuantity    int64
	OrderQuantity        int64
	ExchangeInstrumentID int64
}

type ResponseObject[T any] struct {
	Type   string
	Result T
}

type UserProfile struct {
	ClientName         string
	EmailID            string
	MobileNo           string
	PAN                string
	ResidentialAddress string
	ClientBankInfoList struct {
		AccountNumber   string
		AccountType     string
		BankName        string
		BankBranchName  string
		BankCity        string
		CustomerId      string
		BankCityPincode string
		BankIFSCCode    string
	}
	ClientExchangeDetailsList []struct {
		ParticipantCode   string
		ExchangeSegNumber int64
		Enabled           bool
	}
}

func decodeResponseBody[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()

	var value T

	if err := json.NewDecoder(resp.Body).Decode(&value); err != nil {
		return value, fmt.Errorf("failed decoding response body: %w", err)
	}

	return value, nil
}

// makes POST,PUT,UPDATE,DELETE http calls without adding authentication headers
func newRequestWithBodyAndContext(ctx context.Context, callUrl, method string, rawBody map[string]any) (*http.Response, error) {
	reqUrl, _ := url.JoinPath(XTS_baseURL, callUrl)
	body, err := json.Marshal(rawBody)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

// makes POST,PUT,UPDATE,DELETE http calls with authentication and Content-Type Headers
func newRequestWithBodyAndContextWithAuth(ctx context.Context, callUrl, method, token string, rawBody any) (*http.Response, error) {
	reqUrl, _ := url.JoinPath(XTS_baseURL, callUrl)
	body, err := json.Marshal(rawBody)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	addAuthHeaders(req, token)
	return client.Do(req)
}

func addAuthHeaders(req *http.Request, token string) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("authorization", token)
}

func updateOrder(originalOrder *XTSOrder, modParams *ModificationParams) error {
	return nil
}

/*
	EXPORTED FUNCTIONS
*/

/*
	GENERATOR FUNCTIONS
*/

func NewXTSClient(skey, apikey, clientID string) (*XTSClient, error) {
	if XTS_baseURL == "" {
		log.Fatal("missing base url. check config")
	}
	// If Dealer account clientID must be spoecified

	return &XTSClient{
		secretKey: skey,
		appKey:    apikey,
		ClientID:  clientID,
	}, nil
}

func (x *XTSClient) NewOrder(ctx context.Context, exch string, instID int64, productType string, orderType string, orderSide string, timeInForce string, disclosedQuantity int64, orderQty int64, limitPrice float64, stopPrice float64, orderID string) (*XTSOrder, error) {
	if !x.isInvestorClient {
		return &XTSOrder{
			ClientID:             x.ClientID,
			OrderUID:             orderID,
			ExchangeSegment:      exch,
			ExchangeInstrumentID: instID,
			ProductType:          productType,
			OrderType:            orderType,
			OrderSide:            orderSide,
			TimeInForce:          timeInForce,
			DisclosedQuantity:    disclosedQuantity,
			OrderQuantity:        orderQty,
			LimitPrice:           limitPrice,
			StopPRice:            stopPrice,
		}, nil

	} else {
		return &XTSOrder{
			OrderUID:             orderID,
			ExchangeSegment:      exch,
			ExchangeInstrumentID: instID,
			ProductType:          productType,
			OrderType:            orderType,
			OrderSide:            orderSide,
			TimeInForce:          timeInForce,
			DisclosedQuantity:    disclosedQuantity,
			OrderQuantity:        orderQty,
			LimitPrice:           limitPrice,
			StopPRice:            stopPrice,
		}, nil
	}
}

/*
	API FUNCTIONS
*/

func (x *XTSClient) SessionLogin(ctx context.Context) error {
	res, err := newRequestWithBodyAndContext(ctx, "/interactive/user/session", http.MethodPost, map[string]any{
		"secretKey": x.secretKey,
		"appKey":    x.appKey,
		"source":    "WebAPI",
	})
	if err != nil {
		return err
	}
	type responseBody struct {
		Token            string `json:"token"`
		UserID           string `json:"userID"`
		IsInvestorClient bool   `json:"isInvestorClient"`
	}
	resStruct, err := decodeResponseBody[ResponseObject[responseBody]](res)
	if err != nil {
		return err
	}
	x.sessionToken = resStruct.Result.Token
	x.isInvestorClient = resStruct.Result.IsInvestorClient
	return nil
}

func (x *XTSClient) SessionLogout(ctx context.Context) error {
	res, err := newRequestWithBodyAndContextWithAuth(ctx, "/interactive/user/session", http.MethodDelete, x.sessionToken, nil)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("api response expected 200 but got: %s", res.Status)
	}
	return nil
}

func (x *XTSClient) FetchProfile(ctx context.Context) (*UserProfile, error) {
	reqUrl, _ := url.JoinPath(XTS_baseURL, "/interactive/user/profile")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	if !x.isInvestorClient {
		q.Add("clientID", x.ClientID)
	}
	req.URL.RawQuery = q.Encode()
	addAuthHeaders(req, x.sessionToken)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("API responded with status code other than 200: %s", res.Status)
	}
	resStruct, err := decodeResponseBody[ResponseObject[UserProfile]](res)
	if err != nil {
		return nil, err
	}
	return &resStruct.Result, nil
}

func (x *XTSClient) FetchBalance(ctx context.Context) (any, error) {
	if x.isInvestorClient {
		reqUrl, _ := url.JoinPath(XTS_baseURL, "/interactive/user/profile")
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Add("clientID", x.ClientID)
		req.URL.RawQuery = q.Encode()
		addAuthHeaders(req, x.sessionToken)
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("API responded with status code other than 200: %s", res.Status)
		}
		resStruct, err := decodeResponseBody[ResponseObject[map[string]any]](res)
		if err != nil {
			return nil, err
		}
		return resStruct.Result, nil

	} else {
		return -1, fmt.Errorf("Balance API available for retail API users only, dealers can see the same on dealer terminal")
	}
}

// Returns the AppOrderID from response else the error in the response
func (x *XTSClient) PlaceOrder(ctx context.Context, order *XTSOrder) (float64, error) {
	res, err := newRequestWithBodyAndContextWithAuth(ctx, "/interactive/orders", http.MethodPost, x.sessionToken, order)
	if err != nil {
		return -1, err
	}
	if res.StatusCode != 200 {
		return -1, fmt.Errorf("API responded with status code other than 200: %s", res.Status)
	}
	type respsonseBody struct {
		AppOrderID float64
	}
	resStruct, err := decodeResponseBody[ResponseObject[respsonseBody]](res)
	if err != nil {
		return -1, err
	}
	order.AppOrderID = resStruct.Result.AppOrderID
	return resStruct.Result.AppOrderID, nil
}

type ModificationParams struct {
	ModifiedProductType       string  `json:"modifiedProductType,omitempty"`
	ModifiedOrderType         string  `json:"modifiedOrderType,omitempty"`
	ModifiedTimeInForce       string  `json:"modifiedTimeInForce,omitempty"`
	ModifiedOrderUID          string  `json:"modifiedOrderUID,omitempty"`
	AppOrderID                float64 `json:"appOrderID,omitempty"`
	ModifiedLimitPrice        float64 `json:"modifiedLimitPrice,omitempty"`
	ModifiedStopPrice         float64 `json:"modifiedStopPrice,omitempty"`
	ModifiedOrderQuantity     int64   `json:"modifiedOrderQuantity,omitempty"`
	ModifiedDisclosedQuantity int64   `json:"modifiedDisclosedQuantity,omitempty"`
}

func (params *ModificationParams) AsMap() (map[string]any, error) {
	bdy, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	if err := json.Unmarshal(bdy, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (x *XTSClient) ModifyOpenOrder(ctx context.Context, appOrderID float64, originalOrder *XTSOrder, modified *ModificationParams) (float64, error) {
	res, err := newRequestWithBodyAndContextWithAuth(ctx, "insteractive/orders", http.MethodPut, x.sessionToken, modified)
	if err != nil {
		return -1, err
	}
	type responseBody struct {
		AppOrderID float64
	}
	if res.StatusCode != 200 {
		return -1, fmt.Errorf("API responded with status code other than 200: %s", res.Status)
	}
	resStruct, err := decodeResponseBody[ResponseObject[responseBody]](res)
	if err != nil {
		return -1, err
	}
	newAppOrderID := resStruct.Result.AppOrderID
	originalOrder.AppOrderID = newAppOrderID
	newOrder := &XTSOrder{
		ClientID:             "",
		OrderUID:             "",
		ExchangeSegment:      "",
		ProductType:          "",
		OrderType:            "",
		OrderSide:            "",
		TimeInForce:          "",
		AppOrderID:           appOrderID,
		LimitPrice:           0,
		StopPRice:            0,
		DisclosedQuantity:    0,
		OrderQuantity:        0,
		ExchangeInstrumentID: 0,
	}
	originalOrder = newOrder
	return newAppOrderID, nil
}
