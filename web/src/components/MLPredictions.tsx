import { useEffect, useRef, useState, type FormEvent, type ReactNode } from "react";
import { get, mlGet, mlPost } from "../api";
import { butcheryComparison, predictionError, productionFeatures, type InputLot, type MaterialCard, type OrderInput } from "../predictionInputs";
import { butcheryPayload, FEATURE_LABELS, initialButchery, modelMetrics, numberText, originText,
  PART_LABELS, productionProduct, statusText, warningText, yieldPayload, YIELD_INITIAL,
  type AnimalType, type ButcheryForm, type ButcheryResponse, type MLHealth, type MLReport,
  type ModelInfo, type ModelName, type ProductType, type TargetReport, type YieldForm, type YieldResponse } from "../mlContracts";
import "./mlPredictions.css";

interface Order { id: number; code: string; expected_qty: number; status: string; service_unit_cost: number }
interface Product { id: number; code: string; name: string; unit: string }
interface HistoryRow { code: string; predicted: number | null; actual: number | null; expected?: number; version?: string; error?: string; parts?: { code: string; predicted: number; actual: number }[] }
const TITLES: Record<ModelName, string> = { yield: "Üretim miktarı", butchery_lamb: "Kuzu parçalama", butchery_cattle: "Sığır parçalama" };
const PRODUCTS: Record<ProductType, string> = { sucuk: "Sucuk", salam: "Salam", sausage: "Sosis" };
const SALEABLE = ["leg_weight", "loin_weight", "shoulder_weight", "rib_weight", "mince_weight", "diced_meat_weight"];
function message(error: unknown) { return error instanceof Error ? error.message : String(error); }
function requireModel(health: MLHealth | null, name: ModelName): ModelInfo {
  const model = health?.models.find((entry) => entry.name === name);
  if (!model?.feature_names?.length || !model.input_ranges) throw new Error("Güncel model yüklenemedi. ML servisini ve model dosyalarını kontrol edin.");
  return model;
}
function checkVersion(response: { model_version: string; dataset_version: string }, model: ModelInfo) {
  if (response.model_version !== model.model_version || response.dataset_version !== model.dataset_version) throw new Error("Model veya eğitim verisi sürümü değişti. Verileri yenileyip tekrar tahmin alın.");
}
function Notice({ children, error = false }: { children: ReactNode; error?: boolean }) {
  return <div className={`ai-notice${error ? " ai-error" : ""}`} role={error ? "alert" : "status"}>{children}</div>;
}
function Warnings({ warnings }: { warnings: string[] }) {
  return warnings.length ? <ul className="ai-warnings">{[...new Set(warnings.map(warningText))].map((warning) => <li key={warning}>{warning}</li>)}</ul> : null;
}
function NumberField({ name, value, onChange, unit = "kg", optional = false, min = 0, max, range }: {
  name: string; value: string; onChange: (value: string) => void; unit?: string; optional?: boolean; min?: number; max?: number; range?: { min?: number; max?: number };
}) {
  return <label className="ai-field"><span>{FEATURE_LABELS[name] ?? name} {unit && `(${unit})`}{optional && " · opsiyonel"}</span>
    <input type="number" step="any" min={min} max={max} required={!optional} value={value} onChange={(event) => onChange(event.target.value)} />
    {range?.min !== undefined && range.max !== undefined && <small>Eğitim: {numberText(range.min, unit === "kg" ? 3 : 1)}–{numberText(range.max, unit === "kg" ? 3 : 1)} {unit}</small>}
  </label>;
}
function ErrorTable({ metrics }: { metrics: TargetReport }) {
  return <div className="ai-table-scroll"><table><thead><tr><th>Test metriği</th><th>Model</th><th>Basit oran</th></tr></thead><tbody>
    <tr><th>Ortalama mutlak hata</th><td>{numberText(metrics.test.mae, 3)} kg</td><td>{numberText(metrics.proportional_baseline_test.mae, 3)} kg</td></tr>
    <tr><th>RMSE</th><td>{numberText(metrics.test.rmse, 3)} kg</td><td>{numberText(metrics.proportional_baseline_test.rmse, 3)} kg</td></tr>
    <tr><th>Ortalama yüzde hata (MAPE)</th><td>{numberText(metrics.test.mape)}%</td><td>{numberText(metrics.proportional_baseline_test.mape)}%</td></tr>
    <tr><th>R²</th><td>{numberText(metrics.test.r2, 4)}</td><td>{numberText(metrics.proportional_baseline_test.r2, 4)}</td></tr>
  </tbody></table></div>;
}
function ModelCard({ name, health, report }: { name: ModelName; health: MLHealth | null; report: MLReport | null }) {
  const model = health?.models.find((entry) => entry.name === name);
  const metrics = modelMetrics(report, model, name === "yield" ? "finished_qty" : "total_saleable_weight");
  const yieldMetric = modelMetrics(report, model, "yield_pct");
  const partErrors = SALEABLE.map((part) => modelMetrics(report, model, part)?.test.mape).filter((value): value is number => typeof value === "number" && Number.isFinite(value));
  return <article className="ai-panel"><div className="ai-panel-heading"><h2>{TITLES[name]}</h2><span className="ai-badge">{model ? `v${model.model_version}` : "Yüklü değil"}</span></div>
    {model ? <><p className="ai-muted">{statusText(model.model_status)} · {name === "yield" ? "Mamul miktarı" : "Toplam satılabilir et"} hatası</p>
      <dl className="ai-model-counts"><div><dt>Eğitim</dt><dd>{model.training_sample_count} kayıt</dd></div><div><dt>Bağımsız test</dt><dd>{model.test_sample_count} kayıt</dd></div></dl>
      {metrics ? <><ErrorTable metrics={metrics} /><p className="ai-muted">Eğitim MAE: {numberText(metrics.train.mae, 3)} kg · Test−eğitim farkı: {numberText(metrics.train_test_mae_gap, 3)} kg</p>
        {yieldMetric?.test.mae != null && <p className="ai-muted">Üretim verimi hatası: {numberText(yieldMetric.test.mae * 100)} yüzde puan · R²: {numberText(yieldMetric.test.r2, 4)}</p>}
        {partErrors.length > 0 && <p className="ai-muted">Tek tek satılabilir parçalarda MAPE: %{numberText(Math.min(...partErrors))}–{numberText(Math.max(...partErrors))}.</p>}
      </> : <p className="ai-muted">Bu model ve veri sürümüyle eşleşen test raporu alınamadı.</p>}
      <details className="ai-details"><summary>Eğitim kapsamı ve girdi aralıkları</summary>
        <p>{model.training_date ? `Eğitim: ${new Date(model.training_date).toLocaleString("tr-TR", { timeZone: "Europe/Istanbul" })}` : "Eğitim tarihi sunulmadı."}</p>
        <p>Veri: {originText(model.data_origin)} · Sürüm: <code>{model.dataset_version}</code></p><p className="ai-muted">{model.validation_method}</p>
        <ul>{model.feature_names?.map((feature) => { const range = model.input_ranges?.[feature]; return <li key={feature}>{FEATURE_LABELS[feature] ?? feature}{feature.endsWith("_kg") ? " (kg)" : feature === "age_months" ? " (ay)" : ""}: {range?.allowed?.join(", ") ?? `${numberText(range?.min, 3)}–${numberText(range?.max, 3)}`}</li>; })}</ul>
      </details></> : <p className="ai-muted">Servisten model metadata’sı gelmeden sürüm veya başarı oranı gösterilmez.</p>}
  </article>;
}
function HistoryChart({ rows }: { rows: HistoryRow[] }) {
  const valid = rows.filter((row) => row.predicted !== null && row.actual !== null);
  if (!valid.length) return null;
  const max = Math.max(1, ...valid.flatMap((row) => [row.predicted!, row.actual!, row.expected ?? 0]));
  return <div className="ai-table-scroll"><svg width={Math.max(440, valid.length * 105)} height="190" role="img" aria-label="Güncel tahmin, plan ve kayıtlı mamul miktarı">
    <text x="12" y="18" fontSize="12">Model (mavi) · Plan (turuncu) · Kayıtlı (yeşil), kg</text>
    {valid.map((row, index) => <g key={row.code}><title>{row.code}: tahmin {row.predicted} kg, kayıtlı {row.actual} kg</title>
      {[row.predicted!, row.expected ?? 0, row.actual!].map((value, pos) => <rect key={pos} x={12 + index * 105 + pos * 22} y={155 - value / max * 115} width="18" height={value / max * 115} fill={["#456d9c", "#b97722", "#2c7058"][pos]} />)}
      <text x={12 + index * 105} y="177" fontSize="10">{row.code.slice(-9)}</text></g>)}
  </svg></div>;
}
function HistoryTable({ rows, production = false }: { rows: HistoryRow[]; production?: boolean }) {
  return <div className="ai-table-scroll"><table><thead><tr><th>{production ? "Emir" : "Parti"}</th><th>Tahmin kg</th>{production && <th>Plan kg</th>}<th>Kayıtlı kg</th><th>Mutlak hata kg</th><th>Hata %</th><th>Model / durum</th></tr></thead><tbody>
    {rows.map((row) => { const error = row.actual === null ? null : predictionError(row.predicted, row.actual); return <tr key={row.code}><th>{row.code}</th><td>{numberText(row.predicted, 3)}</td>{production && <td>{numberText(row.expected, 3)}</td>}<td>{numberText(row.actual, 3)}</td><td>{numberText(error?.absolute, 3)}</td><td>{numberText(error?.percentage)}</td><td className={row.error ? "ai-error-text" : ""}>{row.error ?? `v${row.version}`}</td></tr>; })}
  </tbody></table></div>;
}

export default function MLPredictions() {
  const [health, setHealth] = useState<MLHealth | null>(null), [report, setReport] = useState<MLReport | null>(null);
  const [loading, setLoading] = useState(false), [historyLoading, setHistoryLoading] = useState(false);
  const [metadataError, setMetadataError] = useState<string | null>(null), [historyError, setHistoryError] = useState<string | null>(null);
  const [forecast, setForecast] = useState<HistoryRow[]>([]), [cuts, setCuts] = useState<HistoryRow[]>([]);
  const [yieldForm, setYieldForm] = useState<YieldForm>({ ...YIELD_INITIAL });
  const [butcheryForm, setButcheryForm] = useState<ButcheryForm>(initialButchery("lamb"));
  const [yieldResult, setYieldResult] = useState<YieldResponse | null>(null), [cutResult, setCutResult] = useState<ButcheryResponse | null>(null);
  const [yieldError, setYieldError] = useState<string | null>(null), [cutError, setCutError] = useState<string | null>(null);
  const [yieldBusy, setYieldBusy] = useState(false), [cutBusy, setCutBusy] = useState(false);
  const generation = useRef(0);
  const yieldModel = health?.models.find((model) => model.name === "yield");
  const cutModel = health?.models.find((model) => model.name === `butchery_${butcheryForm.animal_type}`);
  const synthetic = health?.models.some((model) => model.data_origin?.startsWith("synthetic") || model.model_status === "synthetic_only");

  async function loadHistory(currentHealth: MLHealth | null, token: number) {
    setHistoryLoading(true);
    try {
      const [orders, lots, cards, batches, products] = await Promise.all([
        get<Order[]>("/production-orders"), get<InputLot[]>("/lots"), get<MaterialCard[]>("/materials"),
        get<{ id: number; code: string; animal_type: string; input_weight: number }[]>("/butchery-batches"), get<Product[]>("/products"),
      ]);
      const [production, butchery] = await Promise.all([
        Promise.all((orders ?? []).filter((order) => order.status === "completed").map(async (order): Promise<HistoryRow> => {
          let actual: number | null = null;
          try {
            const [summary, inputs, outputs] = await Promise.all([
              get<{ finished_qty: number }>(`/production-orders/${order.id}`), get<OrderInput[]>(`/production-orders/${order.id}/inputs`), get<{ product_lot_id: number }[]>(`/production-orders/${order.id}/outputs`),
            ]);
            actual = summary.finished_qty;
            const model = requireModel(currentHealth, "yield");
            const context = productionProduct(outputs ?? [], lots ?? [], products ?? []);
            actual *= context.kg_per_unit;
            const prediction = await mlPost<YieldResponse>("/predict/v2/yield", { ...productionFeatures(inputs ?? [], lots ?? [], cards ?? []), product_type: context.product_type, service_unit_cost: order.service_unit_cost / context.kg_per_unit, labor_cost: 0 });
            checkVersion(prediction, model);
            return { code: order.code, predicted: prediction.prediction.finished_qty, expected: order.expected_qty * context.kg_per_unit, actual, version: prediction.model_version };
          } catch (error) { return { code: order.code, predicted: null, actual: null, error: message(error) }; }
        })),
        Promise.all((batches ?? []).map(async (batch): Promise<HistoryRow> => {
          try {
            const aliases: Record<string, AnimalType> = { kuzu: "lamb", dana: "cattle", lamb: "lamb", cattle: "cattle" };
            const animal = aliases[batch.animal_type.toLocaleLowerCase("tr-TR")];
            if (!animal) throw new Error("Hayvan türü desteklenmiyor.");
            const model = requireModel(currentHealth, `butchery_${animal}`);
            const [result, prediction] = await Promise.all([
              get<{ items: { weight: number; is_waste: boolean; product_id: number }[] }>(`/butchery-batches/${batch.id}`), mlPost<ButcheryResponse>("/predict/v2/butchery", { animal_type: animal, carcass_weight_kg: batch.input_weight }),
            ]);
            checkVersion(prediction, model);
            return { code: batch.code, ...butcheryComparison(result.items ?? [], products ?? [], prediction.prediction), version: prediction.model_version };
          } catch (error) { return { code: batch.code, predicted: null, actual: null, error: message(error) }; }
        })),
      ]);
      if (generation.current === token) { setForecast(production); setCuts(butchery); }
    } catch (error) { if (generation.current === token) setHistoryError(`Kayıt karşılaştırmaları alınamadı: ${message(error)}`); }
    finally { if (generation.current === token) setHistoryLoading(false); }
  }
  async function refresh() {
    const token = ++generation.current;
    setLoading(true); setMetadataError(null); setHistoryError(null); setHealth(null); setReport(null);
    setYieldResult(null); setCutResult(null); setForecast([]); setCuts([]);
    const [h, r] = await Promise.allSettled([mlGet<MLHealth>("/health"), mlGet<MLReport>("/report")]);
    if (generation.current !== token) return;
    const loaded = h.status === "fulfilled" ? h.value : null;
    setHealth(loaded);
    const errors: string[] = [];
    if (h.status === "rejected") errors.push(`ML servisine erişilemiyor: ${message(h.reason)}`);
    if (r.status === "fulfilled" && r.value.model_reports) setReport(r.value);
    else errors.push(r.status === "rejected" ? `Test raporu alınamadı: ${message(r.reason)}` : "ML raporu eski sözleşmeyi kullanıyor; servisi güncelleyin.");
    setMetadataError(errors.length ? errors.join(" ") : null); setLoading(false);
    await loadHistory(loaded, token);
  }
  useEffect(() => { void refresh(); return () => { generation.current++; }; }, []);
  function changeYield<K extends keyof YieldForm>(key: K, value: YieldForm[K]) { setYieldForm((form) => ({ ...form, [key]: value })); setYieldResult(null); setYieldError(null); }
  function changeCut<K extends keyof ButcheryForm>(key: K, value: ButcheryForm[K]) { setButcheryForm((form) => ({ ...form, [key]: value })); setCutResult(null); setCutError(null); }
  async function predictYield(event: FormEvent) {
    event.preventDefault(); setYieldBusy(true); setYieldError(null); setYieldResult(null); const token = generation.current;
    try { const model = requireModel(health, "yield"); const response = await mlPost<YieldResponse>("/predict/v2/yield", yieldPayload(yieldForm)); checkVersion(response, model); if (generation.current === token) setYieldResult(response); }
    catch (error) { if (generation.current === token) setYieldError(message(error)); } finally { setYieldBusy(false); }
  }
  async function predictCut(event: FormEvent) {
    event.preventDefault(); setCutBusy(true); setCutError(null); setCutResult(null); const token = generation.current;
    try { const model = requireModel(health, `butchery_${butcheryForm.animal_type}`); const response = await mlPost<ButcheryResponse>("/predict/v2/butchery", butcheryPayload(butcheryForm)); checkVersion(response, model); if (generation.current === token) setCutResult(response); }
    catch (error) { if (generation.current === token) setCutError(message(error)); } finally { setCutBusy(false); }
  }
  function yieldNumber(name: keyof YieldForm, unit = "kg", max?: number, optional = false) {
    return <NumberField key={name} name={name} value={yieldForm[name]} onChange={(value) => changeYield(name, value)} unit={unit} max={max} optional={optional} min={name === "raw_material_kg" || name === "actual_finished_qty" ? .001 : 0} range={yieldModel?.input_ranges?.[name]} />;
  }
  function cutNumber(name: keyof ButcheryForm, unit = "kg", optional = false, min = .001, max = 1500) {
    return <NumberField key={name} name={name} value={butcheryForm[name]} onChange={(value) => changeCut(name, value)} unit={unit} optional={optional} min={min} max={max} range={cutModel?.input_ranges?.[name]} />;
  }
  const physical = cutResult?.physical_check;
  const physicalOK = physical && !physical.parts_over_carcass && Math.abs(physical.parts_sum_kg - physical.carcass_weight_kg) <= .001 && physical.saleable_sum_kg <= physical.carcass_weight_kg + .001;
  return <section className="ai-page"><header className="ai-page-heading"><div><div className="ai-eyebrow">KARAR DESTEK</div><h1>Yapay Zekâ Merkezi</h1><p className="ai-muted">Üretim ve parçalama tahminleri, yüklü modelin güncel test sonuçlarıyla birlikte.</p></div>
    <button className="ai-button ai-secondary" onClick={() => void refresh()} disabled={loading || historyLoading || yieldBusy || cutBusy}>{loading || historyLoading ? "Yenileniyor…" : "Verileri yenile"}</button></header>
    {metadataError && <Notice error>{metadataError}</Notice>}{loading && <Notice>Model bilgileri servisten alınıyor…</Notice>}
    {health && <Notice><strong>{health.models_loaded}/3 model yüklü.</strong> {synthetic ? "Sentetik veriyle eğitilmiş demo modelleri. Gerçek işletme verileriyle saha doğrulaması yapılmadı; aşağıdaki metrikler sentetik test sonuçlarıdır." : "Eğitim kaynağı ve doğrulama durumu her model kartında gösterilir."}{health.status !== "ok" && " Eksik model nedeniyle bazı tahminler kullanılamıyor."}</Notice>}
    <div className="ai-forms"><article className="ai-panel"><div className="ai-panel-heading"><h2>Üretim verimi ve maliyet</h2><span className="ai-badge">{yieldModel ? `v${yieldModel.model_version}` : "Bağlantı bekleniyor"}</span></div>
      <p className="ai-muted">Et girdisinden beklenen mamul tahmini. EKP, malzeme ve hizmet bedellerinden ayrıca hesaplanır. Başlangıç değerleri demo örneğidir.</p>
      <form onSubmit={(event) => void predictYield(event)}><fieldset disabled={yieldBusy || loading || !yieldModel?.feature_names?.length}><div className="ai-fields">
        <label className="ai-field"><span>Ürün tipi</span><select value={yieldForm.product_type} onChange={(event) => changeYield("product_type", event.target.value as ProductType)}>{Object.entries(PRODUCTS).map(([key, label]) => <option key={key} value={key}>{label}</option>)}</select></label>
        {yieldNumber("raw_material_kg", "kg", 10000)}{yieldNumber("spice_kg", "kg", 500)}{yieldNumber("casing_kg", "kg", 500)}{yieldNumber("other_material_kg", "kg", 500)}</div>
        <details className="ai-details" open><summary>Maliyet ve hizmet bilgileri</summary><div className="ai-fields">
          {yieldNumber("material_unit_price_meat", "€/kg", 1000)}{yieldNumber("material_unit_price_spice", "€/kg", 1000)}{yieldNumber("material_unit_price_casing", "€/kg", 1000)}{yieldNumber("material_unit_price_other", "€/kg", 1000)}
          <label className="ai-field"><span>Hizmet hesaplama</span><select value={yieldForm.service_basis} onChange={(event) => changeYield("service_basis", event.target.value as "unit" | "total")}><option value="unit">Mamul kg başına tarife</option><option value="total">Sabit toplam bütçe</option></select></label>
          {yieldNumber("service_value", yieldForm.service_basis === "unit" ? "€/kg mamul" : "€", yieldForm.service_basis === "unit" ? 1000 : 1000000)}{yieldNumber("labor_cost", "€", 1000000)}
        </div><p className="ai-muted">Hizmet bedeli işçiliği içeriyorsa ek işçiliği 0 bırakın. Mamul birimi kg; PDF’nin 1 kg paket örneğinde paket başına değerle aynıdır.</p></details>
        <div className="ai-fields">{yieldNumber("actual_finished_qty", "kg", 100000, true)}</div><div className="ai-actions"><button className="ai-button" type="submit">{yieldBusy ? "Hesaplanıyor…" : "Üretimi tahmin et"}</button><button className="ai-button ai-secondary" type="button" onClick={() => { setYieldForm({ ...YIELD_INITIAL, actual_finished_qty: "85" }); setYieldResult(null); setYieldError(null); }}>PDF örneğini yükle</button></div>
      </fieldset></form>
      {yieldError && <Notice error>{yieldError}</Notice>}
      {yieldResult && <div className="ai-result" aria-live="polite"><div className="ai-result-metrics"><div><span>Beklenen mamul</span><strong>{numberText(yieldResult.prediction.finished_qty, 3)} kg</strong></div><div><span>Et bazlı verim</span><strong>%{numberText(yieldResult.prediction.yield_pct * 100)}</strong></div><div><span>Tahmini EKP</span><strong>{numberText(yieldResult.prediction.estimated_ekp)} €/kg</strong></div></div>
        <p>Tahmini toplam maliyet: <strong>{numberText(yieldResult.cost_breakdown.total_cost)} €</strong> · Hizmet: {numberText(yieldResult.cost_breakdown.service_cost)} €</p>
        {yieldResult.accounting_ekp !== undefined && <p>Fiili mamule göre muhasebe EKP’si: <strong>{numberText(yieldResult.accounting_ekp)} €/kg</strong> · Fiili toplam maliyet: {numberText(yieldResult.accounting_cost_breakdown?.total_cost)} €</p>}
        {yieldResult.comparison && <p>Kayıtlı mamul: {numberText(yieldResult.comparison.actual, 3)} kg · Mutlak hata: <strong>{numberText(yieldResult.comparison.absolute_error, 3)} kg</strong> · Hata: %{numberText(yieldResult.comparison.percentage_error)}</p>}
        <p className="ai-muted">v{yieldResult.model_version} · {statusText(yieldResult.model_status)} · {originText(yieldResult.data_origin)}</p>
        {yieldResult.uncertainty.finished_qty && <p className="ai-muted">Sentetik testte mamul RMSE: {numberText(yieldResult.uncertainty.finished_qty.test_rmse, 3)} kg. Bu değer tek tahmin için garanti veya kalibre edilmiş güven aralığı değildir.</p>}<Warnings warnings={yieldResult.warnings} /></div>}
    </article>
    <article className="ai-panel"><div className="ai-panel-heading"><h2>Parçalama verimi</h2><span className="ai-badge">{cutModel ? `v${cutModel.model_version}` : "Bağlantı bekleniyor"}</span></div><p className="ai-muted">Kuzu ve sığır ayrı modellerle hesaplanır. Ölçülmüş karkas gereklidir; canlı ağırlıktan otomatik dönüşüm yapılmaz.</p>
      <form onSubmit={(event) => void predictCut(event)}><fieldset disabled={cutBusy || loading}><div className="ai-fields"><label className="ai-field"><span>Hayvan türü</span><select value={butcheryForm.animal_type} onChange={(event) => { setButcheryForm(initialButchery(event.target.value as AnimalType)); setCutResult(null); setCutError(null); }}><option value="lamb">Kuzu</option><option value="cattle">Sığır</option></select></label>{cutNumber("carcass_weight_kg")}{cutNumber("live_weight_kg", "kg", true, .001, 2000)}</div>
        <details className="ai-details" open><summary>Hayvan bilgileri</summary><div className="ai-fields"><label className="ai-field"><span>Irk · opsiyonel</span><input list="ai-breeds" value={butcheryForm.breed} onChange={(event) => changeCut("breed", event.target.value)} placeholder="Belirtilmezse demo varsayımı" /><datalist id="ai-breeds">{cutModel?.input_ranges?.breed?.allowed?.map((breed) => <option key={breed} value={breed} />)}</datalist></label>
          <label className="ai-field"><span>Cinsiyet</span><select value={butcheryForm.sex} onChange={(event) => changeCut("sex", event.target.value as "male" | "female")}><option value="male">Erkek</option><option value="female">Dişi</option></select></label>{cutNumber("age_months", "ay", true, 1, 120)}{cutNumber("conformation_score", "", false, 1, 5)}{cutNumber("fat_score", "", false, 1, 5)}
        </div><p className="ai-muted">Başlangıç skorları 3, cinsiyet erkektir. Eksik ırk/yaş için servis demo varsayımı kullanır; bunları gerçek hayvan ölçümü kabul etmeyin.</p></details><button className="ai-button" type="submit" disabled={!cutModel?.feature_names?.length}>{cutBusy ? "Hesaplanıyor…" : "Parçaları tahmin et"}</button></fieldset></form>
      {cutError && <Notice error>{cutError}</Notice>}
      {cutResult && <div className="ai-result" aria-live="polite"><div className="ai-result-metrics"><div><span>Satılabilir toplam</span><strong>{numberText(cutResult.prediction.total_saleable_weight, 3)} kg</strong></div><div><span>Parçalama verimi</span><strong>%{numberText(cutResult.prediction.cutting_yield_pct)}</strong></div></div>
        <div className="ai-table-scroll"><table><thead><tr><th>Parça</th><th>Tahmini kg</th></tr></thead><tbody>{[...SALEABLE, "bone_weight", "waste_weight"].map((part) => <tr key={part}><th>{PART_LABELS[part]}</th><td>{numberText(cutResult.prediction[part], 3)}</td></tr>)}</tbody></table></div>
        {physical && <Notice error={!physicalOK}>Karkas: {numberText(physical.carcass_weight_kg, 3)} kg · Parçalar + kemik + fire: {numberText(physical.parts_sum_kg, 3)} kg. {physicalOK ? "Ağırlık dengesi tutarlı." : "Ağırlık dengesi kontrol edilmeli."}</Notice>}
        <p className="ai-muted">v{cutResult.model_version} · {statusText(cutResult.model_status)} · {originText(cutResult.data_origin)}</p><Warnings warnings={cutResult.warnings} /></div>}
    </article></div>
    <div className="ai-section-heading"><h2>Model test sonuçları</h2><p className="ai-muted">Metrikler yüklü modelden alınır. R² doğruluk yüzdesi değildir; toplam miktardaki küçük hata her parçanın aynı doğrulukta olduğunu göstermez.</p></div>
    <div className="ai-models">{(["yield", "butchery_lamb", "butchery_cattle"] as const).map((name) => <ModelCard key={name} name={name} health={health} report={report} />)}</div>
    <div className="ai-section-heading"><h2>Kayıtlı sonuçlarla karşılaştırma</h2><p className="ai-muted">Güncel model geçmiş kayıtlara yeniden uygulanır; bunlar üretim öncesinde kaydedilmiş tahminler değildir. Demo kayıtları saha doğrulaması sayılmaz ve bu ekran modeli yeniden eğitmez.</p></div>
    {historyError && <Notice error>{historyError}</Notice>}{historyLoading && <Notice>Kayıtlar ve güncel model tahminleri karşılaştırılıyor…</Notice>}
    <article className="ai-panel"><h3>Tamamlanmış üretim emirleri</h3>{forecast.length ? <><HistoryChart rows={forecast} /><HistoryTable rows={forecast} production /></> : <p className="ai-muted">{historyLoading ? "Kayıtlar yükleniyor…" : historyError ? "Kayıtlar alınamadı." : "Karşılaştırılabilir tamamlanmış üretim emri yok."}</p>}</article>
    <article className="ai-panel"><h3>Parçalama partileri</h3><p className="ai-muted">Yalnız kaydı bulunan aynı satılabilir parçalar karşılaştırılır. Kemik ve fire bu toplama girmez.</p>
      {cuts.length ? <><HistoryTable rows={cuts} /><details className="ai-details"><summary>Parça bazlı hatalar</summary><div className="ai-table-scroll"><table><thead><tr><th>Parti</th><th>Ürün</th><th>Tahmin kg</th><th>Kayıtlı kg</th><th>Mutlak hata kg</th><th>Hata %</th></tr></thead><tbody>{cuts.flatMap((row) => (row.parts ?? []).map((part) => { const error = predictionError(part.predicted, part.actual); return <tr key={`${row.code}-${part.code}`}><th>{row.code}</th><td>{part.code}</td><td>{numberText(part.predicted, 3)}</td><td>{numberText(part.actual, 3)}</td><td>{numberText(error.absolute, 3)}</td><td>{numberText(error.percentage)}</td></tr>; }))}</tbody></table></div></details></>
        : <p className="ai-muted">{historyLoading ? "Kayıtlar yükleniyor…" : historyError ? "Kayıtlar alınamadı." : "Karşılaştırılabilir parçalama kaydı yok."}</p>}
    </article></section>;
}
