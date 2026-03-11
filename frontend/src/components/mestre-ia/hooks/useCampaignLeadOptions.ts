import React from "react"
import type { ObrasPlusRecord, LeadOption } from "../types/create-campaign"
import { api, baseURL, PB } from "../../../store/api"
import { user } from "../../../store/store"
import { fetchLeadIdMap } from "../utils/pb"

function pickBairro(selectedNeighborhood: any[]): string {
  const first = selectedNeighborhood?.[0]
  if (!first) return ""
  if (typeof first === "string") return first.trim()
  return String(first?.bairro ?? "").trim()
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
    fallbackOptions,
  } = params

  const [leadsOptionsLocal, setLeadsOptionsLocal] = React.useState<LeadOption[]>(fallbackOptions)
  const [existingLeadSet, setExistingLeadSet] = React.useState<Set<string>>(new Set())

  const obraRecordMapRef = React.useRef<Map<string, ObrasPlusRecord>>(new Map())

  React.useEffect(() => {
    setLeadsOptionsLocal(fallbackOptions)
  }, [fallbackOptions])

  React.useEffect(() => {
    if (!visible) return
    if (!selectedCity) return
    if (!teamId) return

    const fetchLeads = async () => {
      try {
        const payload = {
          city: selectedCity.city || "",
          bairro: pickBairro(selectedNeighborhood), // ✅ aqui
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
        }

        const resp = await api().get(`${baseURL()}/query/obras-plus`, payload)
        if (resp.error) throw new Error(resp.error)
        const data = await resp.response.json()
        const records: ObrasPlusRecord[] = Array.isArray(data.records) ? data.records : []

        obraRecordMapRef.current = new Map()
        for (const r of records) if (r?.obra_id) obraRecordMapRef.current.set(r.obra_id, r)

        const options: LeadOption[] = records.map((r) => ({
          label: r.owner || r.professional || r.address || r.obra_id,
          value: r.obra_id,
        }))

        setLeadsOptionsLocal(options)

        const obraIds = records.map((r) => r.obra_id).filter(Boolean)
        const pb = PB()
        pb.authStore.save(user.get().token, user.get())

        const leadMap = await fetchLeadIdMap(pb, teamId, obraIds)
        setExistingLeadSet(new Set(Array.from(leadMap.keys())))
      } catch (e) {
        console.error(e)
        setLeadsOptionsLocal([])
        setExistingLeadSet(new Set())
        obraRecordMapRef.current = new Map()
      }
    }

    fetchLeads()
  }, [visible, teamId, selectedCity, selectedNeighborhood, filterValue, startDateFrom, startDateTo, endDateFrom, endDateTo])

  return { leadsOptionsLocal, obraRecordMapRef, existingLeadSet }
}