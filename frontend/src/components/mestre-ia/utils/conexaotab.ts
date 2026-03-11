import type { CryptoMode } from "../types/conexaotab";

export const cryptoOptions: { label: string; value: CryptoMode }[] = [
  { label: "Sem criptografia", value: "NONE" },
  { label: "STARTTLS (TLS)", value: "TLS" },
  { label: "SSL", value: "SSL" },
];
