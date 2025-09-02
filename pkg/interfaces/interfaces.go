package interfaces

import (
	"context"
	"net/http"

	"github.com/hakaitech/xts-go/pkg/types"
)

// TradingAPI defines the interface for trading operations
type TradingAPI interface {
	Login(ctx context.Context) error
	Logout(ctx context.Context) error
	HostLookup(ctx context.Context) error
	IsLoggedIn() bool

	GetProfile(ctx context.Context) (*types.UserProfile, error)
	GetBalance(ctx context.Context) (interface{}, error)

	PlaceOrder(ctx context.Context, order *types.Order) (float64, error)
	ModifyOrder(ctx context.Context, params *types.ModificationParams) (float64, error)
	CancelOrder(ctx context.Context, orderID float64) error
	CancelAllOrders(ctx context.Context, exchangeSegment string, instrumentID int64) (interface{}, error)

	GetOrders(ctx context.Context) (interface{}, error)
	GetTrades(ctx context.Context) (interface{}, error)
	GetPositions(ctx context.Context) (interface{}, error)
	GetHoldings(ctx context.Context) (interface{}, error)

	NewOrder(exchangeSegment string, instrumentID int64, productType, orderType, orderSide, timeInForce string, quantity int64, limitPrice, stopPrice float64, disclosedQuantity int64, orderUniqueIdentifier string) *types.Order

	IsInvestorClient() bool
}

// MarketDataAPI defines the interface for market data operations
type MarketDataAPI interface {
	Login(ctx context.Context) error
	Logout(ctx context.Context) error
	IsLoggedIn() bool

	GetQuote(ctx context.Context, instruments []types.Instrument, messageCode int, format string) (interface{}, error)
	GetTouchlineData(ctx context.Context, instruments []types.Instrument) ([]types.TouchlineData, error)
	GetMarketDepth(ctx context.Context, instruments []types.Instrument) (interface{}, error)
	GetOHLC(ctx context.Context, instruments []types.Instrument, startTime, endTime string, compressionValue int) (interface{}, error)

	Subscribe(ctx context.Context, instruments []types.Instrument, messageCode int) (interface{}, error)
	SubscribeToTouchline(ctx context.Context, instruments []types.Instrument) (interface{}, error)
	SubscribeToMarketDepth(ctx context.Context, instruments []types.Instrument) (interface{}, error)
	Unsubscribe(ctx context.Context, instruments []types.Instrument, messageCode int) (interface{}, error)

	GetInstruments(ctx context.Context, exchangeSegment string) (interface{}, error)
	SearchInstruments(ctx context.Context, searchString, exchangeSegment string) (interface{}, error)
	GetIndexList(ctx context.Context, exchangeSegment string) (interface{}, error)
}

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Post(ctx context.Context, endpoint string, payload interface{}, headers map[string]string) (*http.Response, error)
	Get(ctx context.Context, endpoint string, headers map[string]string) (*http.Response, error)
	SetBaseURL(url string)
}

// XTSClient defines the main client interface
type XTSClient interface {
	LoginToTrading(ctx context.Context) error
	LoginToMarketData(ctx context.Context) error
	LoginToBoth(ctx context.Context) error
	LogoutFromTrading(ctx context.Context) error
	LogoutFromMarketData(ctx context.Context) error
	LogoutFromBoth(ctx context.Context) error

	IsLoggedInToTrading() bool
	IsLoggedInToMarketData() bool
	IsLoggedInToBoth() bool

	SetEnvironment(env string)
	SetDebug(enable bool)

	GetTradingAPI() TradingAPI
	GetMarketDataAPI() MarketDataAPI

	PlaceMarketOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, productType string) (float64, error)
	PlaceLimitOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, limitPrice float64, productType string) (float64, error)
	PlaceStopLossOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, stopPrice float64, productType string) (float64, error)
	GetLiveQuote(ctx context.Context, exchangeSegment int, instrumentID int) (*types.TouchlineData, error)
	SubscribeToLiveData(ctx context.Context, instruments []types.Instrument) (interface{}, error)
	GetInstrumentMaster(ctx context.Context, exchangeSegment string) (interface{}, error)
	SearchInstrument(ctx context.Context, searchString, exchangeSegment string) (interface{}, error)
}
