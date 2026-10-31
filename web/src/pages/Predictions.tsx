import { useEffect, useState, type FormEvent } from "react";
import { mlGet, mlPost } from "../api";
import { ensurePredictionVersion, type YieldResponse, type ButcheryResponse, type MLReport as VersionedReport } from "../mlContracts";

interface YieldResult {
  predicted_qty: number;
  predicted_unit_cost: number;
}

interface ButcheryPart {
  product_code: string;
  weight: number;
}

interface ButcheryResult {
  parts: ButcheryPart[];
}

interface ModelMetrics {
  model_mae: number;
  model_rmse: number;
  baseline_mae: number;
  baseline_rmse: number;
}

interface MlReport extends VersionedReport {
  yield: ModelMetrics;
  butchery: ModelMetrics;
  butchery_lamb: ModelMetrics;
  butchery_cattle: ModelMetrics;
  note: string;
}

export default function Predictions() {
  const [rawMaterialKg, setRawMaterialKg] = useState("100");
  const [yieldResult, setYieldResult] = useState<YieldResult | null>(null);
  const [yieldError, setYieldError] = useState<string | null>(null);

  const [animalType, setAnimalType] = useState("kuzu");
  const [inputWeight, setInputWeight] = useState("25");
  const [butcheryResult, setButcheryResult] = useState<ButcheryResult | null>(null);
  const [butcheryError, setButcheryError] = useState<string | null>(null);

  const [sourceReport, setReport] = useState<MlReport | null>(null);
  const report = sourceReport && { ...sourceReport,
    butchery: ["dana", "cattle"].includes(animalType.trim().toLowerCase()) ? sourceReport.butchery_cattle : sourceReport.butchery_lamb };
  const [reportError, setReportError] = useState<string | null>(null);

  useEffect(() => {
    mlGet<MlReport>("/report")
      .then(setReport)
      .catch((err: unknown) => setReportError(err instanceof Error ? err.message : String(err)));
  }, []);

  async function predictYield(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setYieldError(null);
    setYieldResult(null);

    const raw = Number(rawMaterialKg);
    if (!Number.isFinite(raw) || raw <= 0) {
      setYieldError("Enter numeric raw material kg.");
      return;
    }

    try {
      const result = await mlPost<YieldResponse>("/predict/v2/yield", { raw_material_kg: raw, product_type: "sucuk", service_unit_cost: 5.5 });
      ensurePredictionVersion(sourceReport ?? await mlGet<MlReport>("/report"), "yield", result);
      setYieldResult({ predicted_qty: result.prediction.finished_qty, predicted_unit_cost: result.prediction.estimated_ekp });
    } catch (err) {
      setYieldError(err instanceof Error ? err.message : String(err));
    }
  }

  async function predictButchery(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setButcheryError(null);
    setButcheryResult(null);

    const weight = Number(inputWeight);
    if (!animalType.trim() || !Number.isFinite(weight) || weight <= 0) {
      setButcheryError("Enter an animal type and numeric input weight.");
      return;
    }

    try {
      const aliases: Record<string, "lamb" | "cattle"> = { kuzu: "lamb", dana: "cattle", lamb: "lamb", cattle: "cattle" };
      const animal = aliases[animalType.trim().toLowerCase()];
      if (!animal) throw new Error("Unknown animal type");
      const result = await mlPost<ButcheryResponse>("/predict/v2/butchery", { animal_type: animal, carcass_weight_kg: weight });
      ensurePredictionVersion(sourceReport ?? await mlGet<MlReport>("/report"), animal === "cattle" ? "butchery_cattle" : "butchery_lamb", result);
      const targets = ["leg_weight", "loin_weight", "shoulder_weight", "rib_weight", "mince_weight", "diced_meat_weight", "bone_weight"];
      setButcheryResult({ parts: targets.map((target, index) => ({ product_code: `PRD-${101 + index}`, weight: result.prediction[target] })) });
    } catch (err) {
      setButcheryError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <section>
      <h2>Predictions</h2>

      <h3>Yield prediction</h3>
      <form className="inline-form" onSubmit={predictYield}>
        <input
          value={rawMaterialKg}
          onChange={(event) => setRawMaterialKg(event.target.value)}
          placeholder="Raw material kg"
        />
        <button type="submit">Predict yield</button>
      </form>
      {yieldError && <p className="error">{yieldError}</p>}
      {yieldResult && (
        <table>
          <tbody>
            <tr>
              <th>Predicted quantity</th>
              <td>{yieldResult.predicted_qty}</td>
            </tr>
            <tr>
              <th>Predicted unit cost</th>
              <td>{yieldResult.predicted_unit_cost}</td>
            </tr>
          </tbody>
        </table>
      )}

      <h3>Butchery prediction</h3>
      <form className="inline-form" onSubmit={predictButchery}>
        <input
          value={animalType}
          onChange={(event) => setAnimalType(event.target.value)}
          placeholder="Animal type (kuzu/dana)"
        />
        <input
          value={inputWeight}
          onChange={(event) => setInputWeight(event.target.value)}
          placeholder="Input weight"
        />
        <button type="submit">Predict parts</button>
      </form>
      {butcheryError && <p className="error">{butcheryError}</p>}
      {butcheryResult && (
        <table>
          <thead>
            <tr>
              <th>Product code</th>
              <th>Weight</th>
            </tr>
          </thead>
          <tbody>
            {butcheryResult.parts.map((part) => (
              <tr key={part.product_code}>
                <td>{part.product_code}</td>
                <td>{part.weight}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <h3>Model report</h3>
      {reportError && <p className="error">{reportError}</p>}
      {report && (
        <>
          <h4>Yield</h4>
          <table>
            <thead>
              <tr>
                <th>Metric</th>
                <th>Model</th>
                <th>Baseline</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>MAE</td>
                <td>{report.yield.model_mae}</td>
                <td>{report.yield.baseline_mae}</td>
              </tr>
              <tr>
                <td>RMSE</td>
                <td>{report.yield.model_rmse}</td>
                <td>{report.yield.baseline_rmse}</td>
              </tr>
            </tbody>
          </table>

          <h4>Butchery</h4>
          <table>
            <thead>
              <tr>
                <th>Metric</th>
                <th>Model</th>
                <th>Baseline</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>MAE</td>
                <td>{report.butchery.model_mae}</td>
                <td>{report.butchery.baseline_mae}</td>
              </tr>
              <tr>
                <td>RMSE</td>
                <td>{report.butchery.model_rmse}</td>
                <td>{report.butchery.baseline_rmse}</td>
              </tr>
            </tbody>
          </table>

          <p className="message">{report.note}</p>
        </>
      )}
    </section>
  );
}
