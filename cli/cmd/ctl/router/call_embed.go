package router

import (
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
	Model string `json:"model,omitempty"`
	// Input is the texts, or one image object. The OpenAI body types this as a
	// string or an array of them and Router extends it with the object below,
	// so the field cannot be narrower than what both shapes need.
	Input      any  `json:"input"`
	Dimensions *int `json:"dimensions,omitempty"`
}

// embeddingImageInput is the extension. A picture is one input rather than a
// list because the two towers of a CLIP model answer one thing at a time, and
// `type` is what tells the engine which tower it is being asked for.
type embeddingImageInput struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
}

const (
	// A model application serving CLIP admits a 16 MiB /v1/embeddings body.
	// The allowance is for the JSON around the picture — the field names, the
	// model, the data URL's own prefix — and the check is on the encoded form
	// rather than on the file, because base64 costs four bytes for every three
	// and a file just under the limit does not fit.
	embedRequestBudgetBytes = 16 * 1024 * 1024
	embedEnvelopeAllowance  = 4 * 1024
	embedImageMaxBytes      = embedRequestBudgetBytes - embedEnvelopeAllowance
)

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
		image      string
		apiKey     string
	)
	cmd := &cobra.Command{
		Use:   "embed [text…]",
		Short: "embedding vectors for text or a picture",
		Long: `Turn text into vectors.

Each argument is one input. With no arguments the text comes from standard
input: as one input by default, or one per line with --per-line, which is how a
file of records is embedded in a single call.

--image embeds a picture instead, on the same endpoint. It needs a CLIP
application — one whose row declares supports_embedding_image_input — which
embeds pictures and text into one space, and that shared space is the whole
point: an image vector is only worth having if it can be compared with the text
vectors already stored. A text-only embedding model refuses the picture rather
than describing it. One call carries one picture or the text, never both.

The table form does not print the numbers — a 1024-dimension vector is not
something to read — and shows the dimension count and the first few components
instead, which is enough to confirm the model answered and how wide its output
is. -o json prints the vectors in full, for the caller that wants them.

Embedding needs a model whose mode is embedding; a chat model refuses the route
rather than guessing. "olares-cli router model list --mode embedding" shows the ones
that qualify.

Examples:
  olares-cli router call embed "the quick brown fox"
  olares-cli router call embed "first" "second" "third"
  cat lines.txt | olares-cli router call embed --per-line -o json
  olares-cli router call embed "text" --model bge-m3 --dimensions 512
  olares-cli router call embed --image shot.png --model jinaclipv3/jina-clip-v2
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var inputs []string
			if strings.TrimSpace(image) == "" {
				var err error
				if inputs, err = embedInputs(args, perLine); err != nil {
					return err
				}
			} else if len(args) > 0 || perLine {
				return fmt.Errorf("--image embeds one picture and nothing else; " +
					"embed the text in its own call")
			}
			var dims *int
			if c.Flags().Changed("dimensions") {
				dims = &dimensions
			}
			return runCallEmbed(c.Context(), f, inputs, image,
				callModel(model, categoryEmbedding), dims, apiKey, output)
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelFlagHelp(categoryEmbedding))
	cmd.Flags().IntVar(&dimensions, "dimensions", 0, "ask for a narrower vector, if the model allows it")
	cmd.Flags().BoolVar(&perLine, "per-line", false, "treat each line of stdin as a separate input")
	cmd.Flags().StringVar(&image, "image", "", "embed this picture instead of text: a local file, a data URL or a link")
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
	out, err := readLines(os.Stdin)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no non-empty lines on stdin")
	}
	return out, nil
}

func runCallEmbed(ctx context.Context, f *cmdutil.Factory, inputs []string, image, model string,
	dims *int, apiKey, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	req := embeddingsRequest{
		Model:      strings.TrimSpace(model),
		Input:      inputs,
		Dimensions: dims,
	}
	labels := inputs
	if strings.TrimSpace(image) != "" {
		picture, err := embedImage(image)
		if err != nil {
			return err
		}
		req.Input = picture
		labels = []string{strings.TrimSpace(image)}
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	dp := dataPlane(pc, apiKey)
	var resp embeddingsResponse
	if err := dp.doJSON(ctx, "POST", epEmbeddings, req, &resp); err != nil {
		return callErr(err)
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, resp)
	}
	return renderEmbeddings(os.Stdout, &resp, labels)
}

// embedImage reads what --image was given the way every other picture flag in
// this tree reads one: a local path is encoded, a data URL or a link is sent as
// written. The size is checked here rather than at the file, because what the
// budget applies to is the request.
func embedImage(value string) (embeddingImageInput, error) {
	url, err := mediaInput(value, "image")
	if err != nil {
		return embeddingImageInput{}, err
	}
	if len(url) > embedImageMaxBytes {
		return embeddingImageInput{}, fmt.Errorf("%s encodes to %s (%d bytes); a picture cannot "+
			"exceed %s (%d bytes) once base64 has added a third to it (the application admits %s "+
			"(%d bytes) for the whole request, with %s (%d bytes) reserved for the JSON around it)",
			value, humanBytes(int64(len(url))), len(url),
			humanBytes(embedImageMaxBytes), embedImageMaxBytes,
			humanBytes(embedRequestBudgetBytes), embedRequestBudgetBytes,
			humanBytes(embedEnvelopeAllowance), embedEnvelopeAllowance)
	}
	return embeddingImageInput{Type: "image", ImageURL: url}, nil
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
