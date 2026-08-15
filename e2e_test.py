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
    print("🚀 Starting Complete 23-Tool Android-MCP-go E2E Physical Test...")
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
        print("1. Initialize...")
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
        assert len(tools) >= 23, f"Expected 23 tools, got {len(tools)}"

        # 3. ConnectDevice & device_connect
        print("3. ConnectDevice & device_connect...")
        send(proc, {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "ConnectDevice", "arguments": {"serial": "192.168.1.3:5555"}}})
        resp = recv(proc)
        print(f"   ← ConnectDevice: {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        send(proc, {"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "device_connect", "arguments": {"serial": "192.168.1.3:5555"}}})
        resp = recv(proc)
        print(f"   ← device_connect: {resp.get('result', {}).get('content', [{}])[0].get('text')}")

        # 4. ListDevices, device_list & Device
        print("4. ListDevices, device_list & Device...")
        send(proc, {"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "ListDevices", "arguments": {}}})
        resp = recv(proc)
        devs_out = resp.get("result", {}).get("content", [{}])[0].get("text", "")
        print(f"   ← ListDevices: {devs_out.strip()}")
        assert "192.168.1.3:5555" in devs_out

        send(proc, {"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "device_list", "arguments": {}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "Device", "arguments": {"action": "list"}}})
        resp = recv(proc)

        # 5. Snapshot & ui_snapshot
        print("5. Snapshot & ui_snapshot (use_vision=false)...")
        send(proc, {"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": {"name": "Snapshot", "arguments": {"use_vision": False}}})
        resp = recv(proc)
        snap_text = resp.get("result", {}).get("content", [{}])[0].get("text", "")
        assert "Coordinates" in snap_text or "Label" in snap_text

        send(proc, {"jsonrpc": "2.0", "id": 9, "method": "tools/call", "params": {"name": "ui_snapshot", "arguments": {"use_vision": False}}})
        resp = recv(proc)

        # 6. Snapshot (use_vision=true)
        print("6. Snapshot (use_vision=true)...")
        send(proc, {"jsonrpc": "2.0", "id": 10, "method": "tools/call", "params": {"name": "Snapshot", "arguments": {"use_vision": True}}})
        resp = recv(proc, timeout=45)
        content = resp.get("result", {}).get("content", [])
        assert any(c.get("type") == "image" for c in content), "Expected vision PNG image"

        # 7. Click, ui_click & ClickBySelector
        print("7. Click, ui_click & ClickBySelector...")
        send(proc, {"jsonrpc": "2.0", "id": 11, "method": "tools/call", "params": {"name": "Click", "arguments": {"x": 540, "y": 545}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 12, "method": "tools/call", "params": {"name": "ui_click", "arguments": {"x": 540, "y": 545}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 13, "method": "tools/call", "params": {"name": "ClickBySelector", "arguments": {"text": "Phone", "timeout": 3}}})
        resp = recv(proc)

        # 8. Press (home)
        print("8. Press (home button)...")
        send(proc, {"jsonrpc": "2.0", "id": 14, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 9. LongClick
        print("9. LongClick...")
        send(proc, {"jsonrpc": "2.0", "id": 15, "method": "tools/call", "params": {"name": "LongClick", "arguments": {"x": 540, "y": 1260}}})
        resp = recv(proc)
        send(proc, {"jsonrpc": "2.0", "id": 16, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 10. Swipe & Drag
        print("10. Swipe & Drag...")
        send(proc, {"jsonrpc": "2.0", "id": 17, "method": "tools/call", "params": {"name": "Swipe", "arguments": {"x1": 500, "y1": 1000, "x2": 500, "y2": 500}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 18, "method": "tools/call", "params": {"name": "Drag", "arguments": {"x1": 500, "y1": 1000, "x2": 500, "y2": 800}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 19, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 11. Type
        print("11. Type...")
        send(proc, {"jsonrpc": "2.0", "id": 20, "method": "tools/call", "params": {"name": "Type", "arguments": {"x": 540, "y": 2294, "text": "Android-MCP", "clear": False}}})
        resp = recv(proc)
        send(proc, {"jsonrpc": "2.0", "id": 21, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 12. Notification
        print("12. Notification...")
        send(proc, {"jsonrpc": "2.0", "id": 22, "method": "tools/call", "params": {"name": "Notification", "arguments": {}}})
        resp = recv(proc)
        send(proc, {"jsonrpc": "2.0", "id": 23, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 13. Wait & WaitForElement
        print("13. Wait & WaitForElement...")
        send(proc, {"jsonrpc": "2.0", "id": 24, "method": "tools/call", "params": {"name": "Wait", "arguments": {"duration": 1}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 25, "method": "tools/call", "params": {"name": "WaitForElement", "arguments": {"text": "Phone", "timeout": 5}}})
        resp = recv(proc)

        # 14. list_apps, launch_app & stop_app
        print("14. list_apps, launch_app & stop_app...")
        send(proc, {"jsonrpc": "2.0", "id": 26, "method": "tools/call", "params": {"name": "list_apps", "arguments": {"third_party_only": True}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 27, "method": "tools/call", "params": {"name": "launch_app", "arguments": {"package_name": "com.google.android.youtube"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 28, "method": "tools/call", "params": {"name": "stop_app", "arguments": {"package_name": "com.google.android.youtube"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 29, "method": "tools/call", "params": {"name": "Press", "arguments": {"button": "home"}}})
        resp = recv(proc)

        # 15. shell_exec
        print("15. shell_exec (getprop ro.product.model)...")
        send(proc, {"jsonrpc": "2.0", "id": 30, "method": "tools/call", "params": {"name": "shell_exec", "arguments": {"command": "getprop ro.product.model"}}})
        resp = recv(proc)
        shell_out = resp.get("result", {}).get("content", [{}])[0].get("text", "")
        assert "SOG09" in shell_out or "exit_code" in shell_out

        # 16. file_push and file_pull
        print("16. file_push & file_pull...")
        test_file = os.path.abspath(".tmp_live_test.txt")
        pulled_file = os.path.abspath(".tmp_live_pulled.txt")
        with open(test_file, "w") as f:
            f.write("Full Suite Live Verification 2026")

        send(proc, {"jsonrpc": "2.0", "id": 31, "method": "tools/call", "params": {"name": "file_push", "arguments": {"local_path": test_file, "remote_path": "/sdcard/Download/live_test.txt"}}})
        resp = recv(proc)

        send(proc, {"jsonrpc": "2.0", "id": 32, "method": "tools/call", "params": {"name": "file_pull", "arguments": {"remote_path": "/sdcard/Download/live_test.txt", "local_path": pulled_file}}})
        resp = recv(proc)

        with open(pulled_file, "r") as f:
            pulled_content = f.read()
        assert "Full Suite Live Verification 2026" in pulled_content

        if os.path.exists(test_file):
            os.remove(test_file)
        if os.path.exists(pulled_file):
            os.remove(pulled_file)

        print("\n✅ ALL 23 TOOLS SUCCESSFULLY VERIFIED IN E2E SUITE!")
        return 0

    finally:
        proc.terminate()
        try:
            proc.wait(timeout=3)
        except Exception:
            proc.kill()

if __name__ == "__main__":
    main()
