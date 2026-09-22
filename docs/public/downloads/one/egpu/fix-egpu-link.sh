#!/bin/bash
# fix-egpu-link.sh -- Force the dock-internal PCIe link of ANY Thunderbolt-attached
# NVIDIA GPU down to Gen1, then re-enumerate so nvidia initializes it at Gen1
# (avoids the hard hang caused by the marginal Gen4 TB tunnel during GSP init).
#
# Generic: no hardcoded device ID / BDF. Match criteria = NVIDIA (0x10de) +
#          display class (VGA 0x0300 / 3D 0x0302) + removable=="removable"
#          (TB / hotplug external device; the internal GPU has an empty value,
#          so it is excluded automatically).
# Trigger: udev (add event) -> this script; covers both cold-boot coldplug and
#          hotplug. Can also be run manually / by the boot-time service.
#
# Usage: fix-egpu-link.sh [BDF]   (BDF optional; if omitted, scan all matching GPUs)
set -u
LOG=/var/log/egpu-gen1-fix.log
LOCK=/run/egpu-gen1-fix.lock
log(){ echo "$(date '+%F %T') egpu-gen1-fix[$$]: $*" >> "$LOG"; }

# Serialize: udev may fire once for .0 and once for .1; cold boot + hotplug may overlap.
exec 9>"$LOCK" 2>/dev/null
flock -w 60 9 2>/dev/null || { log "could not acquire lock, exiting"; exit 0; }

is_tb_nvidia_gpu(){
  local d=/sys/bus/pci/devices/$1
  [ -d "$d" ] || return 1
  [ "$(cat "$d/vendor" 2>/dev/null)" = "0x10de" ] || return 1
  case "$(cat "$d/class" 2>/dev/null)" in 0x0300*|0x0302*) ;; *) return 1;; esac
  [ "$(cat "$d/removable" 2>/dev/null)" = "removable" ] || return 1   # TB / hotplug external
  return 0
}

force_gen1_one(){
  local GPU=$1 D=/sys/bus/pci/devices/$1
  local spd; spd=$(cat "$D/current_link_speed" 2>/dev/null)
  local drv; drv=$(readlink "$D/driver" 2>/dev/null | xargs basename 2>/dev/null)
  log "$GPU: TB NVIDIA GPU, speed=$spd driver=${drv:-none}"
  [ "$spd" = "2.5 GT/s PCIe" ] && { log "$GPU: already Gen1, skipping"; return; }

  local BR BRS GPUS
  BR=$(basename "$(dirname "$(readlink -f "$D")")")
  BRS=${BR#0000:}; GPUS=${GPU#0000:}
  log "$GPU: upstream bridge (downstream port) $BR"

  # 1) Set LnkCtl2 Target=Gen1 (read-modify-write) on both bridge and GPU. Do this first:
  #    even if nvidia is already initializing (or stuck on a Gen4 MMIO read), retraining
  #    to Gen1 forces the link to re-train and completes the pending read -- this is the
  #    best rescue for the "nvidia binds at Gen4 first during hotplug" race.
  local dev old new
  for dev in "$BRS" "$GPUS"; do
    old=$(setpci -s "$dev" CAP_EXP+0x30.W 2>/dev/null) || continue
    [ -n "$old" ] || continue
    new=$(printf "%04x" $(( (0x$old & 0xFFF0) | 0x1 )))
    setpci -s "$dev" CAP_EXP+0x30.W="$new"
    log "  $dev LnkCtl2 $old->$new (Target=Gen1)"
  done
  # 2) Trigger Link Retrain on the bridge (LnkCtl bit5)
  local l ln
  l=$(setpci -s "$BRS" CAP_EXP+0x10.W 2>/dev/null)
  if [ -n "$l" ]; then
    ln=$(printf "%04x" $(( 0x$l | 0x20 )))
    setpci -s "$BRS" CAP_EXP+0x10.W="$ln"
    log "  retrain @ $BRS ($l->$ln)"
  fi
  sleep 2
  spd=$(cat "$D/current_link_speed" 2>/dev/null)
  log "  after retrain: $spd"
  [ "$spd" = "2.5 GT/s PCIe" ] || { log "  WARN: not Gen1, aborting re-enumeration of $GPU"; return; }

  # 3) unbind + remove all functions of this device, then rescan -> clean re-enum at Gen1
  local devbus=${GPU%.*} f fb
  for f in /sys/bus/pci/devices/${devbus}.*; do
    [ -e "$f" ] || continue; fb=$(basename "$f")
    if [ -L "$f/driver" ]; then
      log "  unbind $fb (from $(readlink "$f/driver" | xargs basename))"
      echo "$fb" > "$f/driver/unbind" 2>/dev/null
    fi
  done
  for f in /sys/bus/pci/devices/${devbus}.*; do
    [ -e "$f" ] || continue; fb=$(basename "$f")
    log "  remove $fb"; echo 1 > "$f/remove" 2>/dev/null
  done
  sleep 1
  echo 1 > /sys/bus/pci/rescan 2>/dev/null
  sleep 3
  if [ -d "$D" ]; then
    log "  after rescan: $GPU speed=$(cat "$D/current_link_speed" 2>/dev/null) driver=$(readlink "$D/driver" 2>/dev/null | xargs basename 2>/dev/null)"
  else
    log "  after rescan: $GPU is gone (?)"
  fi
}

TARGET="${1:-}"
log "=== start (target='${TARGET:-scan-all}') ==="
if [ -n "$TARGET" ]; then
  TARGET=${TARGET#/}; TARGET=$(basename "$TARGET")   # tolerate whatever form udev passes in
  if is_tb_nvidia_gpu "$TARGET"; then force_gen1_one "$TARGET"; else log "$TARGET: not a TB NVIDIA GPU, skipping"; fi
else
  for d in /sys/bus/pci/devices/*/; do
    b=$(basename "$d")
    is_tb_nvidia_gpu "$b" && force_gen1_one "$b"
  done
fi
log "=== done ==="
