package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hakaitech/xts-go/pkg/client"
	"github.com/hakaitech/xts-go/pkg/config"
	"github.com/hakaitech/xts-go/pkg/errors"
	"github.com/hakaitech/xts-go/pkg/interfaces"
	"github.com/hakaitech/xts-go/pkg/types"
)

// TradingBot demonstrates how to build a trading bot using the XTS library
type TradingBot struct {
	client       interfaces.XTSClient
	tradingAPI   interfaces.TradingAPI
	marketAPI    interfaces.MarketDataAPI
	instruments  []types.Instrument
	strategy     TradingStrategy
}

// TradingStrategy defines the interface for trading strategies
type TradingStrategy interface {
	ShouldBuy(ctx context.Context, quote *types.TouchlineData) (bool, int64, error)
	ShouldSell(ctx context.Context, quote *types.TouchlineData, position *types.Position) (bool, int64, error)
	GetName() string
}

// SimpleMovingAverageStrategy is a sample strategy implementation
type SimpleMovingAverageStrategy struct {
	name      string
	threshold float64
}

func (s *SimpleMovingAverageStrategy) GetName() string {
	return s.name
}

func (s *SimpleMovingAverageStrategy) ShouldBuy(ctx context.Context, quote *types.TouchlineData) (bool, int64, error) {
	// Simple example: buy if price dropped by threshold percentage
	if quote.PercentChange <= -s.threshold {
		return true, 1, nil // Buy 1 quantity
	}
	return false, 0, nil
}

func (s *SimpleMovingAverageStrategy) ShouldSell(ctx context.Context, quote *types.TouchlineData, position *types.Position) (bool, int64, error) {
	// Simple example: sell if price increased by threshold percentage from average price
	avgPrice := position.BuyAveragePrice
	if avgPrice > 0 && ((quote.LastTradedPrice-avgPrice)/avgPrice)*100 >= s.threshold {
		return true, position.NetQuantity, nil // Sell all
	}
	return false, 0, nil
}

// NewTradingBot creates a new trading bot
func NewTradingBot(xtsClient interfaces.XTSClient, instruments []types.Instrument, strategy TradingStrategy) *TradingBot {
	return &TradingBot{
		client:      xtsClient,
		tradingAPI:  xtsClient.GetTradingAPI(),
		marketAPI:   xtsClient.GetMarketDataAPI(),
		instruments: instruments,
		strategy:    strategy,
	}
}

// Start begins the trading bot execution
func (bot *TradingBot) Start(ctx context.Context) error {
	log.Printf("Starting trading bot with strategy: %s", bot.strategy.GetName())

	// Ensure we're logged in
	if !bot.client.IsLoggedInToBoth() {
		if err := bot.client.LoginToBoth(ctx); err != nil {
			return fmt.Errorf("failed to login: %w", err)
		}
	}

	// Main trading loop
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := bot.processTradingCycle(ctx); err != nil {
				log.Printf("Trading cycle error: %v", err)
				
				// Handle different error types
				if xtsErr, ok := err.(*errors.XTSError); ok {
					if xtsErr.IsAuthError() {
						log.Println("Authentication error detected, attempting re-login...")
						if err := bot.client.LoginToBoth(ctx); err != nil {
							log.Printf("Re-login failed: %v", err)
							return err
						}
					}
					
					if !xtsErr.IsRetryable() {
						log.Printf("Non-retryable error: %v", err)
						return err
					}
				}
			}
		}
	}
}

func (bot *TradingBot) processTradingCycle(ctx context.Context) error {
	// Get current positions
	positionsInterface, err := bot.tradingAPI.GetPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	// Since the API returns interface{}, we'll work with it as interface{}
	// In a real implementation, you'd implement proper type assertions or JSON unmarshaling

	// Process each instrument
	for _, instrument := range bot.instruments {
		if err := bot.processInstrument(ctx, instrument, positionsInterface); err != nil {
			log.Printf("Error processing instrument %d: %v", instrument.ExchangeInstrumentID, err)
		}
	}

	return nil
}

func (bot *TradingBot) processInstrument(ctx context.Context, instrument types.Instrument, positions interface{}) error {
	// Get current quote
	quote, err := bot.client.GetLiveQuote(ctx, instrument.ExchangeSegment, instrument.ExchangeInstrumentID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	log.Printf("Processing %d: LTP=%.2f, Change=%.2f%%", 
		instrument.ExchangeInstrumentID, quote.LastTradedPrice, quote.PercentChange)

	// For demonstration purposes, we'll create a mock position
	// In a real implementation, you'd parse the positions interface{} properly
	var currentPosition *types.Position
	_ = positions // positions would be processed here in real implementation

	// Mock position for demonstration (normally you'd parse positions)
	// currentPosition = findPositionInInterface(positions, instrument.ExchangeInstrumentID)

	// Check for sell signal first (if we have a position)
	if currentPosition != nil && currentPosition.NetQuantity > 0 {
		shouldSell, quantity, err := bot.strategy.ShouldSell(ctx, quote, currentPosition)
		if err != nil {
			return fmt.Errorf("strategy sell check failed: %w", err)
		}

		if shouldSell {
			log.Printf("Sell signal detected for %d, quantity: %d", instrument.ExchangeInstrumentID, quantity)
			
			// Place sell order
			orderID, err := bot.client.PlaceMarketOrder(ctx, 
				types.ExchangeNSECM, 
				int64(instrument.ExchangeInstrumentID), 
				types.OrderSideSell, 
				quantity, 
				types.ProductTypeMIS)
			
			if err != nil {
				log.Printf("Failed to place sell order: %v", err)
			} else {
				log.Printf("Sell order placed successfully: %.0f", orderID)
			}
			
			return nil // Don't check buy signal if we just sold
		}
	}

	// Check for buy signal
	shouldBuy, quantity, err := bot.strategy.ShouldBuy(ctx, quote)
	if err != nil {
		return fmt.Errorf("strategy buy check failed: %w", err)
	}

	if shouldBuy {
		log.Printf("Buy signal detected for %d, quantity: %d", instrument.ExchangeInstrumentID, quantity)
		
		// Place buy order
		orderID, err := bot.client.PlaceMarketOrder(ctx, 
			types.ExchangeNSECM, 
			int64(instrument.ExchangeInstrumentID), 
			types.OrderSideBuy, 
			quantity, 
			types.ProductTypeMIS)
		
		if err != nil {
			log.Printf("Failed to place buy order: %v", err)
		} else {
			log.Printf("Buy order placed successfully: %.0f", orderID)
		}
	}

	return nil
}

// CustomMarketDataAnalyzer demonstrates extending market data functionality
type CustomMarketDataAnalyzer struct {
	marketAPI interfaces.MarketDataAPI
	cache     map[int][]types.TouchlineData // Simple price history cache
}

func NewCustomMarketDataAnalyzer(marketAPI interfaces.MarketDataAPI) *CustomMarketDataAnalyzer {
	return &CustomMarketDataAnalyzer{
		marketAPI: marketAPI,
		cache:     make(map[int][]types.TouchlineData),
	}
}

// CalculateVolatility calculates a simple volatility measure
func (analyzer *CustomMarketDataAnalyzer) CalculateVolatility(ctx context.Context, instrumentID int, periods int) (float64, error) {
	history, exists := analyzer.cache[instrumentID]
	if !exists || len(history) < periods {
		return 0, fmt.Errorf("insufficient data for volatility calculation")
	}

	// Simple volatility calculation using standard deviation of returns
	recentHistory := history[len(history)-periods:]
	var returns []float64
	
	for i := 1; i < len(recentHistory); i++ {
		if recentHistory[i-1].LastTradedPrice > 0 {
			ret := (recentHistory[i].LastTradedPrice - recentHistory[i-1].LastTradedPrice) / recentHistory[i-1].LastTradedPrice
			returns = append(returns, ret)
		}
	}

	if len(returns) == 0 {
		return 0, fmt.Errorf("no valid returns calculated")
	}

	// Calculate mean
	var sum float64
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate standard deviation
	var variance float64
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	variance /= float64(len(returns))

	return variance, nil // Return variance as volatility measure
}

// AddDataPoint adds a new data point to the analyzer
func (analyzer *CustomMarketDataAnalyzer) AddDataPoint(instrumentID int, data types.TouchlineData) {
	if analyzer.cache[instrumentID] == nil {
		analyzer.cache[instrumentID] = make([]types.TouchlineData, 0)
	}
	
	analyzer.cache[instrumentID] = append(analyzer.cache[instrumentID], data)
	
	// Keep only last 100 data points
	if len(analyzer.cache[instrumentID]) > 100 {
		analyzer.cache[instrumentID] = analyzer.cache[instrumentID][1:]
	}
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

	// Example 1: Using the library as an interface
	fmt.Println("=== Interface Usage Example ===")
	interfaceExample(ctx, xtsClient)

	// Example 2: Custom Market Data Analyzer
	fmt.Println("\n=== Custom Market Data Analyzer Example ===")
	analyzerExample(ctx, xtsClient)

	// Example 3: Trading Bot (commented out to avoid actual trading)
	fmt.Println("\n=== Trading Bot Example (Simulation Only) ===")
	tradingBotExample(ctx, xtsClient)

	// Cleanup
	if err := xtsClient.LogoutFromBoth(ctx); err != nil {
		log.Printf("Logout failed: %v", err)
	}

	fmt.Println("Extension examples completed!")
}

func interfaceExample(ctx context.Context, xtsClient interfaces.XTSClient) {
	// Demonstrate how the library can be used through interfaces
	if err := xtsClient.LoginToBoth(ctx); err != nil {
		log.Printf("Login failed: %v", err)
		return
	}

	// Get API components through interfaces
	tradingAPI := xtsClient.GetTradingAPI()
	marketAPI := xtsClient.GetMarketDataAPI()

	fmt.Printf("✓ Obtained TradingAPI interface: %T\n", tradingAPI)
	fmt.Printf("✓ Obtained MarketDataAPI interface: %T\n", marketAPI)

	// Use the interfaces
	if tradingAPI.IsLoggedIn() {
		fmt.Println("✓ Trading API is logged in")
	}

	if marketAPI.IsLoggedIn() {
		fmt.Println("✓ Market Data API is logged in")
	}
}

func analyzerExample(ctx context.Context, xtsClient interfaces.XTSClient) {
	if !xtsClient.IsLoggedInToMarketData() {
		if err := xtsClient.LoginToMarketData(ctx); err != nil {
			log.Printf("Failed to login to market data: %v", err)
			return
		}
	}

	// Create custom analyzer
	analyzer := NewCustomMarketDataAnalyzer(xtsClient.GetMarketDataAPI())

	// Simulate adding some data points
	instrumentID := 2885 // RELIANCE
	for i := 0; i < 10; i++ {
		data := types.TouchlineData{
			ExchangeInstrumentID: instrumentID,
			LastTradedPrice:      2500.0 + float64(i)*10, // Simulated increasing price
		}
		analyzer.AddDataPoint(instrumentID, data)
	}

	// Try to calculate volatility
	volatility, err := analyzer.CalculateVolatility(ctx, instrumentID, 5)
	if err != nil {
		log.Printf("Volatility calculation failed: %v", err)
	} else {
		fmt.Printf("✓ Calculated volatility for instrument %d: %.6f\n", instrumentID, volatility)
	}
}

func tradingBotExample(ctx context.Context, xtsClient interfaces.XTSClient) {
	// Define instruments to trade
	instruments := []types.Instrument{
		{ExchangeSegment: 1, ExchangeInstrumentID: 2885}, // RELIANCE
	}

	// Create a simple strategy
	strategy := &SimpleMovingAverageStrategy{
		name:      "Simple Threshold Strategy",
		threshold: 2.0, // 2% threshold
	}

	// Create trading bot
	bot := NewTradingBot(xtsClient, instruments, strategy)

	fmt.Printf("✓ Created trading bot with strategy: %s\n", strategy.GetName())
	fmt.Printf("✓ Monitoring %d instruments\n", len(instruments))
	fmt.Printf("✓ Bot created successfully: %T\n", bot)
	fmt.Printf("✓ Monitoring %d instruments\n", len(instruments))
	fmt.Println("  Note: Actual trading disabled for safety - this is a demonstration")

	// In a real scenario, you would run:
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// defer cancel()
	// if err := bot.Start(ctx); err != nil {
	//     log.Printf("Bot stopped: %v", err)
	// }
}