#!/usr/bin/env python3

import argparse
import hashlib
import os
import requests
import sys
import json
import struct
import gzip
import re

CDN_URL = "https://dc3p1870nn3cj.cloudfront.net"

def get_file_size(objectid, fileid):
    url = f"{CDN_URL}/{objectid}"
    try:
        response = requests.head(url)
        response.raise_for_status()
        content_length = response.headers.get('Content-Length')
        if content_length:
            return int(content_length)
        else:
            print(f"Content-Length header missing for {fileid} from {url}", file=sys.stderr)
            sys.exit(1)
    except requests.RequestException as e:
        print(f"Error getting file size for {fileid} from {url}: {e}", file=sys.stderr)
        sys.exit(1)

def download_checksum(name):
    """Downloads the checksum for a given name."""
    url = f"{CDN_URL}/{name}.checksum.txt"
    try:
        response = requests.get(url)
        response.raise_for_status()
        return response.text.split()[0]
    except requests.exceptions.RequestException as e:
        print(f"Error getting checksum for {name} from {url}: {e}", file=sys.stderr)
        sys.exit(1)

def get_image_manifest(name):
    """Downloads the image manifest for a given name."""
    url = f"{CDN_URL}/{name}.manifest.json"
    try:
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error getting manifest for {name} from {url}: {e}", file=sys.stderr)
        sys.exit(1)


def parse_image_name(image_name):
    """
    Parses a full image name into registry, repository, and reference (tag/digest).
    Handles defaults for Docker Hub.
    """
    # Default to 'latest' tag if no tag or digest is specified
    if ":" not in image_name and "@" not in image_name:
        image_name += ":latest"

    # Split repository from reference (tag or digest)
    if "@" in image_name:
        repo_part, reference = image_name.rsplit("@", 1)
    else:
        repo_part, reference = image_name.rsplit(":", 1)

    # Determine registry and repository
    if "/" not in repo_part:
        # This is an official Docker Hub image, e.g., "ubuntu"
        registry = "registry-1.docker.io"
        repository = f"library/{repo_part}"
    else:
        parts = repo_part.split("/")
        # If the first part looks like a domain name, it's the registry
        if "." in parts[0] or ":" in parts[0]:
            registry = parts[0]
            repository = "/".join(parts[1:])
        else:
            # A scoped Docker Hub image, e.g., "bitnami/nginx"
            registry = "registry-1.docker.io"
            repository = repo_part

    return registry, repository, reference

def get_auth_token(registry, repository):
    """
    Gets an authentication token from the registry's auth service.
    """
    # First, probe the registry to get the auth challenge
    try:
        probe_url = f"https://{registry}/v2/"
        response = requests.get(probe_url, timeout=10)

        if response.status_code == 401:
            # Parse the WWW-Authenticate header
            auth_header = response.headers.get("WWW-Authenticate", "")
            match = re.search(r'realm="([^"]+)"', auth_header)
            if match:
                realm = match.group(1)
                service_match = re.search(r'service="([^"]+)"', auth_header)
                service = service_match.group(1) if service_match else ""

                # Request a token
                token_url = f"{realm}?service={service}&scope=repository:{repository}:pull"
                token_response = requests.get(token_url, timeout=10)
                token_response.raise_for_status()
                token_data = token_response.json()
                return token_data.get("token", token_data.get("access_token"))

        # If no auth required, return None
        return None

    except Exception as e:
        print(f"Error getting auth token for registry {registry}: {e}", file=sys.stderr)
        sys.exit(1)

def get_layer_uncompressed_size_from_registry(registry, repository, layer_digest, auth_token=None):
    """Get the uncompressed size of a gzipped layer from the registry using OCI distribution API."""
    try:
        # Construct the blob URL - digest already contains sha256: prefix
        blob_url = f"https://{registry}/v2/{repository}/blobs/{layer_digest}"

        headers = {}
        if auth_token:
            headers["Authorization"] = f"Bearer {auth_token}"

        # Make a HEAD request to get the compressed size
        response = requests.head(blob_url, headers=headers, timeout=10, allow_redirects=True)
        response.raise_for_status()

        compressed_size = int(response.headers.get('Content-Length', 0))
        if compressed_size < 4:
            print(f"unexpected compressed size of layer {layer_digest} in image {repository}: {compressed_size}, expecting more than 4 bytes", file=sys.stderr)
            sys.exit(1)

        # Request only the last 4 bytes of the gzipped blob
        headers['Range'] = f'bytes={compressed_size - 4}-{compressed_size - 1}'

        response = requests.get(blob_url, headers=headers, timeout=10)
        response.raise_for_status()

        # The last 4 bytes of a gzip file contain the uncompressed size (mod 2^32)
        if len(response.content) != 4:
            print(f"unexpected response size of original size: {len(response.content)}, expecting 4 bytes", file=sys.stderr)
            sys.exit(1)

        uncompressed_size = struct.unpack('<I', response.content)[0]
        return uncompressed_size

    except Exception as e:
        print(f"Error getting uncompressed size for layer {layer_digest} in image {repository}: {e}", file=sys.stderr)
        sys.exit(1)

def process_image_manifest(manifest, image_reference):
    """Process image manifest to add uncompressed sizes for gzipped layers."""
    if not manifest or 'layers' not in manifest:
        return manifest

    # Parse the image reference to get registry and repository
    registry, repository, reference = parse_image_name(image_reference)

    # Get auth token if needed
    auth_token = get_auth_token(registry, repository)

    for layer in manifest['layers']:
        media_type = layer.get('mediaType', '')

        # Check if this is a gzipped layer
        if media_type == 'application/vnd.oci.image.layer.v1.tar+gzip':
            digest = layer.get('digest', '')
            if not digest:
                print(f"Missing digest of layer {layer} in image {repository}", file=sys.stderr)
                sys.exit(1)
            # Get uncompressed size from the registry
            uncompressed_size = get_layer_uncompressed_size_from_registry(
                registry, repository, digest, auth_token
            )

            if not uncompressed_size:
                print(f"got invalid uncompressed size for layer {layer} in image {repository}", file=sys.stderr)
                sys.exit(1)

            layer['uncompressedSize'] = uncompressed_size
            print(f"Added uncompressed size {uncompressed_size} for layer {layer['digest']} in image {repository}", file=sys.stderr)

    return manifest

def main():
    """Main function."""
    parser = argparse.ArgumentParser()
    parser.add_argument("manifest_file", help="The manifest file to write to.")
    args = parser.parse_args()

    manifest_file = args.manifest_file
    version = os.environ.get("VERSION", "")
    repo_path = os.environ.get("REPO_PATH", "/")
    manifest_amd64_data = {}
    manifest_arm64_data = {}

    # Process components
    try:
        with open("components", "r") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue

                # Replace version
                if version:
                    line = line.replace("#__VERSION__", version)

                # Replace repo path
                if repo_path:
                    line = line.replace("#__REPO_PATH__", repo_path)

                fields = line.split(",")
                if len(fields) < 5:
                    print(f"Format error in components file: {line}", file=sys.stderr)
                    sys.exit(1)

                filename, path, deps, _, fileid = fields[:5]
                print(f"Downloading file checksum for {filename}")

                name = hashlib.md5(filename.encode()).hexdigest()
                url_amd64 = name
                url_arm64 = f"arm64/{name}"

                checksum_amd64 = download_checksum(url_amd64)
                checksum_arm64 = download_checksum(url_arm64)

                file_size_amd64 = get_file_size(url_amd64, fileid)
                file_size_arm64 = get_file_size(url_arm64, fileid)

                manifest_amd64_data[filename] = {
                    "type": "component",
                    "path": path,
                    "deps": deps,
                    "url_amd64": url_amd64,
                    "checksum_amd64": checksum_amd64,
                    "fileid": fileid,
                    "size": file_size_amd64,
                }


                manifest_arm64_data[filename] = {
                    "type": "component",
                    "path": path,
                    "deps": deps,
                    "url_arm64": url_arm64,
                    "checksum_arm64": checksum_arm64,
                    "fileid": fileid,
                    "size": file_size_arm64,
                }

    except FileNotFoundError:
        print("Error: 'components' file not found.", file=sys.stderr)
        sys.exit(1)

    # Process images
    path = "images"
    for deps_file in ["images.mf"]:
        try:
            with open(deps_file, "r") as f:
                for line in f:
                    line = line.strip()
                    if not line:
                        continue

                    print(f"Downloading file checksum for {line}")
                    name = hashlib.md5(line.encode()).hexdigest()
                    url_amd64 = f"{name}.tar.gz"
                    url_arm64 = f"arm64/{name}.tar.gz"

                    checksum_amd64 = download_checksum(name)
                    checksum_arm64 = download_checksum(f"arm64/{name}")

                    file_size_amd64 = get_file_size(url_amd64, line)
                    file_size_arm64 = get_file_size(url_arm64, line)

                    # Get the image manifest
                    image_manifest_amd64 = get_image_manifest(name)
                    image_manifest_arm64 = get_image_manifest(f"arm64/{name}")

                    # Process manifests to add uncompressed sizes
                    # Pass the image reference (line) which contains the full image name
                    image_manifest_amd64 = process_image_manifest(image_manifest_amd64, line)
                    image_manifest_arm64 = process_image_manifest(image_manifest_arm64, line)

                    filename = f"{name}.tar.gz"
                    manifest_amd64_data[filename] = {
                        "type": "image",
                        "path": path,
                        "deps": deps_file,
                        "url_amd64": url_amd64,
                        "checksum_amd64": checksum_amd64,
                        "fileid": line,
                        "size": file_size_amd64,
                        "manifest": image_manifest_amd64
                    }

                    manifest_arm64_data[filename] = {
                        "type": "image",
                        "path": path,
                        "deps": deps_file,
                        "url_arm64": url_arm64,
                        "checksum_arm64": checksum_arm64,
                        "fileid": line,
                        "size": file_size_arm64,
                        "manifest": image_manifest_arm64
                    }


        except FileNotFoundError:
            print(f"Warning: '{deps_file}' not found, skipping.", file=sys.stderr)
            sys.exit(1)


    # Write the manifest file
    amd64_manifest_file = f"{manifest_file}.amd64"
    with open(amd64_manifest_file, "w") as mf:
        json.dump(manifest_amd64_data, mf, indent=2)

    arm64_manifest_file = f"{manifest_file}.arm64"
    with open(arm64_manifest_file, "w") as mf:
        json.dump(manifest_arm64_data, mf, indent=2)


    # TODO: compress the manifest files

if __name__ == "__main__":
    main()
