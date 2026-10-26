import { useEffect, useState } from "react";
import { get } from "../api";

interface Product {
  id: number;
  code: string;
  name: string;
  unit: string;
  list_price: number;
  category: string;
  is_finished: boolean;
}

export default function Products() {
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<Product[]>("/products")
      .then(setProducts)
      .catch((err: unknown) => setError(err instanceof Error ? err.message : String(err)))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return <p>Loading products...</p>;
  }
  if (error) {
    return <p className="error">{error}</p>;
  }

  return (
    <section>
      <h2>Products</h2>
      <table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Unit</th>
            <th>List price</th>
            <th>Category</th>
            <th>Finished</th>
          </tr>
        </thead>
        <tbody>
          {products.map((product) => (
            <tr key={product.id}>
              <td>{product.code}</td>
              <td>{product.name}</td>
              <td>{product.unit}</td>
              <td>{product.list_price.toFixed(2)}</td>
              <td>{product.category}</td>
              <td>{product.is_finished ? "Yes" : "No"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
