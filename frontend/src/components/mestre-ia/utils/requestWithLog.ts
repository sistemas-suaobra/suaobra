import { api } from "../../../store/api";

type ReqOpts = {
  label: string;
  method: "GET" | "POST";
  url: string;
  body?: any;
};

export async function requestWithLog<T = any>(opts: ReqOpts) {
  const { label, method, url, body } = opts;

  const startedAt = performance.now();
  const groupTitle = `🔎 [${label}] ${method} ${url}`;

  console.groupCollapsed(groupTitle);
  console.log("request:", { method, url, body });

  try {
    const client = api();

    let resp: any;
    if (method === "GET") resp = await client.get(url, body ?? {});
    else resp = await client.post(url, body ?? {});

    const status = resp?.status ?? resp?.response?.status ?? undefined;

    let data: any = null;
    try {
      data = await resp.json();
    } catch {
      data = null;
    }

    console.log("response:", { status, error: resp?.error, data, rawResp: resp });

    if (resp?.error) {
      const msg =
        typeof resp.error === "string"
          ? resp.error
          : resp.error?.message || "Request error";
      throw new Error(msg);
    }

    return { status, data: data as T };
  } finally {
    console.log(`took: ${(performance.now() - startedAt).toFixed(0)}ms`);
    console.groupEnd();
  }
}