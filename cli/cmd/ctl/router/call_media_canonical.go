package router

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// The body a creative request is spelled in, and the flags that fill it.
//
// Router has three media create routes and one contract behind them.
// /v1/generations takes the canonical body as written; /v1/images/generations
// and /v1/videos take the OpenAI-derived keys they shipped with and translate
// them onto the same canonical fields before anything is charged. So a flag
// here has two spellings, and which is sent depends on the route the verb uses.
//
// Which route a verb uses is not a preference:
//
//   - image and video ride the routes they always have. Every field those
//     routes accept lands on the canonical field the unified route would have
//     given it, and the image route has something the unified one cannot have:
//     a synchronous answer for a provider that keeps no generations, where the
//     picture arrives inline. Moving an image to /v1/generations would refuse
//     the most widely installed provider there is, since that route serves only
//     a provider whose generations can be polled.
//   - music and 3D ride /v1/generations, because they have no other route. They
//     also carry fields no released key spells — a track's lyrics, a mesh's
//     polygon budget — which is why the canonical body exists here at all.
//
// What a family may ask for is Router's table, not this one: a field a family
// cannot express has no flag on that verb, so `--fps` is not a thing to give an
// image and `--lyrics` is not a thing to give a mesh.

// canonicalRequest is the body of a POST /v1/generations.
//
// Pointers mark presence rather than default. Router refuses a field the
// resolved provider has no parameter for, so "n was omitted" and "n was sent as
// 1" are different requests and a zero value must not be sent as a preference
// nobody expressed.
type canonicalRequest struct {
	Model     string `json:"model,omitempty"`
	Operation string `json:"operation,omitempty"`

	Prompt         string `json:"prompt,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	N              *int   `json:"n,omitempty"`
	Seed           *int64 `json:"seed,omitempty"`

	Inputs *canonicalInputs `json:"inputs,omitempty"`
	Output *canonicalOutput `json:"output,omitempty"`
	Music  *canonicalMusic  `json:"music,omitempty"`

	ProviderOptions map[string]json.RawMessage `json:"provider_options,omitempty"`
}

// canonicalInputs are the references a generation works from.
type canonicalInputs struct {
	Images             []string `json:"images,omitempty"`
	Mask               string   `json:"mask,omitempty"`
	Audio              string   `json:"audio,omitempty"`
	SourceGenerationID string   `json:"source_generation_id,omitempty"`
}

// canonicalOutput describes the artifact asked for.
type canonicalOutput struct {
	Format          string   `json:"format,omitempty"`
	Formats         []string `json:"formats,omitempty"`
	Size            string   `json:"size,omitempty"`
	AspectRatio     string   `json:"aspect_ratio,omitempty"`
	Resolution      string   `json:"resolution,omitempty"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	FPS             *int     `json:"fps,omitempty"`
	Quality         string   `json:"quality,omitempty"`

	Texture         *bool `json:"texture,omitempty"`
	PBR             *bool `json:"pbr,omitempty"`
	TargetPolycount *int  `json:"target_polycount,omitempty"`
}

// canonicalMusic carries the two things a track is described by that no other
// family has.
type canonicalMusic struct {
	Lyrics       string `json:"lyrics,omitempty"`
	Instrumental *bool  `json:"instrumental,omitempty"`
}

// The flag names, declared once. A verb lists the ones its family admits and
// the same definition serves all four, so a flag cannot mean one thing on
// `call image` and another on `call video`.
const (
	flagNegative       = "negative"
	flagN              = "n"
	flagSeed           = "seed"
	flagImage          = "image"
	flagMask           = "mask"
	flagAudioIn        = "audio"
	flagSource         = "source-generation"
	flagOperation      = "operation"
	flagFormat         = "format"
	flagFormats        = "formats"
	flagSize           = "size"
	flagAspectRatio    = "aspect-ratio"
	flagResolution     = "resolution"
	flagDuration       = "duration"
	flagFPS            = "fps"
	flagQuality        = "quality"
	flagTexture        = "texture"
	flagPBR            = "pbr"
	flagPolycount      = "polycount"
	flagLyrics         = "lyrics"
	flagInstrumental   = "instrumental"
	flagProviderOption = "provider-option"
)

// mediaFlags holds what the caller asked for. Whether a field was given at all
// is read from the flag set rather than from these values, because for most of
// them the zero value is a legitimate thing to ask for.
type mediaFlags struct {
	negative        string
	n               int
	seed            int64
	images          []string
	mask            string
	audio           string
	source          string
	operation       string
	format          string
	formats         []string
	size            string
	aspect          string
	resolution      string
	duration        float64
	fps             int
	quality         string
	texture         bool
	pbr             bool
	polycount       int
	lyrics          string
	instrumental    bool
	providerOptions []string
}

// register adds the flags for the fields one family admits.
//
// The list a verb passes is that family's row of Router's own table: a mesh has
// no duration and a track has no aspect ratio, and a flag that could only ever
// be refused is worse than no flag, since a caller reads its presence as a
// promise.
func (m *mediaFlags) register(cmd *cobra.Command, names ...string) {
	f := cmd.Flags()
	for _, name := range names {
		switch name {
		case flagNegative:
			f.StringVar(&m.negative, name, "", "what the result should avoid")
		case flagN:
			f.IntVar(&m.n, name, 0, "how many to produce")
		case flagSeed:
			f.Int64Var(&m.seed, name, 0, "the seed, for a result that can be reproduced")
		case flagImage:
			f.StringArrayVar(&m.images, name, nil,
				"an image to work from: a file on this machine, a data URL, or a link (repeatable)")
		case flagMask:
			f.StringVar(&m.mask, name, "", "the mask naming the region to change; a file, a data URL, or a link")
		case flagAudioIn:
			f.StringVar(&m.audio, name, "", "the recording to work from; a file, a data URL, or a link")
		case flagSource:
			f.StringVar(&m.source, name, "",
				"the id of a generation to work from, instead of an input of your own")
		case flagOperation:
			f.StringVar(&m.operation, name, "",
				"what to ask for other than generating from text: edit, inpaint, upscale, extend, lip_sync…")
		case flagFormat:
			f.StringVar(&m.format, name, "", "the file format to ask for")
		case flagFormats:
			f.StringSliceVar(&m.formats, name, nil, "the file formats to ask for, comma-separated")
		case flagSize:
			f.StringVar(&m.size, name, "", "the size in pixels, as <width>x<height>")
		case flagAspectRatio:
			f.StringVar(&m.aspect, name, "", "the shape, as <w>:<h>; give this or --size, not both")
		case flagResolution:
			f.StringVar(&m.resolution, name, "", "a named resolution, such as 720p or 1080p")
		case flagDuration:
			f.Float64Var(&m.duration, name, 0, "how long the result should be, in seconds")
		case flagFPS:
			f.IntVar(&m.fps, name, 0, "frames per second")
		case flagQuality:
			f.StringVar(&m.quality, name, "", "the quality to ask for, in the provider's own vocabulary")
		case flagTexture:
			f.BoolVar(&m.texture, name, false, "texture the mesh")
		case flagPBR:
			f.BoolVar(&m.pbr, name, false, "produce physically based rendering materials")
		case flagPolycount:
			f.IntVar(&m.polycount, name, 0, "the polygon budget to aim for")
		case flagLyrics:
			f.StringVar(&m.lyrics, name, "", "the words to sing")
		case flagInstrumental:
			f.BoolVar(&m.instrumental, name, false, "produce a track with no vocals")
		case flagProviderOption:
			f.StringArrayVar(&m.providerOptions, name, nil,
				"a vendor parameter this contract has no field for, as key=value (repeatable); "+
					"the value is read as JSON when it is JSON, and as a string otherwise")
		default:
			// A verb asking for a flag that does not exist is a mistake in
			// this package, not in a command line.
			panic("router: no media flag named " + name)
		}
	}
}

// canonical builds the unified body.
//
// Only flags the caller actually gave are sent. /v1/generations refuses an
// unknown field and refuses a known field this family does not admit, so a flag
// left alone has to be absent from the JSON rather than present and zero.
func (m *mediaFlags) canonical(cmd *cobra.Command, model, prompt string) (canonicalRequest, error) {
	if err := m.check(cmd); err != nil {
		return canonicalRequest{}, err
	}
	given := func(name string) bool { return cmd.Flags().Changed(name) }
	request := canonicalRequest{
		Model: model, Prompt: prompt, Operation: strings.TrimSpace(m.operation),
	}
	if given(flagNegative) {
		request.NegativePrompt = m.negative
	}
	if given(flagN) {
		n := m.n
		request.N = &n
	}
	if given(flagSeed) {
		seed := m.seed
		request.Seed = &seed
	}
	inputs, err := m.canonicalInputs(cmd)
	if err != nil {
		return canonicalRequest{}, err
	}
	request.Inputs = inputs
	request.Output = m.canonicalOutput(cmd)
	if given(flagLyrics) || given(flagInstrumental) {
		music := canonicalMusic{}
		if given(flagLyrics) {
			music.Lyrics = m.lyrics
		}
		if given(flagInstrumental) {
			instrumental := m.instrumental
			music.Instrumental = &instrumental
		}
		request.Music = &music
	}
	options, err := parseProviderOptions(m.providerOptions)
	if err != nil {
		return canonicalRequest{}, err
	}
	request.ProviderOptions = options
	return request, nil
}

func (m *mediaFlags) canonicalInputs(cmd *cobra.Command) (*canonicalInputs, error) {
	given := func(name string) bool { return cmd.Flags().Changed(name) }
	if !given(flagImage) && !given(flagMask) && !given(flagAudioIn) && !given(flagSource) {
		return nil, nil
	}
	inputs := canonicalInputs{SourceGenerationID: strings.TrimSpace(m.source)}
	for _, image := range m.images {
		resolved, err := mediaInput(image, flagImage)
		if err != nil {
			return nil, err
		}
		inputs.Images = append(inputs.Images, resolved)
	}
	if given(flagMask) {
		mask, err := mediaInput(m.mask, flagMask)
		if err != nil {
			return nil, err
		}
		inputs.Mask = mask
	}
	if given(flagAudioIn) {
		audio, err := mediaInput(m.audio, flagAudioIn)
		if err != nil {
			return nil, err
		}
		inputs.Audio = audio
	}
	return &inputs, nil
}

func (m *mediaFlags) canonicalOutput(cmd *cobra.Command) *canonicalOutput {
	given := func(name string) bool { return cmd.Flags().Changed(name) }
	output := canonicalOutput{}
	filled := false
	set := func(name string, assign func()) {
		if given(name) {
			assign()
			filled = true
		}
	}
	set(flagFormat, func() { output.Format = m.format })
	set(flagFormats, func() { output.Formats = m.formats })
	set(flagSize, func() { output.Size = m.size })
	set(flagAspectRatio, func() { output.AspectRatio = m.aspect })
	set(flagResolution, func() { output.Resolution = m.resolution })
	set(flagDuration, func() { duration := m.duration; output.DurationSeconds = &duration })
	set(flagFPS, func() { fps := m.fps; output.FPS = &fps })
	set(flagQuality, func() { output.Quality = m.quality })
	set(flagTexture, func() { texture := m.texture; output.Texture = &texture })
	set(flagPBR, func() { pbr := m.pbr; output.PBR = &pbr })
	set(flagPolycount, func() { count := m.polycount; output.TargetPolycount = &count })
	if !filled {
		return nil
	}
	return &output
}

// legacyBody is the same intent in the keys the two released routes shipped
// with. Router lifts each of them onto the canonical field it means, so this is
// the same request as the one canonical() builds — spelled in the names those
// routes have always taken.
func (m *mediaFlags) legacyBody(cmd *cobra.Command, model, prompt string) (map[string]any, error) {
	if err := m.check(cmd); err != nil {
		return nil, err
	}
	given := func(name string) bool { return cmd.Flags().Changed(name) }
	body := map[string]any{"model": model}
	if strings.TrimSpace(prompt) != "" {
		body["prompt"] = prompt
	}
	if operation := strings.TrimSpace(m.operation); operation != "" {
		body["operation"] = operation
	}
	pairs := []struct {
		flag  string
		key   string
		value func() any
	}{
		{flagNegative, "negative_prompt", func() any { return m.negative }},
		{flagN, "n", func() any { return m.n }},
		{flagSeed, "seed", func() any { return m.seed }},
		{flagFormat, "output_format", func() any { return m.format }},
		{flagSize, "size", func() any { return m.size }},
		{flagAspectRatio, "aspect_ratio", func() any { return m.aspect }},
		{flagResolution, "resolution", func() any { return m.resolution }},
		{flagDuration, "duration_seconds", func() any { return m.duration }},
		{flagFPS, "fps", func() any { return m.fps }},
		{flagQuality, "quality", func() any { return m.quality }},
		{flagSource, "source_generation_id", func() any { return strings.TrimSpace(m.source) }},
	}
	for _, pair := range pairs {
		if given(pair.flag) {
			body[pair.key] = pair.value()
		}
	}
	if given(flagImage) {
		images := make([]string, 0, len(m.images))
		for _, image := range m.images {
			resolved, err := mediaInput(image, flagImage)
			if err != nil {
				return nil, err
			}
			images = append(images, resolved)
		}
		body["reference_images"] = images
	}
	if given(flagMask) {
		mask, err := mediaInput(m.mask, flagMask)
		if err != nil {
			return nil, err
		}
		// maskImage, not mask: this is FlowStudio's spelling on the released
		// image route, and it is the only one those routes lift.
		body["maskImage"] = mask
	}
	if given(flagAudioIn) {
		audio, err := mediaInput(m.audio, flagAudioIn)
		if err != nil {
			return nil, err
		}
		body["audio"] = audio
	}
	options, err := parseProviderOptions(m.providerOptions)
	if err != nil {
		return nil, err
	}
	// A released route carries a vendor knob at the top level, which is where
	// it already was: anything it does not recognize becomes a provider option
	// on the way in.
	for key, value := range options {
		body[key] = value
	}
	return body, nil
}

// check refuses locally what Router would refuse anyway, when the answer does
// not depend on the provider. A round trip to be told that two flags contradict
// each other is a round trip that could have been the error message.
func (m *mediaFlags) check(cmd *cobra.Command) error {
	given := func(name string) bool { return cmd.Flags().Changed(name) }
	if given(flagSize) && (given(flagAspectRatio) || given(flagResolution)) {
		other := "--" + flagAspectRatio
		if given(flagResolution) {
			other = "--" + flagResolution
		}
		return fmt.Errorf("--%s and %s describe the same output shape; give one", flagSize, other)
	}
	positive := []struct {
		flag string
		ok   bool
	}{
		{flagN, m.n >= 1},
		{flagDuration, m.duration > 0},
		{flagFPS, m.fps > 0},
		{flagPolycount, m.polycount > 0},
	}
	for _, rule := range positive {
		if given(rule.flag) && !rule.ok {
			return fmt.Errorf("--%s has to be a positive number", rule.flag)
		}
	}
	return nil
}

// mediaInput turns what the caller wrote into something the provider can read.
//
// A path is meaningful on this machine and nowhere else, so a local file is
// encoded as a data URL — which is also the only form FlowStudio accepts, and
// it caps an image at 10 MiB. A data URL or a link is passed through: the caller
// already has the form they want, and rewriting it would be this tree guessing.
func mediaInput(value, flag string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("--%s was given nothing", flag)
	}
	switch {
	case strings.HasPrefix(value, "data:"),
		strings.HasPrefix(value, "http://"), strings.HasPrefix(value, "https://"):
		return value, nil
	}
	encoded, err := dataURL(value)
	if err != nil {
		return "", fmt.Errorf("read --%s %s: %w", flag, value, err)
	}
	return encoded, nil
}

// parseProviderOptions reads the key=value pairs into the object Router
// forwards untouched. A value that is JSON is sent as JSON, so a number stays a
// number and a list stays a list; anything else is a string, which is what a
// vendor's enum usually is.
func parseProviderOptions(pairs []string) (map[string]json.RawMessage, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	options := make(map[string]json.RawMessage, len(pairs))
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("--%s takes key=value; %q has no key", flagProviderOption, pair)
		}
		if json.Valid([]byte(value)) {
			options[key] = json.RawMessage(value)
			continue
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("--%s %s: %w", flagProviderOption, key, err)
		}
		options[key] = encoded
	}
	return options, nil
}
