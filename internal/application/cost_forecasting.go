package application

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ForecastPoint represents a single forecasted data point.
type ForecastPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	Value           float64   `json:"value"`
	UpperBound      float64   `json:"upper_bound"`
	LowerBound      float64   `json:"lower_bound"`
	ConfidenceLevel float64   `json:"confidence_level"` // 0.0-1.0
}

// ForecastResult contains forecast data and metadata.
type ForecastResult struct {
	Forecast       []ForecastPoint `json:"forecast"`
	Model          string          `json:"model"`
	Horizon        string          `json:"horizon"`
	BaseValue      float64         `json:"base_value"`
	TrendDirection string          `json:"trend_direction"` // increasing, decreasing, stable
	TrendStrength  float64         `json:"trend_strength"`  // 0.0-1.0
	Seasonality    bool            `json:"seasonality"`
	MAE            float64         `json:"mae,omitempty"`  // Mean Absolute Error
	RMSE           float64         `json:"rmse,omitempty"` // Root Mean Square Error
}

// TimeSeriesPoint represents a single observation in time series.
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// CostForecastingService provides cost prediction capabilities.
type CostForecastingService struct {
	// Configuration
	defaultHorizon     int     // Number of periods to forecast
	smoothingAlpha     float64 // Exponential smoothing parameter (0-1)
	trendBeta          float64 // Trend smoothing parameter (0-1)
	seasonalityGamma   float64 // Seasonality smoothing parameter (0-1)
	confidenceInterval float64 // Confidence interval (e.g., 0.95 for 95%)
}

// NewCostForecastingService creates a new forecasting service.
func NewCostForecastingService() *CostForecastingService {
	return &CostForecastingService{
		defaultHorizon:     30,   // 30 periods ahead
		smoothingAlpha:     0.3,  // Weight for level
		trendBeta:          0.1,  // Weight for trend
		seasonalityGamma:   0.05, // Weight for seasonality
		confidenceInterval: 0.95, // 95% confidence
	}
}

// ForecastOperationsCost predicts future operation costs based on historical data.
func (s *CostForecastingService) ForecastOperationsCost(history []TimeSeriesPoint, horizon int) (*ForecastResult, error) {
	if len(history) < 2 {
		return nil, fmt.Errorf("insufficient historical data: need at least 2 points, got %d", len(history))
	}

	if horizon <= 0 {
		horizon = s.defaultHorizon
	}

	// Sort by timestamp
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.Before(history[j].Timestamp)
	})

	// Calculate time interval between points
	intervals := make([]time.Duration, len(history)-1)
	for i := range len(history) - 1 {
		intervals[i] = history[i+1].Timestamp.Sub(history[i].Timestamp)
	}
	avgInterval := s.averageDuration(intervals)

	// Extract values
	values := make([]float64, len(history))
	for i, point := range history {
		values[i] = point.Value
	}

	// Detect trend and seasonality
	trend := s.detectTrend(values)
	seasonalPeriod := s.detectSeasonality(values)

	// Choose forecasting method based on data characteristics
	var forecast []ForecastPoint
	var model string

	switch {
	case seasonalPeriod > 0 && len(values) >= seasonalPeriod*2:
		// Use Holt-Winters (triple exponential smoothing) for seasonal data
		forecast = s.holtWintersForecasting(history, horizon, avgInterval, seasonalPeriod)
		model = "holt-winters"
	case trend.strength > 0.3:
		// Use Double Exponential Smoothing (Holt's method) for trended data
		forecast = s.doubleExponentialSmoothing(history, horizon, avgInterval, trend.slope)
		model = "double-exponential-smoothing"
	default:
		// Use Simple Exponential Smoothing for stable data
		forecast = s.simpleExponentialSmoothing(history, horizon, avgInterval)
		model = "simple-exponential-smoothing"
	}

	// Calculate forecast accuracy metrics (on training data)
	mae, rmse := s.calculateAccuracyMetrics(history, values)

	result := &ForecastResult{
		Forecast:       forecast,
		Model:          model,
		Horizon:        fmt.Sprintf("%d periods", horizon),
		BaseValue:      values[len(values)-1],
		TrendDirection: trend.direction,
		TrendStrength:  trend.strength,
		Seasonality:    seasonalPeriod > 0,
		MAE:            mae,
		RMSE:           rmse,
	}

	return result, nil
}

// trendInfo contains trend analysis results.
type trendInfo struct {
	direction string
	slope     float64
	strength  float64 // 0.0-1.0
}

// detectTrend analyzes the trend in time series data.
func (s *CostForecastingService) detectTrend(values []float64) trendInfo {
	n := len(values)
	if n < 2 {
		return trendInfo{direction: "stable", slope: 0, strength: 0}
	}

	// Calculate linear regression slope using least squares
	var sumX, sumY, sumXY, sumXX float64
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	nFloat := float64(n)
	slope := (nFloat*sumXY - sumX*sumY) / (nFloat*sumXX - sumX*sumX)

	// Calculate R-squared to measure trend strength
	meanY := sumY / nFloat
	var ssTotal, ssResidual float64
	for i, y := range values {
		predicted := slope*float64(i) + (sumY-slope*sumX)/nFloat
		ssTotal += (y - meanY) * (y - meanY)
		ssResidual += (y - predicted) * (y - predicted)
	}

	rSquared := 1.0 - (ssResidual / ssTotal)
	if ssTotal == 0 {
		rSquared = 0
	}

	// Normalize slope for strength calculation
	absSlope := math.Abs(slope)
	normalizedSlope := absSlope / (meanY + 1.0) // Relative to mean value

	strength := math.Min(rSquared*normalizedSlope, 1.0)

	direction := "stable"
	if slope > 0.01 {
		direction = "increasing"
	} else if slope < -0.01 {
		direction = "decreasing"
	}

	return trendInfo{
		direction: direction,
		slope:     slope,
		strength:  strength,
	}
}

// detectSeasonality detects seasonal patterns in the data.
func (s *CostForecastingService) detectSeasonality(values []float64) int {
	n := len(values)
	if n < 14 { // Need at least 2 weeks for weekly seasonality
		return 0
	}

	// Test common seasonal periods: 7 (weekly), 24 (daily if hourly data), 30 (monthly)
	testPeriods := []int{7, 24, 30}
	maxCorrelation := 0.0
	bestPeriod := 0

	for _, period := range testPeriods {
		if n < period*2 {
			continue
		}

		// Calculate autocorrelation at lag = period
		correlation := s.autocorrelation(values, period)
		if correlation > maxCorrelation {
			maxCorrelation = correlation
			bestPeriod = period
		}
	}

	// Require correlation > 0.6 to consider it seasonal
	if maxCorrelation > 0.6 {
		return bestPeriod
	}

	return 0
}

// autocorrelation calculates autocorrelation at a given lag.
func (s *CostForecastingService) autocorrelation(values []float64, lag int) float64 {
	n := len(values)
	if lag >= n {
		return 0
	}

	// Calculate mean
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(n)

	// Calculate autocorrelation
	var numerator, denominator float64
	for i := range n - lag {
		numerator += (values[i] - mean) * (values[i+lag] - mean)
	}
	for i := range n {
		denominator += (values[i] - mean) * (values[i] - mean)
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// simpleExponentialSmoothing performs basic exponential smoothing.
func (s *CostForecastingService) simpleExponentialSmoothing(history []TimeSeriesPoint, horizon int, interval time.Duration) []ForecastPoint {
	values := make([]float64, len(history))
	for i, point := range history {
		values[i] = point.Value
	}

	// Initialize level with first value
	level := values[0]

	// Apply exponential smoothing to get final level
	for i := 1; i < len(values); i++ {
		level = s.smoothingAlpha*values[i] + (1-s.smoothingAlpha)*level
	}

	// Generate forecast (flat line)
	forecast := make([]ForecastPoint, horizon)
	lastTime := history[len(history)-1].Timestamp

	// Calculate prediction intervals based on historical variance
	variance := s.calculateVariance(values)
	stdDev := math.Sqrt(variance)
	zScore := 1.96 // 95% confidence interval

	for i := range horizon {
		forecastTime := lastTime.Add(interval * time.Duration(i+1))

		// Widen confidence interval as we forecast further
		widthFactor := math.Sqrt(float64(i + 1))
		margin := zScore * stdDev * widthFactor

		forecast[i] = ForecastPoint{
			Timestamp:       forecastTime,
			Value:           level,
			UpperBound:      level + margin,
			LowerBound:      math.Max(0, level-margin),
			ConfidenceLevel: s.confidenceInterval,
		}
	}

	return forecast
}

// doubleExponentialSmoothing performs Holt's linear trend method.
func (s *CostForecastingService) doubleExponentialSmoothing(history []TimeSeriesPoint, horizon int, interval time.Duration, initialSlope float64) []ForecastPoint {
	values := make([]float64, len(history))
	for i, point := range history {
		values[i] = point.Value
	}

	// Initialize level and trend
	level := values[0]
	trend := initialSlope

	// Apply double exponential smoothing
	for i := 1; i < len(values); i++ {
		prevLevel := level
		level = s.smoothingAlpha*values[i] + (1-s.smoothingAlpha)*(level+trend)
		trend = s.trendBeta*(level-prevLevel) + (1-s.trendBeta)*trend
	}

	// Generate forecast with trend
	forecast := make([]ForecastPoint, horizon)
	lastTime := history[len(history)-1].Timestamp

	variance := s.calculateVariance(values)
	stdDev := math.Sqrt(variance)
	zScore := 1.96

	for i := range horizon {
		forecastTime := lastTime.Add(interval * time.Duration(i+1))
		h := float64(i + 1)

		forecastValue := level + h*trend

		// Confidence interval widens with forecast horizon
		widthFactor := math.Sqrt(h * (1 + h*s.trendBeta*s.trendBeta))
		margin := zScore * stdDev * widthFactor

		forecast[i] = ForecastPoint{
			Timestamp:       forecastTime,
			Value:           forecastValue,
			UpperBound:      forecastValue + margin,
			LowerBound:      math.Max(0, forecastValue-margin),
			ConfidenceLevel: s.confidenceInterval,
		}
	}

	return forecast
}

// holtWintersForecasting performs triple exponential smoothing with seasonality.
func (s *CostForecastingService) holtWintersForecasting(history []TimeSeriesPoint, horizon int, interval time.Duration, seasonalPeriod int) []ForecastPoint {
	values := make([]float64, len(history))
	for i, point := range history {
		values[i] = point.Value
	}

	n := len(values)

	// Initialize level, trend, and seasonal components
	level := values[0]
	trend := 0.0
	seasonal := make([]float64, seasonalPeriod)

	// Initialize seasonal indices
	for i := 0; i < seasonalPeriod && i < n; i++ {
		seasonal[i] = values[i] / (level + 1e-10)
	}

	// Apply Holt-Winters smoothing
	for i := 1; i < n; i++ {
		seasonalIdx := i % seasonalPeriod

		prevLevel := level
		deseasonalized := values[i] / (seasonal[seasonalIdx] + 1e-10)
		level = s.smoothingAlpha*deseasonalized + (1-s.smoothingAlpha)*(level+trend)
		trend = s.trendBeta*(level-prevLevel) + (1-s.trendBeta)*trend
		seasonal[seasonalIdx] = s.seasonalityGamma*(values[i]/(level+1e-10)) + (1-s.seasonalityGamma)*seasonal[seasonalIdx]
	}

	// Generate forecast with trend and seasonality
	forecast := make([]ForecastPoint, horizon)
	lastTime := history[len(history)-1].Timestamp

	variance := s.calculateVariance(values)
	stdDev := math.Sqrt(variance)
	zScore := 1.96

	for i := range horizon {
		forecastTime := lastTime.Add(interval * time.Duration(i+1))
		h := float64(i + 1)
		seasonalIdx := (n + i) % seasonalPeriod

		forecastValue := (level + h*trend) * seasonal[seasonalIdx]

		// Account for trend and seasonal uncertainty
		widthFactor := math.Sqrt(h * (1 + h*s.trendBeta*s.trendBeta + s.seasonalityGamma*s.seasonalityGamma))
		margin := zScore * stdDev * widthFactor

		forecast[i] = ForecastPoint{
			Timestamp:       forecastTime,
			Value:           forecastValue,
			UpperBound:      forecastValue + margin,
			LowerBound:      math.Max(0, forecastValue-margin),
			ConfidenceLevel: s.confidenceInterval,
		}
	}

	return forecast
}

// calculateVariance computes variance of values.
func (s *CostForecastingService) calculateVariance(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= n

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= (n - 1)

	return variance
}

// calculateAccuracyMetrics calculates MAE and RMSE.
func (s *CostForecastingService) calculateAccuracyMetrics(history []TimeSeriesPoint, values []float64) (mae, rmse float64) {
	if len(values) < 2 {
		return 0, 0
	}

	// Use simple exponential smoothing to backtest
	level := values[0]
	var errors []float64

	for i := 1; i < len(values); i++ {
		predicted := level
		actual := values[i]
		error := math.Abs(actual - predicted)
		errors = append(errors, error)

		// Update level for next prediction
		level = s.smoothingAlpha*actual + (1-s.smoothingAlpha)*level
	}

	// Calculate MAE
	for _, err := range errors {
		mae += err
	}
	mae /= float64(len(errors))

	// Calculate RMSE
	for _, err := range errors {
		rmse += err * err
	}
	rmse = math.Sqrt(rmse / float64(len(errors)))

	return mae, rmse
}

// averageDuration calculates average duration.
func (s *CostForecastingService) averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return time.Hour
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

// ForecastTokenCost predicts future token usage costs.
func (s *CostForecastingService) ForecastTokenCost(history []TimeSeriesPoint, horizon int) (*ForecastResult, error) {
	// Use same forecasting logic as operations
	return s.ForecastOperationsCost(history, horizon)
}

// ForecastLatency predicts future latency trends.
func (s *CostForecastingService) ForecastLatency(history []TimeSeriesPoint, horizon int) (*ForecastResult, error) {
	// Use same forecasting logic
	return s.ForecastOperationsCost(history, horizon)
}

// ForecastErrorRate predicts future error rate trends.
func (s *CostForecastingService) ForecastErrorRate(history []TimeSeriesPoint, horizon int) (*ForecastResult, error) {
	// Use same forecasting logic
	return s.ForecastOperationsCost(history, horizon)
}
