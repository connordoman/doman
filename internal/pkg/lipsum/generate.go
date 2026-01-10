package lipsum

import (
	_ "embed"
	"fmt"
	"log"
	"math/rand"
	"strings"
)

/*
   Special thanks to https://www.lipsum.com/ for the 10,000 word seed data.
*/

//go:embed lorem-ipsum_10k.txt
var loremIpsum10KRaw string

type State []string
type Chain map[string]map[string]int

const (
	chainOrder = 3
)

const (
	OpeningWordCount = 8 // len(opening) but without punctuation
)

var opening = []string{
	"Lorem",
	"ipsum",
	"dolor",
	"sit",
	"amet",
	",",
	"consectetur",
	"adipiscing",
	"elit",
	".",
}

var initialState = opening[len(opening)-chainOrder:]

func GetOpening() []string {
	return opening
}

func Tokenize(text string) []string {
	lipsumLocal := strings.TrimSpace(text)
	lipsumLocal = strings.ReplaceAll(lipsumLocal, "\n", " ")
	lipsumLocal = strings.ReplaceAll(lipsumLocal, "\r", " ")

	var tokens []string
	var currentToken string

	// fmt.Printf("lipsumLocal: %s\n", lipsumLocal)
	// fmt.Printf("lipsumLocal length: %d\n", len(lipsumLocal))

	for _, c := range lipsumLocal {
		switch c {
		case ' ':
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
		case '.', ',', ';', '?', '!':
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
			tokens = append(tokens, string(c))
		case '\n':
			if currentToken != "" {
				tokens = append(tokens, currentToken)
				currentToken = ""
			}
		default:
			currentToken += string(c)
		}
	}

	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}

	// fmt.Println()

	// fmt.Printf("tokens: %v\n", tokens)

	return tokens
}

func makeStateKey(tokens []string) string {
	return strings.Join(tokens, "|")
}

func BuildChain(tokens []string) *Chain {
	var chain Chain = make(Chain)

	for i := 0; i+chainOrder < len(tokens); i++ {
		stateKey := makeStateKey(tokens[i : i+chainOrder])
		next := tokens[i+chainOrder]

		if chain[stateKey] == nil {
			chain[stateKey] = make(map[string]int)
		}
		chain[stateKey][next]++
	}

	return &chain
}

func weightedChoice(m map[string]int) string {
	total := 0
	for _, c := range m {
		total += c
	}

	if total == 0 {
		log.Printf("warning: weighted choice total is 0\ndist: %v", m)
	}

	r := rand.Intn(total)
	for token, c := range m {
		if r < c {
			return token
		}
		r -= c
	}

	panic("unreachable")
}

func StringifyTokens(tokens []string) string {
	var str strings.Builder

	tokenCount := len(tokens)

	for i, t := range tokens {
		next := tokens[min(tokenCount-1, i+1)]

		switch next {
		case t, ".", ",", ";", "?", "!", "\n":
			str.WriteString(t)
		default:
			str.WriteString(t + " ")
		}
	}

	return strings.TrimSpace(str.String())
}

func isWord(str string) bool {
	switch str {
	case ".", ",", ";", "?", "!":
		return false
	default:
		return true
	}
}

func cleanLastToken(tokens *[]string) {
	if len(*tokens) == 0 {
		return
	}

	last := (*tokens)[len(*tokens)-1]
	if !isWord(last) {
		*tokens = (*tokens)[:len(*tokens)-1]
	}

	if len(*tokens) == 0 || (*tokens)[len(*tokens)-1] != "." {
		*tokens = append(*tokens, ".")
	}
}

func GenerateLoremIpsum(n int) []string {
	tokens := Tokenize(loremIpsum10KRaw)
	chain := BuildChain(tokens)

	var generatedTokens []string
	generatedTokens = append(generatedTokens, opening...)

	var state []string = make([]string, chainOrder)
	copy(state, initialState)

	var count int = OpeningWordCount

	for count < n {
		if len(state) == 0 {
			log.Printf("warning: state window is empty. count: %d", count)
		}

		stateKey := makeStateKey(state)

		dist, exists := (*chain)[stateKey]
		if !exists {
			copy(state, initialState)
			continue
		}

		nextToken := weightedChoice(dist)

		generatedTokens = append(generatedTokens, nextToken)

		if isWord(nextToken) {
			count++
		}

		state = append(state[1:], nextToken)
	}

	// if generatedTokens[len(generatedTokens)-1] != "." {
	// 	generatedTokens = append(generatedTokens, ".")
	// }
	cleanLastToken(&generatedTokens)

	return generatedTokens
}

func randInRange(minimum, maximum int) int {
	return rand.Intn(maximum-minimum+1) + minimum
}

func SplitParagraphs(generatedTokens []string) []string {
	minPeriods := 5
	maxPeriods := 8

	breakNow := randInRange(minPeriods, maxPeriods)

	result := []string{}

	periodCount := 0
	for _, t := range generatedTokens {
		result = append(result, t)

		if t == "." {
			periodCount++
			if periodCount == breakNow {
				result = append(result, "\n", "\n")
				periodCount = 0
				breakNow = randInRange(minPeriods, maxPeriods)
			}
		}

	}

	return result
}

func SplitNParagraphs(generatedTokens []string, n int) ([]string, error) {
	if n < 1 {
		return nil, fmt.Errorf("number of paragraphs must be > 0")
	} else if n == 1 {
		return generatedTokens, nil
	}

	totalPeriods := 0
	for _, t := range generatedTokens {
		if t == "." {
			totalPeriods++
		}
	}

	sentenceCount := totalPeriods / n
	remainder := totalPeriods % n

	result := []string{}

	periodCount := 0
	paragraphCount := 0
	for _, t := range generatedTokens {
		result = append(result, t)

		if t == "." && paragraphCount < n {
			periodCount++

			sc := sentenceCount

			if paragraphCount < remainder {
				sc++
			}

			if periodCount == sc {
				result = append(result, "\n", "\n")
				periodCount = 0
				paragraphCount++
			}
		}
	}

	return result, nil
}
