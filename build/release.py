

import semantic_version
import subprocess
import sys

MAIN_BRANCH = 'main'
MAX_PATCH = 40
MAX_MINOR = 30


def new_version():
    command = ["git", "--no-pager", "tag", "-l"]

    r = subprocess.run(command, capture_output=True)
    if r.returncode != 0:
        print(r.stderr.decode())
        return

    tags = r.stdout.decode().strip().split('\n')
    print("version tags:", tags)

    if not tags:
        print('got none of the tags')
        return

    versions = []

    for tag in tags:
        try:
            v = semantic_version.Version(tag)
            versions.append(v)
        except ValueError:
            continue

    versions.sort()

    latestVer = versions[-1]
    print("got latest version:", latestVer)

    major, minor, patch = latestVer.major, latestVer.minor, latestVer.patch

    if patch >= MAX_PATCH:
        minor += 1
        patch = 0
    elif minor >= MAX_MINOR:
        major += 1
        minor = 0
    else:
        patch += 1

    return semantic_version.Version(major=major, minor=minor, patch=patch)


def release():
    v = new_version()

    if not v:
        print('no new version found')
        sys.exit(1)

    print("got a new version:", v)

    branch = "release-{}.{}.{}".format(v.major, v.minor, v.patch)
    tag = "{}.{}.{}-rc.0".format(v.major, v.minor, v.patch)

    commands = [
        ['git', 'checkout', MAIN_BRANCH],
        ['git', 'checkout', '-b', branch],
        ['git', 'push', 'origin', branch],

        ['git', 'tag', tag],
        ['git', 'push', 'origin', tag],
    ]

    for command in commands:
        r = subprocess.run(command, capture_output=True)
        if r.returncode != 0:
            print(r.stderr.decode())
            sys.exit(r.returncode)

    with open('/tmp/latest_tag_version', 'w') as f:
        f.write(tag)


if __name__ == '__main__':
    release()
