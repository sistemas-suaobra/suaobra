import type { ObrasPlusRecord, ContactRecord } from "../types/create-campaign"

export function safeStr(v: unknown) {
  return String(v ?? "").trim()
}

export function resolveTargets(r: ObrasPlusRecord) {
  const owner = safeStr(r.owner)
  const professional = safeStr(r.professional)
  const phoneName = r.has_owner_phone ? owner : r.has_professional_phone ? professional : owner || professional
  const emailName = r.has_owner_email ? owner : r.has_professional_email ? professional : owner || professional
  return { phoneName, emailName }
}

export function pickTelephone(recs: ContactRecord[]) {
  return safeStr(recs.find((x) => safeStr(x.telephone))?.telephone)
}

export function pickEmail(recs: ContactRecord[]) {
  return safeStr(recs.find((x) => safeStr(x.email))?.email).toLowerCase()
}