import React from "react"
import type { ObrasPlusRecord, LeadOption } from "../types/create-campaign"
import { api, baseURL } from "../../../store/api"

type RecipientKind = "OWNER" | "PROFISSIONAL"

const makeRecipientKey = (obraId: string, contatoTipo: RecipientKind) =>
  `${obraId}::${contatoTipo}`

function pickBairros(selectedNeighborhood: any[]): string {
  if (!Array.isArray(selectedNeighborhood) || !selectedNeighborhood.length) return ""

  return selectedNeighborhood
    .map((item) => {
      if (!item) return ""
      if (typeof item === "string") return item.trim()
      return String(item?.bairro ?? "").trim()
    })
    .filter(Boolean)
    .join("|")
}

function isTruthyContactStatus(value: unknown) {
  if (typeof value === "boolean") return value
  if (value == null) return false

  const raw = String(value).trim().toUpperCase()

  return (
    raw === "1" ||
    raw === "TRUE" ||
    raw === "SIM" ||
    raw === "YES" ||
    raw === "ENVIADO"
  )
}

export function useCampaignLeadOptions(params: {
  visible: boolean
  teamId: string
  selectedCity: any
  selectedNeighborhood: any[]
  filterValue: string
  startDateFrom: string
  startDateTo: string
  endDateFrom: string
  endDateTo: string
  ocultarJaContactados?: boolean
  fallbackOptions: LeadOption[]
}) {
  const {
    visible,
    teamId,
    selectedCity,
    selectedNeighborhood,
    filterValue,
    startDateFrom,
    startDateTo,
    endDateFrom,
    endDateTo,
    ocultarJaContactados = true,
    fallbackOptions,
  } = params

  const [leadsOptionsLocal, setLeadsOptionsLocal] = React.useState<LeadOption[]>(fallbackOptions)
  const [contactedRecipientSet, setContactedRecipientSet] = React.useState<Set<string>>(new Set())

  const obraRecordMapRef = React.useRef<Map<string, ObrasPlusRecord>>(new Map())

  React.useEffect(() => {
    setLeadsOptionsLocal(fallbackOptions)
  }, [fallbackOptions])

  React.useEffect(() => {
    if (!visible) return
    if (!selectedCity?.city) return

    let cancelled = false

    const fetchObras = async () => {
      try {
        const payload = {
          teamId,
          city: selectedCity.city || "",
          bairro: pickBairros(selectedNeighborhood),
          order: "first_listing_date-desc,start_date-desc",
          filter: filterValue,
          sizeMin: "0",
          sizeMax: "9999999",
          offset: "0",
          itemsPerPage: "100",
          enriched: "false",
          startDateFrom,
          startDateTo,
          endDateFrom,
          endDateTo,
          ocultarJaContactados: ocultarJaContactados ? "true" : "false",
        }

        const resp = await api().get(`${baseURL()}/query/leads-plus`, payload)
        if (resp.error) throw new Error(resp.error)

        const data = await resp.response.json()
        const records: ObrasPlusRecord[] = Array.isArray(data?.records) ? data.records : []

        if (cancelled) return

        const nextMap = new Map<string, ObrasPlusRecord>()
        const nextOptions: LeadOption[] = []
        const nextContactedRecipientSet = new Set<string>()

        for (const r of records) {
          const obraId = String((r as any)?.obra_id ?? "").trim()
          if (!obraId) continue

          nextMap.set(obraId, r)

          nextOptions.push({
            label: (r as any)?.owner || (r as any)?.professional || (r as any)?.address || obraId,
            value: obraId,
          })

          const ownerContacted =
            !!(r as any)?.owner_contacted_at ||
            !!(r as any)?.owner_enviado_em ||
            isTruthyContactStatus((r as any)?.owner_contacted) ||
            isTruthyContactStatus((r as any)?.owner_status)

          const professionalContacted =
            !!(r as any)?.professional_contacted_at ||
            !!(r as any)?.professional_enviado_em ||
            isTruthyContactStatus((r as any)?.professional_contacted) ||
            isTruthyContactStatus((r as any)?.professional_status)

          if (ownerContacted) {
            nextContactedRecipientSet.add(makeRecipientKey(obraId, "OWNER"))
          }

          if (professionalContacted) {
            nextContactedRecipientSet.add(makeRecipientKey(obraId, "PROFISSIONAL"))
          }
        }

        obraRecordMapRef.current = nextMap
        setLeadsOptionsLocal(nextOptions)
        setContactedRecipientSet(nextContactedRecipientSet)
      } catch (e) {
        console.error(e)

        if (cancelled) return

        setLeadsOptionsLocal([])
        setContactedRecipientSet(new Set())
        obraRecordMapRef.current = new Map()
      }
    }

    fetchObras()

    return () => {
      cancelled = true
    }
  }, [
    visible,
    teamId,
    selectedCity,
    selectedNeighborhood,
    filterValue,
    startDateFrom,
    startDateTo,
    endDateFrom,
    endDateTo,
    ocultarJaContactados,
  ])

  return { leadsOptionsLocal, obraRecordMapRef, contactedRecipientSet }
}