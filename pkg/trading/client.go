package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hakaitech/xts-go/pkg/api"
	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/types"
)

// Client handles trading/interactive API operations
type Client struct {
	apiClient        *api.Client
	config           *config.Config
	sessionToken     string
	userID           string
	isInvestorClient bool
}

// NewClient creates a new trading client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		apiClient: api.NewClient(cfg),
		config:    cfg,
	}
}

// HostLookup performs host lookup for Interactive API (required before login)
func (c *Client) HostLookup(ctx context.Context) error {
	if c.config.AccessPassword == "" {
		return fmt.Errorf("access password is required for host lookup")
	}

	c.apiClient.SetBaseURL(c.config.HostLookupURL)
	
	body := map[string]string{
		"accesspassword": c.config.AccessPassword,
		"version":        c.config.Version,
	}

	resp, err := c.apiClient.Post(ctx, "/hostlookup", body, nil)
	if err != nil {
		return fmt.Errorf("host lookup failed: %w", err)
	}

	// Decode response to get the actual base URL
	result, err := api.DecodeDirectResponse[map[string]interface{}](c.apiClient, resp)
	if err != nil {
		return fmt.Errorf("failed to decode host lookup response: %w", err)
	}

	// Update base URL with the response
	if baseURL, ok := (*result)["baseURL"].(string); ok && baseURL != "" {
		c.apiClient.SetBaseURL(baseURL)
	}

	return nil
}

// Login authenticates with the XTS Trading API
func (c *Client) Login(ctx context.Context) error {
	loginReq := types.LoginRequest{
		SecretKey: c.config.SecretKey,
		AppKey:    c.config.AppKey,
		Source:    c.config.Source,
	}

	resp, err := c.apiClient.Post(ctx, "/interactive/user/session", loginReq, nil)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	loginResp, err := api.DecodeResponse[types.LoginResponse](c.apiClient, resp)
	if err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.sessionToken = loginResp.Result.Token
	c.userID = loginResp.Result.UserID
	c.isInvestorClient = loginResp.Result.IsInvestorClient

	return nil
}

// Logout terminates the session
func (c *Client) Logout(ctx context.Context) error {
	if c.sessionToken == "" {
		return fmt.Errorf("not logged in")
	}

	resp, err := c.apiClient.Delete(ctx, "/interactive/user/session", api.AuthHeaders(c.sessionToken))
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("logout failed with status: %s", resp.Status)
	}

	// Clear session data
	c.sessionToken = ""
	c.userID = ""
	c.isInvestorClient = false

	return nil
}

// GetProfile retrieves user profile information
func (c *Client) GetProfile(ctx context.Context) (*types.UserProfile, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := make(map[string]string)
	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/user/profile", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	profileResp, err := api.DecodeResponse[types.UserProfile](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode profile response: %w", err)
	}

	return &profileResp.Result, nil
}

// GetBalance retrieves account balance (available only for retail clients)
func (c *Client) GetBalance(ctx context.Context) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	if !c.isInvestorClient {
		return nil, fmt.Errorf("balance API is available for retail API users only")
	}

	params := make(map[string]string)
	if c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/user/balance", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	balanceResp, err := api.DecodeResponse[map[string]interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode balance response: %w", err)
	}

	return balanceResp.Result, nil
}

// NewOrder creates a new order object
func (c *Client) NewOrder(exchangeSegment string, instrumentID int64, productType, orderType, orderSide, timeInForce string, orderQuantity int64, limitPrice, stopPrice float64, disclosedQuantity int64, orderUID string) *types.Order {
	order := &types.Order{
		OrderUID:             orderUID,
		ExchangeSegment:      exchangeSegment,
		ExchangeInstrumentID: instrumentID,
		ProductType:          productType,
		OrderType:            orderType,
		OrderSide:            orderSide,
		TimeInForce:          timeInForce,
		OrderQuantity:        orderQuantity,
		LimitPrice:           limitPrice,
		StopPrice:            stopPrice,
		DisclosedQuantity:    disclosedQuantity,
	}

	// Add client ID for dealer accounts
	if !c.isInvestorClient && c.config.ClientID != "" {
		order.ClientID = c.config.ClientID
	}

	return order
}

// PlaceOrder places a new order
func (c *Client) PlaceOrder(ctx context.Context, order *types.Order) (float64, error) {
	if c.sessionToken == "" {
		return 0, fmt.Errorf("not logged in")
	}

	resp, err := c.apiClient.Post(ctx, "/interactive/orders", order, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return 0, fmt.Errorf("failed to place order: %w", err)
	}

	orderResp, err := api.DecodeResponse[types.OrderResponse](c.apiClient, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to decode order response: %w", err)
	}

	// Update the order with the app order ID
	order.AppOrderID = orderResp.Result.AppOrderID

	return orderResp.Result.AppOrderID, nil
}

// ModifyOrder modifies an existing order
func (c *Client) ModifyOrder(ctx context.Context, modParams *types.ModificationParams) (float64, error) {
	if c.sessionToken == "" {
		return 0, fmt.Errorf("not logged in")
	}

	resp, err := c.apiClient.Put(ctx, "/interactive/orders", modParams, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return 0, fmt.Errorf("failed to modify order: %w", err)
	}

	orderResp, err := api.DecodeResponse[types.OrderResponse](c.apiClient, resp)
	if err != nil {
		return 0, fmt.Errorf("failed to decode modify response: %w", err)
	}

	return orderResp.Result.AppOrderID, nil
}

// CancelOrder cancels an existing order
func (c *Client) CancelOrder(ctx context.Context, appOrderID float64) error {
	if c.sessionToken == "" {
		return fmt.Errorf("not logged in")
	}

	params := map[string]string{
		"appOrderID": strconv.FormatFloat(appOrderID, 'f', 0, 64),
	}

	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/orders", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("cancel order failed with status: %s", resp.Status)
	}

	return nil
}

// CancelAllOrders cancels all orders for a specific instrument
func (c *Client) CancelAllOrders(ctx context.Context, exchangeSegment string, instrumentID int64) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	body := map[string]interface{}{
		"exchangeSegment":      exchangeSegment,
		"exchangeInstrumentID": instrumentID,
	}

	resp, err := c.apiClient.Post(ctx, "/interactive/orders/cancelall", body, map[string]string{
		"Authorization": c.sessionToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel all orders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cancel all orders failed with status: %s", resp.Status)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetOrders retrieves order history
func (c *Client) GetOrders(ctx context.Context) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := make(map[string]string)
	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/orders", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	ordersResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode orders response: %w", err)
	}

	return ordersResp.Result, nil
}

// GetTrades retrieves trade history
func (c *Client) GetTrades(ctx context.Context) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := make(map[string]string)
	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/trades", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	tradesResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode trades response: %w", err)
	}

	return tradesResp.Result, nil
}

// GetPositions retrieves current positions
func (c *Client) GetPositions(ctx context.Context) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := make(map[string]string)
	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/portfolio/positions", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	positionsResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode positions response: %w", err)
	}

	return positionsResp.Result, nil
}

// GetHoldings retrieves current holdings
func (c *Client) GetHoldings(ctx context.Context) (interface{}, error) {
	if c.sessionToken == "" {
		return nil, fmt.Errorf("not logged in")
	}

	params := make(map[string]string)
	if !c.isInvestorClient && c.config.ClientID != "" {
		params["clientID"] = c.config.ClientID
	}

	resp, err := c.apiClient.GetWithQuery(ctx, "/interactive/portfolio/holdings", params, api.AuthHeaders(c.sessionToken))
	if err != nil {
		return nil, fmt.Errorf("failed to get holdings: %w", err)
	}

	holdingsResp, err := api.DecodeResponse[interface{}](c.apiClient, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode holdings response: %w", err)
	}

	return holdingsResp.Result, nil
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

// IsInvestorClient returns true if the logged-in client is an investor client
func (c *Client) IsInvestorClient() bool {
	return c.isInvestorClient
}