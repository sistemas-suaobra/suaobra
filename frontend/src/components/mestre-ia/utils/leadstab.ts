import type { LeadRow } from "../types/leadstab";

export function matchesSearch(lead: LeadRow, q: string) {
  const blob = [
    lead.nome,
    lead.email || "",
    lead.telefone || "",
    lead.cidade,
    lead.bairro,
    lead.obraTipo,
    String(lead.areaM2),
    lead.status,
  ]
    .join(" ")
    .toLowerCase();
  return blob.includes(q);
}
