import { useEffect, useState } from "react";
import { get } from "../api";

interface Material {
  id: number;
  code: string;
  name: string;
  unit: string;
  is_raw: boolean;
  default_purchase_price: number;
}

export default function Materials() {
  const [materials, setMaterials] = useState<Material[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<Material[]>("/materials")
      .then(setMaterials)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : String(err)))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <p>Loading materials...</p>;
  }
  if (error) {
    return <p className="error">{error}</p>;
  }

  return (
    <section>
      <h2>Materials</h2>
      <table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Unit</th>
            <th>Raw</th>
            <th>Default purchase price</th>
          </tr>
        </thead>
        <tbody>
          {materials.map((material) => (
            <tr key={material.id}>
              <td>{material.code}</td>
              <td>{material.name}</td>
              <td>{material.unit}</td>
              <td>{material.is_raw ? "Yes" : "No"}</td>
              <td>{material.default_purchase_price.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
