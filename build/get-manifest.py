#!/usr/bin/env python3

import requests
import json
import argparse
import re
import sys
import platform

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
    except requests.exceptions.RequestException as e:
        print(f"Error: Could not connect to registry at {probe_url}. Details: {e}", file=sys.stderr)
        sys.exit(1)

    if response.status_code != 401:
        # Either public or something is wrong, we can try without a token
        return None

    auth_header = response.headers.get("Www-Authenticate")
    if not auth_header:
        print(f"Error: Registry {registry} returned 401 but did not provide Www-Authenticate header.", file=sys.stderr)
        sys.exit(1)

    # Parse the Www-Authenticate header to find realm, service, and scope
    try:
        realm = re.search('realm="([^"]+)"', auth_header).group(1)
        service = re.search('service="([^"]+)"', auth_header).group(1)
        # Scope for the specific repository is needed
        scope = f"repository:{repository}:pull"
    except AttributeError:
        print(f"Error: Could not parse Www-Authenticate header: {auth_header}", file=sys.stderr)
        sys.exit(1)

    # Request the actual token from the auth realm
    auth_params = {
        "service": service,
        "scope": scope
    }
    
    try:
        auth_response = requests.get(realm, params=auth_params, timeout=10)
        auth_response.raise_for_status()
        return auth_response.json().get("token")
    except requests.exceptions.RequestException as e:
        print(f"Error: Failed to get auth token from {realm}. Details: {e}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError:
        print(f"Error: Failed to decode JSON response from auth server: {auth_response.text}", file=sys.stderr)
        sys.exit(1)


def get_manifest(registry, repository, reference, token):
    """
    Fetches the image manifest from the registry.
    """
    manifest_url = f"https://{registry}/v2/{repository}/manifests/{reference}"
    
    headers = {
        # Request multiple manifest types, the registry will return the correct one
        "Accept": "application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json"
    }

    if token:
        headers["Authorization"] = f"Bearer {token}"

    try:
        response = requests.get(manifest_url, headers=headers, timeout=10)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 401 and not token:
             print("Error: Received 401 Unauthorized. Attempting to get a token...", file=sys.stderr)
             # The initial probe might have passed, but manifest access requires auth.
             # We re-run the token acquisition logic.
             new_token = get_auth_token(registry, repository)
             if new_token:
                 return get_manifest(registry, repository, reference, new_token)
        print(f"Error: Failed to fetch manifest from {manifest_url}. Status: {e.response.status_code}", file=sys.stderr)
        print(f"Response: {e.response.text}", file=sys.stderr)
        sys.exit(1)
    except requests.exceptions.RequestException as e:
        print(f"Error: A network error occurred. Details: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="Fetch an OCI/Docker image manifest from a container registry.",
        epilog="""Examples:
  python get_manifest.py ubuntu:22.04
  python get_manifest.py quay.io/brancz/kube-rbac-proxy:v0.18.1 -o manifest.json
  python get_manifest.py gcr.io/google-containers/pause:3.9""",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument("image_name", help="Full name of the container image (e.g., 'ubuntu:latest' or 'quay.io/prometheus/node-exporter:v1.7.0')")
    parser.add_argument("-o", "--output-file", help="Optional. Path to write the final manifest JSON to. If not provided, prints to stdout.")
    args = parser.parse_args()

    registry, repository, reference = parse_image_name(args.image_name)

    # Suppress informational prints if writing to a file
    verbose_print = print if not args.output_file else lambda *a, **k: None

    verbose_print(f"Registry:   {registry}")
    verbose_print(f"Repository: {repository}")
    verbose_print(f"Reference:  {reference}", end='\n\n', flush=True)

    token = get_auth_token(registry, repository)

    if not token and not args.output_file:
        print("No authentication token needed or could be retrieved. Proceeding without token...", file=sys.stderr)

    manifest = get_manifest(registry, repository, reference, token)
    final_manifest = None

    media_type = manifest.get("mediaType", "")
    if "manifest.list" in media_type or "image.index" in media_type:
        verbose_print("Detected a multi-platform image index. Finding manifest for current architecture...")

        system_arch = platform.machine()
        arch_map = {"x86_64": "amd64", "aarch64": "arm64"}
        target_arch = arch_map.get(system_arch, system_arch)

        verbose_print(f"System architecture: {system_arch} -> Target: linux/{target_arch}")

        target_digest = None
        for m in manifest.get("manifests", []):
            plat = m.get("platform", {})
            if plat.get("os") == "linux" and plat.get("architecture") == target_arch:
                target_digest = m.get("digest")
                break

        if target_digest:
            verbose_print(f"Found manifest for linux/{target_arch} with digest: {target_digest}\n")
            final_manifest = get_manifest(registry, repository, target_digest, token)
        else:
            print(f"Error: Could not find a manifest for 'linux/{target_arch}' in the index.", file=sys.stderr)
            if not args.output_file:
                print("Available platforms:", file=sys.stderr)
                for m in manifest.get("manifests", []):
                    print(f"  - {m.get('platform', {}).get('os')}/{m.get('platform', {}).get('architecture')}", file=sys.stderr)
            sys.exit(1)
    else:
        final_manifest = manifest

    if final_manifest:
        if args.output_file:
            try:
                with open(args.output_file, 'w') as f:
                    json.dump(final_manifest, f, indent=2)
                print(f"Successfully wrote manifest to {args.output_file}")
            except IOError as e:
                print(f"Error: Could not write to file {args.output_file}. Details: {e}", file=sys.stderr)
                sys.exit(1)
        else:
            print(json.dumps(final_manifest, indent=2))


if __name__ == "__main__":
    main()
