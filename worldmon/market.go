package worldmon

import (
	"context"
	"encoding/json"
	"net/url"
)

// Market is GET /api/market/v1/…
type Market struct{ *Service }

// Market returns the market v1 service.
func (c *Client) Market() *Market { return &Market{Service: c.Service("market", "v1")} }

// ListMarketQuotes is GET /api/market/v1/list-market-quotes
func (m *Market) ListMarketQuotes(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-market-quotes", q)
}

// ListCryptoQuotes is GET /api/market/v1/list-crypto-quotes
func (m *Market) ListCryptoQuotes(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-crypto-quotes", q)
}

// ListCommodityQuotes is GET /api/market/v1/list-commodity-quotes
func (m *Market) ListCommodityQuotes(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-commodity-quotes", q)
}

// GetSectorSummary is GET /api/market/v1/get-sector-summary
func (m *Market) GetSectorSummary(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-sector-summary", q)
}

// ListStablecoinMarkets is GET /api/market/v1/list-stablecoin-markets
func (m *Market) ListStablecoinMarkets(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-stablecoin-markets", q)
}

// ListEtfFlows is GET /api/market/v1/list-etf-flows
func (m *Market) ListEtfFlows(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-etf-flows", q)
}

// GetCountryStockIndex is GET /api/market/v1/get-country-stock-index
func (m *Market) GetCountryStockIndex(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-country-stock-index", q)
}

// ListGulfQuotes is GET /api/market/v1/list-gulf-quotes
func (m *Market) ListGulfQuotes(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-gulf-quotes", q)
}

// AnalyzeStock is GET /api/market/v1/analyze-stock
func (m *Market) AnalyzeStock(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "analyze-stock", q)
}

// GetStockAnalysisHistory is GET /api/market/v1/get-stock-analysis-history
func (m *Market) GetStockAnalysisHistory(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-stock-analysis-history", q)
}

// BacktestStock is GET /api/market/v1/backtest-stock
func (m *Market) BacktestStock(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "backtest-stock", q)
}

// ListStoredStockBacktests is GET /api/market/v1/list-stored-stock-backtests
func (m *Market) ListStoredStockBacktests(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-stored-stock-backtests", q)
}

// ListCryptoSectors is GET /api/market/v1/list-crypto-sectors
func (m *Market) ListCryptoSectors(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-crypto-sectors", q)
}

// ListDefiTokens is GET /api/market/v1/list-defi-tokens
func (m *Market) ListDefiTokens(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-defi-tokens", q)
}

// ListAiTokens is GET /api/market/v1/list-ai-tokens
func (m *Market) ListAiTokens(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-ai-tokens", q)
}

// ListOtherTokens is GET /api/market/v1/list-other-tokens
func (m *Market) ListOtherTokens(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-other-tokens", q)
}

// GetFearGreedIndex is GET /api/market/v1/get-fear-greed-index
func (m *Market) GetFearGreedIndex(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-fear-greed-index", q)
}

// ListEarningsCalendar is GET /api/market/v1/list-earnings-calendar
func (m *Market) ListEarningsCalendar(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "list-earnings-calendar", q)
}

// GetCotPositioning is GET /api/market/v1/get-cot-positioning
func (m *Market) GetCotPositioning(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-cot-positioning", q)
}

// GetInsiderTransactions is GET /api/market/v1/get-insider-transactions
func (m *Market) GetInsiderTransactions(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-insider-transactions", q)
}

// GetMarketBreadthHistory is GET /api/market/v1/get-market-breadth-history
func (m *Market) GetMarketBreadthHistory(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-market-breadth-history", q)
}

// GetGoldIntelligence is GET /api/market/v1/get-gold-intelligence
func (m *Market) GetGoldIntelligence(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-gold-intelligence", q)
}

// GetHyperliquidFlow is GET /api/market/v1/get-hyperliquid-flow
func (m *Market) GetHyperliquidFlow(ctx context.Context, q url.Values) (json.RawMessage, error) {
	return m.Fetch(ctx, "get-hyperliquid-flow", q)
}
