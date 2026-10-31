// Render documentation layouts from the original screenshots; application UI is unchanged.
// Run after `npm.cmd --prefix web ci`: node scripts/build-readme-visuals.mjs
import { chromium } from '../web/node_modules/playwright-core/index.mjs';
import { mkdirSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const output = path.join(root, 'docs/screenshots');
mkdirSync(output, { recursive: true });
const web = (time, label, crop = [280, 80, 1610, 810]) => {
  const folder = 'web-arayuz-gorsel';
  const name = readdirSync(path.join(root, folder)).find(name => name.endsWith(` ${time}.png`));
  if (!name) throw new Error(`Missing web screenshot: ${time}`);
  return { file: `${folder}/${name}`, label, crop };
};
const mobile = (name, label) => ({ file: `mobil-arayuz-gorsel/WhatsApp Image 2026-11-01 at ${name}.jpeg`, label });
const layouts = [
  { name: 'web-yapay-zeka', title: 'Yapay Zekâ Merkezi', subtitle: 'Üretim tahmini · Parçalama tahmini · Model ve gerçekleşen kayıt karşılaştırması', columns: 1, screenshots: [
    web('012720', '1 · Beklenen mamul, verim, tahmini EKP ve parça ağırlıkları', [296, 16, 1574, 780]),
    web('012730', '2 · Üretim tahmini ve gerçekleşen miktar aynı grafikte', [296, 76, 1574, 610]),
    web('012738', '3 · Parçalama sapmaları ve model hata metrikleri', [296, 16, 1574, 854]),
  ] },
  { name: 'web-stok-uretim', title: 'Stoktan üretim emrine', subtitle: 'Kartlar, lotlar ve stok girişleri · Üretim kayıtları ve maliyet dökümü', columns: 2, screenshots: [
    web('012446', 'Hammadde, lot ve stok girişi'),
    web('012536', 'Üretim emirleri ve gerçek maliyet', [280, 0, 1610, 890]),
  ] },
  { name: 'web-parcalama-satis', title: 'Parçalamadan satışa', subtitle: 'Giriş partisi ve parça ağırlıkları · Mamul lotundan markete satış', columns: 2, screenshots: [
    web('012545', 'Parçalama kaydı ve maliyet dağılımı'),
    web('012610', 'Lot ve market seçerek satış'),
  ] },
  { name: 'web-lot-geri-cagirma', title: 'Lot izlenebilirliği ve geri çağırma', subtitle: 'Kaynak lotu izle · Etkilenen ürünleri ve alıcıları bul · Satış blokajını uygula', columns: 2, screenshots: [
    web('012625', 'İleri izleme: hammaddeden marketlere'),
    web('012703', 'Geri çağırma: alıcı listesi ve satış blokajı', [280, 50, 1610, 820]),
  ] },
  { name: 'web-rapor-ocr', title: 'Raporlama ve belge metni', subtitle: 'Kayıtları raporla ve dışa aktar · Belge kalemlerini kontrol ederek stoğa işle', columns: 2, screenshots: [
    web('012759', 'Alış raporu, diğer rapor sekmeleri ve PDF/CSV'),
    web('013147', 'Fatura metninden çıkarılan ve onaylanan kalem'),
  ] },
  { name: 'mobil-stok-senkron', title: 'Mobil operasyon ve senkronizasyon', subtitle: 'Sunucu bağlantısı · Ürün ve stok sorgulama · Satış geçmişi', columns: 3, screenshots: [
    mobile('01.04.07', 'Ana sayfa ve sunucu bağlantısı'),
    mobile('01.04.12 (2)', 'Ürünler ve stok görünümü'),
    mobile('01.04.12 (1)', 'Geçmiş satışlar ve senkronizasyon'),
  ] },
  { name: 'mobil-satis-blokaj', title: 'Bloklu lotlar satılamaz', subtitle: 'Uygun lotun satışı senkronize edilir · Bloklu lotta işlem reddedilir ve uyarı gösterilir', columns: 2, screenshots: [
    mobile('01.06.08', 'Uygun lot: Satış senkronize edildi.'),
    mobile('01.05.17', 'Bloklu lot: satış reddedildi.'),
  ] },
  { name: 'mobil-tahmin-rapor', title: 'Mobil tahminler ve PDF raporu', subtitle: 'Kuzu/sığır parça tahminleri · Merkezi kayıtlardan rapor · Cihaza PDF indirme', columns: 2, screenshots: [
    mobile('01.04.08 (2)', 'Parçalama tahmini ve fiziksel tutarlılık'),
    mobile('01.04.11', 'Rapor tablosu ve PDF indirme bildirimi'),
  ] },
];
const escape = value => value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('"', '&quot;');
const browser = await chromium.launch({ executablePath: process.env.CHROME_PATH ?? 'C:/Program Files/Google/Chrome/Application/chrome.exe', headless: true });
try {
  const page = await browser.newPage({ viewport: { width: 1800, height: 1000 }, deviceScaleFactor: 1 });
  for (const layout of layouts) {
    const blocks = layout.screenshots.map(item => {
      const type = item.file.endsWith('.png') ? 'png' : 'jpeg';
      const source = readFileSync(path.join(root, item.file));
      const data = `data:image/${type};base64,${source.toString('base64')}`;
      const [x, y, width, height] = item.crop ?? [0, 0, 0, 0];
      const naturalWidth = item.crop ? source.readUInt32BE(16) : 0;
      const naturalHeight = item.crop ? source.readUInt32BE(20) : 0;
      if (item.crop && (x + width > naturalWidth || y + height > naturalHeight)) throw new Error(`Crop exceeds source: ${item.file}`);
      const visual = item.crop
        ? `<svg viewBox="${x} ${y} ${width} ${height}" role="img"><image href="${data}" width="${naturalWidth}" height="${naturalHeight}" /></svg>`
        : `<img src="${data}" alt="${escape(item.label)}">`;
      return `<article><div class="label">${escape(item.label)}</div>${visual}</article>`;
    }).join('');
    await page.setContent(`<!doctype html><html lang="tr"><meta charset="utf-8"><style>
      *{box-sizing:border-box}body{margin:0;background:#f5f7f4;color:#183f35;font-family:Arial,sans-serif}
      main{padding:32px}header{padding:26px 30px;background:#183f35;color:white;border-radius:18px;margin-bottom:24px}
      .tag{font-size:17px;letter-spacing:2px;color:#c1dfd1;margin-bottom:12px}h1{font-size:38px;margin:0 0 12px}p{font-size:22px;line-height:1.4;margin:0;color:#dfebe5}
      .grid{display:grid;grid-template-columns:repeat(${layout.columns},minmax(0,1fr));gap:22px;align-items:start}
      article{overflow:hidden;border:1px solid #d5dfd8;border-radius:14px;background:white}.label{padding:18px 20px;font-size:22px;font-weight:bold;border-bottom:1px solid #d5dfd8;min-height:64px}
      svg{display:block;width:100%;height:auto}article img{display:block;width:100%;height:auto}
    </style><main><header><div class="tag">AVATAR · ${layout.name.startsWith('web') ? 'WEB' : 'ANDROID'}</div><h1>${escape(layout.title)}</h1><p>${escape(layout.subtitle)}</p></header><section class="grid">${blocks}</section></main></html>`);
    await page.evaluate(async () => {
      await document.fonts.ready;
      await Promise.all([...document.images].map(image => image.decode()));
      await Promise.all([...document.querySelectorAll('svg image')].map(image => new Promise((resolve, reject) => {
        const check = new Image(); check.onload = resolve; check.onerror = reject; check.src = image.getAttribute('href');
      })));
    });
    await page.locator('main').screenshot({ path: path.join(output, `${layout.name}.png`) });
    console.log(`Rendered ${layout.name}: ${layout.screenshots.length} original screenshots`);
  }
} finally { await browser.close(); }
