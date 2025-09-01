# xts-go : Official Go API Client for XTS Trading APIs

![License](https://img.shields.io/github/license/hakaitech/xts-go)
![GitHub issues](https://img.shields.io/github/issues-raw/hakaitech/xts-go)
![GitHub go.mod Go version (branch & subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/hakaitech/xts-go/main)

A comprehensive, modular Go wrapper for the XTS Trading APIs by Symphony FinTech. This library provides a clean, easy-to-use interface for both Market Data and Interactive (Trading) APIs with full WebSocket streaming support.

## Features

- **Modular Design**: Separate packages for client functionality, HTTP API calls, and configuration management
- **Complete API Coverage**: Both Market Data and Interactive/Trading APIs
- **WebSocket Streaming**: Real-time market data with automatic reconnection
- **Type Safety**: Strongly typed Go structs for all API responses
- **Error Handling**: Comprehensive error handling with retry mechanisms
- **Authentication**: Automatic session management and token handling
- **Host Lookup**: Support for XTS host lookup functionality
- **Environment Support**: Easy switching between development, sandbox, and production environments
- **Debug Support**: Built-in logging and debugging capabilities

## Installation

```bash
go get github.com/hakaitech/xts-go
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "xts-go/pkg/client"
    "xts-go/pkg/config"
    "xts-go/pkg/types"
)

func main() {
    // Create configuration
    cfg := config.NewConfig("your-secret-key", "your-app-key", "your-client-id")
    cfg.SetEnvironment("development") // or "production"
    
    // Create XTS client
    xtsClient, err := client.NewXTSClient(cfg)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Login to both APIs
    if err := xtsClient.LoginToBoth(ctx); err != nil {
        log.Fatal(err)
    }
    defer xtsClient.LogoutFromBoth(ctx)
    
    // Get live quote
    quote, err := xtsClient.GetLiveQuote(ctx, 1, 2885) // RELIANCE on NSE
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("RELIANCE LTP: %.2f\n", quote.LastTradedPrice)
    
    // Place a market order (example - use with caution)
    orderID, err := xtsClient.PlaceMarketOrder(ctx, 
        types.ExchangeNSECM, 2885, types.OrderSideBuy, 1, types.ProductTypeMIS)
    if err != nil {
        log.Printf("Order failed: %v", err)
    } else {
        fmt.Printf("Order placed: %.0f\n", orderID)
    }
}
```

### Trading API Only

```go
import (
    "xts-go/pkg/config"
    "xts-go/pkg/trading"
)

cfg := config.NewConfig("secret", "appkey", "clientid")
tradingClient := trading.NewClient(cfg)

// Login
if err := tradingClient.Login(ctx); err != nil {
    log.Fatal(err)
}

// Get profile
profile, err := tradingClient.GetProfile(ctx)
if err != nil {
    log.Fatal(err)
}

// Create and place order
order := tradingClient.NewOrder(
    types.ExchangeNSECM, 2885, types.ProductTypeMIS,
    types.OrderTypeMarket, types.OrderSideBuy, types.TimeInForceDAY,
    1, 0, 0, 0, "unique-order-id")

orderID, err := tradingClient.PlaceOrder(ctx, order)
```

### Market Data API Only

```go
import (
    "xts-go/pkg/config"
    "xts-go/pkg/marketdata"
    "xts-go/pkg/types"
)

cfg := config.NewConfig("secret", "appkey", "clientid")
marketClient := marketdata.NewClient(cfg)

// Login
if err := marketClient.Login(ctx); err != nil {
    log.Fatal(err)
}

// Get quotes
instruments := []types.Instrument{
    {ExchangeSegment: 1, ExchangeInstrumentID: 2885}, // RELIANCE
}

quotes, err := marketClient.GetQuote(ctx, instruments, types.MessageCodeTouchline, "JSON")
```

### WebSocket Streaming

```go
import (
    "xts-go/pkg/websocket"
)

// Implement event handler
type MyHandler struct {
    *websocket.DefaultEventHandler
}

func (h *MyHandler) OnTouchlineData(data *types.TouchlineData) {
    fmt.Printf("Live Update: %d = %.2f\n", 
        data.ExchangeInstrumentID, data.LastTradedPrice)
}

// Create WebSocket client
handler := &MyHandler{}
wsClient := websocket.NewClient(cfg, handler)

// Connect and subscribe
wsClient.Connect(ctx, "wss://xts-websocket-url", token)
wsClient.Subscribe(instruments, types.MessageCodeTouchline)
```

## API Reference

### Configuration

The configuration package provides flexible configuration management:

```go
cfg := config.NewConfig("secretKey", "appKey", "clientID")

// Environment settings
cfg.SetEnvironment("development") // "development", "production", "sandbox"

// Custom settings
cfg.BaseURL = "https://custom-url.com"
cfg.Timeout = 60 * time.Second
cfg.RetryAttempts = 5
cfg.Debug = true
```

### Trading API

#### Authentication
- `Login(ctx)` - Login to trading API
- `HostLookup(ctx)` - Perform host lookup (required for some setups)
- `Logout(ctx)` - Logout from trading API

#### Account Management
- `GetProfile(ctx)` - Get user profile
- `GetBalance(ctx)` - Get account balance (retail clients only)

#### Order Management
- `NewOrder(...)` - Create a new order object
- `PlaceOrder(ctx, order)` - Place an order
- `ModifyOrder(ctx, modParams)` - Modify an existing order
- `CancelOrder(ctx, orderID)` - Cancel an order
- `CancelAllOrders(ctx, exchange, instrumentID)` - Cancel all orders for an instrument

#### Portfolio
- `GetOrders(ctx)` - Get order history
- `GetTrades(ctx)` - Get trade history
- `GetPositions(ctx)` - Get current positions
- `GetHoldings(ctx)` - Get current holdings

### Market Data API

#### Authentication
- `Login(ctx)` - Login to market data API
- `Logout(ctx)` - Logout from market data API

#### Market Data
- `GetQuote(ctx, instruments, messageCode, format)` - Get quotes
- `GetTouchlineData(ctx, instruments)` - Get touchline data
- `GetMarketDepth(ctx, instruments)` - Get market depth
- `GetOHLC(ctx, instruments, startTime, endTime, compression)` - Get OHLC data

#### Subscriptions
- `Subscribe(ctx, instruments, messageCode)` - Subscribe to real-time data
- `SubscribeToTouchline(ctx, instruments)` - Subscribe to touchline data
- `SubscribeToMarketDepth(ctx, instruments)` - Subscribe to market depth
- `Unsubscribe(ctx, instruments, messageCode)` - Unsubscribe from data

#### Instruments
- `GetInstruments(ctx, exchangeSegment)` - Get instrument master
- `SearchInstruments(ctx, searchString, exchange)` - Search instruments
- `GetIndexList(ctx, exchangeSegment)` - Get list of indices

### WebSocket Streaming

#### Event Handler Interface
```go
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
```

#### WebSocket Client Methods
- `Connect(ctx, wsURL, token)` - Connect to WebSocket
- `Disconnect()` - Disconnect from WebSocket
- `Subscribe(instruments, messageCode)` - Subscribe to data stream
- `Unsubscribe(instruments, messageCode)` - Unsubscribe from data stream
- `IsConnected()` - Check connection status

## Constants and Types

### Exchange Segments
- `types.ExchangeNSECM` - NSE Cash Market
- `types.ExchangeNSEFO` - NSE Futures & Options
- `types.ExchangeBSECM` - BSE Cash Market
- `types.ExchangeBSEFO` - BSE Futures & Options

### Order Types
- `types.OrderTypeMarket` - Market Order
- `types.OrderTypeLimit` - Limit Order
- `types.OrderTypeStopLimit` - Stop Limit Order
- `types.OrderTypeStopMarket` - Stop Market Order

### Product Types
- `types.ProductTypeNRML` - Normal
- `types.ProductTypeMIS` - Margin Intraday Square-off
- `types.ProductTypeCNC` - Cash and Carry
- `types.ProductTypeCO` - Cover Order
- `types.ProductTypeBO` - Bracket Order

### Message Codes
- `types.MessageCodeTouchline` (1501) - Touchline data
- `types.MessageCodeMarketDepth` (1502) - Market depth
- `types.MessageCodeIndexData` (1504) - Index data
- `types.MessageCodeCandleData` (1505) - Candle data
- `types.MessageCodeOpenInterest` (1510) - Open interest data

## Examples

The `examples/` directory contains comprehensive examples:

- [`examples/basic_usage/`](examples/basic_usage/) - Basic API usage examples
- [`examples/websocket_streaming/`](examples/websocket_streaming/) - WebSocket streaming example

To run the examples:

```bash
# Set environment variables
export XTS_SECRET_KEY="your-secret-key"
export XTS_APP_KEY="your-app-key"
export XTS_CLIENT_ID="your-client-id"  # Optional

# Run basic usage example
go run examples/basic_usage/main.go

# Run WebSocket streaming example
go run examples/websocket_streaming/main.go
```

## Error Handling

The library provides comprehensive error handling:

```go
import "xts-go/pkg/types"

// API errors implement the error interface
if apiErr, ok := err.(*types.APIError); ok {
    fmt.Printf("API Error - Code: %s, Message: %s\n", 
        apiErr.Code, apiErr.Message)
}
```

## Environment Configuration

### Development
```go
cfg.SetEnvironment("development")
// Uses: https://developers.symphonyfintech.in
```

### Production
```go
cfg.SetEnvironment("production")
// Uses: https://xts-api.trading
```

### Custom
```go
cfg.BaseURL = "https://your-custom-url.com"
cfg.MarketDataBaseURL = "https://your-marketdata-url.com"
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This is an unofficial Go wrapper for the XTS Trading APIs. Please use it at your own risk and ensure you comply with all applicable regulations and terms of service when trading.

## Support

For issues related to this Go wrapper, please open an issue on GitHub.
For XTS API-related questions, please contact Symphony FinTech support.

## Related Links

- [XTS Market Data API Documentation](https://symphonyfintech.com/xts-market-data-front-end-api/)
- [XTS Trading API Documentation](https://symphonyfintech.com/xts-trading-front-end-api-v2/)
- [Symphony FinTech](https://symphonyfintech.com/)
