#!/usr/bin/env python3

import argparse
import hashlib
import os
import requests
import sys
import json

CDN_URL = "https://cdn.olares.com"

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

def main():
    """Main function."""
    parser = argparse.ArgumentParser()
    parser.add_argument("manifest_file", help="The manifest file to write to.")
    args = parser.parse_args()

    manifest_file = args.manifest_file
    version = os.environ.get("VERSION", "")
    release_id = os.environ.get("RELEASE_ID", "")
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

                if release_id:
                    line = line.replace("#__RELEASE_ID_SUFFIX__", "."+release_id)

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
