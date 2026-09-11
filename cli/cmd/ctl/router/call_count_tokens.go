package router

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/messages/count_tokens
//
// How large a turn is before sending it. Claude Code asks this on every turn to
// budget its context window, and it is useful from a shell for the same reason:
// deciding whether a file fits before paying to find out that it does not.
//
// Two things separate it from every other call verb. It is mounted above the
// quota line and writes no spend row, so counting is free and repeatable. And
// the number can come from either of two places — the provider's own counter
// when it has one, or a local estimate when it does not — with nothing on the
// wire saying which. A local model is always the estimate, since Router asks no
// engine to count.

func newCallCountTokensCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output string
		model  string
		system string
		file   string
		apiKey string
	)
	cmd := &cobra.Command{
		Use:   "count-tokens [text…]",
		Short: "how many input tokens a turn would be",
		Long: `Count the input tokens of a turn without sending it.

The text is the arguments, or standard input when there are none. --system
counts a system prompt alongside it, since that is part of what a turn costs.
--file counts a request you already have: a JSON document in the Anthropic
shape, with "messages" and optionally "system", which is what a client that
built the turn itself would send.

Nothing is generated and nothing is billed. This route is mounted above the
quota line and writes no usage row, so counting the same document repeatedly
costs nothing and shows up nowhere in "router usage".

Where the number comes from depends on the model, and the answer does not say
which. A provider with an official counter is asked and answers exactly. Every
other case — a local model, a provider without one, a provider that failed or
was slow — is a local estimate close enough to budget with and not exact.

Examples:
  olares-cli router call count-tokens "summarise this" --model default-chat
  olares-cli router call count-tokens < long-document.txt
  olares-cli router call count-tokens --file turn.json --model anthropic/claude-sonnet-4
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if strings.TrimSpace(file) != "" {
				if len(args) > 0 {
					return fmt.Errorf("--file counts the turn in that document; it takes no text")
				}
				if strings.TrimSpace(system) != "" {
					return fmt.Errorf("--file counts the document as it stands, including its own " +
						"\"system\"; --system would be a second one")
				}
			}
			return runCountTokens(c.Context(), f, countTokensOptions{
				Args: args, File: strings.TrimSpace(file),
				Model: callModel(model, categoryChat), System: strings.TrimSpace(system),
				APIKey: apiKey, OutputIn: output,
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryChat))
	cmd.Flags().StringVar(&system, "system", "", "a system prompt to count alongside the text")
	cmd.Flags().StringVar(&file, "file", "", "count the Anthropic-shaped request in this JSON file")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type countTokensOptions struct {
	Args     []string
	File     string
	Model    string
	System   string
	APIKey   string
	OutputIn string
}

type countTokensResponse struct {
	InputTokens int64 `json:"input_tokens"`
}

func runCountTokens(ctx context.Context, f *cmdutil.Factory, opts countTokensOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	body, err := countTokensBody(opts)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var resp countTokensResponse
	if err := dataPlane(pc, opts.APIKey).doJSON(ctx, "POST", epMessagesCountTokens, body, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	_, err = fmt.Printf("%d input tokens\n", resp.InputTokens)
	return err
}

// countTokensBody builds the Anthropic shape, or takes the caller's own and
// puts the resolved model on it. The model matters even though nothing is
// generated: which provider is asked decides whether the count is official or
// estimated, and a document written for one vendor names a model this Router
// may not have.
func countTokensBody(opts countTokensOptions) (map[string]any, error) {
	if opts.File != "" {
		raw, err := os.ReadFile(opts.File)
		if err != nil {
			return nil, fmt.Errorf("read the turn: %w", err)
		}
		var body map[string]any
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, fmt.Errorf("%s is not a JSON object in the Anthropic shape: %w", opts.File, err)
		}
		if _, ok := body["messages"]; !ok {
			return nil, fmt.Errorf("%s has no \"messages\"; this route counts a turn, "+
				"and a turn is its messages", opts.File)
		}
		body["model"] = opts.Model
		return body, nil
	}
	text, err := readPromptArgs(opts.Args, "text")
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"model":    opts.Model,
		"messages": []map[string]any{{"role": "user", "content": text}},
	}
	if opts.System != "" {
		body["system"] = opts.System
	}
	return body, nil
}
