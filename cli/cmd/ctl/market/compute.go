package market

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// giBytes / miBytes are the binary memory units users may type for a
// MemorySlice allocation (--compute-binding node:device:8 == 8 Gi,
// node:device:512Mi == 512 Mi). The wire field is always bytes. A bare
// number defaults to Gi to match the column wording the rest of the compute
// surface prints.
const (
	giBytes = int64(1) << 30
	miBytes = int64(1) << 20
)

// Check-type discriminators for the compute 422 failed-check payloads. These
// mirror app-service api.CheckTypeCompute* (handler_installer_install.go /
// handler_suspend.go). They are defined locally because the CLI's pinned
// app-service module predates these constants, and — like the support-type /
// compute-mode strings the rest of the market & settings trees already mirror
// — they are a stable wire contract, not an internal symbol.
const (
	checkTypeComputeModeSelect         = "computeModeSelect"
	checkTypeComputeBindingRequired    = "computeBindingRequired"
	checkTypeComputeBindingUnavailable = "computeBindingUnavailable"
)

// parseFailedCheck extracts the structured 422 "failed check" the
// app-store / app-service surface returns inside an otherwise-failed
// APIResponse. install (computeModeSelect / appenv) and resume
// (computeBindingRequired / computeBindingUnavailable) share one envelope:
// app-service's HandleFailedCheck writes {code:422, data:{type, Data}} and
// the market backend either forwards it directly under APIResponse.Data or
// nested under data.backend_response.data — the exact shapes
// parseServerEnvError already copes with. It returns the check type and the
// raw JSON of the inner `Data` payload for the caller to decode into a
// type-specific struct. Empty checkType means "not a failed-check response".
func parseFailedCheck(resp *APIResponse) (string, json.RawMessage) {
	data := parseResponseData(resp)
	if data == nil {
		return "", nil
	}
	checkPayload := data
	if backendResp, ok := data["backend_response"].(map[string]interface{}); ok {
		backendData, ok := backendResp["data"].(map[string]interface{})
		if !ok {
			return "", nil
		}
		checkPayload = backendData
	}
	checkType, _ := checkPayload["type"].(string)
	if checkType == "" {
		return "", nil
	}
	var inner interface{} = checkPayload
	if d, ok := checkPayload["Data"]; ok && d != nil {
		inner = d
	}
	raw, err := json.Marshal(inner)
	if err != nil {
		return checkType, nil
	}
	return checkType, raw
}

// ---------------------------------------------------------------------------
// install: compute mode selection (computeModeSelect)
// ---------------------------------------------------------------------------

// computeModeSelectError is returned when an install needs a compute mode but
// none was supplied and the session can't prompt (no TTY / -q / -o json). It
// lists the installable modes so the user can re-run with --compute-mode.
type computeModeSelectError struct {
	appName     string
	installable []string
}

func (e *computeModeSelectError) Error() string {
	if len(e.installable) == 0 {
		return fmt.Sprintf("app %q needs a compute mode but none is currently installable on this cluster", e.appName)
	}
	return fmt.Sprintf("app %q supports multiple compute modes; re-run with --compute-mode <type> (installable: %s)",
		e.appName, strings.Join(e.installable, ", "))
}

// resolveComputeMode turns a computeModeSelect 422 payload into the
// selectedGpuType to retry the install with. A non-empty preset
// (--compute-mode) is validated against the plan; otherwise an interactive
// TTY prompts a choice and a non-interactive session returns
// computeModeSelectError.
func resolveComputeMode(raw json.RawMessage, appName, preset string, interactive bool) (string, error) {
	var plan []computeModePlan
	if err := json.Unmarshal(raw, &plan); err != nil {
		return "", fmt.Errorf("parse compute mode options: %w", err)
	}

	installable := make([]string, 0, len(plan))
	for _, p := range plan {
		if p.Status == "installable" {
			installable = append(installable, p.ComputeType)
		}
	}

	if preset != "" {
		norm := normalizeMarketComputeMode(preset)
		for _, p := range plan {
			if normalizeMarketComputeMode(p.ComputeType) == norm {
				if p.Status == "installable" {
					return p.ComputeType, nil
				}
				reason := strings.TrimSpace(p.Reason)
				if reason == "" {
					reason = p.Status
				}
				return "", fmt.Errorf("compute mode %q is not installable for %q: %s", preset, appName, reason)
			}
		}
		return "", fmt.Errorf("compute mode %q is not a declared mode for %q (installable: %s)",
			preset, appName, strings.Join(installable, ", "))
	}

	if !interactive || len(installable) == 0 {
		return "", &computeModeSelectError{appName: appName, installable: installable}
	}
	return promptComputeMode(appName, installable)
}

func promptComputeMode(appName string, installable []string) (string, error) {
	fmt.Fprintf(os.Stderr, "App %q supports multiple compute modes. Select one to install:\n", appName)
	for i, mode := range installable {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, marketComputeModeTitle(mode))
	}
	fmt.Fprintf(os.Stderr, "Enter choice [1-%d]: ", len(installable))
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read compute mode selection: %w", err)
	}
	idx, ok := parseChoiceIndex(strings.TrimSpace(line), len(installable))
	if !ok {
		return "", fmt.Errorf("invalid selection %q (expected 1-%d)", strings.TrimSpace(line), len(installable))
	}
	return installable[idx], nil
}

// ---------------------------------------------------------------------------
// resume: compute binding selection (computeBindingRequired / Unavailable)
// ---------------------------------------------------------------------------

// computeBindingError is returned when a resume needs a device binding but
// none was supplied and the session can't prompt. It surfaces the rejection
// reason (on *Unavailable) and the selectable devices so the user can re-run
// with --compute-binding.
type computeBindingError struct {
	appName     string
	reason      string
	requirement string
	options     []string
}

func (e *computeBindingError) Error() string {
	msg := fmt.Sprintf("app %q requires a compute binding", e.appName)
	if e.requirement != "" {
		msg += fmt.Sprintf(" (%s)", e.requirement)
	}
	if e.reason != "" {
		msg += ": " + e.reason
	}
	if len(e.options) > 0 {
		msg += fmt.Sprintf("; re-run with --compute-binding <node>:<device>[:<mem>] (available: %s)",
			strings.Join(e.options, ", "))
	} else {
		msg += "; no device is currently available to bind (free a device or stop another app, then retry)"
	}
	return msg
}

// resolveComputeBinding turns a computeBindingRequired / computeBindingUnavailable
// 422 payload into the computeBinding to retry the resume with, when the caller
// supplied NO --compute-binding. An interactive TTY prompts a selection;
// a non-interactive session returns computeBindingError listing the options.
func resolveComputeBinding(raw json.RawMessage, appName string, interactive bool) ([]BindingSelection, error) {
	var prompt computeBindingPrompt
	if err := json.Unmarshal(raw, &prompt); err != nil {
		return nil, fmt.Errorf("parse compute binding options: %w", err)
	}

	operable := operableDevices(prompt.Availability)
	reason := bindingRejectReason(&prompt)

	if !interactive || len(operable) == 0 {
		return nil, &computeBindingError{
			appName:     appName,
			reason:      reason,
			requirement: availabilityRequirementSummary(prompt.Availability),
			options:     deviceOptionLabels(operable),
		}
	}
	if reason != "" {
		fmt.Fprintf(os.Stderr, "Previous compute binding was rejected: %s\n", reason)
	}
	return promptComputeBinding(appName, prompt.Availability, operable)
}

// computeBindingRejected builds the error surfaced when an explicit
// --compute-binding was sent and the backend rejected it. It reports the
// backend's reason and the currently-available devices instead of silently
// retrying the same (rejected) selection.
func computeBindingRejected(raw json.RawMessage, appName string, attempted []BindingSelection) error {
	var prompt computeBindingPrompt
	_ = json.Unmarshal(raw, &prompt)
	reason := bindingRejectReason(&prompt)
	options := deviceOptionLabels(operableDevices(prompt.Availability))

	labels := make([]string, 0, len(attempted))
	for _, b := range attempted {
		labels = append(labels, fmt.Sprintf("%s:%s", b.NodeName, b.DeviceID))
	}
	msg := fmt.Sprintf("the supplied --compute-binding (%s) was rejected for %q", strings.Join(labels, ", "), appName)
	if reason != "" {
		msg += ": " + reason
	}
	if len(options) > 0 {
		msg += fmt.Sprintf("; available: %s", strings.Join(options, ", "))
	}
	return fmt.Errorf("%s", msg)
}

func promptComputeBinding(appName string, av *computeAvailability, operable []computeDeviceOption) ([]BindingSelection, error) {
	multi := av != nil && (av.Scope == "single-node-cards" || av.Scope == "cross-node-cards")
	if summary := availabilityRequirementSummary(av); summary != "" {
		fmt.Fprintf(os.Stderr, "App %q needs a compute device binding (%s). Available devices:\n", appName, summary)
	} else {
		fmt.Fprintf(os.Stderr, "App %q needs a compute device binding. Available devices:\n", appName)
	}
	for i, d := range operable {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, deviceOptionLine(d))
	}
	if multi {
		fmt.Fprintf(os.Stderr, "Enter one or more choices (comma-separated) [1-%d]: ", len(operable))
	} else {
		fmt.Fprintf(os.Stderr, "Enter choice [1-%d]: ", len(operable))
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read compute binding selection: %w", err)
	}
	indices, err := parseChoiceIndices(strings.TrimSpace(line), len(operable), multi)
	if err != nil {
		return nil, err
	}
	bindings := make([]BindingSelection, 0, len(indices))
	for _, idx := range indices {
		d := operable[idx]
		sel := BindingSelection{NodeName: d.NodeName, DeviceID: d.DeviceID}
		if normalizeSupportTypeKey(d.SupportType) == "memoryslice" {
			mem, err := promptBindingMemory(reader, d)
			if err != nil {
				return nil, err
			}
			sel.Memory = mem
		}
		bindings = append(bindings, sel)
	}
	return bindings, nil
}

func promptBindingMemory(reader *bufio.Reader, d computeDeviceOption) (int64, error) {
	fmt.Fprintf(os.Stderr, "  Memory to allocate on %s:%s (Gi by default; or suffix Gi/Mi) [available %s Gi]: ",
		d.NodeName, d.DeviceID, marketMemoryGi(d.Available))
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read memory allocation: %w", err)
	}
	return parseBindingMemory(strings.TrimSpace(line))
}

// ---------------------------------------------------------------------------
// shared helpers
// ---------------------------------------------------------------------------

// parseComputeBindingFlags decodes repeated --compute-binding values of the
// form <node>:<device>[:<mem>] into BindingSelection. The optional third
// segment is a MemorySlice allocation (Gi by default, or a Gi/Mi suffix)
// converted to bytes.
func parseComputeBindingFlags(raw []string) ([]BindingSelection, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]BindingSelection, 0, len(raw))
	for _, item := range raw {
		parts := strings.Split(strings.TrimSpace(item), ":")
		if len(parts) < 2 || len(parts) > 3 {
			return nil, fmt.Errorf("invalid --compute-binding %q: expected <node>:<device>[:<mem>] (mem e.g. 8, 8Gi, 512Mi)", item)
		}
		node := strings.TrimSpace(parts[0])
		device := strings.TrimSpace(parts[1])
		if node == "" || device == "" {
			return nil, fmt.Errorf("invalid --compute-binding %q: node and device must be non-empty", item)
		}
		sel := BindingSelection{NodeName: node, DeviceID: device}
		if len(parts) == 3 {
			mem, err := parseBindingMemory(parts[2])
			if err != nil {
				return nil, fmt.Errorf("invalid --compute-binding %q: %w", item, err)
			}
			sel.Memory = mem
		}
		out = append(out, sel)
	}
	return out, nil
}

// parseBindingMemory parses a MemorySlice allocation amount into bytes. It
// accepts an optional Gi / Mi suffix (case-insensitive), matching the
// frontend's two-unit VRAM input (VramAmountInput.vue); a bare number is
// treated as Gi. Examples: "8" / "8Gi" -> 8 Gi, "512Mi" -> 512 Mi.
func parseBindingMemory(s string) (int64, error) {
	orig := strings.TrimSpace(s)
	if orig == "" {
		return 0, fmt.Errorf("memory is required for this device (e.g. 8, 8Gi, or 512Mi)")
	}
	num := orig
	unit := giBytes
	switch lower := strings.ToLower(orig); {
	case strings.HasSuffix(lower, "gi"):
		num = strings.TrimSpace(orig[:len(orig)-2])
		unit = giBytes
	case strings.HasSuffix(lower, "mi"):
		num = strings.TrimSpace(orig[:len(orig)-2])
		unit = miBytes
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid memory %q: expected a positive number optionally suffixed with Gi or Mi (e.g. 8, 8Gi, 512Mi)", orig)
	}
	return int64(math.Round(v * float64(unit))), nil
}

func operableDevices(av *computeAvailability) []computeDeviceOption {
	if av == nil {
		return nil
	}
	var out []computeDeviceOption
	for _, n := range av.Nodes {
		for _, d := range n.Devices {
			if d.Operable {
				out = append(out, d)
			}
		}
	}
	return out
}

// bindingRejectReason renders a rejected BindingValidationResult into the same
// human-readable wording the SPA shows. The backend puts the actionable reason
// in Code ("node-pressure:<node>", "device-vram-insufficient:<dev>",
// "gpu-type-mismatch", ...) and only a generic "invalid" bucket in Reason, so
// we must read Code — not Reason — to tell the user anything useful.
func bindingRejectReason(p *computeBindingPrompt) string {
	if p == nil {
		return ""
	}
	if v := p.Validation; v != nil && !v.OK {
		if msg := humanizeBindingValidation(v); msg != "" {
			return msg
		}
	}
	if p.Availability != nil {
		return strings.TrimSpace(p.Availability.Reason)
	}
	return ""
}

// Backend BindingValidationResult.Reason bucket values (compute.types.go).
const (
	bindingReasonValid   = "valid"
	bindingReasonInvalid = "invalid"
)

// Validation code prefixes the SPA localizes (constant/compute.ts
// COMPUTE_VALIDATION_PREFIX). The full code is "<prefix>:<targetId>".
const (
	computeCodeAggregateVRAMInsufficient = "aggregate-vram-insufficient"
	computeCodeDeviceVRAMInsufficient    = "device-vram-insufficient"
	computeCodeDeviceMemoryInsufficient  = "device-memory-insufficient"
	computeCodeNodePressure              = "node-pressure"
)

// SPA copy for each localized validation prefix (i18n/market/en-US
// compute.validation_code.*). Kept verbatim so the CLI and the dialog say the
// same thing for the same backend code.
const (
	msgAggregateVRAMInsufficient = "The selected GPUs do not have enough combined VRAM. Add more GPUs or pick another node."
	msgDeviceVRAMInsufficient    = "This GPU does not have enough VRAM. Remove apps from this GPU to free up VRAM first."
	msgDeviceMemoryInsufficient  = "The node for this GPU does not have enough memory. Pick another GPU or node."
	msgNodePressure              = "Unable to bind to this accelerator. Scheduling the app to this node will cause overload. Please retry later or switch to another node."
)

// humanizeBindingValidation mirrors SelectComputeBindingDialog's
// topValidationMessage / nodeValidationMessage: the resource/pressure codes map
// to the localized sentence (node-pressure additionally lists the per-resource
// breakdown), while the structural codes the dialog prevents but a hand-typed
// --compute-binding can hit (gpu-type-mismatch, memory-required:<dev>,
// exclusive-already-bound:<dev>, multi-card-not-supported, ...) surface the raw
// Code — far more actionable than the generic "invalid" the backend leaves in
// Reason.
func humanizeBindingValidation(v *computeBindingValidation) string {
	prefix, _ := splitValidationCode(v.Code)
	switch prefix {
	case computeCodeNodePressure:
		if detail := nodePressureDetail(v.NodePressure); detail != "" {
			return detail
		}
		return msgNodePressure
	case computeCodeAggregateVRAMInsufficient:
		return msgAggregateVRAMInsufficient
	case computeCodeDeviceVRAMInsufficient:
		return msgDeviceVRAMInsufficient
	case computeCodeDeviceMemoryInsufficient:
		return msgDeviceMemoryInsufficient
	}
	if code := strings.TrimSpace(v.Code); code != "" && code != bindingReasonInvalid && code != bindingReasonValid {
		return code
	}
	return strings.TrimSpace(v.Reason)
}

// splitValidationCode splits "<prefix>:<targetId>" on the FIRST colon (mirrors
// the SPA parseValidationCode: deviceIds contain hyphens but no colons,
// nodeNames are plain). targetId is "" when there is no colon.
func splitValidationCode(code string) (prefix, targetID string) {
	code = strings.TrimSpace(code)
	if i := strings.IndexByte(code, ':'); i >= 0 {
		return code[:i], code[i+1:]
	}
	return code, ""
}

// nodePressureDetail mirrors SelectComputeBindingDialog.nodePressureMessage:
// the base node-pressure sentence followed by one line per pressured resource
// dimension ("Memory: Total X, Used Y, Needed Z"). Empty when there is no
// pressured dimension to report (caller then falls back to the base sentence).
func nodePressureDetail(np *computeNodePressure) string {
	if np == nil {
		return ""
	}
	lines := make([]string, 0, len(np.Dimensions))
	for _, d := range np.Dimensions {
		if !d.Pressured {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s: Total %s, Used %s, Needed %s",
			pressureResourceLabel(d.Resource),
			formatPressureAmount(d.Resource, d.Capacity),
			formatPressureAmount(d.Resource, d.Used),
			formatPressureAmount(d.Resource, d.Required)))
	}
	if len(lines) == 0 {
		return ""
	}
	return msgNodePressure + "\n" + strings.Join(lines, "\n")
}

// pressureResourceLabel mirrors compute.pressure_resource.* (en-US).
func pressureResourceLabel(resource string) string {
	switch resource {
	case "memory":
		return "Memory"
	case "cpu":
		return "CPU"
	case "disk":
		return "Disk"
	default:
		return resource
	}
}

// formatPressureAmount mirrors formatComputePressureAmount: cpu values are
// millicores rendered as cores, memory/disk are bytes rendered as Gi.
func formatPressureAmount(resource string, value int64) string {
	if resource == "cpu" {
		return marketCPUCores(value)
	}
	return marketMemoryGi(value) + " Gi"
}

// marketCPUCores mirrors formatComputeCpuCores: millicores -> "<cores> Core",
// trimmed to at most 2 decimals.
func marketCPUCores(milli int64) string {
	if milli <= 0 {
		return "0 Core"
	}
	cores := float64(milli) / 1000
	rounded := math.Round(cores*100) / 100
	if rounded == math.Trunc(rounded) {
		return fmt.Sprintf("%d Core", int64(rounded))
	}
	return fmt.Sprintf("%s Core", strconv.FormatFloat(rounded, 'f', 2, 64))
}

func deviceOptionLabels(devices []computeDeviceOption) []string {
	labels := make([]string, 0, len(devices))
	for _, d := range devices {
		labels = append(labels, fmt.Sprintf("%s:%s", d.NodeName, d.DeviceID))
	}
	return labels
}

func deviceOptionLine(d computeDeviceOption) string {
	name := strings.TrimSpace(d.CardModel)
	if name == "" {
		name = d.DeviceID
	}
	parts := []string{fmt.Sprintf("%s:%s", d.NodeName, d.DeviceID), name}
	if d.SupportType != "" {
		parts = append(parts, d.SupportType)
	}
	parts = append(parts, fmt.Sprintf("avail %s / %s Gi", marketMemoryGi(d.Available), marketMemoryGi(d.Capacity)))
	if d.FitLevel != "" {
		parts = append(parts, "fit="+d.FitLevel)
	}
	return strings.Join(parts, "  ")
}

// parseChoiceIndex parses a single 1-based selection into a 0-based index.
func parseChoiceIndex(s string, n int) (int, bool) {
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 || v > n {
		return 0, false
	}
	return v - 1, true
}

// parseChoiceIndices parses a 1-based selection (single, or comma-separated
// when multi) into deduplicated 0-based indices preserving input order.
func parseChoiceIndices(s string, n int, multi bool) ([]int, error) {
	fields := []string{s}
	if multi {
		fields = strings.Split(s, ",")
	}
	seen := make(map[int]struct{})
	out := make([]int, 0, len(fields))
	for _, f := range fields {
		idx, ok := parseChoiceIndex(strings.TrimSpace(f), n)
		if !ok {
			return nil, fmt.Errorf("invalid selection %q (expected 1-%d)", strings.TrimSpace(f), n)
		}
		if _, dup := seen[idx]; dup {
			continue
		}
		seen[idx] = struct{}{}
		out = append(out, idx)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no selection provided")
	}
	return out, nil
}

func marketMemoryGi(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	gi := float64(bytes) / float64(giBytes)
	floored := math.Floor(gi*100) / 100
	if floored == math.Trunc(floored) {
		return strconv.FormatInt(int64(floored), 10)
	}
	return strconv.FormatFloat(floored, 'f', 2, 64)
}

// giRoundEpsilon mirrors the SPA GI_ROUND_EPSILON: absorbs byte<->Gi float
// drift before ceil at 2 decimals so a value that is a hair over an exact Gi
// (rounding noise) doesn't bump the displayed need up by 0.01.
const giRoundEpsilon = 1e-6

// marketMemoryGiCeil mirrors the SPA formatComputeMemoryCeil: bytes -> Gi
// rounded UP to 2 decimals (minus the drift epsilon), so a displayed "requires"
// never understates the real need. Returns "0" for non-positive input.
func marketMemoryGiCeil(bytes int64) string {
	if bytes <= 0 {
		return "0"
	}
	gi := float64(bytes) / float64(giBytes)
	ceiled := math.Ceil(gi*100-giRoundEpsilon) / 100
	if ceiled == math.Trunc(ceiled) {
		return strconv.FormatInt(int64(ceiled), 10)
	}
	return strconv.FormatFloat(ceiled, 'f', 2, 64)
}

// requiredResourceBytes mirrors getComputeRequirementResourceBytes: the
// threshold the SPA shows as "required" — RequiredGpu for an nvidia
// requirement, RequiredMemory otherwise.
func requiredResourceBytes(req *computeRequirement) int64 {
	if req == nil {
		return 0
	}
	if normalizeMarketComputeMode(req.Mode) == "nvidia" {
		return req.RequiredGpu
	}
	return req.RequiredMemory
}

// requirementSummary mirrors the SPA useComputeSummary.requirementTags plus the
// required amount: "<mode label>, <scope chip>[, requires <ceil> Gi]". The
// scope chip follows requirementTags: nvidia + multi-node -> "multi-node",
// else multi-card -> "multi-card", else "single card".
func requirementSummary(req *computeRequirement) string {
	if req == nil {
		return ""
	}
	parts := make([]string, 0, 3)
	if label := marketComputeModeTitle(req.Mode); label != "" {
		parts = append(parts, label)
	}
	switch {
	case normalizeMarketComputeMode(req.Mode) == "nvidia" && req.SupportMultiNodes:
		parts = append(parts, "multi-node")
	case req.SupportMultiCards:
		parts = append(parts, "multi-card")
	default:
		parts = append(parts, "single card")
	}
	if b := requiredResourceBytes(req); b > 0 {
		parts = append(parts, fmt.Sprintf("requires %s Gi", marketMemoryGiCeil(b)))
	}
	return strings.Join(parts, ", ")
}

// availabilityRequirementSummary is requirementSummary for an availability
// payload, guarding the nil chain.
func availabilityRequirementSummary(av *computeAvailability) string {
	if av == nil {
		return ""
	}
	return requirementSummary(av.Requirement)
}

func normalizeMarketComputeMode(mode string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(mode)), "_", "-")
}

func normalizeSupportTypeKey(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
}

// marketComputeModeTitle mirrors the brand display names used by
// `settings compute`; unknown values fall back to the raw mode string.
var marketComputeModeTitleMap = map[string]string{
	"cpu":         "CPU",
	"nvidia":      "NVIDIA GPU",
	"nvidia-gb10": "NVIDIA GB10",
	"apple-m":     "Apple Silicon-M",
	"intel":       "Intel",
	"amd":         "AMD",
	"intel-gpu":   "Intel GPU",
	"amd-gpu":     "AMD GPU",
	"moore-soc":   "Moore Threads",
}

func marketComputeModeTitle(mode string) string {
	if t, ok := marketComputeModeTitleMap[normalizeMarketComputeMode(mode)]; ok {
		return fmt.Sprintf("%s (%s)", t, mode)
	}
	return strings.TrimSpace(mode)
}

// isComputeModeSelect / isComputeBindingPrompt classify a parsed check type.
func isComputeModeSelect(checkType string) bool {
	return checkType == checkTypeComputeModeSelect
}

func isComputeBindingPrompt(checkType string) bool {
	return checkType == checkTypeComputeBindingRequired ||
		checkType == checkTypeComputeBindingUnavailable
}
