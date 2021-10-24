#!/usr/bin/python3

import os

import requests

DEBUG = False
REPO_ROOT = "/usr/share/caddy/"
PLATFORMS = ["arista", "cisco", "juniper", "mikrotik"]


def mkdir(path):
    if not os.path.exists(path):
        os.makedirs(path)


def from_platform(name):
    for platform in PLATFORMS:
        if platform in name:
            return platform
        elif "deb" in name:
            return "apt"
        elif "rpm" in name:
            return "yum"
    return None


# Create platform directories
mkdir(REPO_ROOT)
for platform in PLATFORMS + ["apt", "yum"]:
    mkdir(os.path.join(REPO_ROOT, platform))


def sys_exec(command):
    if DEBUG:
        print(command)
    else:
        os.system(command)


# Download assets
release = requests.get("https://api.github.com/repos/natesales/pathvector/releases/latest").json()
for asset in release["assets"]:
    plat = from_platform(asset["name"])
    if plat is not None:  # Skip files that don't contain a platform identifier
        local_asset_path = os.path.join(REPO_ROOT, plat, asset['name'])
        if not os.path.exists(local_asset_path):
            # Download
            print(f"Downloading {asset['name']}")
            r = requests.get(asset['browser_download_url'])
            with open(local_asset_path, "wb") as output_file:
                output_file.write(r.content)

        # Platform specifics
        if plat == "apt":
            sys_exec(f"reprepro --keepunreferencedfiles -b {REPO_ROOT}/apt/ includedeb stable {local_asset_path}")
        elif plat == "yum":
            sys_exec(f"cp {local_asset_path} {REPO_ROOT}/yum/Packages/")

        # Add external signature file
        sys_exec(f"gpg --batch --yes --detach-sign --armor {local_asset_path}")

sys_exec(f"rpm --addsign {REPO_ROOT}/yum/Packages/*rpm")
sys_exec(f"createrepo {REPO_ROOT}/yum/")
sys_exec(f"gpg --batch --yes --detach-sign --armor {REPO_ROOT}/yum/repodata/repomd.xml")
