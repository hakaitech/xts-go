package marketdata

import (
	"context"
	"fmt"

	"github.com/hakaitech/xts-go/pkg/api"
	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/types"
)

// Client handles market data API operations
type Client struct {
	apiClient    *api.Client
	config       *config.Config
	sessionToken string
	userID       string
}

// NewClient creates a new market data client
func NewClient(cfg *config.Config) *Client {
	apiClient := api.NewClient(cfg)
	apiClient.SetBaseURL(cfg.MarketDataBaseURL)

	return &Client{
		apiClient: apiClient,
		config:    cfg,
	}
}

// Login authenticates with the XTS Market Data API
func (c *Client) Login(ctx context.Context) error {
	loginReq := types.LoginRequest{
		SecretKey: c.config.SecretKey,
		AppKey:    c.config.AppKey,
		Source:    c.config.Source,
	}

	resp, err := c.apiClient.Post(ctx, "/marketdata/auth/login", loginReq, nil)
	if err != nil {
		return fmt.Errorf("market data login failed: %w", err)
	}

	loginResp, err := api.DecodeResponse[types.LoginResponse](c.apiClient, resp)
	if err != nil {
		return fmt.Errorf("failed to decode market data login response: %w", err)
	}

	c.sessionToken = loginResp.Result.Token
	c.userID = loginResp.Result.UserID

	return nil
}

// Logout terminates the market data session
func (c *Client) Logout(ctx context.Context) error {
	if c.sessionToken == "" {
		return fmt.Errorf("not logged in")
	}

	resp, err := c.apiClient.Delete(ctx, "/marketdata/auth/logout", api.AuthHeaders(c.sessionToken))
	if err != nil {
		return fmt.Errorf("market data logout failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("market data logout failed with status: %s", resp.Status)
	}

	c.sessionToken = ""
	c.userID = ""

	return nil
}

// GetQuote retrieves quotes for specified instruments
func (c *Client) GetQuote(ctx context.Context, instruments []types.Instrument, messageCode int, publishFormat string) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	if publishFormat == "" {
		publishFormat = "JSON"
	}

	quoteReq := types.QuoteRequest{
		Instruments:    instruments,
		XTSMessageCode: messageCode,
		PublishFormat:  publishFormat,
	}

	resp, err := c.apiClient.Post(ctx, "/marketdata/instruments/quotes", quoteReq, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}

	quoteResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode quote response: %w", err)
	}

	return quoteResp.Result, nil
}

// GetTouchlineData retrieves touchline data for instruments
func (c *Client) GetTouchlineData(ctx context.Context, instruments []types.Instrument) ([]types.TouchlineData, error) {
	result, err := c.GetQuote(ctx, instruments, types.MessageCodeTouchline, "JSON")
	if err != nil {
		return nil, err
	}

	if dataSlice, ok := result.([]interface{}); ok {
		touchlineData := make([]types.TouchlineData, len(dataSlice))
		for i, item := range dataSlice {
			if data, ok := item.(map[string]interface{}); ok {
				touchlineData[i] = mapToTouchlineData(data)
			}
		}
		return touchlineData, nil
	}

	return nil, fmt.Errorf("unexpected response format for touchline data")
}

// GetMarketDepth retrieves market depth data for instruments
func (c *Client) GetMarketDepth(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.GetQuote(ctx, instruments, types.MessageCodeMarketDepth, "JSON")
}

// GetIndexData retrieves index data
func (c *Client) GetIndexData(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.GetQuote(ctx, instruments, types.MessageCodeIndexData, "JSON")
}

// GetCandleData retrieves candle data
func (c *Client) GetCandleData(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.GetQuote(ctx, instruments, types.MessageCodeCandleData, "JSON")
}

// GetOpenInterest retrieves open interest data
func (c *Client) GetOpenInterest(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.GetQuote(ctx, instruments, types.MessageCodeOpenInterest, "JSON")
}

// Subscribe subscribes to real-time data for specified instruments
func (c *Client) Subscribe(ctx context.Context, instruments []types.Instrument, messageCode int) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	subReq := types.SubscriptionRequest{
		Instruments:    instruments,
		XTSMessageCode: messageCode,
	}

	resp, err := c.apiClient.Post(ctx, "/marketdata/instruments/subscription", subReq, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	subResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode subscription response: %w", err)
	}

	return subResp.Result, nil
}

// SubscribeToTouchline subscribes to touchline data
func (c *Client) SubscribeToTouchline(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.Subscribe(ctx, instruments, types.MessageCodeTouchline)
}

// SubscribeToMarketDepth subscribes to market depth data
func (c *Client) SubscribeToMarketDepth(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.Subscribe(ctx, instruments, types.MessageCodeMarketDepth)
}

// SubscribeToIndexData subscribes to index data
func (c *Client) SubscribeToIndexData(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.Subscribe(ctx, instruments, types.MessageCodeIndexData)
}

// SubscribeToCandleData subscribes to candle data
func (c *Client) SubscribeToCandleData(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.Subscribe(ctx, instruments, types.MessageCodeCandleData)
}

// SubscribeToOpenInterest subscribes to open interest data
func (c *Client) SubscribeToOpenInterest(ctx context.Context, instruments []types.Instrument) (interface{}, error) {
	return c.Subscribe(ctx, instruments, types.MessageCodeOpenInterest)
}

// Unsubscribe unsubscribes from real-time data for specified instruments
func (c *Client) Unsubscribe(ctx context.Context, instruments []types.Instrument, messageCode int) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	subReq := types.SubscriptionRequest{
		Instruments:    instruments,
		XTSMessageCode: messageCode,
	}

	resp, err := c.apiClient.Put(ctx, "/marketdata/instruments/subscription", subReq, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to unsubscribe: %w", err)
	}

	subResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode unsubscription response: %w", err)
	}

	return subResp.Result, nil
}

// GetInstruments retrieves instrument master data
func (c *Client) GetInstruments(ctx context.Context, exchangeSegment string) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := map[string]string{
		"exchangeSegment": exchangeSegment,
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/marketdata/instruments/master", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get instruments: %w", err)
	}

	instrumentsResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode instruments response: %w", err)
	}

	return instrumentsResp.Result, nil
}

// GetOHLC retrieves OHLC data
func (c *Client) GetOHLC(ctx context.Context, instruments []types.Instrument, startTime, endTime string, compressionValue int) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	body := map[string]interface{}{
		"instruments":      instruments,
		"startTime":        startTime,
		"endTime":          endTime,
		"compressionValue": compressionValue,
	}

	resp, err := c.apiClient.Post(ctx, "/marketdata/instruments/ohlc", body, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get OHLC data: %w", err)
	}

	ohlcResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode OHLC response: %w", err)
	}

	return ohlcResp.Result, nil
}

// GetIndexList retrieves list of indices
func (c *Client) GetIndexList(ctx context.Context, exchangeSegment string) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := map[string]string{
		"exchangeSegment": exchangeSegment,
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/marketdata/instruments/indexlist", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get index list: %w", err)
	}

	indexResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode index list response: %w", err)
	}

	return indexResp.Result, nil
}

// SearchInstruments searches for instruments by symbol
func (c *Client) SearchInstruments(ctx context.Context, searchString, exchangeSegment string) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	body := map[string]string{
		"searchString":    searchString,
		"exchangeSegment": exchangeSegment,
	}

	resp, err := c.apiClient.Post(ctx, "/marketdata/search/instruments", body, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to search instruments: %w", err)
	}

	searchResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return searchResp.Result, nil
}

// IsLoggedIn returns true if the client is logged in
func (c *Client) IsLoggedIn() bool {
	return c.sessionToken != ""
}

// GetSessionToken returns the current session token
func (c *Client) GetSessionToken() string {
	return c.sessionToken
}

// GetUserID returns the current user ID
func (c *Client) GetUserID() string {
	return c.userID
}

// Helper function to convert map to TouchlineData
func mapToTouchlineData(data map[string]interface{}) types.TouchlineData {
	var touchline types.TouchlineData

	if val, ok := data["exchangeSegment"].(float64); ok {
		touchline.ExchangeSegment = int(val)
	}
	if val, ok := data["exchangeInstrumentID"].(float64); ok {
		touchline.ExchangeInstrumentID = int(val)
	}
	if val, ok := data["lastTradedPrice"].(float64); ok {
		touchline.LastTradedPrice = val
	}
	if val, ok := data["lastTradedTime"].(float64); ok {
		touchline.LastTradedTime = int64(val)
	}
	if val, ok := data["percentChange"].(float64); ok {
		touchline.PercentChange = val
	}
	if val, ok := data["lastTradedQuantity"].(float64); ok {
		touchline.LastTradedQuantity = int(val)
	}
	if val, ok := data["volumeTraded"].(float64); ok {
		touchline.VolumeTraded = int64(val)
	}
	if val, ok := data["bestBuyPrice"].(float64); ok {
		touchline.BestBuyPrice = val
	}
	if val, ok := data["bestSellPrice"].(float64); ok {
		touchline.BestSellPrice = val
	}
	if val, ok := data["totalTrades"].(float64); ok {
		touchline.TotalTrades = int(val)
	}
	if val, ok := data["open"].(float64); ok {
		touchline.Open = val
	}
	if val, ok := data["high"].(float64); ok {
		touchline.High = val
	}
	if val, ok := data["low"].(float64); ok {
		touchline.Low = val
	}
	if val, ok := data["close"].(float64); ok {
		touchline.Close = val
	}

	return touchline
}
