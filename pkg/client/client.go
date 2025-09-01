package client

import (
	"context"
	"fmt"

	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/interfaces"
	"github.com/hakaitech/xts-go/pkg/marketdata"
	"github.com/hakaitech/xts-go/pkg/trading"
	"github.com/hakaitech/xts-go/pkg/types"
)

// XTSClient is the main client that provides access to both trading and market data APIs
type XTSClient struct {
	config     *config.Config
	Trading    *trading.Client
	MarketData *marketdata.Client
}

// Ensure XTSClient implements the XTSClient interface
var _ interfaces.XTSClient = (*XTSClient)(nil)

// NewXTSClient creates a new XTS client with the provided configuration
func NewXTSClient(cfg *config.Config) (*XTSClient, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &XTSClient{
		config:     cfg,
		Trading:    trading.NewClient(cfg),
		MarketData: marketdata.NewClient(cfg),
	}, nil
}

// NewXTSClientWithCredentials creates a new XTS client with credentials
func NewXTSClientWithCredentials(secretKey, appKey, clientID string) (*XTSClient, error) {
	cfg := config.NewConfig(secretKey, appKey, clientID)
	return NewXTSClient(cfg)
}

// NewXTSClientWithConfig creates a new XTS client with a custom configuration
func NewXTSClientWithConfig(cfg *config.Config) (*XTSClient, error) {
	return NewXTSClient(cfg)
}

// LoginToTrading performs host lookup (if needed) and logs into the trading API
func (c *XTSClient) LoginToTrading(ctx context.Context) error {
	// Perform host lookup if access password is configured
	if c.config.AccessPassword != "" {
		if err := c.Trading.HostLookup(ctx); err != nil {
			return fmt.Errorf("host lookup failed: %w", err)
		}
	}

	if err := c.Trading.Login(ctx); err != nil {
		return fmt.Errorf("trading login failed: %w", err)
	}

	return nil
}

// LoginToMarketData logs into the market data API
func (c *XTSClient) LoginToMarketData(ctx context.Context) error {
	if err := c.MarketData.Login(ctx); err != nil {
		return fmt.Errorf("market data login failed: %w", err)
	}

	return nil
}

// LoginToBoth logs into both trading and market data APIs
func (c *XTSClient) LoginToBoth(ctx context.Context) error {
	if err := c.LoginToTrading(ctx); err != nil {
		return err
	}

	if err := c.LoginToMarketData(ctx); err != nil {
		return err
	}

	return nil
}

// LogoutFromTrading logs out from the trading API
func (c *XTSClient) LogoutFromTrading(ctx context.Context) error {
	if !c.Trading.IsLoggedIn() {
		return nil // Already logged out
	}

	return c.Trading.Logout(ctx)
}

// LogoutFromMarketData logs out from the market data API
func (c *XTSClient) LogoutFromMarketData(ctx context.Context) error {
	if !c.MarketData.IsLoggedIn() {
		return nil // Already logged out
	}

	return c.MarketData.Logout(ctx)
}

// LogoutFromBoth logs out from both APIs
func (c *XTSClient) LogoutFromBoth(ctx context.Context) error {
	var tradingErr, marketDataErr error

	if c.Trading.IsLoggedIn() {
		tradingErr = c.Trading.Logout(ctx)
	}

	if c.MarketData.IsLoggedIn() {
		marketDataErr = c.MarketData.Logout(ctx)
	}

	if tradingErr != nil {
		return fmt.Errorf("trading logout failed: %w", tradingErr)
	}

	if marketDataErr != nil {
		return fmt.Errorf("market data logout failed: %w", marketDataErr)
	}

	return nil
}

// GetConfig returns the client configuration
func (c *XTSClient) GetConfig() *config.Config {
	return c.config.Clone()
}

// IsLoggedInToTrading returns true if logged into trading API
func (c *XTSClient) IsLoggedInToTrading() bool {
	return c.Trading.IsLoggedIn()
}

// IsLoggedInToMarketData returns true if logged into market data API
func (c *XTSClient) IsLoggedInToMarketData() bool {
	return c.MarketData.IsLoggedIn()
}

// IsLoggedInToBoth returns true if logged into both APIs
func (c *XTSClient) IsLoggedInToBoth() bool {
	return c.Trading.IsLoggedIn() && c.MarketData.IsLoggedIn()
}

// Convenience methods for common operations

// PlaceMarketOrder places a market order
func (c *XTSClient) PlaceMarketOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, productType string) (float64, error) {
	if !c.Trading.IsLoggedIn() {
		return 0, fmt.Errorf("not logged into trading API")
	}

	order := c.Trading.NewOrder(
		exchangeSegment,
		instrumentID,
		productType,
		types.OrderTypeMarket,
		orderSide,
		types.TimeInForceDAY,
		quantity,
		0, // limit price not needed for market order
		0, // stop price not needed for market order
		0, // disclosed quantity
		"", // order UID will be generated
	)

	return c.Trading.PlaceOrder(ctx, order)
}

// PlaceLimitOrder places a limit order
func (c *XTSClient) PlaceLimitOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, limitPrice float64, productType string) (float64, error) {
	if !c.Trading.IsLoggedIn() {
		return 0, fmt.Errorf("not logged into trading API")
	}

	order := c.Trading.NewOrder(
		exchangeSegment,
		instrumentID,
		productType,
		types.OrderTypeLimit,
		orderSide,
		types.TimeInForceDAY,
		quantity,
		limitPrice,
		0, // stop price not needed for limit order
		0, // disclosed quantity
		"", // order UID will be generated
	)

	return c.Trading.PlaceOrder(ctx, order)
}

// PlaceStopLossOrder places a stop loss order
func (c *XTSClient) PlaceStopLossOrder(ctx context.Context, exchangeSegment string, instrumentID int64, orderSide string, quantity int64, stopPrice float64, productType string) (float64, error) {
	if !c.Trading.IsLoggedIn() {
		return 0, fmt.Errorf("not logged into trading API")
	}

	order := c.Trading.NewOrder(
		exchangeSegment,
		instrumentID,
		productType,
		types.OrderTypeStopMarket,
		orderSide,
		types.TimeInForceDAY,
		quantity,
		0, // limit price not needed for stop market order
		stopPrice,
		0, // disclosed quantity
		"", // order UID will be generated
	)

	return c.Trading.PlaceOrder(ctx, order)
}

// GetLiveQuote gets live quote for a single instrument
func (c *XTSClient) GetLiveQuote(ctx context.Context, exchangeSegment int, instrumentID int) (*types.TouchlineData, error) {
	if !c.MarketData.IsLoggedIn() {
		return nil, fmt.Errorf("not logged into market data API")
	}

	instruments := []types.Instrument{
		{
			ExchangeSegment:      exchangeSegment,
			ExchangeInstrumentID: instrumentID,
		},
	}

	touchlineData, err := c.MarketData.GetTouchlineData(ctx, instruments)
	if err != nil {
		return nil, err
	}

	if len(touchlineData) == 0 {
		return nil, fmt.Errorf("no data received for instrument")
	}

	return &touchlineData[0], nil
}

// SubscribeToLiveData subscribes to live data for instruments
func (c *XTSClient) SubscribeToLiveData(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	if !c.MarketData.IsLoggedIn() {
		return nil, fmt.Errorf("not logged into market data API")
	}

	return c.MarketData.SubscribeToTouchline(ctx, instruments)
}

// GetInstrumentMaster gets instrument master for an exchange
func (c *XTSClient) GetInstrumentMaster(ctx context.Context, exchangeSegment string) (interface{}, error) {
	if !c.MarketData.IsLoggedIn() {
		return nil, fmt.Errorf("not logged into market data API")
	}

	return c.MarketData.GetInstruments(ctx, exchangeSegment)
}

// SearchInstrument searches for instruments by symbol
func (c *XTSClient) SearchInstrument(ctx context.Context, searchString, exchangeSegment string) (interface{}, error) {
	if !c.MarketData.IsLoggedIn() {
		return nil, fmt.Errorf("not logged into market data API")
	}

	return c.MarketData.SearchInstruments(ctx, searchString, exchangeSegment)
}

// SetEnvironment configures the client for different environments
func (c *XTSClient) SetEnvironment(env string) {
	c.config.SetEnvironment(env)
	
	// Update the clients with new configuration
	c.Trading = trading.NewClient(c.config)
	c.MarketData = marketdata.NewClient(c.config)
}

// SetDebug enables or disables debug logging
func (c *XTSClient) SetDebug(enable bool) {
	c.config.Debug = enable
}

// GetTradingAPI returns the trading API interface
func (c *XTSClient) GetTradingAPI() interfaces.TradingAPI {
	return c.Trading
}

// GetMarketDataAPI returns the market data API interface
func (c *XTSClient) GetMarketDataAPI() interfaces.MarketDataAPI {
	return c.MarketData
}