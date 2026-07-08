import React from "react"
import type { NotifyFn } from "../types/create-campaign"
import { PB, api, baseURL } from "../../../store/api"
import { user } from "../../../store/store"
import { buildChannels, generateCampaignName } from "../utils/campaign"

type RecipientKind = "OWNER" | "PROFISSIONAL"

type CreateCampaignArgs = {
  destinatarios: Array<{
    obra_id: string
    contato_tipo: RecipientKind | "PROFESSIONAL"
  }>
  channelWa: boolean
  channelEmail: boolean
  iaContinuar: boolean
  emailSubject: string
  messageText: string
  selectedCity?: any
  cidade?: string
  bairro?: string
  conexaoWhatsAppId?: string
  conexaoEmailId?: string
  ocultarJaContactados: boolean
  onCreate: (created: any) => void
  onClose: () => void
}

const normalizeRecipientKind = (value: unknown): RecipientKind | null => {
  const kind = String(value ?? "")
    .toUpperCase()
    .trim()

  if (kind === "OWNER") return "OWNER"
  if (kind === "PROFISSIONAL" || kind === "PROFESSIONAL") return "PROFISSIONAL"

  return null
}

export function useCreateCampaign(params: { teamId: string; userId: string; notify: NotifyFn }) {
  const { teamId, userId, notify } = params
  const [saving, setSaving] = React.useState(false)

  const createCampaign = React.useCallback(
    async (args: CreateCampaignArgs) => {
      const {
        destinatarios,
        channelWa,
        channelEmail,
        iaContinuar,
        emailSubject,
        messageText,
        conexaoWhatsAppId,
        conexaoEmailId,
        ocultarJaContactados,
        onCreate,
        onClose,
      } = args

      const destinatariosValidos = Array.isArray(destinatarios)
        ? destinatarios
            .map((d) => {
              const obraId = String(d?.obra_id ?? "").trim()
              const contatoTipo = normalizeRecipientKind(d?.contato_tipo)

              if (!obraId || !contatoTipo) return null

              return {
                obra_id: obraId,
                contato_tipo: contatoTipo as RecipientKind,
              }
            })
            .filter(Boolean) as Array<{
            obra_id: string
            contato_tipo: RecipientKind
          }>
        : []

      if (!channelWa && !channelEmail) {
        return notify("warn", "Canais", "Selecione ao menos um canal (WhatsApp ou E-mail).")
      }

      if (channelEmail && !emailSubject.trim()) {
        return notify("warn", "Assunto", "Informe o assunto do e-mail.")
      }

      if (!destinatariosValidos.length) {
        return notify("warn", "Destinatários", "Selecione pelo menos 1 destinatário.")
      }

      if (channelWa && destinatariosValidos.length > 50) {
        return notify(
          "warn",
          "Destinatários",
          `WhatsApp permite até 50 destinatários por disparo. Você selecionou ${destinatariosValidos.length}.`
        )
      }

      if (!messageText.trim()) {
        return notify("warn", "Mensagem", "A mensagem é obrigatória.")
      }

      setSaving(true)

      try {
        const pb = PB()
        pb.authStore.save(user.get().token, user.get())

        const now = new Date()
        const nome = generateCampaignName(now)
        const canais = buildChannels(channelWa, channelEmail)
        const template = messageText.trim()

        const campanhaPayload: any = {
          team_id: teamId,
          nome,
          canal: canais,
          status: "RASCUNHO",
          mensagem_template: template,
          assunto_email: channelEmail ? emailSubject.trim() : "",
          criado_por: userId,
          manter_ia: iaContinuar,
        }

        if (conexaoWhatsAppId) campanhaPayload.conexao_whatsapp_id = conexaoWhatsAppId
        if (conexaoEmailId) campanhaPayload.conexao_email_id = conexaoEmailId

        const campanhaRecord = await pb.collection("campanhas").create(campanhaPayload)

        const resp = await api().post(
          `${baseURL()}/campanhas/${campanhaRecord.id}/destinatarios/obras-plus`,
          {
            destinatarios: destinatariosValidos,
            ocultar_ja_contactados: ocultarJaContactados,
          }
        )

        if (resp?.error) {
          throw new Error(resp.error)
        }

        const json = await resp.response.json().catch(() => null)
        const criados = Number(json?.criados || 0)
        const ignorados = Number(json?.ignorados || 0)

        if (criados <= 0) {
          try {
            await pb.collection("campanhas").delete(campanhaRecord.id)
          } catch (rollbackErr) {
            console.warn("Falha ao remover campanha sem destinatários", rollbackErr)
          }

          throw new Error(
            "Nenhum destinatário válido foi criado para a campanha. Revise os contatos selecionados e tente novamente."
          )
        }

        notify(
          "success",
          "Campanha criada",
          `Campanha criada com ${criados} destinatário(s)${
            ignorados ? ` e ${ignorados} ignorado(s)` : ""
          }.`
        )

        onCreate({
          id: campanhaRecord.id,
          team_id: teamId,
          nome,
          status: "RASCUNHO",
          mensagem_template: template,
          criado_por: userId,
          assunto_email: channelEmail ? emailSubject.trim() : "",
          manter_ia: iaContinuar,
          canal: canais,
          created: campanhaRecord.created,
          updated: campanhaRecord.updated,
          destinatarios: destinatariosValidos,
          channelWa,
          channelEmail,
          ocultarJaContactados,
        })

        onClose()
      } catch (e: any) {
        console.error(e)
        notify("error", "Erro", e?.message || "Erro ao criar campanha. Tente novamente.")
      } finally {
        setSaving(false)
      }
    },
    [teamId, userId, notify]
  )

  return { saving, createCampaign }
}