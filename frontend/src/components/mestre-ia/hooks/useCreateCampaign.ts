import React from "react"
import type { NotifyFn, ObrasPlusRecord, LeadComms } from "../types/create-campaign"
import { PB, makeURL, api } from "../../../store/api"
import { user } from "../../../store/store"
import { buildChannels, generateCampaignName } from "../utils/campaign"
import { safeStr } from "../utils/obrasPlus"
import { fetchLeadIdMap } from "../utils/pb"

export function useCreateCampaign(params: { teamId: string; userId: string; notify: NotifyFn }) {
  const { teamId, userId, notify } = params
  const [saving, setSaving] = React.useState(false)

  const toggleFavorite = async (obraId: string) => {
    const resp = await api().patch(makeURL("/patch/lead-toggle"), {
      team_id: teamId,
      obra_id: obraId,
      toggle_col: "favorited_at",
    })
    if (resp?.error) throw new Error(resp.error)
  }

  const createCampaign = React.useCallback(
    async (args: {
      selectedLeads: string[]
      channelWa: boolean
      channelEmail: boolean
      iaContinuar: boolean
      emailSubject: string
      messageText: string
      selectedCity: any
      selectedNeighborhood: any[] // MultiSelect retorna array de objetos
      conexaoWhatsAppId?: string
      conexaoEmailId?: string
      obraRecordMapRef: React.MutableRefObject<Map<string, ObrasPlusRecord>>
      fetchLeadComms: (obraId: string) => Promise<LeadComms>
      onCreate: (created: any) => void
      onClose: () => void
    }) => {
      const {
        selectedLeads,
        channelWa,
        channelEmail,
        iaContinuar,
        emailSubject,
        messageText,
        selectedCity,
        selectedNeighborhood,
        conexaoWhatsAppId,
        conexaoEmailId,
        obraRecordMapRef,
        fetchLeadComms,
        onCreate,
        onClose,
      } = args

      if (!channelWa && !channelEmail) return notify("warn", "Canais", "Selecione ao menos um canal (WhatsApp ou E-mail).")
      if (channelEmail && !emailSubject.trim()) return notify("warn", "Assunto", "Informe o assunto do e-mail.")
      if (!selectedLeads.length) return notify("warn", "Leads", "Selecione pelo menos 1 lead.")
      if (channelWa && selectedLeads.length > 50)
        return notify("warn", "Leads", "WhatsApp permite até 50 leads por disparo. Você selecionou " + selectedLeads.length + ".")
      if (!messageText.trim()) return notify("warn", "Mensagem", "A mensagem é obrigatória.")

      // ✅ bairro do filtro (1 só, como você quer)
      const bairroFiltroStr =
        safeStr(selectedNeighborhood?.[0]?.bairro) || ""

      setSaving(true)
      let campanhaId: string | null = null

      try {
        const pb = PB()
        pb.authStore.save(user.get().token, user.get())

        const now = new Date()
        const nome = generateCampaignName(now)
        const canais = buildChannels(channelWa, channelEmail)
        const conexaoId = channelWa ? conexaoWhatsAppId : conexaoEmailId
        const obraIds = (channelWa ? selectedLeads.slice(0, 50) : selectedLeads).filter(Boolean)
        const template = messageText.trim()

        const commsByObra = new Map<string, LeadComms>()
        for (const obraId of obraIds) commsByObra.set(obraId, await fetchLeadComms(obraId))

        const leadIdMapBefore = await fetchLeadIdMap(pb, teamId, obraIds)
        const missing = obraIds.filter((id) => !leadIdMapBefore.has(id))
        for (const obraId of missing) await toggleFavorite(obraId)

        const leadIdMap = await fetchLeadIdMap(pb, teamId, obraIds)

        for (const obraId of obraIds) {
          const leadId = leadIdMap.get(obraId)
          if (!leadId) continue

          const rec = obraRecordMapRef.current.get(obraId)
          const comms = commsByObra.get(obraId)

          const props = {
            email: safeStr(comms?.email),
            telefone_e164: safeStr(comms?.telefone_e164),
            nome_contato: safeStr(comms?.nome_contato) || "Cliente",
            obra_id: obraId,
            address: safeStr(rec?.address),

            // ✅ AQUI: bairro sempre string, com fallback do filtro
            bairro: safeStr(rec?.bairro) || bairroFiltroStr,

            // cidade já tava ok
            city: safeStr(rec?.city) || safeStr(selectedCity?.city),
            state: safeStr(rec?.state) || safeStr(selectedCity?.state),
            source: "obras-plus",
          }

          await pb.collection("lead").update(leadId, { properties: props })
        }

        const validObras = obraIds.filter((obraId) => {
          const c = commsByObra.get(obraId)
          if (!c) return false
          if (channelWa && !safeStr(c.telefone_e164)) return false
          if (channelEmail && !safeStr(c.email)) return false
          return true
        })

        if (!validObras.length) return notify("warn", "Leads", "Nenhum lead possui telefone/e-mail suficiente para os canais escolhidos.")

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
        if (conexaoId) campanhaPayload.conexao_id = conexaoId

        const campanhaRecord = await pb.collection("campanhas").create(campanhaPayload)
        campanhaId = campanhaRecord.id

        let destinatariosCriados = 0
        for (const obraId of validObras) {
          const leadId = leadIdMap.get(obraId)
          const comms = commsByObra.get(obraId)
          if (!leadId || !comms) continue

          const destPayload: any = {
            team_id: teamId,
            campanha_id: campanhaRecord.id,
            lead_id: leadId,
            status: "PENDENTE",
            tentativas: 0,
            nome_contato: comms.nome_contato || "Cliente",
          }
          if (safeStr(comms.telefone_e164)) destPayload.telefone_e164 = comms.telefone_e164
          if (safeStr(comms.email)) destPayload.email = comms.email

          await pb.collection("campanha_destinatarios").create(destPayload)
          destinatariosCriados++
        }

        notify("success", "Campanha criada", `Campanha criada com ${destinatariosCriados} destinatário(s)`)

        onCreate({
          id: campanhaRecord.id,
          team_id: teamId,
          nome,
          conexao_id: conexaoId || "",
          status: "RASCUNHO",
          mensagem_template: template,
          criado_por: userId,
          iniciado_em: campanhaRecord.iniciado_em,
          finalizado_em: campanhaRecord.finalizado_em,
          created: campanhaRecord.created,
          updated: campanhaRecord.updated,
          leads: validObras,
          channelWa,
          channelEmail,
          iaContinuar,
        })

        onClose()
      } catch (e) {
        console.error(e)
        if (!campanhaId) notify("error", "Erro", "Erro ao criar campanha. Tente novamente.")
        else args.onClose()
      } finally {
        setSaving(false)
      }
    },
    [teamId, userId, notify]
  )

  return { saving, createCampaign }
}