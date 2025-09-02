package websocket

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/types"

	"github.com/gorilla/websocket"
)

// EventHandler defines the interface for handling WebSocket events
type EventHandler interface {
	OnConnect()
	OnDisconnect()
	OnError(err error)
	OnMessage(data []byte)
	OnTouchlineData(data *types.TouchlineData)
	OnMarketDepthData(data interface{})
	OnIndexData(data interface{})
	OnCandleData(data interface{})
	OnOpenInterestData(data interface{})
}

// DefaultEventHandler provides a default implementation of EventHandler
type DefaultEventHandler struct{}

func (h *DefaultEventHandler) OnConnect()                                {}
func (h *DefaultEventHandler) OnDisconnect()                             {}
func (h *DefaultEventHandler) OnError(err error)                         {}
func (h *DefaultEventHandler) OnMessage(data []byte)                     {}
func (h *DefaultEventHandler) OnTouchlineData(data *types.TouchlineData) {}
func (h *DefaultEventHandler) OnMarketDepthData(data interface{})        {}
func (h *DefaultEventHandler) OnIndexData(data interface{})              {}
func (h *DefaultEventHandler) OnCandleData(data interface{})             {}
func (h *DefaultEventHandler) OnOpenInterestData(data interface{})       {}

// WebSocketMessage represents a message received from WebSocket
type WebSocketMessage struct {
	MessageCode int         `json:"messageCode"`
	Data        interface{} `json:"data"`
	Timestamp   int64       `json:"timestamp"`
}

// Client handles WebSocket connections for real-time data
type Client struct {
	config      *config.Config
	conn        *websocket.Conn
	handler     EventHandler
	isConnected bool
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	reconnect   bool
	wsURL       string
}

// NewClient creates a new WebSocket client
func NewClient(cfg *config.Config, handler EventHandler) *Client {
	if handler == nil {
		handler = &DefaultEventHandler{}
	}

	return &Client{
		config:    cfg,
		handler:   handler,
		reconnect: true,
	}
}

// Connect establishes a WebSocket connection
func (c *Client) Connect(ctx context.Context, wsURL, token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return fmt.Errorf("already connected")
	}

	c.wsURL = wsURL
	c.ctx, c.cancel = context.WithCancel(ctx)

	u, err := url.Parse(wsURL)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	dialer := websocket.DefaultDialer
	if c.config.DisableSSL {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.isConnected = true

	go c.readLoop()
	go c.pingLoop()

	c.handler.OnConnect()

	return nil
}

// Disconnect closes the WebSocket connection
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return nil
	}

	c.reconnect = false

	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.isConnected = false
		c.handler.OnDisconnect()
		return err
	}

	return nil
}

// IsConnected returns true if the WebSocket is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// Subscribe sends a subscription message
func (c *Client) Subscribe(instruments []types.Instrument, messageCode int) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	subscriptionMsg := map[string]interface{}{
		"correlationID": fmt.Sprintf("sub_%d_%d", messageCode, time.Now().Unix()),
		"action":        1, // Subscribe action
		"params": map[string]interface{}{
			"mode":           messageCode,
			"instrumentKeys": instruments,
		},
	}

	return c.sendMessage(subscriptionMsg)
}

// Unsubscribe sends an unsubscription message
func (c *Client) Unsubscribe(instruments []types.Instrument, messageCode int) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}

	unsubscriptionMsg := map[string]interface{}{
		"correlationID": fmt.Sprintf("unsub_%d_%d", messageCode, time.Now().Unix()),
		"action":        0, // Unsubscribe action
		"params": map[string]interface{}{
			"mode":           messageCode,
			"instrumentKeys": instruments,
		},
	}

	return c.sendMessage(unsubscriptionMsg)
}

// sendMessage sends a message through the WebSocket
func (c *Client) sendMessage(message interface{}) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// readLoop handles incoming messages
func (c *Client) readLoop() {
	defer func() {
		c.mu.Lock()
		c.isConnected = false
		c.mu.Unlock()
		c.handler.OnDisconnect()

		if c.reconnect {
			go c.attemptReconnect()
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.handler.OnError(fmt.Errorf("WebSocket read error: %w", err))
				}
				return
			}

			c.handler.OnMessage(message)
			c.handleMessage(message)
		}
	}
}

// pingLoop sends periodic ping messages to keep the connection alive
func (c *Client) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.handler.OnError(fmt.Errorf("failed to send ping: %w", err))
				return
			}
		}
	}
}

// handleMessage processes incoming messages and routes them to appropriate handlers
func (c *Client) handleMessage(message []byte) {
	var wsMsg WebSocketMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		c.handler.OnError(fmt.Errorf("failed to parse WebSocket message: %w", err))
		return
	}

	switch wsMsg.MessageCode {
	case types.MessageCodeTouchline:
		if touchlineData := c.parseTouchlineData(wsMsg.Data); touchlineData != nil {
			c.handler.OnTouchlineData(touchlineData)
		}
	case types.MessageCodeMarketDepth:
		c.handler.OnMarketDepthData(wsMsg.Data)
	case types.MessageCodeIndexData:
		c.handler.OnIndexData(wsMsg.Data)
	case types.MessageCodeCandleData:
		c.handler.OnCandleData(wsMsg.Data)
	case types.MessageCodeOpenInterest:
		c.handler.OnOpenInterestData(wsMsg.Data)
	default:
		// Unhandled message code - silently ignore
	}
}

// parseTouchlineData converts WebSocket data to TouchlineData
func (c *Client) parseTouchlineData(data interface{}) *types.TouchlineData {
	if dataMap, ok := data.(map[string]interface{}); ok {
		var touchline types.TouchlineData

		if val, ok := dataMap["exchangeSegment"].(float64); ok {
			touchline.ExchangeSegment = int(val)
		}
		if val, ok := dataMap["exchangeInstrumentID"].(float64); ok {
			touchline.ExchangeInstrumentID = int(val)
		}
		if val, ok := dataMap["lastTradedPrice"].(float64); ok {
			touchline.LastTradedPrice = val
		}
		if val, ok := dataMap["lastTradedTime"].(float64); ok {
			touchline.LastTradedTime = int64(val)
		}
		if val, ok := dataMap["percentChange"].(float64); ok {
			touchline.PercentChange = val
		}
		if val, ok := dataMap["lastTradedQuantity"].(float64); ok {
			touchline.LastTradedQuantity = int(val)
		}
		if val, ok := dataMap["volumeTraded"].(float64); ok {
			touchline.VolumeTraded = int64(val)
		}
		if val, ok := dataMap["bestBuyPrice"].(float64); ok {
			touchline.BestBuyPrice = val
		}
		if val, ok := dataMap["bestSellPrice"].(float64); ok {
			touchline.BestSellPrice = val
		}
		if val, ok := dataMap["totalTrades"].(float64); ok {
			touchline.TotalTrades = int(val)
		}
		if val, ok := dataMap["open"].(float64); ok {
			touchline.Open = val
		}
		if val, ok := dataMap["high"].(float64); ok {
			touchline.High = val
		}
		if val, ok := dataMap["low"].(float64); ok {
			touchline.Low = val
		}
		if val, ok := dataMap["close"].(float64); ok {
			touchline.Close = val
		}

		return &touchline
	}

	return nil
}

// attemptReconnect tries to reconnect to the WebSocket
func (c *Client) attemptReconnect() {
	maxRetries := c.config.RetryAttempts
	if maxRetries <= 0 {
		maxRetries = 5
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if !c.reconnect {
			return
		}

		time.Sleep(time.Duration(attempt) * time.Second)

		fmt.Printf("Attempting WebSocket reconnection (attempt %d/%d)\n", attempt, maxRetries)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := c.Connect(ctx, c.wsURL, "") // Token would need to be refreshed
		cancel()

		if err == nil {
			fmt.Println("WebSocket reconnection successful")
			return
		}

		c.handler.OnError(fmt.Errorf("reconnection attempt %d failed: %w", attempt, err))
	}

	c.handler.OnError(fmt.Errorf("failed to reconnect after %d attempts", maxRetries))
}

// SetReconnect enables or disables automatic reconnection
func (c *Client) SetReconnect(enable bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = enable
}
