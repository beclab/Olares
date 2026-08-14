#!/bin/sh
#
# Copyright (C) 2026 Intel Corporation
# SPDX-License-Identifier: MIT

set -e

# Copyright/license for the converted file
COPYRIGHT="Copyright (C) 2026 Intel Corporation"
LICENSE="SPDX-License-Identifier: MIT"

# Label needed in configMap to get (Helm installed) Grafana to load it as dashboard
LABEL="grafana_dashboard=1"

# XPUMD Helm define to use as prefix for configMap name & Grafana dashboard URL
prefix_macro="xpumd.fullname"

# XPUMD Helm chart value names
enable_dashboards="grafana.dashboards"
grafana_namespace="grafana.namespace"

# dummy for 'kubectl', so that output will include namespace field (overridden later on)
DUMMY_NS="default"

error_exit ()
{
	name=${0##*/}
	cat << EOF

ERROR: $1!

Convert given Grafana *.json dashboard specs to Grafana dashboard configMap,
compatible with XPUMD Helm chart.

Dashboard 'uid' is overridden with value composed of the XPUMD Helm
chart release and JSON file name, similarly to the produced configMap
name and namespace.

Usage:
	$name <JSON files>

ERROR: $1!
EOF
	exit 1
}

if [ -z "$(which jq)" ]; then
	error_exit "'jq' required for dashboard checks, please install 'jq' first"
fi

if [ -z "$(which kubectl)" ]; then
	error_exit "'kubectl' required for dashboard conversion, please install 'kubernetes-client' first"
fi

if ! kubectl version; then
	error_exit "Broken/missing 'kubectl' cluster config (script does not need it, 'kubectl' does)"
fi

uid=""
title=""

echo
echo "Got following Grafana dashboards:"
for file in "$@"; do
	if [ ! -f "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi
	if [ "${file%.json}" = "$file" ]; then
		error_exit "JSON file '$file' does not exist"
	fi

	# Both dashboard 'uid'...
	uid=$(jq .uid "$file" | tail -1 | tr -d '"')
	if [ -z "$uid" ]; then
		error_exit "'$file' dashboard has invalid JSON"
	elif [ "$uid" = "null" ]; then
		error_exit "'$file' dashboard has no 'uid' field (to override with Helm chart specific one)"
	fi

	# ...and 'title' needed.
	title=$(jq .title "$file" | tail -1 | tr -d '"')
	if [ "$title" = "null" ]; then
		error_exit "'$file' dashboard has no 'title' field (to show in Grafana Dashboards list)"
	fi

	echo "- file: $file, uid: '$uid', title: '$title'"
done

echo
echo "Converting:"
for file in "$@"; do
	base=${file##*/}
	name=${base%.json}
	dst="configmap-${name}.yaml"

	uid=$(jq .uid "$file" | tail -1 | tr -d '"')
	title=$(jq .title "$file" | tail -1 | tr -d '"')

	# convert to k8s object name ("[a-z0-9][-a-z0-9]*[a-z0-9]"):
	# - upper-case -> lowercase, '_' -> '-'
	# - drop anything outside [-a-z]
	# - drop '-' prefix & suffix and successive '-' chars
	k8name=$(echo "$name" | tr A-Z_ a-z- | tr -d -c a-z- | sed -e 's/^-*//' -e 's/-*$//' -e 's/--*/-/g')

	echo "- $base -> $dst"

	echo "{{- if .Values.$enable_dashboards }}" > "$dst"

	# shellcheck disable=SC2129
	echo "# $COPYRIGHT" >> "$dst"
	echo "# $LICENSE" >> "$dst"
	echo "#" >> "$dst"
	echo "# ${0##*/}: $base -> $dst" >> "$dst"

	kubectl create cm -n "$DUMMY_NS" --from-file "$file" --dry-run=client -o yaml "$k8name" |\
	  kubectl label -f- --local --dry-run=client -o yaml "$LABEL" |\
	  grep -v -e "^  creationTimestamp:" >> "$dst"

	echo "{{- end }}" >> "$dst"

	# convert JSON content conflicting with Helm to Helm compatible format
	# and add suitable Dashboard chart Helm variables to the configMap
	sed -i \
	  -e 's/\({{[a-z_]\+}}\)/{{ printf "\1" }}/g' \
	  -e 's/name:.*$/name: {{ include "'$prefix_macro'" . }}'"-${k8name}/" \
	  -e 's/space:.*$/space: {{ .Values.'$grafana_namespace' }}/' \
	  -e "s/${uid}/"'{{ include "'$prefix_macro'" . }}'"-${k8name}/" \
	  "$dst"
	  # override dashboard title
	  # -e "s/${title}/XPUMD $name/" \
done

echo
echo "DONE!"
