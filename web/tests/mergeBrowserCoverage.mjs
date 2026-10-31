import fs from 'node:fs';
import path from 'node:path';
import v8ToIstanbul from 'v8-to-istanbul';
import libCoverage from 'istanbul-lib-coverage';
import libReport from 'istanbul-lib-report';
import reports from 'istanbul-reports';

const root = path.resolve(import.meta.dirname, "../..");
const output = path.join(root, 'coverage/web');
const map = libCoverage.createCoverageMap(JSON.parse(fs.readFileSync(path.join(output, 'coverage-final.json'), 'utf8')));
const entries = JSON.parse(fs.readFileSync(path.join(output, 'browser-v8.json'), 'utf8'));
for (const entry of entries) {
  const name = path.basename(new URL(entry.url).pathname);
  const bundle = path.join(root, 'coverage/web/app/assets', name);
  if (!fs.existsSync(`${bundle}.map`)) continue;
  const sourceMap = { sourcemap: JSON.parse(fs.readFileSync(`${bundle}.map`, 'utf8')) };
  const converter = v8ToIstanbul(bundle, 0, { source: entry.source, sourceMap });
  await converter.load();
  converter.applyCoverage(entry.functions);
  for (const data of Object.values(converter.toIstanbul())) {
    data.path = path.resolve(data.path);
    if (data.path.startsWith(path.join(root, 'web/src') + path.sep)) map.addFileCoverage(data);
  }
}
const context = libReport.createContext({ dir: output, coverageMap: map });
reports.create('lcovonly', { projectRoot: root }).execute(context);
reports.create('json', { file: 'coverage-combined.json' }).execute(context);
reports.create('text-summary').execute(context);
