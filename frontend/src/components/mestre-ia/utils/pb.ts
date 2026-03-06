import { chunk } from "./array"

export function buildOrEq(field: string, values: string[]) {
  return values.map((v) => `${field} = "${v}"`).join(" || ")
}

export async function fetchLeadIdMap(pb: any, teamId: string, obraIds: string[]) {
  const map = new Map<string, string>()
  if (!obraIds.length) return map

  const parts = chunk(obraIds, 60)
  for (const part of parts) {
    const orFilter = buildOrEq("obra_id", part)
    const filter = `team_id = "${teamId}" && (${orFilter})`
    const recs = await pb.collection("lead").getFullList({ filter, fields: "id,obra_id,favorited_at" })
    for (const r of recs as any[]) map.set(String(r.obra_id), r.id)
  }

  return map
}