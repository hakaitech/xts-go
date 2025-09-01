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

// Additional Response Types for API Operations

// APIResponse represents a generic API response
type APIResponse struct {
	Type   string      `json:"type,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Code   string      `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// ProfileResponse represents user profile API response
type ProfileResponse struct {
	Type   string      `json:"type,omitempty"`
	Result UserProfile `json:"result,omitempty"`
}

// BalanceResponse represents account balance API response
type BalanceResponse struct {
	Type   string        `json:"type,omitempty"`
	Result AccountBalance `json:"result,omitempty"`
}

// AccountBalance represents account balance information
type AccountBalance struct {
	AvailableMargin    float64 `json:"availableMargin,omitempty"`
	UsedMargin        float64 `json:"usedMargin,omitempty"`
	TotalMargin       float64 `json:"totalMargin,omitempty"`
	AvailableCash     float64 `json:"availableCash,omitempty"`
	UnutilizedAmount  float64 `json:"unutilizedAmount,omitempty"`
}

// OrderBookResponse represents order book API response
type OrderBookResponse struct {
	Type   string         `json:"type,omitempty"`
	Result []OrderDetails `json:"result,omitempty"`
}

// OrderDetails represents order information
type OrderDetails struct {
	ClientID             string  `json:"clientID,omitempty"`
	AppOrderID           float64 `json:"appOrderID,omitempty"`
	OrderID              string  `json:"orderID,omitempty"`
	ExchangeInstrumentID int64   `json:"exchangeInstrumentID,omitempty"`
	ExchangeSegment      string  `json:"exchangeSegment,omitempty"`
	ProductType          string  `json:"productType,omitempty"`
	OrderType            string  `json:"orderType,omitempty"`
	OrderSide            string  `json:"orderSide,omitempty"`
	TimeInForce          string  `json:"timeInForce,omitempty"`
	OrderQuantity        int64   `json:"orderQuantity,omitempty"`
	LimitPrice           float64 `json:"limitPrice,omitempty"`
	StopPrice            float64 `json:"stopPrice,omitempty"`
	OrderStatus          string  `json:"orderStatus,omitempty"`
	OrderTime            string  `json:"orderTime,omitempty"`
}

// TradeBookResponse represents trade book API response
type TradeBookResponse struct {
	Type   string         `json:"type,omitempty"`
	Result []TradeDetails `json:"result,omitempty"`
}

// TradeDetails represents trade information
type TradeDetails struct {
	ClientID             string  `json:"clientID,omitempty"`
	AppOrderID           float64 `json:"appOrderID,omitempty"`
	OrderID              string  `json:"orderID,omitempty"`
	ExchangeInstrumentID int64   `json:"exchangeInstrumentID,omitempty"`
	ExchangeSegment      string  `json:"exchangeSegment,omitempty"`
	ProductType          string  `json:"productType,omitempty"`
	OrderSide            string  `json:"orderSide,omitempty"`
	OrderType            string  `json:"orderType,omitempty"`
	TradedQuantity       int64   `json:"tradedQuantity,omitempty"`
	TradedPrice          float64 `json:"tradedPrice,omitempty"`
	TradeTime            string  `json:"tradeTime,omitempty"`
	TradeID              string  `json:"tradeID,omitempty"`
}

// PositionBookResponse represents position book API response
type PositionBookResponse struct {
	Type   string     `json:"type,omitempty"`
	Result []Position `json:"result,omitempty"`
}

// Position represents position information
type Position struct {
	ClientID              string  `json:"clientID,omitempty"`
	ExchangeInstrumentID  int64   `json:"exchangeInstrumentID,omitempty"`
	ExchangeSegment       string  `json:"exchangeSegment,omitempty"`
	ProductType           string  `json:"productType,omitempty"`
	NetQuantity           int64   `json:"netQuantity,omitempty"`
	BuyAveragePrice       float64 `json:"buyAveragePrice,omitempty"`
	SellAveragePrice      float64 `json:"sellAveragePrice,omitempty"`
	UnrealizedMTM         float64 `json:"unrealizedMTM,omitempty"`
	RealizedMTM           float64 `json:"realizedMTM,omitempty"`
	BuyQuantity           int64   `json:"buyQuantity,omitempty"`
	SellQuantity          int64   `json:"sellQuantity,omitempty"`
}

// HoldingResponse represents holdings API response
type HoldingResponse struct {
	Type   string    `json:"type,omitempty"`
	Result []Holding `json:"result,omitempty"`
}

// Holding represents holding information
type Holding struct {
	ClientID             string  `json:"clientID,omitempty"`
	ExchangeInstrumentID int64   `json:"exchangeInstrumentID,omitempty"`
	ExchangeSegment      string  `json:"exchangeSegment,omitempty"`
	ProductType          string  `json:"productType,omitempty"`
	Quantity             int64   `json:"quantity,omitempty"`
	AveragePrice         float64 `json:"averagePrice,omitempty"`
	CurrentPrice         float64 `json:"currentPrice,omitempty"`
	MTM                  float64 `json:"mtm,omitempty"`
	PNL                  float64 `json:"pnl,omitempty"`
}

// ModifyOrderParams represents parameters for modifying an order
type ModifyOrderParams struct {
	AppOrderID                float64 `json:"appOrderID"`
	ModifiedProductType       string  `json:"modifiedProductType,omitempty"`
	ModifiedOrderType         string  `json:"modifiedOrderType,omitempty"`
	ModifiedTimeInForce       string  `json:"modifiedTimeInForce,omitempty"`
	ModifiedOrderUID          string  `json:"modifiedOrderUID,omitempty"`
	ModifiedLimitPrice        float64 `json:"modifiedLimitPrice,omitempty"`
	ModifiedStopPrice         float64 `json:"modifiedStopPrice,omitempty"`
	ModifiedOrderQuantity     int64   `json:"modifiedOrderQuantity,omitempty"`
	ModifiedDisclosedQuantity int64   `json:"modifiedDisclosedQuantity,omitempty"`
}