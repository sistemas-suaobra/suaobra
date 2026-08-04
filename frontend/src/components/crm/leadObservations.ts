/**
 * Helpers para persistência de observações do Lead (Venda Mais).
 * Isolados para permitir testes unitários sem montar o Dialog/Editor.
 */

/** Evento típico do PrimeReact Editor / Quill. */
export type EditorHtmlChange = {
  htmlValue?: string | null
  textValue?: string | null
}

/**
 * Quill/PrimeReact frequentemente emite htmlValue=null ao desmontar o Editor
 * (ex.: fechar o Dialog). Se aplicarmos isso no state antes do save no onHide,
 * a observação digitada é apagada e gravada como vazia.
 */
export function normalizeEditorHtmlValue(
  change: EditorHtmlChange,
  previous: string = '',
): string | undefined {
  if (change.htmlValue == null) {
    // null/undefined = artefato de unmount ou evento inválido — manter anterior
    return undefined
  }
  return change.htmlValue
}

/**
 * Monta o payload JSON de lead.properties para o PocketBase.
 * Garante que observations (e demais campos) vão como objeto plain serializável.
 */
export function buildLeadPropertiesUpdatePayload(
  leadProperties: Record<string, unknown> | null | undefined,
  obraPayload: Record<string, unknown> | null | undefined,
): Record<string, unknown> {
  const raw = leadProperties && typeof leadProperties === 'object' ? leadProperties : {}
  let payload: Record<string, unknown>
  try {
    payload = JSON.parse(JSON.stringify(raw)) as Record<string, unknown>
  } catch {
    payload = { ...raw }
  }

  payload.obra = obraPayload && typeof obraPayload === 'object' ? obraPayload : {}
  payload.observations = normalizeObservationsField(payload.observations)
  payload.rating = typeof payload.rating === 'number' ? payload.rating : Number(payload.rating) || 0
  payload.valor = typeof payload.valor === 'number' ? payload.valor : Number(payload.valor) || 0
  payload.starred_contacts = Array.isArray(payload.starred_contacts) ? payload.starred_contacts : []
  payload.contacts = Array.isArray(payload.contacts) ? payload.contacts : []

  return payload
}

export function normalizeObservationsField(value: unknown): string {
  if (value == null) return ''
  if (typeof value === 'string') return value
  return String(value)
}

/** Indica se o HTML do editor tem texto útil (ignora &lt;p&gt;&lt;br&gt;&lt;/p&gt; vazio). */
export function hasMeaningfulObservation(html: string | null | undefined): boolean {
  if (!html) return false
  const text = html
    .replace(/<[^>]*>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/\s+/g, ' ')
    .trim()
  return text.length > 0
}
