import { describe, expect, it } from 'vitest'
import {
  buildLeadPropertiesUpdatePayload,
  hasMeaningfulObservation,
  normalizeEditorHtmlValue,
  normalizeObservationsField,
} from './leadObservations'

describe('normalizeEditorHtmlValue', () => {
  it('ignora htmlValue null para não apagar observação no unmount do Editor', () => {
    expect(normalizeEditorHtmlValue({ htmlValue: null }, '<p>ok</p>')).toBeUndefined()
    expect(normalizeEditorHtmlValue({ htmlValue: undefined }, '<p>ok</p>')).toBeUndefined()
  })

  it('aceita HTML válido digitado pelo usuário', () => {
    expect(normalizeEditorHtmlValue({ htmlValue: '<p>Ligou e pediu orçamento</p>' }, '')).toBe(
      '<p>Ligou e pediu orçamento</p>',
    )
  })

  it('aceita string vazia quando o usuário limpa o campo de propósito', () => {
    expect(normalizeEditorHtmlValue({ htmlValue: '' }, '<p>antes</p>')).toBe('')
  })
})

describe('normalizeObservationsField', () => {
  it('converte null/undefined em string vazia', () => {
    expect(normalizeObservationsField(null)).toBe('')
    expect(normalizeObservationsField(undefined)).toBe('')
  })

  it('preserva string', () => {
    expect(normalizeObservationsField('<p>nota</p>')).toBe('<p>nota</p>')
  })
})

describe('buildLeadPropertiesUpdatePayload', () => {
  it('inclui observations no payload enviado ao PocketBase', () => {
    const payload = buildLeadPropertiesUpdatePayload(
      {
        rating: 5,
        observations: '<p>Cliente pediu retorno amanhã</p>',
        valor: 1000,
        starred_contacts: ['c1'],
        contacts: [{ name: 'A', telephone: '11999999999' }],
        alert_at: 123,
        alerted: false,
        obra: { owner: 'ANTIGO' },
      },
      { owner: 'TALISE', city: 'BRASILIA', state: 'DF' },
    )

    expect(payload.observations).toBe('<p>Cliente pediu retorno amanhã</p>')
    expect(payload.rating).toBe(5)
    expect(payload.valor).toBe(1000)
    expect(payload.obra).toEqual({ owner: 'TALISE', city: 'BRASILIA', state: 'DF' })
    expect(payload.starred_contacts).toEqual(['c1'])
    expect(Array.isArray(payload.contacts)).toBe(true)
  })

  it('não perde observations quando obraPayload substitui obra', () => {
    const payload = buildLeadPropertiesUpdatePayload(
      { observations: '<p>manter</p>', obra: { owner: 'X' } },
      { address: 'Rua Y' },
    )
    expect(payload.observations).toBe('<p>manter</p>')
    expect(payload.obra).toEqual({ address: 'Rua Y' })
  })

  it('normaliza observations null para string vazia (evita gravar null no JSON)', () => {
    const payload = buildLeadPropertiesUpdatePayload({ observations: null }, {})
    expect(payload.observations).toBe('')
  })

  it('funciona com leadProperties vazio/nulo', () => {
    expect(buildLeadPropertiesUpdatePayload(null, null).observations).toBe('')
    expect(buildLeadPropertiesUpdatePayload(undefined, undefined).obra).toEqual({})
  })

  it('serializa como plain object JSON-safe (sem getters de classe)', () => {
    const payload = buildLeadPropertiesUpdatePayload(
      { observations: '<p>x</p>', rating: 1 },
      { city: 'SP' },
    )
    const roundtrip = JSON.parse(JSON.stringify(payload))
    expect(roundtrip.observations).toBe('<p>x</p>')
    expect(roundtrip.obra.city).toBe('SP')
  })
})

describe('hasMeaningfulObservation', () => {
  it('detecta texto real vs editor vazio do Quill', () => {
    expect(hasMeaningfulObservation('<p>oi</p>')).toBe(true)
    expect(hasMeaningfulObservation('<p><br></p>')).toBe(false)
    expect(hasMeaningfulObservation('')).toBe(false)
    expect(hasMeaningfulObservation(null)).toBe(false)
  })
})
