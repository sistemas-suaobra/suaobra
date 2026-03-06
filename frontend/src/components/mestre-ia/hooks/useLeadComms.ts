import React from "react"
import type { ContactRecord, LeadComms, ObrasPlusRecord } from "../types/create-campaign"
import { api, baseURL } from "../../../store/api"
import { resolveTargets, pickTelephone, pickEmail, safeStr } from "../utils/obrasPlus"

export function useLeadComms(selectedCity: any, obraRecordMapRef: React.MutableRefObject<Map<string, ObrasPlusRecord>>) {
  const commsCacheRef = React.useRef<Map<string, LeadComms>>(new Map())

  const getContactsForName = async (nome: string, cidade: string, uf: string) => {
    const n = safeStr(nome)
    const c = safeStr(cidade)
    const s = safeStr(uf)
    if (!n || !c || !s) return [] as ContactRecord[]
    try {
      const resp = await api().get(`${baseURL()}/query/obras-plus-contacts`, { nome: n, cidade: c, uf: s })
      if (resp.error) throw new Error(resp.error)
      const data = await resp.response.json()
      return Array.isArray(data.records) ? (data.records as ContactRecord[]) : []
    } catch {
      return [] as ContactRecord[]
    }
  }

  const fetchLeadComms = React.useCallback(
    async (obraId: string): Promise<LeadComms> => {
      const cached = commsCacheRef.current.get(obraId)
      if (cached) return cached

      const rec = obraRecordMapRef.current.get(obraId)
      const city = safeStr(rec?.city) || safeStr(selectedCity?.city)
      const uf = safeStr(rec?.state) || safeStr(selectedCity?.state)

      const { phoneName, emailName } = resolveTargets(rec || ({ obra_id: obraId } as ObrasPlusRecord))

      const phoneRecs = phoneName ? await getContactsForName(phoneName, city, uf) : []
      const emailRecs = emailName && emailName !== phoneName ? await getContactsForName(emailName, city, uf) : phoneRecs

      const telefone_e164 = pickTelephone(phoneRecs)
      const email = pickEmail(emailRecs)
      const nome_contato = safeStr(phoneName) || safeStr(emailName) || "[NOME]"

      const comms: LeadComms = { telefone_e164, email, nome_contato }
      commsCacheRef.current.set(obraId, comms)
      return comms
    },
    [selectedCity, obraRecordMapRef]
  )

  return { fetchLeadComms }
}