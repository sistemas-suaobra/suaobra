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

export type DueReminderRecord = {
  lead_id?: string
  owner?: string
  city?: string
  state?: string
  alert_at?: number | string | null
}

const UNIT_MS: Record<ReminderTimeUnit, number> = {
  Horas: 1000 * 3600,
  Dias: 1000 * 3600 * 24,
  Semanas: 1000 * 3600 * 24 * 7,
  Meses: 1000 * 3600 * 24 * 30,
}

/**
 * Calcula o timestamp do lembrete a partir de unidade relativa.
 * Mantido por compatibilidade; a UI atual usa data personalizada.
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

/** Normaliza e valida a data personalizada do lembrete. */
export function resolveReminderDate(
  value: Date | string | number | null | undefined,
  opts?: { requireFuture?: boolean; nowMs?: number },
): Date | undefined {
  if (value == null || value === '') return undefined

  const date = value instanceof Date ? value : new Date(value)
  if (!Number.isFinite(date.getTime())) return undefined

  const nowMs = opts?.nowMs ?? Date.now()
  if (opts?.requireFuture !== false && date.getTime() <= nowMs) return undefined

  return date
}

/** Data padrão ao abrir o agendamento (amanhã, mesma hora, segundos zerados). */
export function defaultReminderDate(nowMs: number = Date.now()): Date {
  const d = new Date(nowMs)
  d.setDate(d.getDate() + 1)
  d.setSeconds(0, 0)
  return d
}

export function isReminderDue(
  alertAt: number | null | undefined,
  alerted: boolean | null | undefined,
  nowMs: number = Date.now(),
  opts?: { ignoreAlerted?: boolean },
): boolean {
  // alerted=true = e-mail do cron já enviado; a UI da plataforma notifica
  // independentemente (toast/banner) usando ignoreAlerted.
  if (!opts?.ignoreAlerted && alerted) return false
  if (alertAt == null || !Number.isFinite(Number(alertAt))) return false
  return Number(alertAt) <= nowMs
}

export function findDueReminders(
  leads: ReminderLeadLike[],
  nowMs: number = Date.now(),
  opts?: { ignoreAlerted?: boolean },
): ReminderLeadLike[] {
  return (leads || []).filter((lead) =>
    isReminderDue(
      lead.lead_properties?.alert_at,
      lead.lead_properties?.alerted,
      nowMs,
      opts,
    ),
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

export function reminderBannerText(rec: DueReminderRecord | ReminderLeadLike | null | undefined): string {
  if (!rec) return 'Hora de retornar o contato com o lead.'

  const like = rec as ReminderLeadLike
  if (like.title) return reminderToastDetail(like)

  const due = rec as DueReminderRecord
  const owner = String(due.owner || '').trim()
  const city = String(due.city || '').trim()
  const state = String(due.state || '').trim()
  const place = [city, state].filter(Boolean).join(', ')
  const title = [owner, place].filter(Boolean).join(' - ') || 'Lead'
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
