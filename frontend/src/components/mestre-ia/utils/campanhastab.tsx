import React from "react";
import type { LeadOption } from "../types/campanhastab";

function extractNomeFromLead(lead: LeadOption | undefined) {
  if (!lead) return "Cliente";
  return lead.owner || lead.professional || "Cliente";
}

function saudacaoFromNow() {
  const h = new Date().getHours();
  if (h < 12) return "bom dia";
  if (h < 18) return "boa tarde";
  return "boa noite";
}

export function renderPreview(text: string, selectedLeads: string[], leadsOptions: LeadOption[], city: string, p0: any[]) {
  const firstId = selectedLeads?.[0];
  const first = leadsOptions.find((l) => l.value === firstId);

  const nome = extractNomeFromLead(first);
  const primeiroNome = nome.split(" ")[0];
  const saudacao = saudacaoFromNow();

  const preview = (text || "")
    .replaceAll("{{nome}}", nome)
    .replaceAll("{{primeiroNome}}", primeiroNome)
    .replaceAll("{{saudacao}}", saudacao);

  return <span style={{ whiteSpace: "pre-wrap" }}>{preview || "—"}</span>;
}