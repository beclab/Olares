#!/bin/bash
# collect-egpu-info.sh -- Collect everything needed for an eGPU support report on Olares OS.
#
# Read-only: it only reads system state. The single output file is written into the
# current directory. Nothing is uploaded and no system setting is changed.
#
# Usage: sudo ./collect-egpu-info.sh
# Output: ./egpu-report-<hostname>-<timestamp>.txt   (attach this to the forum thread)

set -u
OUT="./egpu-report-$(hostname -s 2>/dev/null || echo host)-$(date +%Y%m%d-%H%M%S).txt"

sec(){ echo; echo "=============================================================="; echo "## $*"; echo "=============================================================="; }
have(){ command -v "$1" >/dev/null 2>&1; }

# Log lines that indicate a real failure (as opposed to normal boot chatter).
# NOTE: NVIDIA Xid is matched via its "NVRM: Xid" / "Xid (PCI:" form on purpose --
# a bare "Xid" also matches the Realtek NIC's chip id line ("RTL8125B, ..., XID 641").
CRIT='NVRM: Xid|Xid \(PCI|fallen off the bus|GPU has fallen|NV_ERR_GPU_IS_LOST|GPU is lost|CmpltTO|UnsupReq|severity=Uncorrectable|Multiple Uncorrectable|RmInitAdapter failed|rm_init_adapter failed|BAR1 is 0M|WPR2|BAR [0-9].*(failed to assign|can.t assign)|bridge window.*(failed to assign|can.t assign)|Unable to change power state|no response from device|soft lockup|hung_task|watchdog: BUG|Kernel panic'
# Known-harmless lines that would otherwise trip the filter above on this platform
# (the SBIOS temp/power-mode assertions appear on every boot and mean nothing here).
BENIGN='PlatformRequestHandler|platform_request_handler|from SBIOS|XID [0-9]+, IRQ'
# Lines that are useful context even when nothing is wrong.
CTX='nvidia|NVRM|thunderbolt|pciehp|AER|available PCIe bandwidth|10de:|JHL9[45]|Link (Up|Down)|BAR [0-9].*assigned'

# ---- Detect the external (Thunderbolt-attached) NVIDIA GPU, if any ----
EXT_BDF=""; EXT_SPEED=""; EXT_DRV=""
for d in /sys/bus/pci/devices/*/; do
  [ "$(cat "$d/vendor" 2>/dev/null)" = "0x10de" ] || continue
  case "$(cat "$d/class" 2>/dev/null)" in 0x0300*|0x0302*) ;; *) continue;; esac
  [ "$(cat "$d/removable" 2>/dev/null)" = "removable" ] || continue
  EXT_BDF=$(basename "$d")
  EXT_SPEED=$(cat "$d/current_link_speed" 2>/dev/null)
  EXT_DRV=$(readlink "$d/driver" 2>/dev/null | xargs basename 2>/dev/null)
  break
done
CRIT_N=$(dmesg 2>/dev/null | grep -iE "$CRIT" | grep -ivE "$BENIGN" | wc -l | tr -d " ")

collect(){

echo "eGPU report generated at $(date -Is 2>/dev/null || date)"

sec "0. Summary (read this first)"
if [ -n "$EXT_BDF" ]; then
  echo "External (Thunderbolt) GPU detected : YES  -> $EXT_BDF"
  echo "  link speed now                   : ${EXT_SPEED:-unknown}   (2.5 GT/s = Gen1, 16.0 GT/s = Gen4)"
  echo "  driver bound                     : ${EXT_DRV:-none}"
else
  echo "External (Thunderbolt) GPU detected : NO"
  echo "  -> the GPU was never enumerated on the PCI bus. Check dock power, cable,"
  echo "     the host port, and section 3 (Thunderbolt) / section 4 (PCI topology)."
fi
GEN1_EN=$(systemctl is-enabled egpu-gen1-fix.service 2>/dev/null | head -1)
echo "Gen1 workaround installed          : $( [ -x /usr/local/sbin/fix-egpu-link.sh ] && echo YES || echo NO )"
echo "Gen1 workaround enabled            : ${GEN1_EN:-not-installed}"
echo "Critical error lines in current dmesg: ${CRIT_N:-0}   (see section 7a)"
echo "Internal GPU(s):"
for d in /sys/bus/pci/devices/*/; do
  [ "$(cat "$d/vendor" 2>/dev/null)" = "0x10de" ] || continue
  case "$(cat "$d/class" 2>/dev/null)" in 0x0300*|0x0302*) ;; *) continue;; esac
  [ "$(cat "$d/removable" 2>/dev/null)" = "removable" ] && continue
  echo "  $(basename "$d")  speed=$(cat "$d/current_link_speed" 2>/dev/null)  driver=$(readlink "$d/driver" 2>/dev/null | xargs basename 2>/dev/null || echo none)"
done

sec "1. System"
uname -a
[ -r /etc/os-release ] && grep -E '^(PRETTY_NAME|VERSION)=' /etc/os-release
for f in /etc/olares-release /etc/olares/version /opt/olares/VERSION; do
  [ -r "$f" ] && { echo "--- $f ---"; cat "$f"; }
done
have olares-cli && { echo "--- olares-cli version ---"; olares-cli --version 2>&1 | head -3; }
echo "--- kernel cmdline ---"
cat /proc/cmdline 2>/dev/null
echo "--- uptime ---"
uptime 2>/dev/null

sec "2. NVIDIA driver"
if have nvidia-smi; then
  nvidia-smi 2>&1 | head -20
  echo "--- per-GPU link info ---"
  nvidia-smi --query-gpu=index,name,pci.bus_id,pcie.link.gen.current,pcie.link.gen.max,pcie.link.width.current,memory.total \
             --format=csv 2>&1
else
  echo "nvidia-smi NOT FOUND (driver not installed, or not in PATH)"
fi
echo "--- loaded nvidia modules ---"
lsmod 2>/dev/null | grep -i nvidia || echo "(no nvidia module loaded)"
echo "--- module version ---"
modinfo nvidia 2>/dev/null | grep -E '^(version|filename)' || echo "(modinfo nvidia unavailable)"

sec "3. Thunderbolt"
if have boltctl; then
  bc=$(boltctl list 2>&1)
  if [ -n "$bc" ]; then echo "$bc"; else echo "(boltctl list returned nothing -- no Thunderbolt peripheral is currently connected/authorized)"; fi
else
  echo "boltctl NOT FOUND"
fi
echo "--- USB4/TB tunnel rate per device ---"
found_tb=0
for d in /sys/bus/thunderbolt/devices/*/; do
  [ -d "$d" ] || continue
  n=$(basename "$d")
  nm=$(cat "$d/device_name" 2>/dev/null)
  gen=$(cat "$d/generation" 2>/dev/null)
  rxs=$(cat "$d/rx_speed" 2>/dev/null); txs=$(cat "$d/tx_speed" 2>/dev/null)
  rxl=$(cat "$d/rx_lanes" 2>/dev/null); txl=$(cat "$d/tx_lanes" 2>/dev/null)
  auth=$(cat "$d/authorized" 2>/dev/null)
  [ -n "${rxs}${txs}${gen}" ] && { echo "  [$n] name=${nm:-?} gen=${gen:-?} authorized=${auth:-?} rx=${rxs:-?}/${rxl:-?}lane tx=${txs:-?}/${txl:-?}lane"; found_tb=1; }
done
[ "$found_tb" = "0" ] && echo "  (only the host router, or no Thunderbolt device present)"

sec "4. PCI topology"
lspci -tvnn 2>/dev/null || lspci -tv 2>/dev/null
echo
echo "--- NVIDIA / Thunderbolt / bridge devices ---"
lspci -nn 2>/dev/null | grep -iE "10de:|thunderbolt|JHL|PCI bridge|USB4"

sec "5. GPU link speed (per device, and the chain up to the root)"
for d in /sys/bus/pci/devices/*/; do
  bdf=$(basename "$d")
  [ "$(cat "$d/vendor" 2>/dev/null)" = "0x10de" ] || continue
  case "$(cat "$d/class" 2>/dev/null)" in 0x0300*|0x0302*) ;; *) continue;; esac
  rm_=$(cat "$d/removable" 2>/dev/null)
  echo "GPU $bdf  removable='${rm_:-<empty>}'  $( [ "$rm_" = "removable" ] && echo '<-- EXTERNAL (Thunderbolt)' || echo '<-- internal' )"
  echo "  device : $(lspci -s "${bdf#0000:}" 2>/dev/null | sed 's/^[^ ]* //')"
  echo "  driver : $(readlink "$d/driver" 2>/dev/null | xargs basename 2>/dev/null || echo none)"
  echo "  speed  : cur=$(cat "$d/current_link_speed" 2>/dev/null) max=$(cat "$d/max_link_speed" 2>/dev/null)"
  echo "  width  : cur=$(cat "$d/current_link_width" 2>/dev/null) max=$(cat "$d/max_link_width" 2>/dev/null)"
  echo "  --- link chain from this GPU up to the root ---"
  p="$d"
  while :; do
    p=$(dirname "$(readlink -f "$p")")
    [ -r "$p/vendor" ] || break
    pb=$(basename "$p")
    echo "    $pb  cur=$(cat "$p/current_link_speed" 2>/dev/null) max=$(cat "$p/max_link_speed" 2>/dev/null)  $(lspci -s "${pb#0000:}" 2>/dev/null | sed 's/^[^ ]* //' | cut -c1-56)"
  done
  echo
done

sec "6. BAR / memory windows of the external GPU"
if [ -n "$EXT_BDF" ]; then
  echo "--- $EXT_BDF ---"
  lspci -vv -s "${EXT_BDF#0000:}" 2>/dev/null | grep -iE "Region|Resizable BAR|BAR [0-9]|LnkCap:|LnkSta:|LnkCtl2:" | head -25
else
  echo "(no external GPU detected -- nothing to report here)"
fi

sec "7a. CRITICAL errors in the current boot"
cl=$(dmesg 2>/dev/null | grep -iE "$CRIT" | grep -ivE "$BENIGN" | tail -40)
if [ -n "$cl" ]; then echo "$cl"; else echo "(none found -- no GPU-lost / AER-uncorrectable / BAR-assign failure in this boot)"; fi

sec "7b. PCIe / Thunderbolt / NVIDIA log of the current boot (context, may be normal)"
dmesg 2>/dev/null | grep -iE "$CTX" | tail -60 || echo "(dmesg unreadable -- rerun with sudo)"

sec "8a. CRITICAL errors in the PREVIOUS boot (use this if the previous boot hung)"
pl=$(journalctl -b -1 -k --no-pager 2>/dev/null | grep -iE "$CRIT" | grep -ivE "$BENIGN" | tail -40)
if [ -n "$pl" ]; then echo "$pl"
else echo "(none found, or no persistent journal for the previous boot)"; fi

sec "8b. PCIe / Thunderbolt / NVIDIA log of the PREVIOUS boot"
journalctl -b -1 -k --no-pager 2>/dev/null | grep -iE "$CTX" | tail -50 || echo "(no persistent journal for the previous boot)"

sec "8c. Last lines of the PREVIOUS boot (where it stopped)"
journalctl -b -1 -k --no-pager 2>/dev/null | tail -15 || echo "(no persistent journal for the previous boot)"

sec "9. Gen1 workaround status"
echo "enabled: $(systemctl is-enabled egpu-gen1-fix.service 2>&1)"
echo "active : $(systemctl is-active egpu-gen1-fix.service 2>&1)"
echo "script : $( [ -x /usr/local/sbin/fix-egpu-link.sh ] && echo present || echo missing )"
echo "udev   : $( [ -r /etc/udev/rules.d/99-egpu-gen1-fix.rules ] && echo present || echo missing )"
echo "--- /var/log/egpu-gen1-fix.log (tail) ---"
tail -n 40 /var/log/egpu-gen1-fix.log 2>/dev/null || echo "(not installed / no log)"

sec "10. IOMMU / container runtime"
dmesg 2>/dev/null | grep -iE "DMAR:|IOMMU" | head -8
echo "--- GPUs visible to the container runtime ---"
have nvidia-container-cli && nvidia-container-cli info 2>&1 | head -25 || echo "(nvidia-container-cli not present)"

echo
echo "=============================================================="
echo "## End of report"
echo "=============================================================="

}

collect > "$OUT" 2>&1

echo "Report written to: $OUT"
if [ -n "$EXT_BDF" ]; then
  echo "External GPU detected at $EXT_BDF (link ${EXT_SPEED:-unknown})."
else
  echo "NOTE: no external Thunderbolt GPU was detected in this boot."
fi
echo "Please attach that file to your forum thread."
