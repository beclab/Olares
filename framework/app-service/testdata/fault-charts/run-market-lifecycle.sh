#!/usr/bin/env bash
# Run on an Olares node. Cover the 42 application lifecycle states and write a full report.
#
#   cd /code/beclab/app-service/framework/app-service
#   bash testdata/fault-charts/run-market-lifecycle.sh
#
# Report directory (a latest symlink is created):
#   /tmp/market-lifecycle-reports/latest/report.html
#
# Run only some scenarios (comma-separated):
#   ONLY=faultpull,pressure bash testdata/fault-charts/run-market-lifecycle.sh
# Skip the final fixture pass:
#   SKIP_FREEZE=1 bash testdata/fault-charts/run-market-lifecycle.sh
# Do not clean existing test apps and do not upload again:
#   SKIP_RESET=1 SKIP_UPLOAD=1 bash testdata/fault-charts/run-market-lifecycle.sh
#
# Each version bump adds only 0.0.1 (0.0.1 -> 0.0.2, 0.3.1 -> 0.3.2).
# When the catalog requires a version higher than the published one, the next version is that version plus 0.0.1.
# Before upload, stopcpu, stopmem, stopdisk, and stopk8scpu are rewritten to fit node capacity so install is not rejected for exceeding the machine.
# A failed upload or install prints the CLI output and exits immediately.
# The first step checks login with profile whoami --refresh. On failure it tells you to profile login and exits.
#
# This script deletes only the test apps listed in APPS. The fixture section scales
# os-framework/app-service to 0 and back up so the first reconcile sees an already expired state.

set -uo pipefail

CHART_DIR=$(cd "$(dirname "$0")" && pwd)
STAMP=$(date -u +%Y%m%dT%H%M%SZ)
REPORT_DIR=${REPORT_DIR:-/tmp/market-lifecycle-reports/$STAMP}
LOG="$REPORT_DIR/run.log"
TRACE="$REPORT_DIR/trace.jsonl"
OWNER=${OWNER:-}
ONLY=${ONLY:-}
SKIP_RESET=${SKIP_RESET:-0}
SKIP_UPLOAD=${SKIP_UPLOAD:-0}
SKIP_FREEZE=${SKIP_FREEZE:-0}
WAIT_FAST=${WAIT_FAST:-180}
WAIT_INSTALL=${WAIT_INSTALL:-600}
WAIT_LONG=${WAIT_LONG:-480}
COMPUTE_MODE=${COMPUTE_MODE:-cpu}
FROZEN=0
POLL_PID=""

APPS="faultgood faultpull faultinstall holdinst stopinit stopdisk stopmem stopk8scpu stopcpu upgradetest ginit gpull gcont applyenv faultunready faulthanginit faultslowhelm faultdelhook faultmw faultgpu"

STATES="pending downloading installing initializing running stopping stopped resuming uninstalling uninstalled applyingEnv upgradePending upgradeDownloading upgrading pendingCanceling downloadingCanceling installingCanceling initializingCanceling resumingCanceling applyingEnvCanceling upgradePendingCanceling upgradeDownloadingCanceling upgradingCanceling pendingCanceled downloadingCanceled installingCanceled downloadFailed installFailed stopFailed uninstallFailed resumeFailed upgradeFailed applyEnvFailed pendingCancelFailed downloadingCancelFailed installingCancelFailed initializingCancelFailed resumingCancelFailed applyingEnvCancelFailed upgradePendingCancelFailed upgradeDownloadingCancelFailed upgradingCancelFailed"

say() { echo "$*" | tee -a "$LOG"; }

die() { say "ERROR: $*"; exit 1; }

# On command failure, print the CLI output and exit. Do not swallow the error and wait on a state.
cli_must() {
  local desc=$1
  shift
  local out
  say "+ $*"
  if out=$("$@" 2>&1); then
    printf '%s\n' "$out" >>"$LOG"
    return 0
  fi
  printf '%s\n' "$out" | tee -a "$LOG"
  die "$desc"
}

need() { command -v "$1" >/dev/null 2>&1 || die "missing command $1"; }

enabled() {
  local name=$1
  [[ -z "$ONLY" ]] && return 0
  local ifs=$IFS x
  IFS=','
  for x in $ONLY; do
    [[ "$x" == "$name" ]] && { IFS=$ifs; return 0; }
  done
  IFS=$ifs
  return 1
}

ago_stamp() {
  date -u -d "$1" +%Y-%m-%dT%H:%M:%SZ
}

set_method() { echo "$1" >"$REPORT_DIR/method"; }
set_case() { echo "$1" >"$REPORT_DIR/case"; }

method_now() { cat "$REPORT_DIR/method" 2>/dev/null || echo "natural"; }
case_now() { cat "$REPORT_DIR/case" 2>/dev/null || echo "-"; }

record_case_result() {
  local app=$1 wanted=$2 actual=$3 passed=$4 reason=${5:-}
  jq -nc \
    --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg case "$(case_now)" \
    --arg app "$app" \
    --arg wanted "$wanted" \
    --arg actual "$actual" \
    --arg reason "$reason" \
    --argjson passed "$passed" \
    '{ts:$ts,case:$case,app:$app,wanted:$wanted,actual:$actual,reason:$reason,passed:$passed}' \
    >>"$REPORT_DIR/case-results.jsonl" || true
}

with_lock() {
  local lock="$REPORT_DIR/report.lock" i=0
  while ! mkdir "$lock" 2>/dev/null; do
    sleep 0.1
    i=$((i + 1))
    [[ "$i" -gt 100 ]] && return 0
  done
  "$@"
  local rc=$?
  rmdir "$lock" 2>/dev/null || true
  return $rc
}

am_of() {
  local app=$1 cache="$REPORT_DIR/am-$app" guess found
  if [[ -s "$cache" ]]; then
    cat "$cache"
    return 0
  fi
  guess="$app-$OWNER-$app"
  if kubectl get applicationmanager "$guess" -o name >/dev/null 2>&1; then
    echo "$guess" | tee "$cache"
    return 0
  fi
  found=$(kubectl get applicationmanager -o json 2>/dev/null | jq -r --arg a "$app" 'first(.items[] | select(.spec.appName==$a) | .metadata.name) // empty')
  [[ -n "$found" ]] || return 1
  echo "$found" | tee "$cache"
}

read_snap() {
  local app=$1 am
  am=$(am_of "$app") || return 1
  kubectl get applicationmanager "$am" -o json
}

current_state() {
  local snap
  snap=$(read_snap "$1" 2>/dev/null) || { echo ""; return 1; }
  jq -r '.status.state // empty' <<<"$snap"
}

seen_state() {
  [[ -f "$TRACE" ]] && grep -F "\"state\":\"$1\"" "$TRACE" >/dev/null 2>&1
}

ns_of_snap() {
  local snap=$1 app=$2 ns
  ns=$(jq -r '.spec.appNamespace // empty' <<<"$snap")
  [[ -n "$ns" ]] || ns="$app-$OWNER"
  echo "$ns"
}

spec_op() {
  local app=$1 op=$2 am id
  am=$(am_of "$app") || return 1
  id="lifecycle-$(date +%s)-$RANDOM"
  say "spec $app opType=$op request=$id"
  kubectl patch applicationmanager "$am" --type=merge -p "$(jq -n --arg op "$op" --arg id "$id" '{spec:{opType:$op},metadata:{annotations:{"bytetrade.io/operation-request-id":$id}}}')"
}

cli_try() {
  say "+ $*"
  "$@" >>"$LOG" 2>&1
}

market_op() {
  local verb=$1 app=$2 op=$3
  if cli_try olares-cli market "$verb" "$app" -o json; then
    return 0
  fi
  spec_op "$app" "$op"
}

app_version() {
  local app=$1 dir name ver found=""
  for dir in "$CHART_DIR"/*/; do
    [[ -f "$dir/Chart.yaml" ]] || continue
    read -r name ver <<<"$(chart_meta "$dir/Chart.yaml")"
    [[ "$name" == "$app" ]] || continue
    if [[ -z "$found" || "$ver" < "$found" ]]; then
      found=$ver
    fi
  done
  echo "${found:-0.1.0}"
}

note_uploaded_version() {
  mkdir -p "$REPORT_DIR/uploaded"
  printf '%s\n' "$2" >"$REPORT_DIR/uploaded/$1"
}

uploaded_version() {
  local f="$REPORT_DIR/uploaded/$1"
  [[ -f "$f" ]] && cat "$f"
}

chart_dir_for() {
  local app=$1 dir name base found=""
  for dir in "$CHART_DIR"/*/; do
    [[ -f "$dir/Chart.yaml" ]] || continue
    read -r name _ <<<"$(chart_meta "$dir/Chart.yaml")"
    [[ "$name" == "$app" ]] || continue
    base=$(basename "$dir")
    if [[ "$base" == "$app" ]]; then
      echo "${dir%/}"
      return 0
    fi
    [[ -z "$found" ]] && found=${dir%/}
  done
  [[ -n "$found" ]] || return 1
  echo "$found"
}

missing_catalog_version() {
  local out=$1
  [[ "$out" == *"HTTP 404"* && ( "$out" == *"has no version"* || "$out" == *"cataloged version"* || "$out" == *"App not found"* ) ]]
}

install_market() {
  local app=$1 ver=$2 mode=$3
  local -a cmd=(olares-cli market install "$app" -s upload -o json --compute-mode "$mode")
  [[ -n "$ver" ]] && cmd+=(--version "$ver")
  say "+ ${cmd[*]}"
  if INSTALL_OUT=$("${cmd[@]}" 2>&1); then
    printf '%s\n' "$INSTALL_OUT" >>"$LOG"
    return 0
  fi
  printf '%s\n' "$INSTALL_OUT" | tee -a "$LOG"
  return 1
}

install_app() {
  local app=$1 ver=${2:-} mode=${3:-$COMPUTE_MODE} dir before after
  # 0.1.0 is a caller placeholder. The catalog accepts only the current uploaded version, so use that.
  if [[ -z "$ver" || "$ver" == "0.1.0" ]]; then
    ver=$(uploaded_version "$app")
    [[ -n "$ver" ]] || ver=$(app_version "$app")
  fi
  if install_market "$app" "$ver" "$mode"; then
    return 0
  fi
  if ! missing_catalog_version "$INSTALL_OUT"; then
    die "install $app failed"
  fi
  dir=$(chart_dir_for "$app") || die "install $app failed, and no chart directory was found"
  say "$app has no version $ver, uploading again"
  read -r _ before <<<"$(chart_meta "$dir/Chart.yaml")"
  upload_one "$dir"
  after=$(cat "$REPORT_DIR/raised-version")
  if [[ "$after" == "$before" ]]; then
    say "$app is still $before, bumping another 0.0.1"
    raise_chart_version "$dir"
    upload_one "$dir"
    after=$(cat "$REPORT_DIR/raised-version")
  fi
  ver=$after
  install_market "$app" "$ver" "$mode" || die "install $app failed"
}

upgrade_app() {
  local app=$1 ver=$2
  cli_must "upgrade $app to $ver failed" \
    olares-cli market upgrade "$app" -s upload --version "$ver" -o json
}

env_set() {
  local app=$1 var=$2
  say "+ settings apps env set $app --var $var"
  olares-cli settings apps env set "$app" --var "$var" >>"$LOG" 2>&1 || true
}

annotation_of() {
  local snap
  snap=$(read_snap "$1" 2>/dev/null) || { echo ""; return 1; }
  jq -r --arg key "$2" '.metadata.annotations[$key] // empty' <<<"$snap"
}

deploy_replicas() {
  local app=$1 ns
  ns=$(kubectl get applicationmanager "$(am_of "$app")" -o jsonpath='{.spec.appNamespace}' 2>/dev/null || echo "$app-$OWNER")
  kubectl -n "$ns" get deploy "$app" -o jsonpath='{.spec.replicas}' 2>/dev/null || echo ""
}

expect_eq() {
  local app=$1 wanted=$2 actual=$3
  if [[ "$actual" == "$wanted" ]]; then
    record_case_result "$app" "$wanted" "$actual" true
    return 0
  fi
  record_case_result "$app" "$wanted" "${actual:-empty}" false
  say "case failed: $app wanted $wanted, got ${actual:-empty}"
  return 1
}

backdate() {
  local app=$1 ago=$2 am stamp
  am=$(am_of "$app") || return 1
  stamp=$(ago_stamp "$ago") || return 1
  say "backdate $app statusTime=$stamp ($ago)"
  set_method "backdate"
  kubectl patch applicationmanager "$am" --type=merge -p "{\"status\":{\"statusTime\":\"$stamp\"}}"
}

hold_ns() {
  local ns=$1
  kubectl get ns "$ns" >/dev/null 2>&1 || kubectl create ns "$ns" >>"$LOG" 2>&1 || return 1
  local cur
  cur=$(kubectl get ns "$ns" -o json)
  if jq -e '.metadata.finalizers // [] | index("test.olares/hold")' <<<"$cur" >/dev/null; then
    return 0
  fi
  jq '.metadata.finalizers = ((.metadata.finalizers // []) + ["test.olares/hold"] | unique)' <<<"$cur" \
    | kubectl replace -f - >>"$LOG" 2>&1
}

release_holds() {
  local app ns obj
  for app in $APPS; do
    ns="$app-$OWNER"
    kubectl get ns "$ns" >/dev/null 2>&1 || continue
    obj=$(kubectl get ns "$ns" -o json 2>/dev/null) || continue
    if jq -e '.metadata.finalizers // [] | index("test.olares/hold")' <<<"$obj" >/dev/null; then
      say "remove finalizer $ns"
      jq '.metadata.finalizers |= map(select(. != "test.olares/hold"))' <<<"$obj" \
        | kubectl replace --raw "/api/v1/namespaces/$ns/finalize" -f - >>"$LOG" 2>&1 || true
    fi
  done
}

freeze() {
  say "freeze app-service (scale replicas to 0)"
  # Scaling the StatefulSet to 0 deletes the Pod and its logs.
  # Keep the state-change lines before that happens.
  harvest_state_changes || true
  kubectl -n os-framework scale sts/app-service --replicas=0 >>"$LOG" 2>&1
  kubectl -n os-framework rollout status sts/app-service --timeout=180s >>"$LOG" 2>&1
  FROZEN=1
}

thaw() {
  local n
  n=$(kubectl -n os-framework get sts app-service -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "1")
  if [[ "$n" == "0" ]]; then
    say "thaw app-service"
    kubectl -n os-framework scale sts/app-service --replicas=1 >>"$LOG" 2>&1 || true
    kubectl -n os-framework rollout status sts/app-service --timeout=240s >>"$LOG" 2>&1 || true
  fi
  FROZEN=0
}

on_exit() {
  [[ -n "$POLL_PID" ]] && kill "$POLL_PID" >/dev/null 2>&1 || true
  release_holds
  thaw || true
  if [[ -f "$REPORT_DIR/meta.env" ]]; then
    rebuild_report || true
  fi
}
trap on_exit EXIT

init_docs() {
  local st enter leave
  mkdir -p "$REPORT_DIR/states"
  : >"$TRACE"
  : >"$REPORT_DIR/state-plans.jsonl"
  : >"$REPORT_DIR/case-results.jsonl"
  : >"$REPORT_DIR/state-changes.txt"
  set_method "natural"
  set_case "-"
  while IFS='|' read -r st enter leave; do
    [[ -z "${st:-}" ]] && continue
    jq -nc --arg state "$st" --arg enter "$enter" --arg leave "$leave" \
      '{state:$state,enter:$enter,leave:$leave}' >>"$REPORT_DIR/state-plans.jsonl"
  done <<'EOF'
pending|T01 install creates the AM|downloading; pendingCanceling
downloading|T02 image pull|installing; downloadingCanceling; downloadFailed
installing|T02 helm install|initializing; running; installingCanceling; stopping; installFailed
initializing|T02 startup|running; initializingCanceling
running|T02 install finished|stopping; uninstalling; upgradePending; applyingEnv
stopping|T04 stop, or resource check / startup failure|stopped; stopFailed
stopped|T04 stop finished|resuming; uninstalling; upgradePending; applyingEnv
resuming|T06 resume|initializing; resumingCanceling; stopping; resumeFailed
uninstalling|T05 or T07 uninstall|uninstalled; uninstallFailed
uninstalled|T05 uninstall finished|pending
applyingEnv|T10 or T09 env change|initializing; stopped; stopping; applyingEnvCanceling; applyEnvFailed
upgradePending|T18 or T08 upgrade|upgradeDownloading; upgradePendingCanceling
upgradeDownloading|T18 upgrade image pull|upgrading; upgradeDownloadingCanceling; upgradeFailed
upgrading|T19 upgrade helm|initializing; stopped; upgradingCanceling; upgradeFailed
pendingCanceling|T12 cancel pending|pendingCanceled; pendingCancelFailed
downloadingCanceling|T13 or T57 cancel download|downloadingCanceled; downloadingCancelFailed
installingCanceling|T14, T34, T45|installingCanceled; installFailed; installingCancelFailed
initializingCanceling|T15 or T46|stopping; initializingCancelFailed
resumingCanceling|T16 or T60|stopping; resumingCancelFailed
applyingEnvCanceling|T17 or T59|stopping; applyingEnvCancelFailed
upgradePendingCanceling|T22|running; stopped; upgradePendingCancelFailed
upgradeDownloadingCanceling|T25, T58, or S1|running; stopped; upgradeDownloadingCancelFailed
upgradingCanceling|T27 or T61|upgradeFailed; upgradingCancelFailed
pendingCanceled|T12|pending
downloadingCanceled|T13|pending
installingCanceled|T14|pending
downloadFailed|T33 image pull failed|pending
installFailed|T34 or T36|uninstalled; uninstalling
stopFailed|T41 or T47|stopping; uninstalling
uninstallFailed|T42 or T48|uninstalling
resumeFailed|T40 resume after deleting the Helm release|resuming; applyingEnv; uninstalling
upgradeFailed|T28, T37, T38|upgradePending; uninstalling
applyEnvFailed|T39 BREAK=1|applyingEnv; stopping; uninstalling
pendingCancelFailed|T50 cancel timed out|pendingCanceling
downloadingCancelFailed|T51|downloadingCanceling
installingCancelFailed|T49|installingCanceling; uninstalling
initializingCancelFailed|T52|initializingCanceling; uninstalling
resumingCancelFailed|T53|stay; resumingCanceling
applyingEnvCancelFailed|T54|applyingEnvCanceling; uninstalling
upgradePendingCancelFailed|T55|upgradePendingCanceling
upgradeDownloadingCancelFailed|T43 or S1|upgradeDownloadingCanceling
upgradingCancelFailed|T56|upgradingCanceling; uninstalling
EOF
}

record_visit() {
  local app=$1 snap st reason msg am ns phase method casen ext pods helm logs market body n
  snap=$(cat "$REPORT_DIR/snap-$app.json")
  st=$(jq -r '.status.state // empty' <<<"$snap")
  [[ -n "$st" ]] || return 0
  reason=$(jq -r '.status.reason // ""' <<<"$snap")
  msg=$(jq -r '.status.message // ""' <<<"$snap")
  am=$(jq -r '.metadata.name' <<<"$snap")
  ns=$(ns_of_snap "$snap" "$app")
  phase=$(jq -r '(.status.payload // {}) | to_entries | map(select(.key | test("phase$|terminal$|cleanup$"))) | map("\(.key)=\(.value)") | join(", ")' <<<"$snap")
  method=$(method_now)
  casen=$(case_now)
  ext=$(jq '{
    am:.metadata.name,
    appNamespace:.spec.appNamespace,
    specOp:.spec.opType,
    type:.spec.type,
    configVersion: ((.spec.config // "{}") | if type=="string" then (fromjson? // {}) else . end | .Version // .version // null),
    chart:.metadata.annotations["bytetrade.io/chart-version"],
    target:.metadata.annotations["bytetrade.io/target-chart-version"],
    preState:.metadata.annotations["bytetrade.io/pre-upgrade-state"],
    preConfig:.metadata.annotations["bytetrade.io/pre-upgrade-config"],
    requestID:.metadata.annotations["bytetrade.io/operation-request-id"],
    status:.status
  }' <<<"$snap")
  if kubectl get ns "$ns" >/dev/null 2>&1; then
    pods=$(kubectl get pods -n "$ns" -o json 2>/dev/null | jq -c '[.items[]? | {name:.metadata.name,phase:.status.phase,reason:.status.reason}]' || echo "[]")
    if command -v helm >/dev/null 2>&1; then
      helm=$(timeout 20 helm -n "$ns" history "$app" --max 5 2>&1 | tail -20 || true)
    else
      helm="helm is not installed on this node"
    fi
  else
    pods="namespace does not exist"
    helm="namespace does not exist"
  fi
  logs=$(kubectl -n os-framework logs sts/app-service --since=30m --tail=4000 2>/dev/null \
    || kubectl -n os-framework logs sts/app-service -c app-service --since=30m --tail=4000 2>/dev/null \
    || true)
  logs=$(printf '%s\n' "$logs" | grep 'state change' | grep -F "$app" | tail -30 || true)
  market=$(timeout 20 olares-cli market status "$app" -o json 2>/dev/null | head -c 3000 || true)
  jq -nc --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg app "$app" --arg state "$st" --arg reason "$reason" --arg case "$casen" --arg method "$method" \
    '{ts:$ts,app:$app,state:$state,reason:$reason,case:$case,method:$method}' >>"$TRACE"
  if [[ -z "$ext" ]]; then
    ext='{}'
  fi
  jq -nc \
    --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg app "$app" \
    --arg method "$method" \
    --arg case "$casen" \
    --arg am "$am" \
    --arg ns "$ns" \
    --arg reason "$reason" \
    --arg message "$msg" \
    --arg phase "${phase:-none}" \
    --argjson ext "$ext" \
    --arg pods "$pods" \
    --arg helm "$helm" \
    --arg logs "${logs:-none}" \
    --arg market "${market:-none}" \
    '{ts:$ts,app:$app,method:$method,case:$case,am:$am,ns:$ns,reason:$reason,message:$message,phase:$phase,ext:$ext,pods:$pods,helm:$helm,logs:$logs,market:$market}' \
    >>"$REPORT_DIR/states/$st.jsonl"
  say "$(date -u +%H:%M:%S) $app -> $st reason=${reason:--} case=$casen ($method)"
}

poll_app() {
  local app=$1 timeout=$2 until=${3:-} leave=${4:-}
  local end=$((SECONDS + timeout)) last="" snap st fp wanted="" reason=""
  if [[ -n "$leave" ]]; then
    wanted="leave $leave"
  elif [[ -n "$until" ]]; then
    wanted="$until"
  fi
  while (( SECONDS < end )); do
    snap=$(read_snap "$app" 2>/dev/null) || { sleep 0.5; continue; }
    st=$(jq -r '.status.state // empty' <<<"$snap")
    reason=$(jq -r '.status.reason // empty' <<<"$snap" 2>/dev/null || true)
    fp=$(jq -c '[.status.state,.status.reason,.status.opID,(.status.payload // {})]' <<<"$snap" 2>/dev/null || echo "")
    if [[ -n "$fp" && "$fp" != "$last" ]]; then
      printf '%s\n' "$snap" >"$REPORT_DIR/snap-$app.json"
      with_lock record_visit "$app"
      last=$fp
    fi
    fire_hook "$app" "$st" || true
    if [[ -n "$leave" && -n "$st" && "$st" != "$leave" ]]; then
      [[ -n "$wanted" ]] && record_case_result "$app" "$wanted" "$st" true "$reason"
      return 0
    fi
    if [[ -z "$leave" && -n "$until" && "$st" =~ $until ]]; then
      [[ -n "$wanted" ]] && record_case_result "$app" "$wanted" "$st" true "$reason"
      return 0
    fi
    sleep 0.5
  done
  [[ -n "$wanted" ]] && record_case_result "$app" "$wanted" "${st:-absent}" false "$reason"
  say "wait timed out: $app wanted /${until:-leave $leave}/, last state ${st:-absent}"
  return 1
}

watch_install() {
  local app=$1 ver=$2 mode=$3 timeout=$4 until=$5
  rm -f "$REPORT_DIR/am-$app"
  ensure_slot "$app"
  poll_app "$app" "$timeout" "$until" &
  POLL_PID=$!
  install_app "$app" "$ver" "$mode"
  wait "$POLL_PID" || true
  POLL_PID=""
}

ensure_slot() {
  local self=$1 app st
  for app in $APPS; do
    [[ "$app" == "$self" ]] && continue
    st=$(current_state "$app" 2>/dev/null || true)
    if [[ "$st" == "downloading" || "$st" == "upgradeDownloading" ]]; then
      say "$app holds the download slot ($st), cancel it first"
      market_op cancel "$app" cancel || true
      poll_app "$app" 90 'Canceled|Failed|running|stopped|uninstalled' || true
    fi
  done
}

arm_hook() {
  jq -n --arg when "$2" --arg action "$3" --arg arg "${4:-}" '{when:$when,action:$action,arg:$arg}' >"$REPORT_DIR/hook-$1.json"
}

fire_hook() {
  local app=$1 st=$2 f="$REPORT_DIR/hook-$1.json" when action arg
  [[ -f "$f" ]] || return 0
  when=$(jq -r '.when' "$f")
  [[ "$st" == "$when" ]] || return 0
  action=$(jq -r '.action' "$f")
  arg=$(jq -r '.arg' "$f")
  rm -f "$f"
  say "hook $app state=$st action=$action arg=$arg"
  case "$action" in
    cancel) market_op cancel "$app" cancel || true ;;
    stop) market_op stop "$app" stop || true ;;
    backdate) backdate "$app" "$arg" || true ;;
    hold-ns) hold_ns "$app-$OWNER" || true ;;
  esac
}

clear_app() {
  local app=$1 am ns
  say "clean $app"
  am=$(kubectl get applicationmanager -o json 2>/dev/null | jq -r --arg a "$app" 'first(.items[] | select(.spec.appName==$a) | .metadata.name) // empty')
  if [[ -n "$am" ]]; then
    ns=$(kubectl get applicationmanager "$am" -o jsonpath='{.spec.appNamespace}' 2>/dev/null || true)
    market_op uninstall "$app" uninstall || true
    sleep 2
    kubectl delete applicationmanager "$am" --wait=false >>"$LOG" 2>&1 || true
  else
    ns="$app-$OWNER"
  fi
  if [[ -n "$ns" ]] && kubectl get ns "$ns" >/dev/null 2>&1; then
    kubectl get ns "$ns" -o json 2>/dev/null | jq '.metadata.finalizers=[]' \
      | kubectl replace --raw "/api/v1/namespaces/$ns/finalize" -f - >>"$LOG" 2>&1 || true
    kubectl delete ns "$ns" --wait=false >>"$LOG" 2>&1 || true
  fi
  rm -f "$REPORT_DIR/am-$app"
  olares-cli market delete "$app" -o json >>"$LOG" 2>&1 || true
}

chart_meta() {
  python3 - "$1" <<'PY'
import re, sys
text = open(sys.argv[1], encoding="utf-8").read()
name = re.search(r"(?m)^name:\s*(\S+)", text)
ver = re.search(r"(?m)^version:\s*(\S+)", text)
print((name.group(1) if name else ""), (ver.group(1) if ver else ""))
PY
}

# Bump only the last version component by 0.0.1. If the second argument is set, write the chart to that version instead of bumping.
# Do not change olaresManifest.version or dependency constraints. The new version is written to raised-version.
apply_chart_version() {
  local dir=${1%/} explicit=${2:-} out
  if ! out=$(python3 - "$dir" "$explicit" 2>&1 <<'PY'
import pathlib, re, sys
root = pathlib.Path(sys.argv[1])
explicit = sys.argv[2].strip()
chart_path = root / "Chart.yaml"
man_path = root / "OlaresManifest.yaml"
chart = chart_path.read_text(encoding="utf-8")
match = re.search(r"(?m)^version:\s*(\S+)", chart)
if not match:
    raise SystemExit(f"{chart_path} has no version")

def next_patch(version):
    version = version.strip("'\"")
    parts = version.split(".")
    if not parts or not parts[-1].isdigit():
        raise SystemExit(f"cannot bump version {version}")
    parts[-1] = str(int(parts[-1]) + 1)
    return ".".join(parts)

old = match.group(1).strip("'\"")
new = explicit or next_patch(old)
chart = re.sub(r"(?m)^version:\s*\S+", f"version: {new}", chart, count=1)
chart = re.sub(r'(?m)^appVersion:\s*\S+', f'appVersion: "{new}"', chart, count=1)
chart_path.write_text(chart, encoding="utf-8")
man = man_path.read_text(encoding="utf-8")
man = re.sub(r"(?m)^  version:\s*['\"][^'\"]+['\"]", f"  version: '{new}'", man, count=1)
man = re.sub(r"(?m)^  versionName:\s*['\"][^'\"]+['\"]", f"  versionName: '{new}'", man, count=1)
man_path.write_text(man, encoding="utf-8")
print(new)
PY
); then
    printf '%s\n' "$out"
    die "version bump failed: $dir"
  fi
  printf '%s\n' "$out" >"$REPORT_DIR/raised-version"
  if [[ -n "$explicit" ]]; then
    say "$dir version aligned to $out"
  else
    say "$dir version plus 0.0.1, now $out"
  fi
}

raise_chart_version() {
  apply_chart_version "$1"
}

copy_chart() {
  local src=$1 dest=$2
  rm -rf "$dest"
  mkdir -p "$(dirname "$dest")"
  cp -a "$src/." "$dest/"
}

# Pressure charts whose required* exceeds schedulable node capacity are rejected at submit time.
# Read node CPU, memory, and ephemeral disk, then lower any excess below schedulable capacity while staying above 90% of the total,
# so install is accepted and the later cluster-headroom check still fails.
fit_install_resources() {
  local dir=${1%/} base out
  base=$(basename "$dir")
  case "$base" in
    stopcpu|stopmem|stopdisk|stopk8scpu) ;;
    *) return 0 ;;
  esac
  mkdir -p "$REPORT_DIR"
  kubectl get nodes -o json >"$REPORT_DIR/nodes.json" 2>>"$LOG" || die "failed to read node capacity"
  if ! out=$(python3 - "$dir" "$base" "$REPORT_DIR/nodes.json" 2>&1 <<'PY'
import json, re, sys
from pathlib import Path

root = Path(sys.argv[1])
chart = sys.argv[2]
nodes = json.loads(Path(sys.argv[3]).read_text(encoding="utf-8"))

def parse_qty(text, kind):
    text = text.strip().strip("'\"")
    match = re.fullmatch(r"([0-9]+(?:\.[0-9]+)?)([A-Za-z]*)", text)
    if not match:
        raise SystemExit(f"cannot parse quantity {text}")
    number, suffix = float(match.group(1)), match.group(2)
    if kind == "cpu":
        if suffix == "m":
            return int(number)
        if suffix == "":
            return int(number * 1000)
        raise SystemExit(f"cannot parse CPU {text}")
    scale = {"": 1, "Ki": 1024, "Mi": 1024**2, "Gi": 1024**3, "Ti": 1024**4,
             "K": 1000, "M": 1000**2, "G": 1000**3, "T": 1000**4}
    if suffix not in scale:
        raise SystemExit(f"cannot parse capacity {text}")
    return int(number * scale[suffix])

def sum_resource(field, key, kind):
    total = 0
    for item in nodes.get("items", []):
        raw = (((item.get("status") or {}).get(field) or {}).get(key))
        if raw:
            total += parse_qty(raw, kind)
    return total

def target_between(capacity, allocatable):
    if capacity <= 0 and allocatable <= 0:
        return 0
    ceiling = min(v for v in (capacity, allocatable) if v > 0)
    floor = int(max(capacity, allocatable) * 9 / 10)
    if ceiling <= floor + 1:
        return max(ceiling - 1, 1)
    return (floor + ceiling) // 2

def format_cpu(milli):
    return f"{milli}m"

def format_bytes(value):
    amount = max(value // (1024 ** 2), 1)
    return f"{amount}Mi"

cpu_cap = sum_resource("capacity", "cpu", "cpu")
cpu_alloc = sum_resource("allocatable", "cpu", "cpu")
mem_cap = sum_resource("capacity", "memory", "mem")
mem_alloc = sum_resource("allocatable", "memory", "mem")
disk_cap = sum_resource("capacity", "ephemeral-storage", "mem")
disk_alloc = sum_resource("allocatable", "ephemeral-storage", "mem")

want = {
    "stopcpu": ("cpu",),
    "stopk8scpu": ("cpu",),
    "stopmem": ("memory",),
    "stopdisk": ("disk",),
}[chart]
limits = {
    "cpu": ("requiredCpu", "limitedCpu", target_between(cpu_cap, cpu_alloc), "cpu"),
    "memory": ("requiredMemory", "limitedMemory", target_between(mem_cap, mem_alloc), "memory"),
    "disk": ("requiredDisk", "limitedDisk", target_between(disk_cap, disk_alloc), "disk"),
}
path = root / "OlaresManifest.yaml"
text = path.read_text(encoding="utf-8")
changed = []
for kind in want:
    required_key, limited_key, target, qty_kind = limits[kind]
    if target <= 0:
        raise SystemExit(f"node has no {kind} capacity")
    match = re.search(rf"(?m)^  {required_key}:\s*(\S+)", text)
    if not match:
        raise SystemExit(f"{path} has no {required_key}")
    current = parse_qty(match.group(1), qty_kind)
    shown_now = match.group(1).strip("'\"")
    if current <= target:
        print(f"{chart} {required_key}={shown_now} is within the install limit, left unchanged")
        continue
    shown = format_cpu(target) if kind == "cpu" else format_bytes(target)
    text = re.sub(rf"(?m)^  {required_key}:\s*\S+", f"  {required_key}: {shown}", text, count=1)
    if re.search(rf"(?m)^  {limited_key}:\s*\S+", text):
        text = re.sub(rf"(?m)^  {limited_key}:\s*\S+", f"  {limited_key}: {shown}", text, count=1)
    changed.append(f"{required_key} {shown_now} -> {shown}")
path.write_text(text, encoding="utf-8")
if changed:
    print(chart + " rewritten to node capacity: " + ", ".join(changed))
PY
); then
    printf '%s\n' "$out"
    die "rewrite to node capacity failed: $dir"
  fi
  [[ -n "$out" ]] && say "$out"
}

version_already_exists() {
  local out=$1
  [[ "$out" == *"HTTP 409"* || "$out" == *"already exists"* || "$out" == *"does not supersede"* || "$out" == *"higher than"* ]]
}

published_floor() {
  python3 - "$1" <<'PY'
import re, sys
text = sys.argv[1]
match = re.search(r"higher than ([0-9]+(?:\.[0-9A-Za-z]+)*)", text)
if not match:
    match = re.search(r"does not supersede version ([0-9]+(?:\.[0-9A-Za-z]+)*)", text)
print(match.group(1) if match else "")
PY
}

# Bump 0.0.1, then upload. If the version already exists, bump 0.0.1 again.
# When the catalog requires a higher version than the published one, align the chart to that version; the next loop adds 0.0.1.
# Any other upload error prints the CLI output and exits.
package_and_upload() {
  local dir=${1%/} pkg=$2
  local attempt ver name tgz out floor
  mkdir -p "$pkg"
  fit_install_resources "$dir"
  for attempt in $(seq 1 30); do
    raise_chart_version "$dir"
    ver=$(cat "$REPORT_DIR/raised-version")
    read -r name _ <<<"$(chart_meta "$dir/Chart.yaml")"
    say "package $name $ver"
    cli_must "package $dir failed" olares-cli chart package "$dir" -o "$pkg" --force
    tgz="$pkg/${name}-${ver}.tgz"
    [[ -f "$tgz" ]] || die "package did not produce $tgz"
    say "+ olares-cli market upload $tgz -o json"
    if out=$(olares-cli market upload "$tgz" -o json 2>&1); then
      printf '%s\n' "$out" >>"$LOG"
      note_uploaded_version "$name" "$ver"
      say "uploaded $name $ver"
      return 0
    fi
    printf '%s\n' "$out" | tee -a "$LOG"
    if version_already_exists "$out"; then
      floor=$(published_floor "$out")
      if [[ -n "$floor" ]]; then
        say "$name $ver does not supersede published $floor, next version is $floor plus 0.0.1"
        apply_chart_version "$dir" "$floor"
      else
        say "$name $ver already exists, bump 0.0.1 and upload again"
      fi
      continue
    fi
    die "upload $tgz failed"
  done
  die "upload $dir failed to supersede the catalog version after 30 attempts"
}

publish_charts() {
  [[ "$SKIP_UPLOAD" == "1" ]] && { say "skip upload"; return 0; }
  local pkg="$REPORT_DIR/charts" dir base
  rm -rf "$pkg"
  mkdir -p "$pkg"
  for dir in "$CHART_DIR"/*/; do
    [[ -f "$dir/Chart.yaml" ]] || continue
    base=$(basename "$dir")
    if [[ "$base" == "fault-good-missimg" || "$base" == "fault-good-upgrade" ]]; then
      say "skip upload of $base; the upgrade scenario uploads it later"
      continue
    fi
    package_and_upload "$dir" "$pkg"
  done
}

upload_one() {
  package_and_upload "$1" "$REPORT_DIR/charts-extra"
}

require_olares_login() {
  local out id role
  say "check olares-cli login"
  if ! out=$(olares-cli profile whoami --refresh -o json 2>&1); then
    printf '%s\n' "$out" | tee -a "$LOG"
    die "olares-cli is not logged in or the session expired. Run: olares-cli profile login --olares-id <account>"
  fi
  printf '%s\n' "$out" >>"$LOG"
  id=$(printf '%s\n' "$out" | jq -r 'select(type=="object") | .olaresId // empty' | head -1)
  role=$(printf '%s\n' "$out" | jq -r 'select(type=="object") | .role // empty' | head -1)
  if [[ -z "$id" ]]; then
    printf '%s\n' "$out"
    die "olares-cli did not return a login account. Run: olares-cli profile login --olares-id <account>"
  fi
  say "logged in: $id role=${role:-unknown}"
}

preflight() {
  need kubectl
  need jq
  need python3
  need olares-cli
  require_olares_login
  date -u -d '1 minute ago' +%Y-%m-%dT%H:%M:%SZ >/dev/null 2>&1 || die "GNU date is required (run this on the Olares node)"
  kubectl get --raw=/readyz >>"$LOG" 2>&1 || die "kubectl cannot reach the cluster"
  kubectl -n os-framework get sts app-service >>"$LOG" 2>&1 || die "os-framework/app-service was not found"
  IMAGE=$(kubectl -n os-framework get sts app-service -o jsonpath='{.spec.template.spec.containers[0].image}')
  if [[ -z "$OWNER" ]]; then
    OWNER=$(kubectl get applicationmanager -o jsonpath='{range .items[*]}{.spec.appOwner}{"\n"}{end}' 2>/dev/null | awk 'NF' | sort | uniq -c | sort -nr | awk 'NR==1{print $2}')
  fi
  OWNER=${OWNER:-olaresid}
  say "image: $IMAGE"
  say "owner: $OWNER"
  say "report: $REPORT_DIR/report.html"
  cat >"$REPORT_DIR/meta.env" <<EOF
time=$(date -u +%Y-%m-%dT%H:%M:%SZ)
image=$IMAGE
owner=$OWNER
host=$(hostname)
EOF
}

reset_apps() {
  [[ "$SKIP_RESET" == "1" ]] && { say "skip reset"; return 0; }
  local app
  for app in $APPS; do
    clear_app "$app"
  done
  say "wait for test namespaces to disappear"
  local i left app
  for i in 1 2 3 4 5 6 7 8 9 10; do
    left=0
    for app in $APPS; do
      kubectl get ns "$app-$OWNER" >/dev/null 2>&1 && left=1
    done
    [[ "$left" == "0" ]] && return 0
    sleep 3
  done
}

scenario_faultpull() {
  enabled faultpull || return 0
  set_case "T33 downloadFailed"
  set_method "natural"
  say "=== faultpull missing image ==="
  watch_install faultpull 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" 'downloadFailed|installFailed'
  if [[ "$(current_state faultpull 2>/dev/null || true)" == "downloadFailed" ]]; then
    set_case "T65 reinstall from downloadFailed and try to cancel the download"
    arm_hook faultpull downloading cancel
    rm -f "$REPORT_DIR/am-faultpull"
    watch_install faultpull 0.1.0 "$COMPUTE_MODE" "$WAIT_FAST" 'downloadingCanceled|downloadFailed|pendingCanceled|installing'
  fi
}

scenario_faultinstall() {
  enabled faultinstall || return 0
  set_case "T34 helm install failed"
  set_method "natural"
  say "=== faultinstall ==="
  watch_install faultinstall 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" 'installFailed|uninstalled|installingCanceled'
}

scenario_faultunready() {
  enabled faultunready || return 0
  set_case "T35 faultunready stays Pending and enters stopping, reason Unschedulable"
  set_method "natural"
  say "=== faultunready ==="
  # The pod mounts a Secret that does not exist, so it stays Pending.
  # WaitForStartUp returns ErrPodPending after 1 minute and the install stops.
  upload_one "$CHART_DIR/fault-unready"
  watch_install faultunready 0.1.0 "$COMPUTE_MODE" "$WAIT_LONG" 'stopped|stopping'
}

scenario_holdinst() {
  enabled holdinst || return 0
  set_case "T44 backdate installing by 1 minute and it stays, then T14 cancel"
  set_method "natural"
  say "=== holdinst ==="
  arm_hook holdinst installing backdate "1 minute ago"
  watch_install holdinst 0.1.0 "$COMPUTE_MODE" "$WAIT_FAST" 'installing|stopping|stopped|installFailed'
  if [[ "$(current_state holdinst 2>/dev/null || true)" == "installing" ]]; then
    say "T44: still installing after a 1 minute backdate, then cancel"
    sleep 10
    local snap
    snap=$(read_snap holdinst 2>/dev/null) || true
    if [[ -n "${snap:-}" ]]; then
      printf '%s\n' "$snap" >"$REPORT_DIR/snap-holdinst.json"
      record_visit holdinst
    fi
    set_case "T14 cancel installing"
    market_op cancel holdinst cancel || true
    poll_app holdinst "$WAIT_FAST" 'installingCanceled|installFailed|stopped|installingCancelFailed'
  fi
}

scenario_slowhelm() {
  enabled slowhelm || return 0
  set_case "T45 installing timed out"
  set_method "natural"
  say "=== faultslowhelm ==="
  arm_hook faultslowhelm installing backdate "31 minutes ago"
  watch_install faultslowhelm 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" 'installFailed|installingCancelFailed|installingCanceled|stopped'
}

scenario_hanginit() {
  enabled hanginit || return 0
  set_case "T46 initializing timed out, expect initializeCancelBySystem"
  set_method "natural"
  say "=== faulthanginit ==="
  arm_hook faulthanginit initializing backdate "2 hours ago"
  watch_install faulthanginit 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" 'stopped|running|initializingCancelFailed'
}

scenario_ginit() {
  enabled ginit || return 0
  set_case "T15 cancel initializing"
  set_method "natural"
  say "=== ginit cancel initializing ==="
  arm_hook ginit initializing cancel
  watch_install ginit 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" 'stopped|running|initializingCancelFailed'
  local st
  st=$(current_state ginit 2>/dev/null || true)
  if [[ "$st" == "running" ]]; then
    set_case "T04 stop"
    market_op stop ginit stop || true
    poll_app ginit "$WAIT_FAST" 'stopped|stopFailed'
  fi
  st=$(current_state ginit 2>/dev/null || true)
  if [[ "$st" == "stopped" ]]; then
    set_case "T40 resume after deleting the Helm release"
    local ns
    ns=$(kubectl get applicationmanager "$(am_of ginit)" -o jsonpath='{.spec.appNamespace}' 2>/dev/null || echo "ginit-$OWNER")
    say "delete the ginit Helm release secret in $ns"
    kubectl -n "$ns" delete secret -l name=ginit,owner=helm --wait=false >>"$LOG" 2>&1 || true
    set_method "natural"
    market_op resume ginit resume || true
    poll_app ginit "$WAIT_FAST" 'resumeFailed|running|stopped'
  fi
}

scenario_faultgood() {
  enabled faultgood || return 0
  set_case "T01 T02 install faultgood to running"
  set_method "natural"
  say "=== faultgood main path ==="
  watch_install faultgood 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" '^running$'
  if [[ "$(current_state faultgood 2>/dev/null || true)" != "running" ]]; then
    say "faultgood did not reach running, skip later stop and upgrade"
    return 0
  fi
  set_case "T04 stop"
  market_op stop faultgood stop || true
  poll_app faultgood "$WAIT_FAST" 'stopped|stopFailed'
  if [[ "$(current_state faultgood 2>/dev/null || true)" != "stopped" ]]; then
    say "faultgood did not stop"
    return 0
  fi
  set_case "T06 resume"
  market_op resume faultgood resume || true
  poll_app faultgood "$WAIT_INSTALL" '^running$|resumeFailed'
  if [[ "$(current_state faultgood 2>/dev/null || true)" != "running" ]]; then
    return 0
  fi
  set_case "T32 stop during resume"
  market_op stop faultgood stop || true
  poll_app faultgood "$WAIT_FAST" 'stopped|stopFailed'
  arm_hook faultgood resuming stop
  market_op resume faultgood resume || true
  poll_app faultgood "$WAIT_FAST" 'stopped|running|stopFailed'
  if [[ "$(current_state faultgood 2>/dev/null || true)" != "running" ]]; then
    set_case "T06 resume again"
    market_op resume faultgood resume || true
    poll_app faultgood "$WAIT_INSTALL" '^running$|resumeFailed|stopped'
  fi
  [[ "$(current_state faultgood 2>/dev/null || true)" == "running" ]] || return 0

  set_case "T38 upgrade image pull failed with fault-good-missimg"
  upload_one "$CHART_DIR/fault-good-missimg"
  upgrade_app faultgood "$(cat "$REPORT_DIR/raised-version")"
  poll_app faultgood "$WAIT_INSTALL" 'upgradeFailed|running|stopped'

  set_case "T37 helm upgrade failed. Copy the failing template and bump 0.0.1 on upload"
  local bumped="$REPORT_DIR/src/fault-good-upgrade"
  copy_chart "$CHART_DIR/fault-good-upgrade" "$bumped"
  upload_one "$bumped"
  upgrade_app faultgood "$(cat "$REPORT_DIR/raised-version")"
  poll_app faultgood "$WAIT_INSTALL" 'upgradeFailed|running|stopped'

  set_case "T73 uninstall from upgradeFailed"
  market_op uninstall faultgood uninstall || true
  poll_app faultgood "$WAIT_FAST" 'uninstalled|uninstallFailed'
  if [[ "$(current_state faultgood 2>/dev/null || true)" == "uninstalled" ]]; then
    set_case "T11 install again from uninstalled"
    rm -f "$REPORT_DIR/am-faultgood"
    # After an upgrade chart is uploaded, the catalog accepts only the latest version. Upload a working fault-good again, then install that version.
    upload_one "$CHART_DIR/fault-good"
    watch_install faultgood "$(cat "$REPORT_DIR/raised-version")" "$COMPUTE_MODE" 90 'pending|downloading|installing|running|downloadFailed'
    market_op uninstall faultgood uninstall || market_op cancel faultgood cancel || true
  fi
}

scenario_upgradetest() {
  enabled upgradetest || return 0
  set_case "T20/T21 successful upgrade. Stop then upgrade one step, resume, then upgrade one more step"
  set_method "natural"
  say "=== upgradetest versions ==="
  watch_install upgradetest 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" '^running$'
  [[ "$(current_state upgradetest 2>/dev/null || true)" == "running" ]] || return 0
  set_case "T04 stop upgradetest before T21"
  market_op stop upgradetest stop || true
  poll_app upgradetest "$WAIT_FAST" 'stopped|stopFailed'
  [[ "$(current_state upgradetest 2>/dev/null || true)" == "stopped" ]] || return 0
  local v2="$REPORT_DIR/src/upgradetest-next" v3="$REPORT_DIR/src/upgradetest-next2"
  local bad="$REPORT_DIR/src/upgradetest-fail"
  copy_chart "$CHART_DIR/upgradetest" "$bad"
  # Upload runs helm render, so a template fail is rejected before the chart
  # is stored. An unknown kind renders, then fails when the upgrade applies it.
  cat >>"$bad/templates/app.yaml" <<'EOF'
---
apiVersion: example.com/v1
kind: NoSuch
metadata:
  name: upgrade-boom
  namespace: {{ .Release.Namespace }}
spec:
  reason: upgrade-boom
EOF
  set_case "S4 upgrade from stopped fails"
  upload_one "$bad"
  upgrade_app upgradetest "$(cat "$REPORT_DIR/raised-version")"
  poll_app upgradetest "$WAIT_INSTALL" '^upgradeFailed$'
  if [[ "$(current_state upgradetest 2>/dev/null || true)" == "upgradeFailed" ]]; then
    set_case "S4 failed upgrade keeps pre-upgrade-state stopped"
    expect_eq upgradetest stopped "$(annotation_of upgradetest bytetrade.io/pre-upgrade-state || true)" || true
    copy_chart "$CHART_DIR/upgradetest" "$v2"
    upload_one "$v2"
    upgrade_app upgradetest "$(cat "$REPORT_DIR/raised-version")"
    sleep 2
    set_case "S4 retry does not replace stopped with upgradeFailed"
    expect_eq upgradetest stopped "$(annotation_of upgradetest bytetrade.io/pre-upgrade-state || true)" || true
    set_case "T21 S4 retry from upgradeFailed stays stopped"
    poll_app upgradetest "$WAIT_INSTALL" '^stopped$|^running$|upgradeFailed'
    expect_eq upgradetest stopped "$(current_state upgradetest 2>/dev/null || true)" || true
    if [[ "$(current_state upgradetest 2>/dev/null || true)" == "stopped" ]]; then
      set_case "S4 stopped upgrade retry keeps replicas at 0"
      st=$(current_state upgradetest 2>/dev/null || true)
      if [[ "$(deploy_replicas upgradetest)" == "0" ]]; then
        expect_eq upgradetest stopped "$st" || true
      else
        record_case_result upgradetest stopped "${st:-empty}" false "deployment was scaled up" || true
      fi
    fi
  fi
  [[ "$(current_state upgradetest 2>/dev/null || true)" == "stopped" ]] || return 0
  set_case "T06 resume to running before T20"
  market_op resume upgradetest resume || true
  poll_app upgradetest "$WAIT_INSTALL" '^running$|resumeFailed'
  [[ "$(current_state upgradetest 2>/dev/null || true)" == "running" ]] || return 0
  copy_chart "$CHART_DIR/upgradetest" "$v3"
  set_case "T18 T20 upgrade again from running, version plus 0.0.1"
  upload_one "$v3"
  upgrade_app upgradetest "$(cat "$REPORT_DIR/raised-version")"
  poll_app upgradetest "$WAIT_INSTALL" '^running$|upgradeFailed'
}

scenario_applyenv() {
  enabled applyenv || return 0
  set_case "install applyenv"
  set_method "natural"
  say "=== applyenv ==="
  # Upload again with the pre-upgrade hook that disables sidecar injection, or install never leaves installing.
  upload_one "$CHART_DIR/applyenv"
  watch_install applyenv 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" '^running$'
  [[ "$(current_state applyenv 2>/dev/null || true)" == "running" ]] || return 0
  set_case "T04 stop"
  market_op stop applyenv stop || true
  poll_app applyenv "$WAIT_FAST" 'stopped|stopFailed'
  if [[ "$(current_state applyenv 2>/dev/null || true)" == "stopped" ]]; then
    # BREAK is already 0 in the chart. Setting it to 0 again does not
    # mark NeedApply, so the state stays stopped and the snapshot is
    # never written. BREAK=2 renders; only BREAK=1 fails the template.
    set_case "S3 applyenv from stopped leaves stopped"
    env_set applyenv "BREAK=2"
    if poll_app applyenv "$WAIT_FAST" "" stopped; then
      set_case "S3 T09 T30 applyenv from stopped stays stopped"
      poll_app applyenv "$WAIT_INSTALL" '^stopped$|^running$|applyEnvFailed'
      expect_eq applyenv stopped "$(current_state applyenv 2>/dev/null || true)" || true
      if [[ "$(current_state applyenv 2>/dev/null || true)" == "stopped" ]]; then
        set_case "S3 applyenv from stopped snapshots pre-upgrade-state"
        expect_eq applyenv stopped "$(annotation_of applyenv bytetrade.io/pre-upgrade-state || true)" || true
        set_case "S3 stopped applyenv keeps replicas at 0"
        st=$(current_state applyenv 2>/dev/null || true)
        if [[ "$(deploy_replicas applyenv)" == "0" ]]; then
          expect_eq applyenv stopped "$st" || true
        else
          record_case_result applyenv stopped "${st:-empty}" false "deployment was scaled up" || true
        fi
      fi
    fi
  fi
  if [[ "$(current_state applyenv 2>/dev/null || true)" == "stopped" ]]; then
    set_case "T06 resume"
    market_op resume applyenv resume || true
    poll_app applyenv "$WAIT_INSTALL" '^running$|resumeFailed'
  fi
  if [[ "$(current_state applyenv 2>/dev/null || true)" == "running" ]]; then
    set_case "S3 applyenv from running leaves running"
    env_set applyenv "BREAK=0"
    if poll_app applyenv "$WAIT_FAST" "" running; then
      set_case "S3 applyenv from running replaces leftover stopped"
      poll_app applyenv "$WAIT_INSTALL" '^running$|^stopped$|applyEnvFailed'
      expect_eq applyenv running "$(current_state applyenv 2>/dev/null || true)" || true
      if [[ "$(current_state applyenv 2>/dev/null || true)" == "running" ]]; then
        set_case "S3 applyenv from running snapshots pre-upgrade-state"
        expect_eq applyenv running "$(annotation_of applyenv bytetrade.io/pre-upgrade-state || true)" || true
      fi
    fi
  fi
  if [[ "$(current_state applyenv 2>/dev/null || true)" == "running" ]]; then
    set_case "T10 T39 set BREAK=1 while running, expect applyEnvFailed"
    env_set applyenv "BREAK=1"
    poll_app applyenv "$WAIT_FAST" 'applyEnvFailed|running|stopped'
  fi
  if [[ "$(current_state applyenv 2>/dev/null || true)" == "applyEnvFailed" ]]; then
    set_case "T75 applyEnvFailed stop"
    market_op stop applyenv stop || true
    poll_app applyenv "$WAIT_FAST" 'stopped|stopping|stopFailed'
    set_case "T76 applyEnvFailed uninstall"
    market_op uninstall applyenv uninstall || true
    poll_app applyenv "$WAIT_FAST" 'uninstalled|uninstallFailed'
  fi
}

scenario_simple() {
  local name=$1 app=$2 note=$3 until=$4
  enabled "$name" || return 0
  set_case "$note"
  set_method "natural"
  say "=== $app $note ==="
  watch_install "$app" 0.1.0 "$COMPUTE_MODE" "$WAIT_LONG" "$until"
}

scenario_pressure() {
  scenario_simple pressure stopmem "stopmem expects SystemMemoryPressure" 'stopped|running|installFailed|downloadFailed'
  scenario_simple pressure stopcpu "stopcpu expects SystemCPUPressure" 'stopped|running|installFailed|downloadFailed'
  scenario_simple pressure stopdisk "stopdisk expects DiskPressure" 'stopped|running|installFailed|downloadFailed'
  scenario_simple pressure stopk8scpu "stopk8scpu expects K8sRequestCPUPressure; SystemCPUPressure if cluster headroom stops it first" 'stopped|running|installFailed|downloadFailed'
}

scenario_gpu() {
  enabled gpu || return 0
  set_case "faultgpu expects HamiUnschedulable; record Unschedulable when the event is FailedScheduling"
  set_method "natural"
  say "=== faultgpu ==="
  # Publish the chart whose pod consume policy is single, so HAMi records InsufficientGPU.
  upload_one "$CHART_DIR/fault-gpu"
  watch_install faultgpu 0.1.0 nvidia "$WAIT_LONG" 'stopped|running|installFailed|downloadFailed'
}

scenario_stopinit() {
  scenario_simple stopinit stopinit "stopinit expects InitFailed. The container must keep failing for more than 5 minutes after exit" 'stopped|running|installFailed'
}

scenario_gcont() {
  scenario_simple gcont gcont "gcont pending unschedulable returns immediately to stopping, reason Unschedulable" 'stopped|stopping'
}

scenario_middleware() {
  scenario_simple middleware faultmw "T03 middleware, expect installing to go straight to running" 'running|installFailed|stopped|downloadFailed'
}

scenario_deletehook() {
  enabled deletehook || return 0
  set_case "install faultdelhook to running then uninstall, expect T42 uninstallFailed"
  set_method "natural"
  say "=== faultdelhook ==="
  watch_install faultdelhook 0.1.0 "$COMPUTE_MODE" "$WAIT_INSTALL" '^running$|installFailed|stopped'
  if [[ "$(current_state faultdelhook 2>/dev/null || true)" == "running" || "$(current_state faultdelhook 2>/dev/null || true)" == "stopped" ]]; then
    market_op uninstall faultdelhook uninstall || true
    # The uninstall timeout is 5 minutes. If the hook pod never starts, wait past those 5 minutes to reach uninstallFailed.
    poll_app faultdelhook 420 'uninstallFailed|uninstalled'
  fi
}

scenario_cancel_pending() {
  enabled cancel-pending || return 0
  set_case "faultpull holds the download slot, install gcont and cancel it immediately, expect pendingCanceled"
  set_method "natural"
  say "=== install gcont while the download slot is held, then cancel it immediately, expect pendingCanceled ==="
  local st i holder_pid=""
  rm -f "$REPORT_DIR/am-gcont" "$REPORT_DIR/hook-gcont.json"

  # gcont is usually still stopped from the previous scenario and cannot be installed again. Uninstall it back to uninstalled first.
  st=$(current_state gcont 2>/dev/null || true)
  case "$st" in
    ""|uninstalled|downloadFailed|downloadingCanceled|installingCanceled|pendingCanceled) ;;
    *)
      market_op uninstall gcont uninstall || true
      poll_app gcont 180 'uninstalled|uninstallFailed' || true
      st=$(current_state gcont 2>/dev/null || true)
      ;;
  esac
  case "$st" in
    ""|uninstalled|downloadFailed|downloadingCanceled|installingCanceled|pendingCanceled) ;;
    *)
      say "gcont is ${st:-absent} and cannot be reinstalled, skip pendingCanceled"
      return 0
      ;;
  esac

  # faultpull's image does not exist, so the pull stays in downloading and the download slot stays busy.
  st=$(current_state faultpull 2>/dev/null || true)
  case "$st" in
    downloading) ;;
    ""|uninstalled|downloadFailed|downloadingCanceled|installingCanceled|pendingCanceled)
      poll_app faultpull 120 'downloading|downloadFailed|installing' &
      holder_pid=$!
      install_app faultpull 0.1.0 "$COMPUTE_MODE"
      ;;
    pending) ;;
    *)
      say "faultpull is $st and cannot hold the download slot"
      return 0
      ;;
  esac
  st=""
  for ((i = 0; i < 50; i++)); do
    st=$(current_state faultpull 2>/dev/null || true)
    [[ "$st" == "downloading" ]] && break
    sleep 0.2
  done
  if [[ "$st" != "downloading" ]]; then
    say "faultpull did not stay in downloading (now ${st:-absent}), gcont will not stay in pending"
    [[ -n "$holder_pid" ]] && wait "$holder_pid" || true
    return 0
  fi

  set_case "faultpull holds the download slot; install gcont and cancel it immediately"
  install_app gcont 0.1.0 "$COMPUTE_MODE"
  market_op cancel gcont cancel || true
  poll_app gcont "$WAIT_FAST" 'pendingCanceled|pendingCancelFailed|downloading|installing|running|stopped' || true
  [[ -n "$holder_pid" ]] && wait "$holder_pid" || true
  market_op cancel faultpull cancel || market_op uninstall faultpull uninstall || true
}

write_fixture() {
  local app=$1 state=$2 ago=$3 payload=$4 am stamp
  am=$(am_of "$app") || return 1
  stamp=$(ago_stamp "$ago") || return 1
  say "fixture $app -> $state statusTime=$stamp"
  set_method "fixture"
  kubectl get applicationmanager "$am" -o json | jq \
    --arg st "$state" --arg ts "$stamp" --argjson payload "$payload" \
    '.status.state=$st | .status.statusTime=$ts | .status.reason="fixture" | .status.message="lifecycle fixture" | .status.payload=$payload' \
    | kubectl replace -f - >>"$LOG" 2>&1
}

scenario_freeze() {
  enabled freeze || return 0
  [[ "$SKIP_FREEZE" == "1" ]] && { say "skip fixture pass"; return 0; }
  local app="" candidate
  for candidate in faultpull faultinstall holdinst faultgood ginit applyenv; do
    if am_of "$candidate" >/dev/null 2>&1; then
      app=$candidate
      break
    fi
  done
  if [[ -z "$app" ]]; then
    say "no ApplicationManager to write a fixture, skip fixture pass"
    return 0
  fi
  say "fixture pass uses $app"
  fixture_pending_cancel "$app"
  local row state ago payload need
  while IFS='|' read -r state ago payload need; do
    [[ -z "${state:-}" ]] && continue
    if [[ -n "${need:-}" ]]; then
      if seen_state "$need"; then
        say "already entered $need, skip fixture"
        continue
      fi
    elif seen_state "$state"; then
      say "already entered $state, skip fixture"
      continue
    fi
    set_case "fixture $state, backdate statusTime by $ago so the runner takes the timeout edge"
    rm -f "$REPORT_DIR/hook-$app.json"
    if ! freeze; then
      say "freeze failed"
      thaw
      continue
    fi
    if ! write_fixture "$app" "$state" "$ago" "$payload"; then
      say "fixture write failed $state"
      thaw
      continue
    fi
    local snap
    snap=$(read_snap "$app" 2>/dev/null) || true
    if [[ -n "${snap:-}" ]]; then
      printf '%s\n' "$snap" >"$REPORT_DIR/snap-$app.json"
      record_visit "$app"
    fi
    thaw
    set_method "fixture"
    poll_app "$app" 180 "" "$state" || true
  done <<'EOF'
pendingCanceling|11 minutes ago|{}|pendingCancelFailed
downloadingCanceling|11 minutes ago|{}
installingCanceling|11 minutes ago|{"installing-canceling.phase":"wait-ns","installing-canceling.terminal":"installFailed"}
initializingCanceling|11 minutes ago|{}
resumingCanceling|11 minutes ago|{}
applyingEnvCanceling|11 minutes ago|{}
upgradePendingCanceling|11 minutes ago|{}
upgradeDownloadingCanceling|11 minutes ago|{}|
upgradingCanceling|11 minutes ago|{}|
stopping|31 minutes ago|{}|stopFailed
uninstalling|31 minutes ago|{}|uninstallFailed
EOF
}

# When the download slot is free, pending lasts only a few milliseconds and a 0.5s poll misses the cancel hook.
# Freeze, write pending with cancel, then thaw so the first reconcile takes the cancel path.
fixture_pending_cancel() {
  local app=$1 am id snap
  seen_state pendingCanceled && return 0
  am=$(am_of "$app") || return 0
  set_case "fixture cancel from pending, expect pendingCanceled"
  set_method "fixture"
  say "fixture $app written as pending, then cancel"
  if ! freeze; then
    thaw
    return 0
  fi
  id="lifecycle-pending-cancel-$RANDOM"
  if ! snap=$(kubectl get applicationmanager "$am" -o json | jq \
    --arg id "$id" --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '.status.state="pending"
     | .status.statusTime=$ts
     | .status.reason="fixture"
     | .status.message="pending cancel fixture"
     | .status.opID="fixture-old"
     | .status.payload={}
     | .spec.opType="cancel"
     | .metadata.annotations = ((.metadata.annotations // {}) + {"bytetrade.io/operation-request-id":$id})'); then
    thaw
    return 0
  fi
  printf '%s\n' "$snap" | kubectl replace -f - >>"$LOG" 2>&1 || { thaw; return 0; }
  printf '%s\n' "$snap" >"$REPORT_DIR/snap-$app.json"
  record_visit "$app" || true
  thaw
  set_method "fixture"
  poll_app "$app" 180 'pendingCanceled|pendingCancelFailed|downloading|pendingCanceling' || true
}

harvest_state_changes() {
  local fresh merged
  fresh=$(mktemp)
  merged=$(mktemp)
  # Append. A later harvest runs after freeze deletes the Pod, and
  # overwriting here would keep only the new Pod's lines.
  kubectl -n os-framework logs sts/app-service --since=8h 2>/dev/null \
    | grep 'state change ' >"$fresh" || true
  if [[ -f "$REPORT_DIR/state-changes.txt" ]]; then
    awk '!seen[$0]++' "$REPORT_DIR/state-changes.txt" "$fresh" >"$merged"
  else
    awk '!seen[$0]++' "$fresh" >"$merged"
  fi
  mv "$merged" "$REPORT_DIR/state-changes.txt"
  rm -f "$fresh"
}

rebuild_report() {
  harvest_state_changes || true
  python3 - "$REPORT_DIR" "$STATES" <<'PY'
import html, json, pathlib, re, sys
from datetime import datetime

root = pathlib.Path(sys.argv[1])
states = sys.argv[2].split()

def esc(value):
    return html.escape("" if value is None else str(value), quote=True)

def load_jsonl(path):
    rows = []
    if not path.exists():
        return rows
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            rows.append(json.loads(line))
        except json.JSONDecodeError:
            continue
    return rows

meta = {}
meta_path = root / "meta.env"
if meta_path.exists():
    for line in meta_path.read_text(encoding="utf-8").splitlines():
        if "=" in line:
            k, v = line.split("=", 1)
            meta[k] = v

plans = {}
for row in load_jsonl(root / "state-plans.jsonl"):
    st = row.get("state")
    if st:
        plans[st] = row

def parse_stamp(name):
    m = re.search(r"(\d{4})(\d{2})(\d{2})T(\d{2})(\d{2})(\d{2})Z", name)
    if not m:
        return None
    return datetime(*(int(x) for x in m.groups()))

def short_app(am, owner):
    if owner:
        token = f"-{owner}-"
        if token in am:
            left, right = am.split(token, 1)
            if left == right:
                return left
    parts = am.split("-")
    if len(parts) >= 3 and parts[0] == parts[-1]:
        return parts[0]
    return am

def with_next(rows, same_key):
    out = []
    for i, row in enumerate(rows):
        nxt = ""
        for later in rows[i + 1:]:
            if later.get(same_key) == row.get(same_key) and later.get("state") != row.get("state"):
                nxt = later.get("state") or ""
                break
        item = dict(row)
        item["next"] = nxt
        out.append(item)
    return out

CHANGE_RE = re.compile(
    r"I(\d{2})(\d{2}) (\d{2}):(\d{2}):(\d{2})\.\d+.*state change (\S+)\s+(\S*)\s*->\s+(\S+)\s+reason=(\S+)"
)

def parse_log_text(text, owner, stamp):
    rows = []
    year = stamp.year if stamp else datetime.utcnow().year
    for line in (text or "").splitlines():
        m = CHANGE_RE.search(line)
        if not m:
            continue
        month, day, hh, mm, ss, am, src, dst, reason = m.groups()
        try:
            ts = datetime(year, int(month), int(day), int(hh), int(mm), int(ss))
        except ValueError:
            continue
        if stamp and ts < stamp:
            continue
        rows.append({
            "ts": ts.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "app": short_app(am, owner),
            "am": am,
            "state": dst,
            "reason": reason,
            "from": src,
            "method": "log",
            "case": "controller state change",
            "message": "",
            "phase": "",
        })
    return rows

def load_log_rows(path, owner, stamp):
    if not path.exists():
        return []
    return parse_log_text(path.read_text(encoding="utf-8", errors="replace"), owner, stamp)

trace = load_jsonl(root / "trace.jsonl")
by = {s: [] for s in states}
for item in with_next(trace, "app"):
    st = item.get("state")
    if st in by:
        by[st].append(item)

seen_poll = {
    (item.get("app"), st)
    for st, rows in by.items()
    for item in rows
}
stamp = parse_stamp(root.name)
log_rows = load_log_rows(root / "state-changes.txt", meta.get("owner", ""), stamp)
transitions = {}
for row in log_rows:
    key = (row.get("ts"), row.get("am") or row.get("app"), row.get("from") or "", row.get("state"), row.get("reason") or "")
    transitions.setdefault(key, row)
for path in sorted((root / "states").glob("*.jsonl")):
    for visit in load_jsonl(path):
        for row in parse_log_text(visit.get("logs") or "", meta.get("owner", ""), stamp):
            key = (row.get("ts"), row.get("am") or row.get("app"), row.get("from") or "", row.get("state"), row.get("reason") or "")
            transitions.setdefault(key, row)

def remember(row, st):
    app = row.get("app") or ""
    if st not in by or (app, st) in seen_poll:
        return
    item = dict(row)
    item["state"] = st
    by[st].append(item)

for row in transitions.values():
    remember(row, row.get("state") or "")
filled = {st for st, rows in by.items() if rows}
for row in transitions.values():
    src = row.get("from") or ""
    if src in filled or src not in by:
        continue
    remember(row, src)
for st in by:
    by[st].sort(key=lambda row: row.get("ts") or "")

missing = [st for st in states if not by.get(st)]
seen = len(states) - len(missing)

def leave_set(state):
    text = (plans.get(state) or {}).get("leave") or ""
    return {part.strip() for part in re.split(r"[;\uff1b]", text) if part.strip() and part.strip() not in ("stay", "\u4fdd\u6301")}

# Method names stored by older runs, mapped to the English labels.
METHOD_EN = {
    "\u81ea\u7136": "natural",
    "\u65e5\u5fd7": "log",
    "\u62e8\u56de": "backdate",
    "\u5939\u5177": "fixture",
}

def method_en(name):
    return METHOD_EN.get(name or "", name or "")

def verdict(src, dst):
    if not src:
        return "expected" if dst == "pending" else "unexpected"
    if dst in leave_set(src):
        return "expected"
    return "unexpected"

def poll_method(app, dst, ts):
    try:
        when = datetime.strptime(ts, "%Y-%m-%dT%H:%M:%SZ")
    except (TypeError, ValueError):
        return "log"
    best = ""
    best_dt = None
    for row in trace:
        if row.get("app") != app or row.get("state") != dst:
            continue
        try:
            polled = datetime.strptime(row.get("ts") or "", "%Y-%m-%dT%H:%M:%SZ")
        except ValueError:
            continue
        delta = abs((polled - when).total_seconds())
        if delta > 180:
            continue
        if best_dt is None or delta < best_dt:
            best_dt = delta
            best = row.get("method") or ""
    return best or "log"

order_index = {name: i for i, name in enumerate(states)}
grouped = {}
for row in transitions.values():
    src = row.get("from") or ""
    dst = row.get("state") or ""
    reason = row.get("reason") or ""
    method = poll_method(row.get("app") or "", dst, row.get("ts") or "")
    key = (src, reason, dst)
    item = grouped.get(key)
    if item is None:
        item = {"from": src, "reason": reason, "to": dst, "apps": [], "methods": []}
        grouped[key] = item
    app = row.get("app") or ""
    if app and app not in item["apps"]:
        item["apps"].append(app)
    if method and method not in item["methods"]:
        item["methods"].append(method)

def edge_sort(item):
    src = item["from"]
    return (
        order_index.get(src, -1) if src else -1,
        order_index.get(item["to"], 10**6),
        item["reason"],
    )

edges = sorted(grouped.values(), key=edge_sort)
ok_n = 0
bad_n = 0
rows_html = []
for item in edges:
    result = verdict(item["from"], item["to"])
    if result == "expected":
        ok_n += 1
        css = "ok"
    else:
        bad_n += 1
        css = "miss"
    src = item["from"] or "none"
    src_cell = f"<code>{esc(src)}</code>"
    if item["from"] in order_index:
        src_cell = f"<a href=\"#state-{esc(item['from'])}\">{src_cell}</a>"
    dst_cell = f"<code>{esc(item['to'])}</code>"
    if item["to"] in order_index:
        dst_cell = f"<a href=\"#state-{esc(item['to'])}\">{dst_cell}</a>"
    methods = []
    for raw in item["methods"]:
        label = method_en(raw)
        if label and label != "log" and label not in methods:
            methods.append(label)
    if not methods:
        methods = ["log"]
    rows_html.append(
        f"<tr class=\"{css}\">"
        f"<td>{src_cell}</td>"
        f"<td>{esc(', '.join(item['apps']))}</td>"
        f"<td>{esc(item['reason'])}</td>"
        f"<td>{dst_cell}</td>"
        f"<td>{esc(', '.join(methods))}</td>"
        f"<td>{esc(result)}</td></tr>"
    )

case_rows = load_jsonl(root / "case-results.jsonl")
case_rows.sort(key=lambda row: row.get("ts") or "")
case_ok = sum(1 for row in case_rows if row.get("passed"))
case_bad = len(case_rows) - case_ok
case_html = []
for row in case_rows:
    passed = bool(row.get("passed"))
    result = "expected" if passed else "unexpected"
    css = "ok" if passed else "miss"
    case_html.append(
        f"<tr class=\"{css}\">"
        f"<td>{esc(row.get('case') or '')}</td>"
        f"<td>{esc(row.get('app') or '')}</td>"
        f"<td><code>{esc(row.get('wanted') or '')}</code></td>"
        f"<td><code>{esc(row.get('actual') or '')}</code></td>"
        f"<td>{esc(row.get('reason') or '')}</td>"
        f"<td>{esc(result)}</td></tr>"
    )
if not case_html:
    case_html.append("<tr><td colspan=\"6\">No case results.</td></tr>")

sections = []
for st in states:
    plan = plans.get(st) or {}
    visits = by.get(st) or []
    blocks = [
        f"<section id=\"state-{esc(st)}\">",
        f"<h2><code>{esc(st)}</code></h2>",
        "<ul>",
        f"<li>Planned entry: {esc(plan.get('enter') or '')}</li>",
        f"<li>Planned exit: {esc(plan.get('leave') or '')}</li>",
        "</ul>",
        "<h3>Observations</h3>",
    ]
    if not visits:
        blocks.append("<p class=\"miss\">Not entered.</p>")
    for i, visit in enumerate(visits, 1):
        blocks.append("<article class=\"visit\">")
        blocks.append(
            f"<h3>Visit {i} {esc(visit.get('ts'))} "
            f"<code>{esc(visit.get('app'))}</code> {esc(method_en(visit.get('method')))}</h3>"
        )
        blocks.append("<dl>")
        for label, key in (
            ("Case", "case"),
            ("ApplicationManager", "am"),
            ("Namespace", "ns"),
            ("reason", "reason"),
        ):
            val = visit.get(key)
            shown = val if val not in (None, "") else ""
            blocks.append(f"<dt>{esc(label)}</dt><dd>{esc(shown)}</dd>")
        if visit.get("from"):
            blocks.append(f"<dt>From</dt><dd><code>{esc(visit.get('from'))}</code></dd>")
        blocks.append("</dl>")
        blocks.append("</article>")
    blocks.append("<p><a href=\"#summary\">Back to summary</a></p>")
    blocks.append("</section>")
    sections.append("\n".join(blocks))

missing_text = "none" if not missing else ", ".join(f"<code>{esc(s)}</code>" for s in missing)
page = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Market lifecycle test report</title>
<style>
  body {{ margin: 0; color: #1c1c1c; background: #f7f7f5; font: 15px/1.5 -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }}
  main {{ max-width: 1100px; margin: 0 auto; padding: 28px 20px 64px; }}
  h1 {{ font-size: 28px; margin: 0 0 12px; }}
  h2 {{ font-size: 22px; margin: 36px 0 12px; }}
  h3 {{ font-size: 16px; margin: 18px 0 8px; }}
  h4 {{ font-size: 13px; margin: 14px 0 6px; color: #444; }}
  a {{ color: #0b57d0; }}
  code {{ font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 0.92em; }}
  .meta, .note {{ color: #333; }}
  .note {{ background: #fff; border: 1px solid #e4e4e0; border-radius: 8px; padding: 12px 14px; }}
  table {{ width: 100%; border-collapse: collapse; background: #fff; }}
  th, td {{ border: 1px solid #e4e4e0; padding: 7px 8px; text-align: left; vertical-align: top; }}
  th {{ background: #f0f0ec; position: sticky; top: 0; }}
  tr.ok td:last-child {{ color: #176b2a; font-weight: 600; }}
  tr.miss td:last-child, p.miss {{ color: #9a3412; font-weight: 600; }}
  section {{ background: #fff; border: 1px solid #e4e4e0; border-radius: 8px; padding: 8px 18px 16px; margin-top: 16px; }}
  article.visit {{ border-top: 1px solid #eee; padding-top: 8px; }}
  dl {{ display: grid; grid-template-columns: 160px 1fr; gap: 4px 12px; margin: 8px 0; }}
  dt {{ color: #555; }}
  dd {{ margin: 0; }}
  pre {{ margin: 0; padding: 12px; overflow: auto; background: #f6f8fa; border-radius: 6px; white-space: pre-wrap; word-break: break-word; }}
</style>
</head>
<body>
<main>
<h1>Market lifecycle test report</h1>
<ul class="meta">
<li>Time: {esc(meta.get("time", ""))}</li>
<li>Host: {esc(meta.get("host", ""))}</li>
<li>owner: {esc(meta.get("owner", ""))}</li>
<li>app-service image: <code>{esc(meta.get("image", ""))}</code></li>
<li>Trace: <code>{esc(root / "trace.jsonl")}</code></li>
</ul>
<h2>Methods</h2>
<div class="note">
<p>Each row is one transition recorded by the controller. The same source state with different destinations is split into separate rows. The result is <strong>expected</strong> when the destination is in that source state's planned exits. An empty source arriving at <code>pending</code> is an install creating the app, and that counts as expected.</p>
<dl>
<dt>natural</dt>
<dd>The script drove a real install, stop, cancel, resume, or env change, and the 0.5s poll recorded the state.</dd>
<dt>backdate</dt>
<dd>Only <code>statusTime</code> is moved into the past (1 minute, 31 minutes, or 2 hours). app-service keeps running, and the next reconcile treats the state as timed out.</dd>
<dt>fixture</dt>
<dd>app-service is scaled to 0, the state and an already-expired <code>statusTime</code> are written onto the ApplicationManager, then app-service is scaled back up. The first reconcile takes that edge.</dd>
<dt>log</dt>
<dd>The controller wrote a <code>state change</code> line, but the 0.5s poll did not record that arrival. The report fills the transition from the log. If the same edge was also polled, the table keeps natural, backdate, or fixture and does not also show log.</dd>
</dl>
<p>A cell that lists more than one method means different apps took the same edge in different ways.</p>
</div>
<h2 id="summary">Summary</h2>
<p>{len(edges)} transitions. expected {ok_n}, unexpected {bad_n}. States sampled {seen} / {len(states)}. Not sampled: {missing_text}</p>
<table>
<thead><tr><th>From</th><th>App</th><th>reason</th><th>To</th><th>Method</th><th>Result</th></tr></thead>
<tbody>
{chr(10).join(rows_html)}
</tbody>
</table>
<h2 id="cases">Test cases</h2>
<p>Each row is one wait in the active test case. The result is <strong>expected</strong> when that wait reached its pattern, and <strong>unexpected</strong> when it timed out. This table does not change the transition results above.</p>
<p>{len(case_rows)} cases. expected {case_ok}, unexpected {case_bad}.</p>
<table>
<thead><tr><th>Case</th><th>App</th><th>Expected</th><th>Actual</th><th>reason</th><th>Result</th></tr></thead>
<tbody>
{chr(10).join(case_html)}
</tbody>
</table>
<h2>States</h2>
{chr(10).join(sections)}
</main>
</body>
</html>
"""
out = root / "report.html"
out.write_text(page, encoding="utf-8")
old = root / "report.md"
if old.exists():
    old.unlink()
print(f"transitions {len(edges)}. expected {ok_n}, unexpected {bad_n}")
print(f"cases {len(case_rows)}. expected {case_ok}, unexpected {case_bad}")
print(f"report {out}")
PY
}

if [[ "${REBUILD_ONLY:-}" == "1" ]]; then
  rebuild_report
  exit 0
fi

main() {
  mkdir -p "$REPORT_DIR/states" /tmp/market-lifecycle-reports
  ln -sfn "$REPORT_DIR" /tmp/market-lifecycle-reports/latest
  : >"$LOG"
  say "report directory $REPORT_DIR"
  preflight
  init_docs
  reset_apps
  publish_charts
  local step
  for step in scenario_faultpull scenario_faultinstall scenario_faultunready scenario_holdinst scenario_slowhelm scenario_hanginit scenario_ginit scenario_faultgood scenario_upgradetest scenario_applyenv scenario_pressure scenario_gpu scenario_stopinit scenario_gcont scenario_middleware scenario_deletehook scenario_cancel_pending scenario_freeze; do
    "$step" || say "step failed: $step"
    rebuild_report || true
  done
  rebuild_report
  say "done. report: $REPORT_DIR/report.html"
  say "latest link: /tmp/market-lifecycle-reports/latest/report.html"
}

main "$@"
