import datetime
import json
import os
from pathlib import Path
import re
import uuid
import subprocess

root = Path(__file__).resolve().parents[2]
report = root / 'coverage/go'
report.mkdir(parents=True, exist_ok=True)
config = json.loads(subprocess.check_output(['docker', 'inspect', 'avatar-mysql'], text=True))[0]
variables = dict(value.split('=', 1) for value in config['Config']['Env'] if '=' in value)
password = variables['MYSQL_ROOT_PASSWORD']
database = 'avatar_sonar_test_' + datetime.datetime.now().strftime('%Y%m%d_%H%M%S') + '_' + uuid.uuid4().hex[:8]
assert re.fullmatch(r'avatar_sonar_test_\d{8}_\d{6}_[0-9a-f]{8}', database)
mysql = ['docker', 'exec', '-i', 'avatar-mysql', 'sh', '-c',
         'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" exec mysql -uroot --default-character-set=utf8mb4']
schema = '\n'.join(path.read_text(encoding='utf-8') for path in sorted((root / 'backend/migrations').glob('*.up.sql')))
subprocess.run(mysql, input=f'CREATE DATABASE `{database}`; USE `{database}`;\n' + schema,
    text=True, encoding='utf-8', check=True, capture_output=True)
try:
    port = config['NetworkSettings']['Ports']['3306/tcp'][0]['HostPort']
    env = os.environ.copy()
    base = f'root:{password}@tcp(127.0.0.1:{port})/'
    env['TEST_MYSQL_DSN'] = base + database + '?parseTime=true&charset=utf8mb4'
    env['TEST_MYSQL_ROOT_DSN'] = base
    env['AVATAR_ML_TEST_URL'] = 'http://localhost:8000'
    command = ['go', 'test', '-tags=integration', '-p', '1', '-count=1',
        '-coverprofile=' + str(report / 'coverage.out'), '-covermode=atomic', './...']
    process = subprocess.run(command, cwd=root / 'backend', env=env, capture_output=True, text=True)
    log = (process.stdout + process.stderr).replace(password, '[REDACTED]') if len(password) > 6 else process.stdout + process.stderr
    (report / 'tests.log').write_text(log, encoding='utf-8')
    print(log)
    raise SystemExit(process.returncode)
finally:
    assert database.startswith('avatar_sonar_test_')
    subprocess.run(mysql, input=f'DROP DATABASE `{database}`;', text=True, check=True, capture_output=True)
    print('Isolated test database removed; application database unchanged.')
