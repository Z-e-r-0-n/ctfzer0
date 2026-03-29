#!/usr/bin/env python3

import os
import subprocess

BASE_DIR = "/etc/net"
INT_DIR = BASE_DIR + "/interfaces"
GLOBAL_CONF = BASE_DIR + "/global.conf"

def list_int():
    try:
        return os.listdir(INT_DIR)
    except FileNotFoundError:
        print("directory not accessible or missing:", INT_DIR)
        return []

def parse_files(path):
    intent = {}
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" not in line:
                continue
            key, value = line.split("=", 1)
            intent[key.strip()] = value.strip()
    return intent

def read_global():
    intent = {}
    if not os.path.exists(GLOBAL_CONF):
        return intent
    with open(GLOBAL_CONF) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" not in line:
                continue
            k, v = line.split("=", 1)
            intent[k.strip()] = v.strip()
    return intent

ALLOWED_ROLES = {"uplink", "downlink", "isolated", "monitor"}

def validate_interfaces(interfaces):
    uplinks = []
    for iface, data in interfaces.items():
        if "role" not in data:
            raise RuntimeError(f"{iface}: missing role")

        role = data["role"]
        if role not in ALLOWED_ROLES:
            raise RuntimeError(f"{iface}: invalid role '{role}'")

        if role == "uplink":
            uplinks.append(iface)

    if len(uplinks) != 1:
        raise RuntimeError(
            f"exactly one uplink required, found {len(uplinks)}"
        )

    return uplinks[0]

def system_interfaces():
    result = subprocess.run(
        ["ip", "-o", "link", "show"],
        capture_output=True,
        text=True,
        check=True
    )

    interfaces = set()
    for line in result.stdout.splitlines():
        name = line.split(":")[1].strip()
        interfaces.add(name)

    return interfaces

def check_intent_interfaces_exist(interfaces):
    system_ifaces = system_interfaces()
    for iface in interfaces:
        if iface not in system_ifaces:
            print(f"WARNING: interface '{iface}' declared but not present")

def interface_state(iface):
    result = subprocess.run(
        ["ip", "link", "show", iface],
        capture_output=True,
        text=True
    )

    if result.returncode != 0:
        return "missing"
    if "state UP" in result.stdout:
        return "up"
    return "down"

# ------------------ main flow ------------------

interfaces = {}
files = list_int()

for fname in files:
    if not fname.endswith(".conf"):
        continue

    iface = fname.replace(".conf", "")
    path = os.path.join(INT_DIR, fname)
    interfaces[iface] = parse_files(path)

global_intent = read_global()

print("Global intent:")
print(global_intent)

print("\nInterface intent:")
for iface, data in interfaces.items():
    print(iface, "->", data)

try:
    active_uplink = validate_interfaces(interfaces)
    print("\nActive uplink:", active_uplink)
except RuntimeError as e:
    print("CONFIG ERROR:", e)
    exit(1)

check_intent_interfaces_exist(interfaces)

state = interface_state(active_uplink)
print(f"Uplink {active_uplink} state:", state)
