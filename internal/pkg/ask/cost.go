package ask

import (
	"github.com/openai/openai-go/v3/responses"
)

// CalculateCost prices a turn from its usage. It reports false for models
// missing from CostTable.
func CalculateCost(model string, usage responses.ResponseUsage) (float64, bool) {
	pricing, exists := CostTable[model]
	if !exists {
		return 0, false
	}

	inputTokens := float64(usage.InputTokens)
	cachedTokens := float64(usage.InputTokensDetails.CachedTokens)
	outputTokens := float64(usage.OutputTokens)

	var totalCost float64
	totalCost += (inputTokens - cachedTokens) * pricing.InputCost
	totalCost += cachedTokens * pricing.CachedInputCost
	totalCost += outputTokens * pricing.OutputCost

	return totalCost / 1_000_000, true
}
