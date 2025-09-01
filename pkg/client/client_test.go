package client

import (
	"fmt"
	"testing"

	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/interfaces"
)

// TestXTSClientImplementsInterface ensures our client implements the interface
func TestXTSClientImplementsInterface(t *testing.T) {
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify it implements the interface
	var _ interfaces.XTSClient = client
}

// TestClientCreation tests different ways to create the client
func TestClientCreation(t *testing.T) {
	tests := []struct {
		name        string
		createFunc  func() (*XTSClient, error)
		expectError bool
	}{
		{
			name: "WithCredentials",
			createFunc: func() (*XTSClient, error) {
				return NewXTSClientWithCredentials("secret", "app", "client")
			},
			expectError: false,
		},
		{
			name: "WithConfig",
			createFunc: func() (*XTSClient, error) {
				cfg := config.NewConfig("secret", "app", "client")
				return NewXTSClientWithConfig(cfg)
			},
			expectError: false,
		},
		{
			name: "InvalidConfig",
			createFunc: func() (*XTSClient, error) {
				cfg := config.NewConfig("", "", "") // Invalid config
				return NewXTSClient(cfg)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := tt.createFunc()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Error("Expected client but got nil")
				}
			}
		})
	}
}

// TestInterfaceAccess tests accessing components through interfaces
func TestInterfaceAccess(t *testing.T) {
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test interface access
	tradingAPI := client.GetTradingAPI()
	if tradingAPI == nil {
		t.Error("Expected trading API interface but got nil")
	}

	marketAPI := client.GetMarketDataAPI()
	if marketAPI == nil {
		t.Error("Expected market data API interface but got nil")
	}

	// Verify interfaces have the expected methods
	if !tradingAPI.IsLoggedIn() {
		// This is expected for a test client that hasn't logged in
	}

	if !marketAPI.IsLoggedIn() {
		// This is expected for a test client that hasn't logged in
	}
}

// TestConfiguration tests configuration methods
func TestConfiguration(t *testing.T) {
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test environment setting
	client.SetEnvironment("development")
	
	// Test debug setting
	client.SetDebug(true)
	client.SetDebug(false)

	// Test config retrieval
	retrievedConfig := client.GetConfig()
	if retrievedConfig == nil {
		t.Error("Expected config but got nil")
	}

	// Verify it's a copy (not the same reference)
	if retrievedConfig == cfg {
		t.Error("Expected config copy but got same reference")
	}
}

// TestLoginStatus tests login status methods
func TestLoginStatus(t *testing.T) {
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Initially should not be logged in
	if client.IsLoggedInToTrading() {
		t.Error("Expected not logged in to trading initially")
	}

	if client.IsLoggedInToMarketData() {
		t.Error("Expected not logged in to market data initially")
	}

	if client.IsLoggedInToBoth() {
		t.Error("Expected not logged in to both initially")
	}
}

// TestOrderCreation demonstrates how to use the library for order creation
func TestOrderCreation(t *testing.T) {
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tradingAPI := client.GetTradingAPI()

	// Test order creation
	order := tradingAPI.NewOrder(
		"NSECM",     // exchange segment
		2885,        // instrument ID (RELIANCE)
		"MIS",       // product type
		"MARKET",    // order type
		"BUY",       // order side
		"DAY",       // time in force
		1,           // quantity
		0,           // limit price (not needed for market order)
		0,           // stop price (not needed for market order)
		0,           // disclosed quantity
		"test-order", // order unique identifier
	)

	if order == nil {
		t.Error("Expected order but got nil")
	}

	if order.ExchangeSegment != "NSECM" {
		t.Errorf("Expected exchange segment NSECM but got %s", order.ExchangeSegment)
	}

	if order.ExchangeInstrumentID != 2885 {
		t.Errorf("Expected instrument ID 2885 but got %d", order.ExchangeInstrumentID)
	}

	if order.OrderQuantity != 1 {
		t.Errorf("Expected quantity 1 but got %d", order.OrderQuantity)
	}
}

// Test types for demonstration
type TradingStrategy interface {
	GetName() string
	ShouldBuy(price float64) bool
	ShouldSell(price float64) bool
}

type SimpleStrategy struct {
	name          string
	buyThreshold  float64
	sellThreshold float64
}

func (s *SimpleStrategy) GetName() string {
	return s.name
}

func (s *SimpleStrategy) ShouldBuy(price float64) bool {
	return price < s.buyThreshold
}

func (s *SimpleStrategy) ShouldSell(price float64) bool {
	return price > s.sellThreshold
}

// TestExtensibilityExample demonstrates how users can extend the library
func TestExtensibilityExample(t *testing.T) {
	// Custom trading bot using the XTS library
	type TradingBot struct {
		client   interfaces.XTSClient
		strategy TradingStrategy
	}

	// Test the extensibility
	cfg := config.NewConfig("test-secret", "test-app", "test-client")
	client, err := NewXTSClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	strategy := &SimpleStrategy{
		name:          "Test Strategy",
		buyThreshold:  2400.0,
		sellThreshold: 2600.0,
	}

	bot := &TradingBot{
		client:   client,
		strategy: strategy,
	}

	// Test the bot
	if bot.strategy.GetName() != "Test Strategy" {
		t.Error("Strategy name mismatch")
	}

	if bot.client.IsLoggedInToBoth() {
		t.Error("Bot should not be ready without login")
	}

	if !bot.strategy.ShouldBuy(2300.0) {
		t.Error("Should buy at price below threshold")
	}

	if !bot.strategy.ShouldSell(2700.0) {
		t.Error("Should sell at price above threshold")
	}
}

// BenchmarkClientCreation benchmarks client creation
func BenchmarkClientCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := config.NewConfig("test-secret", "test-app", "test-client")
		_, err := NewXTSClient(cfg)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
	}
}

// ExampleXTSClient demonstrates basic usage
func ExampleXTSClient() {
	// Create configuration
	cfg := config.NewConfig("your-secret", "your-app-key", "your-client-id")
	cfg.SetEnvironment("development")

	// Create client
	client, err := NewXTSClient(cfg)
	if err != nil {
		panic(err)
	}

	// Demonstrate client creation and basic operations (without actual login)
	fmt.Printf("Created XTS client successfully\n")
	fmt.Printf("Trading API available: %t\n", client.GetTradingAPI() != nil)
	fmt.Printf("Market Data API available: %t\n", client.GetMarketDataAPI() != nil)
	fmt.Printf("Configuration set: %t\n", client.GetConfig() != nil)
	
	// Note: In real usage, you would login before operations:
	// ctx := context.Background()
	// if err := client.LoginToBoth(ctx); err != nil {
	//     panic(err)
	// }
	// defer client.LogoutFromBoth(ctx)

	// Output: Created XTS client successfully
	// Trading API available: true
	// Market Data API available: true
	// Configuration set: true
}

// ExampleXTSClient_extensibility demonstrates how to extend the library
func ExampleXTSClient_extensibility() {
	// Custom wrapper around the XTS client
	type MyTradingApp struct {
		xts interfaces.XTSClient
	}

	// Usage
	cfg := config.NewConfig("secret", "app", "client")
	xtsClient, _ := NewXTSClient(cfg)
	
	myApp := &MyTradingApp{xts: xtsClient}
	_ = myApp // Use the custom app
	
	fmt.Println("Extension example completed")
	// Output: Extension example completed
}