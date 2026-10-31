import { chromium } from "playwright-core";
import assert from "node:assert/strict";
import path from "node:path";
import { writeFileSync, mkdirSync } from "node:fs";

const artifactDirectory = path.resolve(import.meta.dirname, "../../coverage/web/screenshots");
const browser = await chromium.launch({ executablePath: process.env.CHROME_PATH ?? "C:/Program Files/Google/Chrome/Application/chrome.exe", headless: true });
mkdirSync(artifactDirectory, { recursive: true });
const errors = [], posts = [];
try {
  const page = await browser.newPage({ viewport: { width: 1534, height: 1000 } });
  page.on("pageerror", (error) => errors.push(error.message));
  page.on("request", (request) => { if (request.method() === "POST") posts.push(new URL(request.url()).pathname); });
  if (process.env.WEB_COVERAGE_FILE) await page.coverage.startJSCoverage({ resetOnNavigation: false });
  await page.goto(process.env.QA_BASE_URL ?? "http://localhost:5174", { waitUntil: "networkidle" });
  await page.getByRole("button", { name: /Yapay Zekâ/ }).click();
  const heading = page.getByRole("heading", { name: "Model tahmini vs Gerçek", exact: true });
  await heading.waitFor();
  await page.getByRole("heading", { name: "Parçalama tahmini vs Gerçek", exact: true }).waitFor();
  await page.getByRole("heading", { name: "Üretim modeli doğruluk", exact: true }).waitFor();
  await page.waitForLoadState("networkidle");
  const history = page.locator(".panel").filter({ has: heading });
  assert.deepEqual(await history.locator("thead th").allTextContents(), ["Üretim emri", "Model tahmini", "Gerçekleşen", "Sapma"]);
  assert.equal(await page.getByRole("columnheader", { name: "Plan", exact: true }).count(), 0);
  assert.equal(await page.getByText(/Model durumu:/).count(), 0);
  const historyCount = await history.locator("tbody tr").count();
  assert.ok(historyCount > 0, "Seeded local backend is required for integration tests");
  const chart = page.locator(".forecast-chart");
  assert.equal(await chart.locator(".forecast-chart__bar--predicted").count(), historyCount);
  assert.equal(await chart.locator(".forecast-chart__bar--actual").count(), historyCount);
  assert.equal(await chart.getByText("Miktar (kg)", { exact: true }).count(), 1);
  const groups = chart.locator(".forecast-chart__group");
  await groups.first().focus();
  await page.waitForFunction(() => document.querySelector(".forecast-chart__detail").textContent.includes("URT-2026-045"));
  await groups.nth(1).hover();
  await page.waitForFunction(() => document.querySelector(".forecast-chart__detail").textContent.includes("URT-2026-046"));
  const colors = await chart.evaluate((element) => [".forecast-chart__bar--predicted", ".forecast-chart__bar--actual"].map((selector) => getComputedStyle(element.querySelector(selector)).fill));
  assert.notEqual(colors[0], colors[1]);
  await heading.hover();
  await history.screenshot({ path: path.join(artifactDirectory, "clean-comparison-desktop.png") });
  const forms = page.locator("form.inline-form");
  assert.equal(await forms.nth(0).locator("input").count(), 5);
  assert.equal(await forms.nth(1).locator("input").count(), 2);
  const yieldResponse = page.waitForResponse((response) => response.url().endsWith("/ml/predict/v2/yield") && response.request().method() === "POST");
  await forms.nth(0).getByRole("button", { name: "Tahmin et", exact: true }).click();
  const production = await (await yieldResponse).json();
  assert.equal(production.model_version, "1.1.0");
  await page.getByText(`${production.prediction.finished_qty} kg`, { exact: true }).waitFor();
  const cutResponse = page.waitForResponse((response) => response.url().endsWith("/ml/predict/v2/butchery") && response.request().method() === "POST");
  await forms.nth(1).getByRole("button", { name: "Tahmin et", exact: true }).click();
  const cut = await (await cutResponse).json();
  assert.equal(cut.model_version, "1.1.0");
  assert.equal(cut.physical_check.parts_sum_kg, 21);
  await page.getByText("Karkas: 21 kg · Parça toplamı: 21 kg · Fiziksel kontrol tutarlı.", { exact: false }).waitFor();
  const body = await page.locator("body").innerText();
  assert.ok(!body.includes("Model sürümü:"));
  assert.ok(!body.includes("synthetic_only"));
  for (const message of ["Model is trained on", "Unspecified fields use", "Missing breed/age/sex", "Part weights reconciled", "prediction is extrapolated"]) assert.ok(!body.includes(message), message);
  assert.deepEqual(await page.locator("section > p.muted").allTextContents(), []);
  await page.screenshot({ path: path.join(artifactDirectory, "clean-ai-desktop.png"), fullPage: true });
  if (process.env.WEB_COVERAGE_FILE) {
    for (const menu of ["Genel Bakış", "Üretim Emirleri", "Kuzu Parçalama", "Raporlar", "Lot İzlenebilirliği", "Fatura OCR"]) {
      if (menu === "Raporlar") await page.route("**/reports/summary", (route) => route.fulfill({ json: {
        generated_at: "2026-10-31", purchases: [{ lot_code: "QA-LOT", material_name: "QA material", material_code: "QA-M", supplier_code: "QA-S", qty: 2, unit_cost: 5, total_cost: 10 }],
        production: [], butchery: [], lot_movements: [], sales: [], recalls: [],
      } }));
      await page.getByRole("button", { name: new RegExp(menu) }).click();
      await page.waitForLoadState("networkidle");
      if (menu === "Üretim Emirleri") {
        await page.locator(".pagehead button").click();
        await page.getByRole("button", { name: /Hammadde ekle/ }).click();
        await page.getByRole("button", { name: /Hammadde ekle/ }).click();
        const quantities = page.locator('.stack-form input[placeholder="Miktar"]');
        await quantities.nth(1).pressSequentially("12");
        assert.equal(await quantities.nth(1).inputValue(), "12");
        await page.locator(".stack-form").getByRole("button", { name: "✕", exact: true }).first().click();
        assert.equal(await quantities.first().inputValue(), "12");
      }
      if (menu === "Kuzu Parçalama") {
        const batch = page.locator('select:has(option[value="1"])').first();
        if (await batch.count()) await batch.selectOption("1");
        await page.waitForLoadState("networkidle");
        await page.locator(".pagehead button").click();
        await page.getByRole("button", { name: /Parça ekle/ }).click();
        await page.getByRole("button", { name: /Parça ekle/ }).click();
        const weights = page.locator('.stack-form input[placeholder="Ağırlık"]');
        await weights.nth(1).pressSequentially("12");
        assert.equal(await weights.nth(1).inputValue(), "12");
        await page.locator(".stack-form").getByRole("button", { name: "✕", exact: true }).first().click();
        assert.equal(await weights.first().inputValue(), "12");
      }
      if (menu === "Raporlar") {
        await page.getByRole("button", { name: "Alış Raporu", exact: true }).waitFor();
        assert.ok(await page.locator('.panel').getByText("QA-LOT", { exact: true }).count() > 0);
        for (const reportName of ["Üretim Maliyeti Raporu", "Parçalama Raporu", "Lot Hareket Geçmişi", "Satış Raporu", "Geri Çağırma Raporu", "Alış Raporu"]) {
          await page.getByRole("button", { name: reportName, exact: true }).click();
        }
      }
      if (menu === "Lot İzlenebilirliği") {
        await page.getByRole("button", { name: /İz sür/ }).click();
        await page.waitForLoadState("networkidle");
      }
      if (menu === "Fatura OCR") {
        await page.locator("textarea").fill("100 kg Et ET-2026-010 9,70 TL");
        await page.getByRole("button", { name: "Metni ayrıştır", exact: true }).click();
        await page.getByRole("heading", { name: "Önerilen kalemler", exact: true }).waitFor();
      }
    }
    await page.getByRole("button", { name: /Yapay Zekâ/ }).click();
    await heading.waitFor();
    await page.waitForLoadState("networkidle");
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await heading.scrollIntoViewIfNeeded();
  const scroll = chart.locator(".forecast-chart__scroll");
  assert.ok(await scroll.evaluate((element) => element.scrollWidth > element.clientWidth));
  await scroll.evaluate((element) => { element.scrollLeft = element.scrollWidth; });
  assert.ok(await scroll.evaluate((element) => element.scrollLeft > 0));
  const overflow = await page.evaluate(() => ({ viewport: document.documentElement.clientWidth, document: document.documentElement.scrollWidth }));
  assert.ok(overflow.document <= overflow.viewport + 1, JSON.stringify(overflow));
  await page.screenshot({ path: path.join(artifactDirectory, "clean-ai-mobile.png"), fullPage: true });
  assert.ok(posts.every((url) => url.startsWith("/ml/predict/v2/") || url === "/ocr/invoice"), "No business writes");
  assert.deepEqual(errors, []);
  if (process.env.WEB_COVERAGE_FILE) writeFileSync(process.env.WEB_COVERAGE_FILE, JSON.stringify(await page.coverage.stopJSCoverage()));
  console.log(JSON.stringify({ result: "PASS", model: production.model_version, production: production.prediction.finished_qty,
    carcassSum: cut.physical_check.parts_sum_kg, chartGroups: await groups.count(), colors, keyboardAndHover: true,
    mobileOverflow: overflow, warningsRemoved: true, planRemoved: true, pageErrors: errors }));
} finally { await browser.close(); }
