# xts-go : Unofficial Go API Client for XTS Trading APIs

![License](https://img.shields.io/github/license/hakaitech/xts-go)
![GitHub issues](https://img.shields.io/github/issues-raw/hakaitech/xts-go)
![GitHub go.mod Go version (branch & subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/hakaitech/xts-go/main)

A comprehensive, modular Go wrapper for the XTS Trading APIs by Symphony FinTech. This **unofficial** library provides a clean, easy-to-use interface for both Market Data and Interactive (Trading) APIs with full WebSocket streaming support.

This project aims to expand XTS capabilities by leveraging Go's versatility and performance characteristics, making it ideal for building high-performance algorithmic trading systems, automated trading bots, and real-time market data processing applications.

> **Note**: This is an unofficial client library. For official documentation and support, please refer to [Symphony FinTech's official XTS API documentation](https://symphonyfintech.com/).

## Features

- [x] **Complete API Coverage**: Both Market Data and Interactive/Trading APIs
- [x] **WebSocket Streaming**: Real-time market data with automatic reconnection
- [x] **Type Safety**: Strongly typed Go structs for all API responses
- [x] **Authentication**: Automatic session management and token handling
- [x] **Host Lookup**: Support for XTS host lookup functionality
- [x] **Environment Support**: Easy switching between development, sandbox, and production environments
- [x] **Library Ready**: Designed to be used as a library for building trading applications and tools
- [x] **Extensible**: Clean interfaces and abstractions for easy extension and customization

## Installation

### Prerequisites

- Go 1.21 or higher
- XTS Trading account with API credentials from Symphony FinTech

### Install using go get

```bash
go get github.com/hakaitech/xts-go
```

### Install using go mod

Add to your `go.mod` file:

```go
require github.com/hakaitech/xts-go v1.0.0
```

Then run:

```bash
go mod tidy
```

### Verify Installation

```go
package main

import (
    "fmt"
    "github.com/hakaitech/xts-go/pkg/config"
)

func main() {
    cfg := config.NewConfig("your-secret", "your-app-key", "your-client-id")
    fmt.Printf("XTS Go client initialized: %+v\n", cfg != nil)
}
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/hakaitech/xts-go/pkg/client"
    "github.com/hakaitech/xts-go/pkg/config"
)

func main() {
    // Initialize configuration
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
    
    // Your trading/market data code here...
}
```

## Official XTS API Resources

- **Official XTS Documentation**: [https://symphonyfintech.com/](https://symphonyfintech.com/)
- **XTS Market Data API**: [https://symphonyfintech.com/xts-market-data-front-end-api/](https://symphonyfintech.com/xts-market-data-front-end-api/)
- **XTS Trading API**: [https://symphonyfintech.com/xts-trading-front-end-api-v2/](https://symphonyfintech.com/xts-trading-front-end-api-v2/)
- **Get XTS API Access**: Contact Symphony FinTech for API credentials and documentation
