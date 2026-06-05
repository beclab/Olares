#!/usr/bin/env python3

import re
import sys
from pathlib import Path


def fail(message: str) -> None:
    print(f"ERROR: {message}", file=sys.stderr)
    sys.exit(1)


def read_text(path: Path) -> str:
    if not path.exists():
        fail(f"missing required file: {path}")
    return path.read_text(encoding="utf-8")


def normalize_release_tag(tag: str) -> str:
    value = tag.strip()
    if not value:
        fail("release tag/version is empty")
    return value[1:] if value.startswith("v") else value


def extract_image_tag(image: str, expected_repo: str) -> str:
    prefix = f"{expected_repo}:"
    if not image.startswith(prefix):
        fail(f"expected image '{expected_repo}:<tag>', got '{image}'")
    tag = image[len(prefix) :].strip()
    if not tag:
        fail(f"missing tag in image '{image}'")
    return tag


def extract_first(pattern: str, text: str, field_name: str) -> str:
    match = re.search(pattern, text, re.MULTILINE)
    if not match:
        fail(f"cannot find {field_name}")
    value = match.group(1).strip().strip("\"'")
    if not value:
        fail(f"{field_name} is empty")
    return value


def extract_env_value(yaml_text: str, env_name: str) -> str:
    pattern = (
        rf"^\s*-\s*name:\s*{re.escape(env_name)}\s*$"
        rf"\n^\s*value:\s*([^\n#]+)"
    )
    return extract_first(pattern, yaml_text, f"{env_name} value")


def ensure_pinned_image(image: str, field_name: str) -> None:
    value = image.strip()
    if value.endswith(":latest"):
        fail(f"{field_name} must not use ':latest' ({value})")
    if "@sha256:" in value:
        return
    if re.match(r"^[^:@\s]+(?:/[^:@\s]+)+:[^:\s]+$", value):
        return
    fail(f"{field_name} must be pinned by tag or digest ({value})")


def main() -> None:
    if len(sys.argv) != 2:
        fail("usage: build/verify_release_image_pins.py <release-tag>")

    release_tag = normalize_release_tag(sys.argv[1])
    repo_root = Path(__file__).resolve().parents[1]

    appservice_file = repo_root / "framework/app-service/.olares/config/cluster/deploy/appservice_deploy.yaml"
    sysevent_file = repo_root / "platform/tapr/.olares/config/cluster/deploy/sys_event_deploy.yaml"
    upgrade_file = repo_root / "framework/upgrade/.olares/Olares.yaml"

    appservice_yaml = read_text(appservice_file)
    sysevent_yaml = read_text(sysevent_file)
    upgrade_yaml = read_text(upgrade_file)

    app_service_image = extract_first(
        r"^\s*image:\s*(beclab/app-service:[^\s#]+)\s*$",
        appservice_yaml,
        "app-service image",
    )
    app_service_tag = extract_image_tag(app_service_image, "beclab/app-service")
    if app_service_tag != release_tag:
        fail(
            f"app-service image tag mismatch: expected '{release_tag}', got '{app_service_tag}'"
        )

    sys_event_image = extract_first(
        r"^\s*image:\s*(beclab/sys-event:[^\s#]+)\s*$",
        sysevent_yaml,
        "sys-event image",
    )
    sys_event_tag = extract_image_tag(sys_event_image, "beclab/sys-event")
    if sys_event_tag != release_tag:
        fail(
            f"sys-event image tag mismatch: expected '{release_tag}', got '{sys_event_tag}'"
        )

    d2_sidecar_image = extract_env_value(appservice_yaml, "D2_SIDECAR_IMAGE")
    if "@sha256:" not in d2_sidecar_image:
        fail(
            "D2_SIDECAR_IMAGE must be pinned with immutable digest '@sha256:' "
            f"(got '{d2_sidecar_image}')"
        )

    ws_image = extract_env_value(appservice_yaml, "WS_CONTAINER_IMAGE")
    upload_image = extract_env_value(appservice_yaml, "UPLOAD_CONTAINER_IMAGE")
    job_image_env = extract_env_value(appservice_yaml, "JOB_IMAGE")
    upgrade_job_image = extract_first(
        r"^\s*name:\s*(beclab/upgrade-job:[^\s#]+)\s*$",
        upgrade_yaml,
        "upgrade Olares.yaml output.containers[].name",
    )

    ensure_pinned_image(ws_image, "WS_CONTAINER_IMAGE")
    ensure_pinned_image(upload_image, "UPLOAD_CONTAINER_IMAGE")
    ensure_pinned_image(job_image_env, "JOB_IMAGE")

    if job_image_env != upgrade_job_image:
        fail(
            "JOB_IMAGE mismatch between appservice_deploy.yaml and "
            f"framework/upgrade/.olares/Olares.yaml: '{job_image_env}' != '{upgrade_job_image}'"
        )

    print("release image pin verification passed")
    print(f"  release tag: {release_tag}")
    print(f"  app-service: {app_service_image}")
    print(f"  sys-event:   {sys_event_image}")
    print(f"  d2-sidecar:  {d2_sidecar_image}")
    print(f"  ws-image:    {ws_image}")
    print(f"  upload-image:{upload_image}")
    print(f"  job-image:   {job_image_env}")


if __name__ == "__main__":
    main()
