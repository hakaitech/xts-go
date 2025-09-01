package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"xts-go/pkg/client"
	"xts-go/pkg/config"
	"xts-go/pkg/types"
	"xts-go/pkg/websocket"
)

// CustomEventHandler implements the WebSocket event handler interface
type CustomEventHandler struct {
	*websocket.DefaultEventHandler
}

func (h *CustomEventHandler) OnConnect() {
	fmt.Println("🔗 WebSocket Connected!")
}

func (h *CustomEventHandler) OnDisconnect() {
	fmt.Println("❌ WebSocket Disconnected!")
}

func (h *CustomEventHandler) OnError(err error) {
	fmt.Printf("⚠️  WebSocket Error: %v\n", err)
}

func (h *CustomEventHandler) OnMessage(data []byte) {
	fmt.Printf("📨 Raw Message: %s\n", string(data))
}

func (h *CustomEventHandler) OnTouchlineData(data *types.TouchlineData) {
	fmt.Printf("📈 Touchline Update - Instrument: %d, LTP: %.2f, Change: %.2f%%, Volume: %d\n",
		data.ExchangeInstrumentID, data.LastTradedPrice, data.PercentChange, data.VolumeTraded)
}

func (h *CustomEventHandler) OnMarketDepthData(data interface{}) {
	fmt.Printf("📊 Market Depth Update: %v\n", data)
}

func (h *CustomEventHandler) OnIndexData(data interface{}) {
	fmt.Printf("📈 Index Update: %v\n", data)
}

func (h *CustomEventHandler) OnCandleData(data interface{}) {
	fmt.Printf("🕯️  Candle Update: %v\n", data)
}

func (h *CustomEventHandler) OnOpenInterestData(data interface{}) {
	fmt.Printf("📊 Open Interest Update: %v\n", data)
}

func main() {
	// Read credentials from environment variables
	secretKey := os.Getenv("XTS_SECRET_KEY")
	appKey := os.Getenv("XTS_APP_KEY")
	clientID := os.Getenv("XTS_CLIENT_ID")

	if secretKey == "" || appKey == "" {
		log.Fatal("Please set XTS_SECRET_KEY and XTS_APP_KEY environment variables")
	}

	// Create configuration
	cfg := config.NewConfig(secretKey, appKey, clientID)
	cfg.SetEnvironment("development")
	cfg.Debug = true

	// Create XTS client
	xtsClient, err := client.NewXTSClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create XTS client: %v", err)
	}

	ctx := context.Background()

	// Login to market data API (required for WebSocket streaming)
	fmt.Println("Logging into Market Data API...")
	if err := xtsClient.LoginToMarketData(ctx); err != nil {
		log.Fatalf("Market data login failed: %v", err)
	}
	fmt.Println("✓ Successfully logged into Market Data API")

	// Create WebSocket client with custom event handler
	handler := &CustomEventHandler{}
	wsClient := websocket.NewClient(cfg, handler)

	// Define instruments to stream (RELIANCE and TCS)
	instruments := []types.Instrument{
		{ExchangeSegment: 1, ExchangeInstrumentID: 2885},  // RELIANCE
		{ExchangeSegment: 1, ExchangeInstrumentID: 11536}, // TCS
	}

	// Connect to WebSocket (Note: You'll need the actual WebSocket URL from XTS)
	// This is a placeholder URL - replace with actual XTS WebSocket endpoint
	wsURL := "wss://developers.symphonyfintech.in/marketdata/socket.io"
	token := xtsClient.MarketData.GetSessionToken()

	fmt.Printf("Connecting to WebSocket: %s\n", wsURL)
	if err := wsClient.Connect(ctx, wsURL, token); err != nil {
		log.Fatalf("WebSocket connection failed: %v", err)
	}

	// Wait a moment for connection to establish
	time.Sleep(2 * time.Second)

	// Subscribe to touchline data
	fmt.Println("Subscribing to touchline data...")
	if err := wsClient.Subscribe(instruments, types.MessageCodeTouchline); err != nil {
		log.Printf("Failed to subscribe to touchline data: %v", err)
	} else {
		fmt.Println("✓ Subscribed to touchline data")
	}

	// Subscribe to market depth data
	fmt.Println("Subscribing to market depth data...")
	if err := wsClient.Subscribe(instruments[:1], types.MessageCodeMarketDepth); err != nil {
		log.Printf("Failed to subscribe to market depth data: %v", err)
	} else {
		fmt.Println("✓ Subscribed to market depth data")
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n🚀 WebSocket streaming started! Press Ctrl+C to stop...")
	fmt.Println("📊 Waiting for real-time data updates...")
	fmt.Println(strings.Repeat("=", 50))

	// Keep the program running and show periodic status
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n🛑 Shutdown signal received...")
			goto cleanup

		case <-ticker.C:
			if wsClient.IsConnected() {
				fmt.Println("💚 WebSocket connection is healthy")
			} else {
				fmt.Println("💔 WebSocket connection lost")
			}

		case <-time.After(5 * time.Minute): // Auto-stop after 5 minutes for demo
			fmt.Println("\n⏰ Demo time limit reached (5 minutes)")
			goto cleanup
		}
	}

cleanup:
	fmt.Println("🧹 Cleaning up...")

	// Unsubscribe from all data
	if wsClient.IsConnected() {
		fmt.Println("Unsubscribing from data streams...")
		if err := wsClient.Unsubscribe(instruments, types.MessageCodeTouchline); err != nil {
			fmt.Printf("Failed to unsubscribe from touchline: %v\n", err)
		}
		if err := wsClient.Unsubscribe(instruments[:1], types.MessageCodeMarketDepth); err != nil {
			fmt.Printf("Failed to unsubscribe from market depth: %v\n", err)
		}
	}

	// Disconnect WebSocket
	if err := wsClient.Disconnect(); err != nil {
		fmt.Printf("WebSocket disconnect error: %v\n", err)
	} else {
		fmt.Println("✓ WebSocket disconnected")
	}

	// Logout from market data API
	if err := xtsClient.LogoutFromMarketData(ctx); err != nil {
		fmt.Printf("Market data logout error: %v\n", err)
	} else {
		fmt.Println("✓ Logged out from Market Data API")
	}

	fmt.Println("👋 WebSocket streaming example completed!")
}

// showUsageInstructions displays usage instructions
func showUsageInstructions() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📖 XTS WebSocket Streaming Example")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("This example demonstrates how to:")
	fmt.Println("1. Connect to XTS Market Data WebSocket")
	fmt.Println("2. Subscribe to real-time data streams")
	fmt.Println("3. Handle different types of market data events")
	fmt.Println("4. Gracefully disconnect and cleanup")
	fmt.Println()
	fmt.Println("Required Environment Variables:")
	fmt.Println("- XTS_SECRET_KEY: Your XTS API secret key")
	fmt.Println("- XTS_APP_KEY: Your XTS API app key")
	fmt.Println("- XTS_CLIENT_ID: Your client ID (optional for investor accounts)")
	fmt.Println()
	fmt.Println("Note: This is a demonstration. Replace the WebSocket URL")
	fmt.Println("with the actual XTS WebSocket endpoint provided by Symphony FinTech.")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
}

func init() {
	// Show usage instructions when the program starts
	showUsageInstructions()
}