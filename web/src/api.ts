const API_BASE: string = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

async function request<T>(base: string, path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${base}${path}`, init);
  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const body = (await response.json()) as { error?: string; detail?: string };
      if (body.error) {
        message = body.error;
      } else if (body.detail) {
        message = body.detail;
      }
    } catch {
      // Keep the status message when the response is not JSON.
    }
    throw new Error(message);
  }
  return (await response.json()) as T;
}

export function get<T>(path: string): Promise<T> {
  return request<T>(API_BASE, path);
}

export function post<T>(path: string, body: unknown): Promise<T> {
  return request<T>(API_BASE, path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

// ML is always reached through the Go backend; the browser never needs the
// Python service address.
export function mlGet<T>(path: string): Promise<T> {
  return request<T>(API_BASE, `/ml${path}`);
}

export function mlPost<T>(path: string, body: unknown): Promise<T> {
  return request<T>(API_BASE, `/ml${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}
