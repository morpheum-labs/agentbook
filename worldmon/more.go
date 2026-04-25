package worldmon

// The following methods return a generic [Service] for the matching directory under
// [server/worldmonitor] on GitHub. Call [Service.Fetch] with the handler file stem as
// the method (kebab-case, same as the .ts name without the extension) when a typed
// wrapper is not yet in this module.
//
// [server/worldmonitor]: https://github.com/koala73/worldmonitor/tree/main/server/worldmonitor

// Aviation is GET /api/aviation/v1/…
func (c *Client) Aviation() *Service { return c.Service("aviation", "v1") }

// ConsumerPrices is GET /api/consumer-prices/v1/…
func (c *Client) ConsumerPrices() *Service { return c.Service("consumer-prices", "v1") }

// Displacement is GET /api/displacement/v1/…
func (c *Client) Displacement() *Service { return c.Service("displacement", "v1") }

// Economic is GET /api/economic/v1/…
func (c *Client) Economic() *Service { return c.Service("economic", "v1") }

// Giving is GET /api/giving/v1/…
func (c *Client) Giving() *Service { return c.Service("giving", "v1") }

// Health is GET /api/health/v1/…
func (c *Client) Health() *Service { return c.Service("health", "v1") }

// Imagery is GET /api/imagery/v1/…
func (c *Client) Imagery() *Service { return c.Service("imagery", "v1") }

// Infrastructure is GET /api/infrastructure/v1/…
func (c *Client) Infrastructure() *Service { return c.Service("infrastructure", "v1") }

// Leads is GET /api/leads/v1/…
func (c *Client) Leads() *Service { return c.Service("leads", "v1") }

// PositiveEvents is GET /api/positive-events/v1/…
func (c *Client) PositiveEvents() *Service { return c.Service("positive-events", "v1") }

// Prediction is GET /api/prediction/v1/…
func (c *Client) Prediction() *Service { return c.Service("prediction", "v1") }

// Radiation is GET /api/radiation/v1/…
func (c *Client) Radiation() *Service { return c.Service("radiation", "v1") }

// Research is GET /api/research/v1/…
func (c *Client) Research() *Service { return c.Service("research", "v1") }

// Resilience is GET /api/resilience/v1/…
func (c *Client) Resilience() *Service { return c.Service("resilience", "v1") }

// Sanctions is GET /api/sanctions/v1/…
func (c *Client) Sanctions() *Service { return c.Service("sanctions", "v1") }

// Scenario is GET /api/scenario/v1/…
func (c *Client) Scenario() *Service { return c.Service("scenario", "v1") }

// SupplyChain is GET /api/supply-chain/v1/…
func (c *Client) SupplyChain() *Service { return c.Service("supply-chain", "v1") }

// Thermal is GET /api/thermal/v1/…
func (c *Client) Thermal() *Service { return c.Service("thermal", "v1") }

// Webcam is GET /api/webcam/v1/…
func (c *Client) Webcam() *Service { return c.Service("webcam", "v1") }

// Wildfire is GET /api/wildfire/v1/…
func (c *Client) Wildfire() *Service { return c.Service("wildfire", "v1") }
