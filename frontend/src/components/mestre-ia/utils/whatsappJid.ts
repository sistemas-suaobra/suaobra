/** Extrai só os dígitos do telefone a partir de um JID WhatsApp (ex: 5511...@s.whatsapp.net). */
export function jidToPhoneDigits(jid: string): string {
  if (!jid) return ""

  const local = jid.split("@")[0] ?? ""
  const phonePart = local.split(":")[0] ?? local

  return phonePart.replace(/\D/g, "")
}

/** Formata dígitos BR/E164 para exibição (+55 DDD XXXXX-XXXX). */
export function formatPhoneDigits(digits: string): string {
  const d = digits.replace(/\D/g, "")
  if (!d) return ""

  if (d.length >= 12 && d.startsWith("55")) {
    return `+${d.slice(0, 2)} ${d.slice(2, 4)} ${d.slice(4, 9)}-${d.slice(9)}`
  }

  if (d.length >= 10) {
    return `+${d}`
  }

  return `+${d}`
}

export function formatWhatsappJid(jid: string): string {
  return formatPhoneDigits(jidToPhoneDigits(jid))
}
