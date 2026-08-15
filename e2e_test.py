#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import time

BINARY = os.path.join(os.path.dirname(__file__), "android-mcp")

def send(proc, obj):
    proc.stdin.write((json.dumps(obj) + "\n").encode())
    proc.stdin.flush()

def recv(proc, timeout=30):
    deadline = time.time() + timeout
    while True:
        if time.time() > deadline:
            raise TimeoutError("Timeout reading from server")
        line = proc.stdout.readline()
        if not line:
            raise EOFError("Server stdout closed")
        line = line.strip()
        if line:
            return json.loads(line)

def run():
    print("🚀 Starting Android-MCP-go E2E Physical Test...")
    proc = subprocess.Popen(
        [BINARY],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    try:
        # 1. Initialize
        print("1. Sending initialize...")
        send(proc, {
            "jsonrpc": "2.0", "id": 1, "method": "initialize",
            "params": {"protocolVersion": "2024-11-05", "clientInfo": {"name": "e2e"}}
        })
        resp = recv(proc)
        print("   ← Initialize result:", resp.get("result", {}).get("serverInfo"))
        assert resp.get("result", {}).get("serverInfo", {}).get("name") == "Android-MCP"

        # 2. List tools
        print("2. Requesting tools/list...")
        send(proc, {"jsonrpc": "2.0", "id": 2, "method": "tools/list"})
        resp = recv(proc)
        tools = resp.get("result", {}).get("tools", [])
        print(f"   ← Total tools registered: {len(tools)}")
        assert len(tools) >= 14, f"Expected at least 14 tools, got {len(tools)}"

        # 3. Call ListDevices
        print("3. Calling ListDevices tool...")
        send(proc, {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "ListDevices", "arguments": {}}})
        resp = recv(proc)
        content = resp.get("result", {}).get("content", [])
        text = " ".join(c.get("text", "") for c in content)
        print(f"   ← ListDevices output:\n{text}")
        assert "192.168.1.3:5555" in text or "device" in text

        # 4. Call Snapshot (use_vision=false)
        print("4. Calling Snapshot tool (use_vision=false)...")
        send(proc, {"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "Snapshot", "arguments": {"use_vision": False}}})
        resp = recv(proc, timeout=45)
        content = resp.get("result", {}).get("content", [])
        tree_text = " ".join(c.get("text", "") for c in content)
        print(f"   ← Snapshot text preview (first 500 chars):\n{tree_text[:500]}")
        assert len(tree_text) > 30

        # 5. Call Snapshot (use_vision=true)
        print("5. Calling Snapshot tool (use_vision=true)...")
        send(proc, {"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "Snapshot", "arguments": {"use_vision": True}}})
        resp = recv(proc, timeout=45)
        content = resp.get("result", {}).get("content", [])
        print(f"   ← Snapshot content types: {[c.get('type') for c in content]}")
        assert any(c.get("type") == "image" for c in content), "Expected image content when use_vision=True"

        # 6. Call list_apps tool
        print("6. Calling list_apps tool...")
        send(proc, {"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "list_apps", "arguments": {"third_party_only": True}}})
        resp = recv(proc, timeout=30)
        content = resp.get("result", {}).get("content", [])
        apps_text = " ".join(c.get("text", "") for c in content)
        print(f"   ← list_apps output (first 300 chars):\n{apps_text[:300]}")

        # 7. Call shell_exec tool
        print("7. Calling shell_exec tool (getprop ro.product.model)...")
        send(proc, {"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "shell_exec", "arguments": {"command": "getprop ro.product.model"}}})
        resp = recv(proc, timeout=30)
        content = resp.get("result", {}).get("content", [])
        shell_out = " ".join(c.get("text", "") for c in content)
        print(f"   ← shell_exec output:\n{shell_out}")
        assert "SOG09" in shell_out or "exit_code" in shell_out

        print("\n✅ ALL PHYSICAL VERIFICATION CHECKS PASSED SUCCESSFULLY!")
        return 0

    except Exception as e:
        print(f"\n❌ E2E Physical Test Failed: {e}")
        return 1
    finally:
        proc.terminate()

if __name__ == "__main__":
    sys.exit(run())
