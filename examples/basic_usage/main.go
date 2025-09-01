package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hakaitech/xts-go/pkg/client"
	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/types"
)

// SimpleExample demonstrates basic usage of the XTS Go wrapper
func main() {
	// Read credentials from environment variables or set them directly
	secretKey := os.Getenv("XTS_SECRET_KEY")
	appKey := os.Getenv("XTS_APP_KEY")
	clientID := os.Getenv("XTS_CLIENT_ID") // Optional for investor clients

	if secretKey == "" || appKey == "" {
		log.Fatal("Please set XTS_SECRET_KEY and XTS_APP_KEY environment variables")
	}

	// Create configuration
	cfg := config.NewConfig(secretKey, appKey, clientID)
	cfg.SetEnvironment("development") // Use development environment
	cfg.Debug = true                   // Enable debug logging

	// Create XTS client
	xtsClient, err := client.NewXTSClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create XTS client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Trading API Operations
	fmt.Println("=== Trading API Example ===")
	if err := tradingExample(ctx, xtsClient); err != nil {
		log.Printf("Trading example failed: %v", err)
	}

	// Example 2: Market Data API Operations
	fmt.Println("\n=== Market Data API Example ===")
	if err := marketDataExample(ctx, xtsClient); err != nil {
		log.Printf("Market data example failed: %v", err)
	}

	// Example 3: Convenience Methods
	fmt.Println("\n=== Convenience Methods Example ===")
	if err := convenienceExample(ctx, xtsClient); err != nil {
		log.Printf("Convenience example failed: %v", err)
	}

	// Cleanup
	if err := xtsClient.LogoutFromBoth(ctx); err != nil {
		log.Printf("Logout failed: %v", err)
	}

	fmt.Println("Examples completed successfully!")
}

// tradingExample demonstrates trading API operations
func tradingExample(ctx context.Context, client *client.XTSClient) error {
	// Login to trading API
	if err := client.LoginToTrading(ctx); err != nil {
		return fmt.Errorf("trading login failed: %w", err)
	}
	fmt.Println("✓ Successfully logged into Trading API")

	// Get user profile
	profile, err := client.Trading.GetProfile(ctx)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}
	fmt.Printf("✓ User Profile: %s (%s)\n", profile.ClientName, profile.EmailID)

	// Get account balance (only for retail clients)
	if client.Trading.IsInvestorClient() {
		balance, err := client.Trading.GetBalance(ctx)
		if err != nil {
			fmt.Printf("  Note: Could not get balance: %v\n", err)
		} else {
			fmt.Printf("✓ Account Balance retrieved: %v\n", balance)
		}
	} else {
		fmt.Println("  Note: Balance API not available for dealer accounts")
	}

	// Get existing orders
	orders, err := client.Trading.GetOrders(ctx)
	if err != nil {
		fmt.Printf("  Note: Could not get orders: %v\n", err)
	} else {
		fmt.Printf("✓ Retrieved orders: %v\n", orders)
	}

	// Get positions
	positions, err := client.Trading.GetPositions(ctx)
	if err != nil {
		fmt.Printf("  Note: Could not get positions: %v\n", err)
	} else {
		fmt.Printf("✓ Retrieved positions: %v\n", positions)
	}

	return nil
}

// marketDataExample demonstrates market data API operations
func marketDataExample(ctx context.Context, client *client.XTSClient) error {
	// Login to market data API
	if err := client.LoginToMarketData(ctx); err != nil {
		return fmt.Errorf("market data login failed: %w", err)
	}
	fmt.Println("✓ Successfully logged into Market Data API")

	// Define some instruments (RELIANCE and TCS on NSE)
	instruments := []types.Instrument{
		{ExchangeSegment: 1, ExchangeInstrumentID: 2885}, // RELIANCE
		{ExchangeSegment: 1, ExchangeInstrumentID: 11536}, // TCS
	}

	// Get quotes
	quotes, err := client.MarketData.GetQuote(ctx, instruments, types.MessageCodeTouchline, "JSON")
	if err != nil {
		return fmt.Errorf("failed to get quotes: %w", err)
	}
	fmt.Printf("✓ Retrieved quotes: %v\n", quotes)

	// Get touchline data
	touchlineData, err := client.MarketData.GetTouchlineData(ctx, instruments)
	if err != nil {
		return fmt.Errorf("failed to get touchline data: %w", err)
	}
	fmt.Printf("✓ Retrieved touchline data for %d instruments\n", len(touchlineData))
	for i, data := range touchlineData {
		fmt.Printf("  Instrument %d: LTP=%.2f, Change=%.2f%%\n", 
			data.ExchangeInstrumentID, data.LastTradedPrice, data.PercentChange)
		if i >= 1 { // Show only first 2 for brevity
			break
		}
	}

	// Search for instruments
	searchResults, err := client.MarketData.SearchInstruments(ctx, "RELIANCE", types.ExchangeNSECM)
	if err != nil {
		fmt.Printf("  Note: Search failed: %v\n", err)
	} else {
		fmt.Printf("✓ Search results for 'RELIANCE': %v\n", searchResults)
	}

	// Subscribe to live data (this would normally be used with WebSocket)
	subResult, err := client.MarketData.SubscribeToTouchline(ctx, instruments[:1]) // Subscribe to just one instrument
	if err != nil {
		fmt.Printf("  Note: Subscription failed: %v\n", err)
	} else {
		fmt.Printf("✓ Subscribed to live data: %v\n", subResult)
	}

	return nil
}

// convenienceExample demonstrates convenience methods
func convenienceExample(ctx context.Context, client *client.XTSClient) error {
	// Ensure both APIs are logged in
	if !client.IsLoggedInToBoth() {
		if err := client.LoginToBoth(ctx); err != nil {
			return fmt.Errorf("failed to login to both APIs: %w", err)
		}
	}

	// Get live quote for a single instrument
	quote, err := client.GetLiveQuote(ctx, 1, 2885) // RELIANCE on NSE
	if err != nil {
		return fmt.Errorf("failed to get live quote: %w", err)
	}
	fmt.Printf("✓ Live quote for RELIANCE: LTP=%.2f, Volume=%d\n", 
		quote.LastTradedPrice, quote.VolumeTraded)

	// Search for an instrument
	searchResults, err := client.SearchInstrument(ctx, "TCS", types.ExchangeNSECM)
	if err != nil {
		fmt.Printf("  Note: Search failed: %v\n", err)
	} else {
		fmt.Printf("✓ Search results for 'TCS': %v\n", searchResults)
	}

	// Example of placing orders (commented out to avoid actual trades)
	fmt.Println("✓ Order placement examples (not executed):")
	fmt.Println("  - Market Buy Order: PlaceMarketOrder(ctx, NSECM, 2885, BUY, 1, MIS)")
	fmt.Println("  - Limit Sell Order: PlaceLimitOrder(ctx, NSECM, 2885, SELL, 1, 2500.0, MIS)")
	fmt.Println("  - Stop Loss Order: PlaceStopLossOrder(ctx, NSECM, 2885, SELL, 1, 2400.0, MIS)")

	/*
	// Uncomment to place actual orders (use with caution!)
	
	// Place a market buy order
	orderID, err := client.PlaceMarketOrder(ctx, types.ExchangeNSECM, 2885, types.OrderSideBuy, 1, types.ProductTypeMIS)
	if err != nil {
		fmt.Printf("  Market order failed: %v\n", err)
	} else {
		fmt.Printf("✓ Market order placed: Order ID = %.0f\n", orderID)
		
		// Cancel the order immediately
		if err := client.Trading.CancelOrder(ctx, orderID); err != nil {
			fmt.Printf("  Failed to cancel order: %v\n", err)
		} else {
			fmt.Printf("✓ Order cancelled successfully\n")
		}
	}
	*/

	return nil
}