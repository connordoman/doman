package ask

const (
	maxConversationTitleLength = 80
	minMeaningfulTitleRunes    = 3
)

type MessageHistory struct {
	Role    string
	Content string
}

type ModelPricing struct {
	InputCost       float64
	CachedInputCost float64
	OutputCost      float64
}

var CostTable = map[string]ModelPricing{
	"gpt-5-nano": {
		InputCost:       0.050,
		CachedInputCost: 0.005,
		OutputCost:      0.400,
	},
	"gpt-5-mini": {
		InputCost:       0.250,
		CachedInputCost: 0.025,
		OutputCost:      2.000,
	},
	"gpt-5": {
		InputCost:       1.250,
		CachedInputCost: 0.125,
		OutputCost:      10.000,
	},
}

var SplashTexts = []string{
	"Talking to robots",
	"Getting the skinny",
	"Finding the hay in the needle stack",
	"Pushing to production",
	"Talking to the little man",
	"Getting a second opinion",
	"Looking it up for you (even though you could probably do it)",
	"Getting the lowdown",
	"Writing a strongly worded letter",
	"Lighting a signal fire",
	"Using a small village's electricity budget for this",
	"Asking the AI to ask the AI",
	"Preparing to shut off PS4",
	"Castling",
	"Throwing it back",
	"Consulting the tea leaves",
	"Praying to god",
	"Shooting the messenger",
	"Texting my mom",
	"Figuring out the hard way",
	"Firing my assistant",
	"Thinking about stuff",
	"Letting the voices in",
	"Getting the instructions out of the garbage",
	"Going back to school",
	"Definitely not just Googling it",
	"Finishing my protein shake first",
	"Playing the long game",
	"Asking my supervisor",
	"F***ing around in hope of finding out",
	"Calling your boss",
	"Reading your diary",
	"Waking up early to get the worm",
	"Making a quick buck",
	"Sending a message in a bottle",
	"Asking Dr. Wilson for a consult",
	"I'm working on it",
	"Swiping right",
	"Finishing my bathroom break",
}

const DeveloperDefinedSystemMessage = `
You are a helpful assistant inside a CLI tool called 'doman'. Users can only ask text-based questions.

Audience assumptions:
- Users are technically literate.
- Most questions are technical (programming/devops/tools), but all topics are allowed.

Core response goals:
- Be concise and direct, but do not omit important caveats, constraints, or "gotchas".
- Optimize for readability in a terminal Markdown renderer.
- Users may follow up (possibly with '--continue' / '-c'), so keep answers scannable.

CRITICAL: Markdown structure is required
- Your entire response MUST be valid Markdown (GitHub-flavored is fine).
- You MUST use headings to structure the body. Do not output an unstructured wall of text.
- Start the response with a level-2 heading ("## ..."). (Do not start with plain text.)
- Use only "##" and "###" headings (avoid "#", and avoid heading levels deeper than "###").
- Use bullet lists where appropriate. Use blank lines between paragraphs/sections.

Required output template (fill the relevant sections; omit only if truly not applicable):
## <A short, descriptive title of the answer>

### Context (optional)
- <1-3 bullets, only if it helps orient the user>

### Details
<1-6 short paragraphs and/or bullets; keep lines/ideas separated>

### Examples (optional)
<If you show commands, config, or code, prefer an example>

## Short answer
<Put the short answer at the end when applicable. If the user explicitly asks for "just the short answer", still keep this section and keep the rest minimal.>

Code formatting rules:
- Use fenced code blocks and ALWAYS include a language identifier (e.g. bash, sh, go, json, yaml, python, typescript, rust, text).
- Inline HTML will render in the terminal: when discussing HTML tags, wrap them in backticks or a fenced code block.
- Do NOT use HTML to format your response.

Self-check before you send:
- If your draft has no "##" heading, rewrite it to match the template.
- If you used a code fence, ensure it has a language tag.

The user may also configure an additional system message. That message can override these rules.
`

const UserDefinedSystemMessagePrefix = "Additional system message, provided by the end user: \n\n"
