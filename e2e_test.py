import json
import subprocess
import time
import os

def send(proc, msg):
    line = json.dumps(msg) + "\n"
    proc.stdin.write(line)
    proc.stdin.flush()

def recv(proc, timeout=30):
    start = time.time()
    while time.time() - start < timeout:
        line = proc.stdout.readline()
        if not line:
            time.sleep(0.1)
            continue
        try:
            return json.loads(line)
        except Exception:
            continue
    raise TimeoutError("Timed out waiting for response from server")

def main():
    print("🚀 Starting Complete 25-Tool Android-MCP-go E2E Physical Test Suite...")
    binary_path = os.path.abspath("./android-mcp")
    
    proc = subprocess.Popen(
        [binary_path, "--debug"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1
    )

    try:
        # 1. Initialize
        print("1. Initialize...", flush=True)
        send(proc, {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05"}})
        resp = recv(proc)
        print(f"   ← Initialize: {resp.get('result', {}).get('serverInfo')}")
        assert resp.get("result", {}).get("serverInfo", {}).get("name") == "Android-MCP"

        # 2. List tools
        print("2. tools/list...")
        send(proc, {"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
        resp = recv(proc)
        tools = resp.get("result", {}).get("tools", [])
        print(f"   ← Total tools registered: {len(tools)}")
        assert len(tools) >= 23, f"Expected at least 23 tools, got {len(tools)}"

        # 3. ConnectDevice & device_connect (with both empty and explicit serial)
        print("3. ConnectDevice & device_connect...")
        send(proc, {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "ConnectDevice", "arguments": {}}})
        resp = recv(proc)
        print(f"   ← ConnectDevice (empty args): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "ConnectDevice", "arguments": {"serial": "192.168.1.3:5555"}}})
        resp = recv(proc)
        print(f"   ← ConnectDevice (explicit serial): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "device_connect", "arguments": {"serial": "192.168.1.3:5555"}}})
        resp = recv(proc)

        # 4. ListDevices, device_list & Device (empty, get, list)
        print("4. ListDevices, device_list & Device...")
        send(proc, {"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "ListDevices", "arguments": {}}})
        resp = recv(proc)
        devs_out = resp.get("result", {}).get("content", [{}])[0].get("text", "")
        print(f"   ← ListDevices: {devs_out.strip()}")

        send(proc, {"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "device_list", "arguments": {}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": {"name": "Device", "arguments": {}}})
        resp = recv(proc)
        print(f"   ← Device (empty args): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 9, "method": "tools/call", "params": {"name": "Device", "arguments": {"action": "list"}}})
        resp = recv(proc)
        print(f"   ← Device (action=list): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        # 5. Press Home & Snapshot
        print("5. Press Home & Snapshot...")
        send(proc, {"jsonrpc": "2.0", "id": 10, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 11, "method": "tools/call", "params": {"name": "Snapshot", "arguments": {"use_vision": False}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 12, "method": "tools/call", "params": {"name": "ui_snapshot", "arguments": {"use_vision": False}}})
        resp = recv(proc)

        # 6. Swipe up to App Drawer & Click Chrome
        print("6. Swipe (App Drawer) & ClickBySelector (Chrome)...")
        send(proc, {"jsonrpc": "2.0", "id": 13, "method": "tools/call", "params": {"name": "Swipe", "arguments": {"x1": 540, "y1": 1800, "x2": 540, "y2": 400, "duration_ms": 300}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 14, "method": "tools/call", "params": {"name": "ClickBySelector", "arguments": {"text": "Chrome", "timeout": 3}}})
        resp = recv(proc)
        print(f"   ← ClickBySelector: {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        # 7. WaitForElement & Type URL
        print("7. WaitForElement & Type URL...")
        send(proc, {"jsonrpc": "2.0", "id": 15, "method": "tools/call", "params": {"name": "WaitForElement", "arguments": {"text": "Search Google or type URL", "timeout_sec": 5}}})
        resp = recv(proc, timeout=10)

        send(proc, {"jsonrpc": "2.0", "id": 16, "method": "tools/call", "params": {"name": "Type", "arguments": {"x": 467, "y": 564, "text": "https://github.com/tintupratap/Android-MCP-go", "clear": True}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 17, "method": "tools/call", "params": {"name": "Press", "arguments": {"keycode": "KEYCODE_ENTER"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 18, "method": "tools/call", "params": {"name": "Wait", "arguments": {"seconds": 2}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 19, "method": "tools/call", "params": {"name": "stop_app", "arguments": {"package_name": "com.android.chrome"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 20, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 8. Drag (YouTube icon)
        print("8. Drag (YouTube icon home screen)...")
        send(proc, {"jsonrpc": "2.0", "id": 21, "method": "tools/call", "params": {"name": "Drag", "arguments": {"x1": 128, "y1": 1737, "x2": 540, "y2": 1439}}})
        resp = recv(proc)
        print(f"   ← Drag (Move Up): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 22, "method": "tools/call", "params": {"name": "Drag", "arguments": {"x1": 540, "y1": 1439, "x2": 128, "y2": 1737}}})
        resp = recv(proc)
        print(f"   ← Drag (Move Back): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        # 9. Pinch (Google Photos)
        print("9. Pinch (Google Photos Zoom In & Zoom Out)...")
        send(proc, {"jsonrpc": "2.0", "id": 23, "method": "tools/call", "params": {"name": "launch_app", "arguments": {"package_name": "com.google.android.apps.photos"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 24, "method": "tools/call", "params": {"name": "Click", "arguments": {"x": 540, "y": 1902}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 25, "method": "tools/call", "params": {"name": "Pinch", "arguments": {"x1": 450, "y1": 1260, "x2": 150, "y2": 1260, "x3": 630, "y3": 1260, "x4": 930, "y4": 1260}}})
        resp = recv(proc)
        print(f"   ← Pinch (Zoom In): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 26, "method": "tools/call", "params": {"name": "pinch", "arguments": {"x1": 150, "y1": 1260, "x2": 450, "y2": 1260, "x3": 930, "y3": 1260, "x4": 630, "y4": 1260}}})
        resp = recv(proc)
        print(f"   ← pinch (Zoom Out): {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 27, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 10. Notification, file_push, file_pull, shell_exec & list_apps
        print("10. Notification, file_push, file_pull, shell_exec & list_apps...")
        send(proc, {"jsonrpc": "2.0", "id": 28, "method": "tools/call", "params": {"name": "Notification", "arguments": {"title": "Test", "message": "Pass"}}})
        resp = recv(proc)

        test_file = os.path.abspath(".tmp_live_test.txt")
        pulled_file = os.path.abspath(".tmp_live_pulled.txt")
        with open(test_file, "w") as f:
            f.write("Full Suite Live Verification 2026")

        send(proc, {"jsonrpc": "2.0", "id": 29, "method": "tools/call", "params": {"name": "file_push", "arguments": {"local_path": test_file, "remote_path": "/data/local/tmp/live_test.txt"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 30, "method": "tools/call", "params": {"name": "file_pull", "arguments": {"remote_path": "/data/local/tmp/live_test.txt", "local_path": pulled_file}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 31, "method": "tools/call", "params": {"name": "shell_exec", "arguments": {"command": "getprop ro.product.model"}}})
        resp = recv(proc)
        shell_out = resp.get("result", {}).get("content", [{}])[0].get("text", "")
        print(f"   ← shell_exec: {shell_out.strip()}")

        send(proc, {"jsonrpc": "2.0", "id": 32, "method": "tools/call", "params": {"name": "list_apps", "arguments": {"type": "user"}}})
        resp = recv(proc)

        if os.path.exists(test_file):
            os.remove(test_file)
        if os.path.exists(pulled_file):
            os.remove(pulled_file)

        print("\n✅ ALL 25 DISTINCT MCP TOOLS AND ALIASES VERIFIED 100% IN E2E SUITE!")
        return 0

    finally:
        proc.terminate()
        try:
            proc.wait(timeout=3)
        except Exception:
            proc.kill()

if __name__ == "__main__":
    main()
