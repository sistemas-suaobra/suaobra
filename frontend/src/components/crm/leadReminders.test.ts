import { describe, expect, it } from 'vitest'
import {
  applyReminderFieldsToPayload,
  computeReminderAt,
  filterUnnotifiedReminders,
  findDueReminders,
  isReminderDue,
  reminderToastDetail,
} from './leadReminders'

describe('computeReminderAt', () => {
  const now = Date.parse('2026-08-03T23:00:00.000Z')

  it('rejeita número 0 (bug: gravava lembrete vazio)', () => {
    expect(computeReminderAt('Dias', 0, now)).toBeUndefined()
  })

  it('rejeita negativo ou NaN', () => {
    expect(computeReminderAt('Dias', -1, now)).toBeUndefined()
    expect(computeReminderAt('Horas', Number.NaN, now)).toBeUndefined()
  })

  it('soma dias/horas corretamente', () => {
    expect(computeReminderAt('Dias', 1, now)?.getTime()).toBe(now + 24 * 3600 * 1000)
    expect(computeReminderAt('Horas', 2, now)?.getTime()).toBe(now + 2 * 3600 * 1000)
    expect(computeReminderAt('Semanas', 1, now)?.getTime()).toBe(now + 7 * 24 * 3600 * 1000)
  })
})

describe('isReminderDue / findDueReminders', () => {
  const now = 1_000_000

  it('considera devido quando alert_at passou e alerted=false', () => {
    expect(isReminderDue(now - 1, false, now)).toBe(true)
    expect(isReminderDue(now + 1, false, now)).toBe(false)
    expect(isReminderDue(now - 1, true, now)).toBe(false)
    expect(isReminderDue(undefined, false, now)).toBe(false)
  })

  it('lista leads com lembrete vencido', () => {
    const due = findDueReminders(
      [
        { lead_id: 'a', title: 'A', lead_properties: { alert_at: now - 10, alerted: false } },
        { lead_id: 'b', title: 'B', lead_properties: { alert_at: now + 10, alerted: false } },
        { lead_id: 'c', title: 'C', lead_properties: { alert_at: now - 10, alerted: true } },
      ],
      now,
    )
    expect(due.map((l) => l.lead_id)).toEqual(['a'])
  })
})

describe('filterUnnotifiedReminders', () => {
  it('não re-notifica o mesmo lead na sessão', () => {
    const due = [
      { lead_id: 'a', title: 'A' },
      { lead_id: 'b', title: 'B' },
    ]
    const seen = new Set(['a'])
    expect(filterUnnotifiedReminders(due, seen).map((l) => l.lead_id)).toEqual(['b'])
  })
})

describe('applyReminderFieldsToPayload', () => {
  it('inclui alert_at e alerted no payload', () => {
    const payload = applyReminderFieldsToPayload({ observations: 'x' }, 12345, false)
    expect(payload.alert_at).toBe(12345)
    expect(payload.alerted).toBe(false)
    expect(payload.observations).toBe('x')
  })

  it('remove alert_at ao cancelar lembrete', () => {
    const payload = applyReminderFieldsToPayload({ alert_at: 99, alerted: false }, undefined, undefined)
    expect(payload.alert_at).toBeUndefined()
    expect(payload.alerted).toBeUndefined()
  })
})

describe('reminderToastDetail', () => {
  it('monta mensagem clara para toast na plataforma', () => {
    expect(reminderToastDetail({ title: 'RAFAEL MENDES - VOTORANTIM, SP' })).toContain(
      'RAFAEL MENDES - VOTORANTIM, SP',
    )
  })
})
