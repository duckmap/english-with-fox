#!/usr/bin/env python3
"""Serve the prebuilt browser game on loopback only; no third-party modules."""
import http.server,os,webbrowser
from pathlib import Path
os.chdir(Path(__file__).resolve().parent/'web')
class Handler(http.server.SimpleHTTPRequestHandler):
 extensions_map={**http.server.SimpleHTTPRequestHandler.extensions_map,'.wasm':'application/wasm','.js':'text/javascript'}
print('Откройте http://localhost:8080 — Ctrl+C для остановки')
server=http.server.ThreadingHTTPServer(('127.0.0.1',8080),Handler)
webbrowser.open('http://localhost:8080')
try: server.serve_forever()
except KeyboardInterrupt: pass
finally: server.server_close()
