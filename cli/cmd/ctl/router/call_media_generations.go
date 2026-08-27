package router

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `router call music` and `router call 3d` — the two families that exist only
// on the unified route.
//
// A track and a mesh are generations like an image or a video: submitted,
// polled, downloaded from the same content route, and recorded in the same
// table. What they do not have is a released OpenAI-shaped route, because
// OpenAI has no music or 3D API to be shaped like — so /v1/generations is the
// only way to ask for either, and their fields have no legacy spelling.
//
// Both require --model, and the reason is worth stating because it is not an
// oversight: Router resolves a default category for image and video generation
// and deliberately has none for these two. Its own registry says why — a
// category is a promise that something sensible answers it, and with one
// implementation apiece (FlowStudio serves both today) a default would name
// that one implementation while reading like a choice. So the model is named or
// the request is refused, and `router model list --mode music_generation` is
// what says which names exist here.

func newCallMusicCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		out      string
		outputID string
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
		flags    mediaFlags
	)
	cmd := &cobra.Command{
		Use:   "music [prompt…]",
		Short: "generate a track",
		Long: `Generate music from a description.

The track is written to --out, or to a file named after the generation. Router
holds the bytes, so the file does not depend on a provider's link staying alive.

--lyrics gives the words to sing; --instrumental asks for a track without any.
A model that cannot honor one of these refuses the request rather than ignoring
the field, which is the point of naming it.

--model is required. Router resolves a default for image and video generation
and none for music: today FlowStudio is the only thing serving it, and a
default would name that one workflow while reading like a choice. "olares-cli
router model list --mode music_generation" lists the names this credential can
send.

--no-wait prints the generation id and stops; "--id <id>" collects it later. A
generation expires, and --no-wait says when.

Examples:
  olares-cli router call music "a slow waltz on a rainy afternoon" --model FlowStudio/ace-step
  olares-cli router call music "an upbeat theme" --model FlowStudio/ace-step --duration 30 --instrumental
  olares-cli router call music "a ballad" --model FlowStudio/ace-step --lyrics "$(cat words.txt)"
  olares-cli router call music --id gen_01H… --model FlowStudio/ace-step
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runCanonicalMedia(c, f, musicKind, canonicalVerb{
				model: model, id: id, out: out, outputID: outputID,
				wait: !noWait, timeout: timeout, apiKey: apiKey, format: output,
				flags: &flags, args: args, mode: "music_generation",
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelRequiredHelp("music_generation"))
	cmd.Flags().StringVar(&out, "out", "", "write the track here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the track")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	flags.register(cmd, musicFields...)
	addOutputFlag(cmd, &output)
	return cmd
}

func newCall3DCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		output   string
		model    string
		out      string
		outputID string
		noWait   bool
		id       string
		timeout  time.Duration
		apiKey   string
		flags    mediaFlags
	)
	cmd := &cobra.Command{
		Use:   "3d [prompt…]",
		Short: "generate a 3D model",
		Long: `Generate a 3D model.

The mesh is written to --out, or to a file named after the generation. --formats
asks for the file formats to produce, --texture and --pbr for surfaces rather
than bare geometry, and --polycount for a polygon budget.

--image is how the subject is usually given: most 3D workflows work from a
picture rather than from words, and unlike image and video generation this
family accepts an input on a plain generate. A file on this machine is encoded
and sent; a data URL or a link is passed through as written.

--model is required. Router resolves a default for image and video generation
and none for 3D: today FlowStudio is the only thing serving it, and a default
would name that one workflow while reading like a choice. "olares-cli router
model list --mode model3d_generation" lists the names this credential can send.

--no-wait prints the generation id and stops; "--id <id>" collects it later. A
generation expires, and --no-wait says when.

Examples:
  olares-cli router call 3d "a brass lantern" --model FlowStudio/hunyuan3d
  olares-cli router call 3d --model FlowStudio/hunyuan3d --image lantern.png --formats glb,obj
  olares-cli router call 3d "a low-poly tree" --model FlowStudio/hunyuan3d --polycount 8000 --texture
  olares-cli router call 3d --id gen_01H… --model FlowStudio/hunyuan3d
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runCanonicalMedia(c, f, model3DKind, canonicalVerb{
				model: model, id: id, out: out, outputID: outputID,
				wait: !noWait, timeout: timeout, apiKey: apiKey, format: output,
				flags: &flags, args: args, mode: "model3d_generation",
				// A picture is a subject, so a mesh can be asked for with no
				// words at all. Every other family is asked in words.
				inputHint: "give an --" + flagImage + " to work from",
			})
		},
	}
	cmd.Flags().StringVar(&model, "model", "", modelRequiredHelp("model3d_generation"))
	cmd.Flags().StringVar(&out, "out", "", "write the model here instead of a name derived from the generation")
	cmd.Flags().StringVar(&outputID, "output-id", "", "which of the generation's outputs to write; the first when omitted")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "print the generation id instead of waiting for the model")
	cmd.Flags().StringVar(&id, "id", "", "collect a generation submitted earlier")
	cmd.Flags().DurationVar(&timeout, "timeout", 15*time.Minute, "give up waiting after this long; the work continues")
	cmd.Flags().StringVar(&apiKey, "api-key", "", dataPlaneKeyFlagUsage)
	flags.register(cmd, model3DFields...)
	addOutputFlag(cmd, &output)
	return cmd
}

// modelRequiredHelp is what --model says on a verb with no category behind it.
// It must not name a default: a caller told to expect one would read the
// resulting 404 as a missing model rather than as a route that never existed.
func modelRequiredHelp(mode string) string {
	return "model to use, as <provider>/<model> or a route name; required, " +
		"since Router resolves no default for " + mode
}

// canonicalVerb is what the two verbs above differ in, which is almost nothing.
type canonicalVerb struct {
	model     string
	id        string
	out       string
	outputID  string
	wait      bool
	timeout   time.Duration
	apiKey    string
	format    string
	flags     *mediaFlags
	args      []string
	mode      string
	inputHint string
}

func runCanonicalMedia(
	c *cobra.Command, f *cmdutil.Factory, kind mediaKind, verb canonicalVerb,
) error {
	model := strings.TrimSpace(verb.model)
	if model == "" {
		return fmt.Errorf("--model is required: Router resolves no default for %s, so there is "+
			"nothing to fall back to\n`olares-cli router model list --mode %s` lists the names "+
			"this credential can send", verb.mode, verb.mode)
	}
	opts := mediaOptions{
		Out: verb.out, OutputID: verb.outputID, Wait: verb.wait, Timeout: verb.timeout,
		APIKey: verb.apiKey, OutputIn: verb.format, ID: strings.TrimSpace(verb.id),
	}
	if opts.ID != "" {
		if len(verb.args) > 0 {
			return fmt.Errorf("--id collects a generation that already exists; it takes no prompt")
		}
		return runMedia(c.Context(), f, kind, opts)
	}
	prompt, err := resolvePrompt(c, verb.flags, verb.args, verb.inputHint)
	if err != nil {
		return err
	}
	body, err := verb.flags.canonical(c, model, prompt)
	if err != nil {
		return err
	}
	opts.Body = body
	return runMedia(c.Context(), f, kind, opts)
}
