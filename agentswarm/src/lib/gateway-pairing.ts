/**
 * HTTP helpers for MiroClaw/ZeroClaw gateway pairing (`/admin/paircode`, `/pair`).
 * Admin routes are often localhost-only on the gateway; browsers may block cross-origin calls without CORS.
 */

export function gatewayHttpBaseFromPublicUrl(publicUrl: string): string {
  const raw = publicUrl.trim();
  if (!raw) {
    throw new Error("Public URL is empty");
  }
  const u = new URL(/^https?:\/\//i.test(raw) ? raw : `https://${raw}`);
  u.hash = "";
  u.search = "";
  u.pathname = u.pathname.replace(/\/+$/, "") || "/";
  let out = u.toString();
  if (u.pathname === "/" && out.endsWith("/")) {
    out = out.slice(0, -1);
  }
  return out;
}

export function joinGatewayHttpPath(httpBase: string, path: string): string {
  const base = httpBase.replace(/\/+$/, "");
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${base}${p}`;
}

export async function fetchGatewayPairingCode(httpBase: string): Promise<string | null> {
  const url = joinGatewayHttpPath(httpBase, "/admin/paircode");
  const res = await fetch(url, { method: "GET", mode: "cors" });
  const data = (await res.json().catch(() => ({}))) as {
    pairing_code?: string | null;
    message?: string;
  };
  if (!res.ok) {
    throw new Error(data.message || `GET /admin/paircode failed (${res.status})`);
  }
  const c = data.pairing_code;
  return typeof c === "string" && c.trim() ? c.trim() : null;
}

export async function postGatewayPairingCodeNew(httpBase: string): Promise<string> {
  const url = joinGatewayHttpPath(httpBase, "/admin/paircode/new");
  const res = await fetch(url, { method: "POST", mode: "cors" });
  const data = (await res.json().catch(() => ({}))) as {
    pairing_code?: string;
    message?: string;
    success?: boolean;
  };
  if (!res.ok) {
    throw new Error(data.message || `POST /admin/paircode/new failed (${res.status})`);
  }
  const code = data.pairing_code?.trim();
  if (!code) {
    throw new Error(data.message || "No pairing_code in response");
  }
  return code;
}

export async function exchangeGatewayPairingCode(httpBase: string, pairingCode: string): Promise<string> {
  const url = joinGatewayHttpPath(httpBase, "/pair");
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "X-Pairing-Code": pairingCode.trim(),
    },
    mode: "cors",
  });
  const data = (await res.json().catch(() => ({}))) as {
    token?: string;
    message?: string;
    error?: string;
  };
  if (!res.ok) {
    throw new Error(data.error || data.message || `POST /pair failed (${res.status})`);
  }
  const tok = data.token?.trim();
  if (!tok) {
    throw new Error(data.message || "No token in /pair response");
  }
  return tok;
}
