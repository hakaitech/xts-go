package types

// Common response structure for XTS APIs
type XTSResponse[T any] struct {
	Type   string `json:"type,omitempty"`
	Result T      `json:"result,omitempty"`
}

// Authentication related types
type LoginRequest struct {
	SecretKey string `json:"secretKey"`
	AppKey    string `json:"appKey"`
	Source    string `json:"source"`
}

type LoginResponse struct {
	Token            string `json:"token,omitempty"`
	UserID           string `json:"userID,omitempty"`
	IsInvestorClient bool   `json:"isInvestorClient,omitempty"`
}

// Order related types
type Order struct {
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

type OrderResponse struct {
	AppOrderID float64 `json:"appOrderID,omitempty"`
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

// User Profile types
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

// Market Data types
type Instrument struct {
	ExchangeSegment      int `json:"exchangeSegment"`
	ExchangeInstrumentID int `json:"exchangeInstrumentID"`
}

type QuoteRequest struct {
	Instruments    []Instrument `json:"instruments"`
	XTSMessageCode int          `json:"xtsMessageCode"`
	PublishFormat  string       `json:"publishFormat,omitempty"`
}

type SubscriptionRequest struct {
	Instruments    []Instrument `json:"instruments"`
	XTSMessageCode int          `json:"xtsMessageCode"`
}

// Market Data Response types
type MarketQuote struct {
	ExchangeSegment      int     `json:"exchangeSegment"`
	ExchangeInstrumentID int     `json:"exchangeInstrumentID"`
	LastTradedPrice      float64 `json:"lastTradedPrice"`
	LastTradedQuantity   int     `json:"lastTradedQuantity"`
	AverageTradePrice    float64 `json:"averageTradePrice"`
	VolumeTraded         int64   `json:"volumeTraded"`
	TotalBuyQuantity     int64   `json:"totalBuyQuantity"`
	TotalSellQuantity    int64   `json:"totalSellQuantity"`
	TotalTrades          int     `json:"totalTrades"`
	Open                 float64 `json:"open"`
	High                 float64 `json:"high"`
	Low                  float64 `json:"low"`
	Close                float64 `json:"close"`
	LastUpdateTime       int64   `json:"lastUpdateTime"`
	ExchangeTimeStamp    int64   `json:"exchangeTimeStamp"`
}

type TouchlineData struct {
	ExchangeSegment      int     `json:"exchangeSegment"`
	ExchangeInstrumentID int     `json:"exchangeInstrumentID"`
	LastTradedPrice      float64 `json:"lastTradedPrice"`
	LastTradedTime       int64   `json:"lastTradedTime"`
	PercentChange        float64 `json:"percentChange"`
	LastTradedQuantity   int     `json:"lastTradedQuantity"`
	VolumeTraded         int64   `json:"volumeTraded"`
	BestBuyPrice         float64 `json:"bestBuyPrice"`
	BestSellPrice        float64 `json:"bestSellPrice"`
	TotalTrades          int     `json:"totalTrades"`
	Open                 float64 `json:"open"`
	High                 float64 `json:"high"`
	Low                  float64 `json:"low"`
	Close                float64 `json:"close"`
}

// Constants for XTS API
const (
	// Exchange Segments
	ExchangeNSECM = "NSECM"
	ExchangeNSEFO = "NSEFO"
	ExchangeBSECM = "BSECM"
	ExchangeBSEFO = "BSEFO"
	ExchangeMCXSX = "MCXSX"

	// Order Types
	OrderTypeMarket = "MARKET"
	OrderTypeLimit  = "LIMIT"
	OrderTypeStopLimit = "STOPLIMIT"
	OrderTypeStopMarket = "STOPMARKET"

	// Order Sides
	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"

	// Product Types
	ProductTypeNRML = "NRML"
	ProductTypeMIS  = "MIS"
	ProductTypeCNC  = "CNC"
	ProductTypeCO   = "CO"
	ProductTypeBO   = "BO"

	// Time in Force
	TimeInForceDAY = "DAY"
	TimeInForceIOC = "IOC"

	// Message Codes for Market Data
	MessageCodeTouchline    = 1501
	MessageCodeMarketDepth  = 1502
	MessageCodeIndexData    = 1504
	MessageCodeCandleData   = 1505
	MessageCodeOpenInterest = 1510

	// Source
	SourceWebAPI = "WEBAPI"
)

// Error types
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Code + ": " + e.Message + " (" + e.Detail + ")"
	}
	return e.Code + ": " + e.Message
}