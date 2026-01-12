package ask

import (
	openai "github.com/openai/openai-go"
)

func CalculateCost(model string, completion *openai.ChatCompletion) (float64, bool) {
	pricing, exists := CostTable[model]
	if !exists {
		return 0, false
	}

	var totalCost float64

	inputTokens := float64(completion.Usage.PromptTokens)
	cachedTokens := float64(completion.Usage.PromptTokensDetails.CachedTokens)
	outputTokens := float64(completion.Usage.CompletionTokens)

	totalCost += (inputTokens - cachedTokens) * pricing.InputCost
	totalCost += cachedTokens * pricing.CachedInputCost
	totalCost += outputTokens * pricing.OutputCost

	return totalCost / 1_000_000, true
}
