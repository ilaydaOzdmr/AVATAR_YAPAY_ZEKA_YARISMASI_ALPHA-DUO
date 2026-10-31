from functools import partial
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import os
import subprocess
import threading

root=Path(__file__).resolve().parents[2]
directory=root/'coverage/web/app'
class QuietHandler(SimpleHTTPRequestHandler):
    def log_message(self,*args): pass
server=ThreadingHTTPServer(('127.0.0.1',0),partial(QuietHandler,directory=str(directory)))
threading.Thread(target=server.serve_forever,daemon=True).start()
try:
    env=os.environ.copy()
    env['QA_BASE_URL']=f'http://localhost:{server.server_port}/m3.html'
    env['WEB_COVERAGE_FILE']=str(root/'coverage/web/browser-v8.json')
    subprocess.run(['node','web/tests/browserCoverage.mjs'],cwd=root,env=env,check=True)
finally:
    server.shutdown()
    server.server_close()
