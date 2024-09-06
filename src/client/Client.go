package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

/*
	COMMON STRUCTS AND METHODS
*/

var (
	XTS_baseURL    string
	ErrOrderFailed error = fmt.Errorf("order failed")
)

type XTSClient struct {
	secretKey        string
	appKey           string
	sessionToken     string
	ClientID         string
	UserID           string
	isInvestorClient bool
}

type XTSOrder struct {
	ClientID             string  `json:"clientID,omitempty"`
	OrderUID             string  `json:"orderUID,omitempty"`
	ExchangeSegment      string  `json:"exchangeSegment,omitempty"`
	ProductType          string  `json:"productType,omitempty"`
	OrderType            string  `json:"orderType,omitempty"`
	OrderSide            string  `json:"orderSide,omitempty"`
	TimeInForce          string  `json:"timeInForce,omitempty"`
	AppOrderID           float64 `json:"appOrderID,omitempty"`
	LimitPrice           float64 `json:"limitPrice,omitempty"`
	StopPrice            float64 `json:"stopPrice,omitempty"`
	DisclosedQuantity    int64   `json:"disclosedQuantity,omitempty"`
	OrderQuantity        int64   `json:"orderQuantity,omitempty"`
	ExchangeInstrumentID int64   `json:"exchangeInstrumentID,omitempty"`
}

type ResponseObject[T any] struct {
	Type   string `json:"type,omitempty"`
	Result T      `json:"result,omitempty"`
}

type UserProfile struct {
	ClientName         string `json:"clientName,omitempty"`
	EmailID            string `json:"emailID,omitempty"`
	MobileNo           string `json:"mobileNo,omitempty"`
	PAN                string `json:"pan,omitempty"`
	ResidentialAddress string `json:"residentialAddress,omitempty"`
	ClientBankInfoList struct {
		AccountNumber   string `json:"accountNumber,omitempty"`
		AccountType     string `json:"accountType,omitempty"`
		BankName        string `json:"bankName,omitempty"`
		BankBranchName  string `json:"bankBranchName,omitempty"`
		BankCity        string `json:"bankCity,omitempty"`
		CustomerId      string `json:"customerId,omitempty"`
		BankCityPincode string `json:"bankCityPincode,omitempty"`
		BankIFSCCode    string `json:"bankIFSCCode,omitempty"`
	} `json:"clientBankInfoList,omitempty"`
	ClientExchangeDetailsList []struct {
		ParticipantCode   string `json:"participantCode,omitempty"`
		ExchangeSegNumber int64  `json:"exchangeSegNumber,omitempty"`
		Enabled           bool   `json:"enabled,omitempty"`
	} `json:"clientExchangeDetailsList,omitempty"`
}

func AsMap[T any](in T) (map[string]any, error) {
	inrec, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	if err := json.Unmarshal(inrec, &res); err != nil {
		return nil, err
	}
	return res, nil
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
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, bytes.NewBuffer(body))
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
	req, err := http.NewRequestWithContext(ctx, method, reqUrl, bytes.NewBuffer(body))
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
			StopPrice:            stopPrice,
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
			StopPrice:            stopPrice,
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
		Token            string `json:"token,omitempty"`
		UserID           string `json:"userID,omitempty"`
		IsInvestorClient bool   `json:"isInvestorClient,omitempty"`
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
		return -1, ErrOrderFailed
	}
	type respsonseBody struct {
		AppOrderID float64 `json:"appOrderID,omitempty"`
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

func (x *XTSClient) ModifyOpenOrder(ctx context.Context, originalOrder *XTSOrder, modified *ModificationParams) (*XTSOrder, error) {
	res, err := newRequestWithBodyAndContextWithAuth(ctx, "insteractive/orders", http.MethodPut, x.sessionToken, modified)
	if err != nil {
		return nil, err
	}
	type responseBody struct {
		AppOrderID float64 `json:"appOrderID,omitempty"`
	}
	if res.StatusCode != 200 {
		return nil, ErrOrderFailed
	}
	resStruct, err := decodeResponseBody[ResponseObject[responseBody]](res)
	if err != nil {
		return nil, err
	}
	newAppOrderID := resStruct.Result.AppOrderID
	newOrder := originalOrder
	newOrder.AppOrderID = newAppOrderID
	modMap, err := AsMap(modified)
	for field, Value := range modMap {
		switch field {
		case "modifiedProductType":
			newOrder.ProductType = Value.(string)
		case "modifiedOrderType":
			newOrder.OrderType = Value.(string)
		case "modifiedTimeInForce":
			newOrder.TimeInForce = Value.(string)
		case "modifiedOrderUID":
			newOrder.OrderUID = Value.(string)
		case "modifiedLimitPrice":
			newOrder.LimitPrice = Value.(float64)
		case "modifiedStopPrice":
			newOrder.StopPrice = Value.(float64)
		case "modifiedOrderQuantity":
			newOrder.OrderQuantity = int64(Value.(float64))
		case "modifiedDisclosedQuantity":
			newOrder.DisclosedQuantity = int64(Value.(float64))
		default:
			return nil, fmt.Errorf("invalid field modifier")
		}
	}
	return newOrder, nil
}

func (x *XTSClient) CancelOpenOrder(ctx context.Context, order *XTSOrder) error {
	req, err := http.NewRequest(http.MethodDelete, "/interactive/orders", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("appOrderID", fmt.Sprintf("%d", int64(order.AppOrderID)))
	if !x.isInvestorClient {
		q.Add("ClientID", x.ClientID)
	}
	req.URL.RawQuery = q.Encode()
	addAuthHeaders(req, x.sessionToken)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return ErrOrderFailed
	}
	return nil
}

func (x *XTSClient) CancelAllOrders(ctx context.Context, xSeg string, xInstID int64) error {
	body := map[string]any{
		"exchangeSegment":      xSeg,
		"exchangeInstrumentID": xInstID,
	}
	res, err := newRequestWithBodyAndContextWithAuth(ctx, "/interactive/orders/cancelall", http.MethodPost, x.sessionToken, body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return ErrOrderFailed
	}
	return nil
}

func (x *XTSClient) PlaceBracketOrder(ctx context.Context) error {
	return nil
}
