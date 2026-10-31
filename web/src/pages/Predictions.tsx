import { useEffect, useState, type FormEvent } from "react";
import { mlGet, mlPost } from "../api";

interface YieldResult {
  predicted_qty: number;
  predicted_unit_cost: number;
  warnings: string[];
}

interface ButcheryPart {
  product_code: string;
  weight: number;
}

interface ButcheryResult {
  parts: ButcheryPart[];
  warnings: string[];
}

interface ModelMetrics {
  model_mae: number;
  model_rmse: number;
  baseline_mae: number;
  baseline_rmse: number;
}

interface MlReport {
  yield: ModelMetrics;
  butchery: ModelMetrics;
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

  const [report, setReport] = useState<MlReport | null>(null);
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

    const raw = Number.parseFloat(rawMaterialKg);
    if (!Number.isFinite(raw) || raw <= 0) {
      setYieldError("Enter numeric raw material kg.");
      return;
    }

    try {
      setYieldResult(
        await mlPost<YieldResult>("/predict/yield", {
          raw_material_kg: raw,
        }),
      );
    } catch (err) {
      setYieldError(err instanceof Error ? err.message : String(err));
    }
  }

  async function predictButchery(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setButcheryError(null);
    setButcheryResult(null);

    const weight = Number.parseFloat(inputWeight);
    if (!animalType.trim() || !Number.isFinite(weight) || weight <= 0) {
      setButcheryError("Enter an animal type and numeric input weight.");
      return;
    }

    try {
      setButcheryResult(
        await mlPost<ButcheryResult>("/predict/butchery", {
          animal_type: animalType.trim(),
          input_weight: weight,
        }),
      );
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

      {yieldResult?.warnings.map((warning) => <p key={warning} className="message">{warning}</p>)}
      <h3>Butchery prediction (measured carcass weight)</h3>
      <form className="inline-form" onSubmit={predictButchery}>
        <input
          value={animalType}
          onChange={(event) => setAnimalType(event.target.value)}
          placeholder="Animal type (kuzu/dana)"
        />
        <input
          value={inputWeight}
          onChange={(event) => setInputWeight(event.target.value)}
          placeholder="Carcass weight (kg)"
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

      {butcheryResult?.warnings.map((warning) => <p key={warning} className="message">{warning}</p>)}
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
