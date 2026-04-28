export const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ||
  "http://localhost:8080/api";

export class ApiError extends Error {
  code?: string;
  status: number;
  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

function parseErrorPayload(data: unknown): string {
  if (data && typeof data === "object" && "errors" in data) {
    const errs = (data as { errors?: { message?: string; code?: string }[] })
      .errors;
    if (errs?.length) {
      return errs[0].message || "Request failed";
    }
  }
  return "Request failed";
}

export async function apiJson<T>(
  path: string,
  opts: RequestInit & { token?: string | null } = {}
): Promise<T> {
  const headers = new Headers(opts.headers);
  if (
    opts.body &&
    !(opts.body instanceof FormData) &&
    !headers.has("Content-Type")
  ) {
    headers.set("Content-Type", "application/json");
  }
  let token: string | null | undefined = opts.token;
  if (token === undefined && typeof window !== "undefined") {
    token = localStorage.getItem("carmanage_token");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, { ...opts, headers });
  if (res.status === 204) {
    return undefined as T;
  }
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const msg = parseErrorPayload(data);
    const code =
      data && typeof data === "object" && "errors" in data
        ? (data as { errors?: { code?: string }[] }).errors?.[0]?.code
        : undefined;
    throw new ApiError(res.status, msg, code);
  }
  return data as T;
}

export async function apiForm<T>(
  path: string,
  form: FormData,
  opts: { method?: string; token?: string | null } = {}
): Promise<T> {
  const headers = new Headers();
  let token: string | null | undefined = opts.token;
  if (token === undefined && typeof window !== "undefined") {
    token = localStorage.getItem("carmanage_token");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_BASE}${path}`, {
    method: opts.method || "POST",
    headers,
    body: form,
  });
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, parseErrorPayload(data));
  }
  return data as T;
}
