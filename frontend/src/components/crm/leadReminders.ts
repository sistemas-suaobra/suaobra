/**
 * Lembretes do Venda Mais (alert_at / alerted em lead.properties).
 */

export type ReminderTimeUnit = 'Horas' | 'Dias' | 'Semanas' | 'Meses'

export type ReminderLeadLike = {
  lead_id?: string
  list_lead_id?: string
  title?: string
  lead_properties?: {
    alert_at?: number | null
    alerted?: boolean | null
  }
}

const UNIT_MS: Record<ReminderTimeUnit, number> = {
  Horas: 1000 * 3600,
  Dias: 1000 * 3600 * 24,
  Semanas: 1000 * 3600 * 24 * 7,
  Meses: 1000 * 3600 * 24 * 30,
}

/**
 * Calcula o timestamp do lembrete.
 * Número deve ser >= 1 — com 0 a UI antiga gravava undefined e o lembrete sumia.
 */
export function computeReminderAt(
  unit: ReminderTimeUnit,
  amount: number,
  nowMs: number = Date.now(),
): Date | undefined {
  const n = Number(amount)
  if (!Number.isFinite(n) || n < 1) return undefined
  const step = UNIT_MS[unit]
  if (!step) return undefined
  return new Date(nowMs + n * step)
}

export function isReminderDue(
  alertAt: number | null | undefined,
  alerted: boolean | null | undefined,
  nowMs: number = Date.now(),
): boolean {
  if (alerted) return false
  if (alertAt == null || !Number.isFinite(Number(alertAt))) return false
  return Number(alertAt) <= nowMs
}

export function findDueReminders(
  leads: ReminderLeadLike[],
  nowMs: number = Date.now(),
): ReminderLeadLike[] {
  return (leads || []).filter((lead) =>
    isReminderDue(lead.lead_properties?.alert_at, lead.lead_properties?.alerted, nowMs),
  )
}

/** Filtra lembretes ainda não notificados nesta sessão. */
export function filterUnnotifiedReminders(
  due: ReminderLeadLike[],
  alreadyNotifiedIds: Set<string>,
): ReminderLeadLike[] {
  return due.filter((lead) => {
    const id = lead.lead_id || lead.list_lead_id || ''
    return id && !alreadyNotifiedIds.has(id)
  })
}

export function reminderToastDetail(lead: ReminderLeadLike): string {
  const title = (lead.title || 'Lead').trim() || 'Lead'
  return `Hora de retornar contato: ${title}`
}

/** Garante alert_at/alerted no payload de properties (não perder no save). */
export function applyReminderFieldsToPayload(
  payload: Record<string, unknown>,
  alertAt: number | null | undefined,
  alerted: boolean | null | undefined,
): Record<string, unknown> {
  const next = { ...payload }
  if (alertAt == null || !Number.isFinite(Number(alertAt))) {
    delete next.alert_at
  } else {
    next.alert_at = Number(alertAt)
  }
  if (alerted == null) {
    delete next.alerted
  } else {
    next.alerted = Boolean(alerted)
  }
  return next
}
