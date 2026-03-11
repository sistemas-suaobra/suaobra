export function normalizeQr(qr?: string) {
  const v = (qr || "").trim();
  if (!v) return "";
  if (v.startsWith("data:image")) return v;
  return `data:image/png;base64,${v}`;
}