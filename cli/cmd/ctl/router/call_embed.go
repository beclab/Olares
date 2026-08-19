package router

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// POST /v1/embeddings

type embeddingsRequest struct {
	Model      string   `json:"model,omitempty"`
	Input      []string `json:"input"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

type embeddingsResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage *struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func newCallEmbedCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output     string
		model      string
		dimensions int
		perLine    bool
		apiKey     string
	)
	cmd := &cobra.Command{
		Use:   "embed [text…]",
		Short: "embedding vectors for text",
		Long: `Turn text into vectors.

Each argument is one input. With no arguments the text comes from standard
input: as one input by default, or one per line with --per-line, which is how a
file of records is embedded in a single call.

The table form does not print the numbers — a 1024-dimension vector is not
something to read — and shows the dimension count and the first few components
instead, which is enough to confirm the model answered and how wide its output
is. -o json prints the vectors in full, for the caller that wants them.

Embedding needs a model whose mode is embedding; a chat model refuses the route
rather than guessing. "olares-cli router list --mode embedding" shows the ones
that qualify.

Examples:
  olares-cli router call embed "the quick brown fox"
  olares-cli router call embed "first" "second" "third"
  cat lines.txt | olares-cli router call embed --per-line -o json
  olares-cli router call embed "text" --model bge-m3 --dimensions 512
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			inputs, err := embedInputs(args, perLine)
			if err != nil {
				return err
			}
			var dims *int
			if c.Flags().Changed("dimensions") {
				dims = &dimensions
			}
			return runCallEmbed(c.Context(), f, inputs, model, dims, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", "model to use; the workspace embedding default when omitted")
	cmd.Flags().IntVar(&dimensions, "dimensions", 0, "ask for a narrower vector, if the model allows it")
	cmd.Flags().BoolVar(&perLine, "per-line", false, "treat each line of stdin as a separate input")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	addOutputFlag(cmd, &output)
	return cmd
}

func embedInputs(args []string, perLine bool) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}
	if isTerminal(os.Stdin) {
		return nil, fmt.Errorf("no text given; pass it as arguments or pipe it in")
	}
	if !perLine {
		text, err := readPromptArgs(nil, "text")
		if err != nil {
			return nil, err
		}
		return []string{text}, nil
	}
	var out []string
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read text from stdin: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no non-empty lines on stdin")
	}
	return out, nil
}

func runCallEmbed(ctx context.Context, f *cmdutil.Factory, inputs []string, model string, dims *int, apiKey, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp, _, err := dataPlane(ctx, pc, apiKey)
	if err != nil {
		return err
	}
	req := embeddingsRequest{
		Model:      strings.TrimSpace(model),
		Input:      inputs,
		Dimensions: dims,
	}
	var resp embeddingsResponse
	if err := dp.doJSON(ctx, "POST", epEmbeddings, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderEmbeddings(os.Stdout, &resp, inputs)
}

func renderEmbeddings(w io.Writer, resp *embeddingsResponse, inputs []string) error {
	if len(resp.Data) == 0 {
		_, err := fmt.Fprintln(w, "the model returned no vectors.")
		return err
	}
	t := newTable(w, "#", "DIMS", "NORM", "FIRST COMPONENTS", "INPUT")
	for i := range resp.Data {
		d := &resp.Data[i]
		input := ""
		if d.Index >= 0 && d.Index < len(inputs) {
			input = inputs[d.Index]
		}
		t.row(
			strconv.Itoa(d.Index), strconv.Itoa(len(d.Embedding)),
			fmt.Sprintf("%.3f", l2Norm(d.Embedding)),
			headComponents(d.Embedding, 3), clip(input, 40))
	}
	if err := t.flush(); err != nil {
		return err
	}
	line := "\n" + nonEmpty(resp.Model)
	if resp.Usage != nil {
		line += fmt.Sprintf("  %d tokens", resp.Usage.TotalTokens)
	}
	line += "  -o json prints the vectors"
	_, err := fmt.Fprintln(os.Stderr, line)
	return err
}

// l2Norm is shown because it is the one number that says something useful about
// a vector at a glance: an embedding model that normalises its output puts it at
// 1.0, and one that does not is a thing to know before comparing with a dot
// product.
func l2Norm(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}

func headComponents(v []float64, n int) string {
	if len(v) < n {
		n = len(v)
	}
	parts := make([]string, 0, n)
	for _, x := range v[:n] {
		parts = append(parts, fmt.Sprintf("%.4f", x))
	}
	s := strings.Join(parts, " ")
	if len(v) > n {
		s += " …"
	}
	return s
}
