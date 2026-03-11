import { safeStr } from "../utils/obrasPlus"

function normalizeBairros(input: any): string[] {
  if (Array.isArray(input)) {
    return input
      .map((x) => (typeof x === "string" ? x : safeStr(x?.bairro)))
      .filter(Boolean)
  }
  if (typeof input === "string") return input.trim() ? [input.trim()] : []
  return []
}

export function generateCampaignName(now: Date) {
  const d = now.toLocaleDateString("pt-BR")
  const t = now.toLocaleTimeString("pt-BR", { hour: "2-digit", minute: "2-digit" })
  return `Campanha ${d} ${t}`
}

export function buildFinalMessageText(messageText: string, city: string, bairro: any) {
  console.log("Debug bairro input:", bairro); // Adicionado para depuração
  const bairros = normalizeBairros(bairro);
  console.log("Debug bairros após normalização:", bairros); // Adicionado para depuração

  return messageText
    .replace(/{{\s*nome\s*}}/gi, "[NOME]")
    .replace(/{{\s*cidade\s*}}/gi, city || "")
    .replace(/{{\s*bairro\s*}}/gi, bairros.join(", ") || "")
    .trim()
}

export function buildChannels(channelWa: boolean, channelEmail: boolean) {
  const canais: string[] = []
  if (channelWa) canais.push("WHATSAPP")
  if (channelEmail) canais.push("EMAIL")
  return canais
}