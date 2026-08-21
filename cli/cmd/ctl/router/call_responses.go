/*
Copyright 2024 bytetrade.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/responses
//
// One request, one answer, and nothing kept. Router's Responses surface is
// larger than this — a conversation continues through previous_response_id,
// and there are routes to retrieve, cancel, compact and delete a stored one —
// but a conversation held across separate CLI invocations is a session, and a
// session is a thing to design rather than a flag to add. What is missing
// without any of that is the ability to tell whether a model configured with
// `--mode responses` works at all, and that is what this answers.

type responsesRequest struct {
	Model           string `json:"model,omitempty"`
	Input           string `json:"input"`
	Instructions    string `json:"instructions,omitempty"`
	MaxOutputTokens *int   `json:"max_output_tokens,omitempty"`
	// Store is sent false so a one-shot verb leaves nothing behind. Router
	// stores a response by default and anchors it on the calling credential,
	// so the default would accumulate rows nothing here can list or delete —
	// and a row stored under a key is not even readable by a later keyless
	// call, which makes the accumulation invisible as well as permanent.
	Store bool `json:"store"`
}

type responsesItem struct {
	Type    string `json:"type"`
	Role    string `json:"role,omitempty"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content,omitempty"`
	Summary []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"summary,omitempty"`
}

type responsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type responsesAnswer struct {
	ID                string          `json:"id"`
	Model             string          `json:"model"`
	Status            string          `json:"status"`
	Output            []responsesItem `json:"output,omitempty"`
	Usage             *responsesUsage `json:"usage,omitempty"`
	IncompleteDetails *struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func newCallResponsesCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output       string
		model        string
		instructions string
		maxTokens    int
		quiet        bool
		apiKey       string
	)
	cmd := &cobra.Command{
		Use:   "responses [input]",
		Short: "a one-shot Responses call",
		Long: `Send one input to a model on the Responses API and print the answer.

The input is the argument, or standard input when there is none.

This is the Responses equivalent of "call chat", and it exists because a model
configured with --mode responses could not be called from here at all: it is
served on a different endpoint from chat, so a chat call to it fails in a way
that says nothing about whether the model works.

--model is required. Every other verb here falls back to a default category
when it is omitted; this mode has none, because Router resolves defaults for
every mode except this one and does so on purpose.
"router model list --mode responses" is where the names are.

One request, one answer. A Responses conversation continues by quoting the id
of the previous answer, and Router keeps the stored ones so it can — but a
conversation spread across separate command invocations is a session, and this
verb does not open one. Each call sends store=false and leaves nothing behind.

Answers stream on this API as they do on chat, and this verb does not stream:
the answer is printed when it is complete. For watching a long answer arrive,
"call chat" against a chat-mode model is the verb that does that today.

Examples:
  olares-cli router call responses "summarise the CAP theorem in three lines"
  git diff | olares-cli router call responses --instructions "write a commit message"
  olares-cli router call responses "hello" --model gpt-5 -o json
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if strings.TrimSpace(model) == "" {
				return fmt.Errorf("--model is required here: Router resolves no default for the " +
					"responses mode, deliberately, so there is no category to fall back to. " +
					"`olares-cli router model list --mode responses` shows the models that answer on it")
			}
			input, err := readPromptArgs(args, "input")
			if err != nil {
				return err
			}
			opts := responsesOptions{
				Model:        model,
				Instructions: instructions,
				Quiet:        quiet,
				APIKey:       apiKey,
				OutputIn:     output,
			}
			if c.Flags().Changed("max-tokens") {
				opts.MaxTokens = &maxTokens
			}
			return runCallResponses(c.Context(), f, input, opts)
		},
	}
	cmd.Flags().StringVar(&model, "model", "",
		"model to use, as <provider>/<model> or a route name; required, as this mode has no default")
	cmd.Flags().StringVar(&instructions, "instructions", "", "instruction sent ahead of the input")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 0, "cap the answer length in tokens")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "print only the answer, without the model and token line")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

type responsesOptions struct {
	Model        string
	Instructions string
	MaxOutputStr string
	MaxTokens    *int
	Quiet        bool
	APIKey       string
	OutputIn     string
}

func runCallResponses(ctx context.Context, f *cmdutil.Factory, input string, opts responsesOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, opts.APIKey)

	req := responsesRequest{
		Model:           strings.TrimSpace(opts.Model),
		Input:           input,
		Instructions:    strings.TrimSpace(opts.Instructions),
		MaxOutputTokens: opts.MaxTokens,
	}
	var answer responsesAnswer
	if err := dp.doJSON(ctx, "POST", epResponses, req, &answer); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, answer)
	}
	return renderResponses(os.Stdout, &answer, opts.Quiet)
}

// renderResponses prints the message text and, when there is one, the
// reasoning summary above it — matching what `call chat` does with
// reasoning_content, since a reader switching between the two verbs should not
// have to learn a second layout for the same two things.
func renderResponses(w io.Writer, a *responsesAnswer, quiet bool) error {
	// A refusal inside a 200 is the shape this API uses for a model that
	// answered by declining. Printing nothing would read as an empty answer.
	if a.Error != nil && strings.TrimSpace(a.Error.Message) != "" {
		return fmt.Errorf("%s: %s", nonEmpty(a.Error.Code), a.Error.Message)
	}

	var reasoning, text []string
	for i := range a.Output {
		item := &a.Output[i]
		for _, s := range item.Summary {
			if t := strings.TrimSpace(s.Text); t != "" {
				reasoning = append(reasoning, t)
			}
		}
		for _, c := range item.Content {
			if t := c.Text; t != "" {
				text = append(text, t)
			}
		}
	}

	if len(reasoning) > 0 {
		if _, err := fmt.Fprintf(w, "[reasoning]\n%s\n\n", strings.Join(reasoning, "\n")); err != nil {
			return err
		}
	}
	if len(text) == 0 {
		if _, err := fmt.Fprintln(w, "the model returned no text."); err != nil {
			return err
		}
	} else if _, err := fmt.Fprintln(w, strings.TrimRight(strings.Join(text, ""), "\n")); err != nil {
		return err
	}
	if quiet {
		return nil
	}
	return printResponsesFooter(a)
}

// printResponsesFooter goes to stderr for the same reason the chat one does:
// the answer on stdout stays pipeable while the accounting stays visible.
func printResponsesFooter(a *responsesAnswer) error {
	line := "\n" + nonEmpty(a.Model)
	if a.Usage != nil {
		line += fmt.Sprintf("  %d input + %d output = %d tokens",
			a.Usage.InputTokens, a.Usage.OutputTokens, a.Usage.TotalTokens)
	}
	// `completed` is the ordinary ending and saying so adds nothing. Anything
	// else changes what the answer above is worth.
	if s := strings.TrimSpace(a.Status); s != "" && s != "completed" {
		line += "  status: " + s
		if a.IncompleteDetails != nil && a.IncompleteDetails.Reason != "" {
			line += " (" + a.IncompleteDetails.Reason + ")"
		}
	}
	_, err := fmt.Fprintln(os.Stderr, line)
	return err
}
