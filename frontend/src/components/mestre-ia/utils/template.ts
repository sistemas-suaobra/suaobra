import { safeStr } from "./obrasPlus"

export function normalizeLeadProperties(properties: any): Record<string, any> {
  if (!properties) return {}
  if (typeof properties === "string") {
    try {
      return JSON.parse(properties)
    } catch {
      return {}
    }
  }
  if (typeof properties === "object") return properties
  return {}
}

export function applyLeadVariables(template: string, vars: Record<string, any>) {
  const nome = (vars.nome ?? vars.nome_contato ?? "").toString()
  const cidade = (vars.cidade ?? vars.city ?? "").toString()
  const bairro = (vars.bairro ?? "").toString()

  console.log("[applyLeadVariables] bairro recebido:", bairro)

  return template
    .replace(/{{\s*nome\s*}}/gi, nome)
    .replace(/{{\s*cidade\s*}}/gi, cidade)
    .replace(/{{\s*bairro\s*}}/gi, bairro)
}