#!/bin/sh
#
# Copyright (C) 2026 Intel Corporation
# SPDX-License-Identifier: MIT

set -e

error_exit ()
{
	name=${0##*/}
	cat << EOF 1>&2

ERROR: $1!

Given Grafana dashboard JSON file that uses old XPUM v1 metric/attrib
names, this does partial conversion to new XPUM v2 metric/attrib names
(with OpenTelemetry semantics).

Conversion is partial if queries use XPUM v1 attributes which names
are too unspecific (e.g. "type"), or metrics that are not supported
yet in v2.

Usage:
	$name <old JSON file> <new JSON file>

ERROR: $1!
EOF
	exit 1
}

if [ $# -ne 2 ]; then
	error_exit "2 JSON file names expected as args"
fi
old=$1
new=$2

if [ ! -f "$old" ] || [ "${old%.json}" = "$old" ]; then
	error_exit "JSON (*.json) file '$old' does not exist"
fi
if [ "$old" = "$new" ]; then
	error_exit "in-place update not supported, provide different file names"
fi

if [ -z "$(which diff)" ]; then
	error_exit "'diff' required for showing dashboard diff, please install 'diff' first"
fi
if [ -z "$(which jq)" ]; then
	error_exit "'jq' required for dashboard checks, please install 'jq' first"
fi

# Both dashboard 'uid'...
uid=$(jq .uid "$old" | tail -1 | tr -d '"')
if [ -z "$uid" ]; then
	error_exit "'$old' dashboard has invalid JSON"
fi
# ...and 'title' needed.
title=$(jq .title "$old" | tail -1 | tr -d '"')
if [ "$title" = "null" ]; then
	error_exit "'$old' dashboard has no 'title' field (to show in Grafana Dashboards list)"
fi

metrics=$(grep xpum_ "$old" | sed 's/^.*\(xpum_[_a-z]*\).*$/\1/' | sort -u)

count=0
unsupported=0
for m in $metrics; do
	case $m in
	*_eu_*|*_fabric_*|*_errors_*|*_resets_*)
		echo "- UNSUPPORTED metric: $m" 1>&2
		unsupported=$((unsupported+1))
		;;
	*)
		count=$((count+1))
		;;
	esac
done

echo
echo "Converting $count metrics in '$old' to '$new'..." 1>&2

# Add to XPUM dashboard, to replace removed metrics:
# - GPU info & HW status metrics
# - mem BW rate vs. rate() - debug
# - PCI rate vs rate() - debug
# - frequency ratio
# - mem BW ratio
# - power ratio
# - PCIe ratio

# Dashboard changes:
# - variable (drop-down) metrics ("hw_gpu_info" is always available)
# - metric names
#   - Mhz -> Hz unit
#   - engine ratio averaging
# - attribute names in queries
# - attribute names in legends
sed "$old" \
-e 's/label_values(xpum_[a-z_]*/label_values(hw_gpu_info/' \
\
-e 's/xpum_energy_joules_total/hw_energy_joules_total/g' \
\
-e 's/xpum_engine_ratio{\(.*\)}/avg by (pci_bdf) (hw_gpu_utilization_ratio{\1})/g' \
-e 's/xpum_per_engine_ratio/hw_gpu_utilization_ratio/g' \
-e 's/xpum_engine_group_ratio{\(.*\)}/avg by (pci_bdf,hw_gpu_task) (hw_gpu_utilization_ratio{\1})/g' \
\
-e 's/xpum_frequency_mhz{\(.*\)}[*]1000000/hw_frequency_hertz{aggregation=\\"avg\\", \1}/g' \
\
-e 's/xpum_memory_bandwidth_ratio/hw_memory_bandwidth_utilization_ratio/g' \
-e 's/xpum_memory_read_bytes_total{/hw_memory_io_bytes_total{network_io_direction=\\"receive\\",/g' \
-e 's/xpum_memory_used_bytes/hw_memory_usage_bytes/g' \
-e 's/xpum_memory_write_bytes_total{/hw_memory_io_bytes_total{network_io_direction=\\"transmit\\",/g' \
-e 's%xpum_memory_ratio{\(.*\)}%hw_memory_usage_bytes{\1}/hw_memory_size_bytes{\1}%g' \
\
-e 's/xpum_pcie_read_bytes_total{/hw_gpu_io_bytes_total{network_io_direction=\\"receive\\",/g' \
-e 's/xpum_pcie_write_bytes_total{/hw_gpu_io_bytes_total{network_io_direction=\\"transmit\\",/g' \
\
-e 's/xpum_power_watts/hw_power_watts/' \
\
-e 's/xpum_max_temperature_celsius{/hw_temperature_celsius{statistic=\\"max\\",/g' \
-e 's/xpum_temperature_celsius/hw_temperature_celsius/g' \
\
-e 's/location=\\"mem\\"/location=\\"memory\\"/g' \
-e 's/location=/hw_sensor_location=/g' \
\
-e 's/instance=/node=/g' \
\
-e 's/{{sub_dev}}/{{com_intel_subdevice_id}}/g' \
-e 's/{{type}}-{{engine_id}}/{{hw_name}}/g' \
> "$new"

if [ $unsupported -gt 0 ]; then
	echo
	echo "WARNING: dashboard contains queries with $unsupported unsupported metrics!"
fi

echo
echo "DONE, for diffs, see 'diff -ub $old $new'!"
