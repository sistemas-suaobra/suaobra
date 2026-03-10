import React from "react"
import { Dialog } from "primereact/dialog"
import { MultiSelect } from "primereact/multiselect"
import { Checkbox } from "primereact/checkbox"
import { InputSwitch } from "primereact/inputswitch"
import { InputText } from "primereact/inputtext"
import { InputTextarea } from "primereact/inputtextarea"
import { Button } from "primereact/button"
import { Dropdown } from "primereact/dropdown"
import { Tag } from "primereact/tag"

import type { CreateCampaignDialogProps } from "../types/create-campaign"
import { applyLeadVariables } from "../utils/template"
import { safeStr } from "../utils/obrasPlus"

import { useObrasPlusFilters } from "../hooks/useObrasPlusFilters"
import { useCampaignLeadOptions } from "../hooks/useCampaignLeadOptions"
import { useLeadComms } from "../hooks/useLeadComms"
import { useCreateCampaign } from "../hooks/useCreateCampaign"
import { api, baseURL } from "../../../store/api"

type CursorPos = { start: number; end: number }

export default function CreateCampaignDialog(props: CreateCampaignDialogProps) {
  const {
    visible,
    onClose,
    leadsOptions,
    onCreate,
    notify,
    conexaoWhatsAppId,
    conexaoEmailId,
    teamId,
    userId,
  } = props

  const [selectedLeads, setSelectedLeads] = React.useState<string[]>([])
  const [channelWa, setChannelWa] = React.useState(true)
  const [channelEmail, setChannelEmail] = React.useState(false)
  const [iaContinuar, setIaContinuar] = React.useState(true)
  const [emailSubject, setEmailSubject] = React.useState("")
  const [messageText, setMessageText] = React.useState(
    "Olá, {{nome}}, tudo bem? Podemos conversar sobre sua obra no {{bairro}} em {{cidade}}?"
  )
  const [objetivo, setObjetivo] = React.useState("")
  const [generating, setGenerating] = React.useState(false)

  const cursorRef = React.useRef<CursorPos>({ start: 0, end: 0 })
  const textareaElRef = React.useRef<HTMLTextAreaElement | null>(null)

  const syncCursorFromEvent = (e: any) => {
    const el = (e?.target || e?.currentTarget) as HTMLTextAreaElement | undefined
    if (!el) return
    textareaElRef.current = el
    const start = typeof el.selectionStart === "number" ? el.selectionStart : el.value.length
    const end = typeof el.selectionEnd === "number" ? el.selectionEnd : el.value.length
    cursorRef.current = { start, end }
  }

  const insertVariable = (token: string) => {
    const base = messageText ?? ""
    const { start, end } = cursorRef.current

    const s = Math.max(0, Math.min(start, base.length))
    const e = Math.max(0, Math.min(end, base.length))

    const next = base.slice(0, s) + token + base.slice(e)
    setMessageText(next)

    setTimeout(() => {
      const el = textareaElRef.current
      if (!el) return
      try {
        el.focus()
        const pos = s + token.length
        el.selectionStart = pos
        el.selectionEnd = pos
        cursorRef.current = { start: pos, end: pos }
      } catch {
        //
      }
    }, 0)
  }

  const VARIABLE_BUTTONS: { label: string; token: string }[] = [
    { label: "Nome", token: "{{nome}}" },
    { label: "Cidade", token: "{{cidade}}" },
    { label: "Bairro", token: "{{bairro}}" },
  ]

  const handleGenerate = async () => {
    if (!objetivo) {
      notify("warn", "Objetivo em branco", "Por favor, descreva o objetivo da campanha para a IA.")
      return
    }

    setGenerating(true)

    try {
      const resp = await api().post(`${baseURL()}/campanhas/gerar-mensagem-ia`, { objetivo })
      if (resp.error) throw new Error(resp.error)

      const data = await resp.response.json()
      setMessageText(data.mensagem)
      notify("success", "Mensagem gerada", "A IA gerou uma nova sugestão de mensagem.")
    } catch (error: any) {
      const detail =
        error?.response?.data?.message || error?.message || "Não foi possível gerar a mensagem."
      notify("error", "Erro de IA", detail)
    } finally {
      setGenerating(false)
    }
  }

  const {
    citiesOptions,
    selectedCity,
    onCityChange,
    selectedNeighborhood,
    setSelectedNeighborhood,
    neighborhoodsOptions,
    filterValue,
    setFilterValue,
    startDateFrom,
    setStartDateFrom,
    startDateTo,
    setStartDateTo,
    endDateFrom,
    setEndDateFrom,
    endDateTo,
    setEndDateTo,
  } = useObrasPlusFilters(visible)

  const { leadsOptionsLocal, obraRecordMapRef, existingLeadSet } = useCampaignLeadOptions({
    visible,
    teamId,
    selectedCity,
    selectedNeighborhood,
    filterValue,
    startDateFrom,
    startDateTo,
    endDateFrom,
    endDateTo,
    fallbackOptions: leadsOptions,
  })

  const { fetchLeadComms } = useLeadComms(selectedCity, obraRecordMapRef)
  const { saving, createCampaign } = useCreateCampaign({ teamId, userId, notify })

  React.useEffect(() => {
    if (!visible) return

    setSelectedLeads([])
    setChannelWa(true)
    setChannelEmail(false)
    setIaContinuar(true)
    setEmailSubject("")
    setMessageText("Olá, {{nome}}, tudo bem? Podemos conversar sobre sua obra no {{bairro}} em {{cidade}}?")
    setObjetivo("")
    cursorRef.current = { start: 0, end: 0 }
    textareaElRef.current = null
  }, [visible])

  const getCidadeBairroStr = React.useCallback(() => {
    const obraId = selectedLeads[0]
    const rec = obraId ? obraRecordMapRef.current.get(obraId) : null

    const cidadeStr = safeStr(rec?.city) || safeStr(selectedCity?.city) || ""
    const bairroStr = safeStr(rec?.bairro) || safeStr(selectedNeighborhood?.[0]?.bairro) || ""

    return { cidadeStr, bairroStr, rec }
  }, [selectedLeads, selectedCity, selectedNeighborhood, obraRecordMapRef])

  const previewText = React.useMemo(() => {
    const obraId = selectedLeads[0]

    if (!obraId) {
      return applyLeadVariables(messageText, {
        nome: "Cliente",
        cidade: "",
        bairro: "",
        nome_contato: "Cliente",
        city: "",
      })
    }

    const { cidadeStr, bairroStr, rec } = getCidadeBairroStr()
    const label = leadsOptionsLocal.find((o) => o.value === obraId)?.label || "Cliente"
    const nomeStr = safeStr(rec?.owner) || safeStr(rec?.professional) || safeStr(label) || "Cliente"

    const vars: any = {
      nome: nomeStr,
      cidade: cidadeStr,
      bairro: bairroStr,
      nome_contato: nomeStr,
      city: cidadeStr,
    }

    return applyLeadVariables(messageText, vars)
  }, [messageText, selectedLeads, leadsOptionsLocal, getCidadeBairroStr])

  const footer = (
    <div className="flex flex-column md:flex-row justify-content-end gap-2 w-full">
      <Button
        label="Cancelar"
        icon="pi pi-times"
        severity="secondary"
        onClick={onClose}
        disabled={saving}
        className="w-full md:w-auto"
      />
      <Button
        label="Criar campanha"
        icon="pi pi-check"
        loading={saving}
        className="w-full md:w-auto"
        onClick={() => {
          const { cidadeStr, bairroStr } = getCidadeBairroStr()

          createCampaign({
            selectedLeads,
            channelWa,
            channelEmail,
            iaContinuar,
            emailSubject,
            messageText,
            selectedCity,
            cidade: cidadeStr,
            bairro: bairroStr,
            conexaoWhatsAppId,
            conexaoEmailId,
            obraRecordMapRef,
            fetchLeadComms,
            onCreate,
            onClose,
          } as any)
        }}
      />
    </div>
  )

  return (
    <Dialog
      header="Criar nova campanha"
      visible={visible}
      onHide={onClose}
      draggable={false}
      dismissableMask
      modal
      breakpoints={{ "1100px": "94vw", "768px": "96vw", "560px": "100vw" }}
      style={{ width: "92vw", maxWidth: "980px" }}
      contentStyle={{ paddingBottom: "1rem" }}
      footer={footer}
    >
      <div className="formgrid grid">
        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Cidade</label>
          <Dropdown
            className="w-full"
            value={selectedCity}
            options={citiesOptions}
            onChange={(e) => {
              onCityChange(e.value)
              setSelectedNeighborhood([])
            }}
            optionLabel="city"
            filter
            placeholder="Selecione uma cidade"
            emptyMessage="Nenhuma cidade"
          />
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Bairro</label>
          <MultiSelect
            className="w-full"
            value={selectedNeighborhood}
            options={neighborhoodsOptions}
            onChange={(e) => setSelectedNeighborhood(e.value)}
            optionLabel="bairro"
            filter
            placeholder="Todos os bairros"
            maxSelectedLabels={2}
          />
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Palavra Chave</label>
          <div className="p-inputgroup w-full">
            <InputText
              className="w-full"
              placeholder="Nome, endereço..."
              value={filterValue}
              onChange={(e) => setFilterValue(e.target.value)}
            />
            <Button icon="pi pi-search" type="button" />
          </div>
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Data de Início (De)</label>
          <InputText
            className="w-full"
            placeholder="Selecione uma data"
            value={startDateFrom}
            onChange={(e) => setStartDateFrom(e.target.value)}
          />
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Data de Início (Até)</label>
          <InputText
            className="w-full"
            placeholder="Selecione uma data"
            value={startDateTo}
            onChange={(e) => setStartDateTo(e.target.value)}
          />
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Data de Fim (De)</label>
          <InputText
            className="w-full"
            placeholder="Selecione uma data"
            value={endDateFrom}
            onChange={(e) => setEndDateFrom(e.target.value)}
          />
        </div>

        <div className="field col-12 md:col-6 xl:col-3 mb-0">
          <label>Data de Fim (Até)</label>
          <InputText
            className="w-full"
            placeholder="Selecione uma data"
            value={endDateTo}
            onChange={(e) => setEndDateTo(e.target.value)}
          />
        </div>

        <div className="field col-12">
          <label style={{ fontWeight: 700 }}>Leads ({leadsOptionsLocal.length} disponíveis)</label>
          <MultiSelect
            value={selectedLeads}
            options={leadsOptionsLocal}
            onChange={(e) => setSelectedLeads(e.value as string[])}
            placeholder="Selecione leads para esta campanha"
            className="w-full"
            filter
            maxSelectedLabels={3}
            emptyFilterMessage="Nenhum lead encontrado."
            itemTemplate={(opt: any) => (
              <div className="flex align-items-center gap-2 flex-wrap">
                <span>{opt.label}</span>
                {existingLeadSet.has(opt.value) ? (
                  <Tag value="Já é lead" severity="info" style={{ fontSize: 11 }} />
                ) : null}
              </div>
            )}
          />
          <div className="text-secondary" style={{ fontSize: 12, marginTop: 4 }}>
            {selectedLeads.length} lead(s) selecionado(s) {channelWa ? "• WhatsApp máx. 50" : ""}
          </div>
        </div>

        <div className="field col-12 lg:col-6">
          <label style={{ fontWeight: 700 }}>Canais</label>
          <div className="flex flex-column md:flex-row gap-3">
            <div className="flex align-items-center gap-2">
              <Checkbox
                inputId="wa"
                checked={channelWa}
                onChange={(e) => setChannelWa(e.checked ?? false)}
              />
              <label htmlFor="wa">WhatsApp</label>
            </div>

            <div className="flex align-items-center gap-2">
              <Checkbox
                inputId="email"
                checked={channelEmail}
                onChange={(e) => setChannelEmail(e.checked ?? false)}
              />
              <label htmlFor="email">E-mail</label>
            </div>
          </div>
        </div>

        <div className="field col-12 lg:col-6">
          <label style={{ fontWeight: 700 }}>Manter IA após resposta</label>
          <div className="flex align-items-center justify-content-between gap-3 flex-wrap">
            <InputSwitch checked={iaContinuar} onChange={(e) => setIaContinuar(!!e.value)} />
            <div className="text-secondary" style={{ flex: 1, minWidth: 220 }}>
              {iaContinuar ? "IA continua a conversa após resposta" : "IA não continuará"}
            </div>
          </div>
        </div>

        {channelEmail ? (
          <div className="field col-12">
            <label style={{ fontWeight: 700 }}>Assunto (E-mail)</label>
            <InputText
              value={emailSubject}
              onChange={(e) => setEmailSubject(e.currentTarget.value)}
              className="w-full"
            />
          </div>
        ) : null}

        <div className="field col-12">
          <div style={{ fontSize: 16, fontWeight: 700, marginBottom: 8 }}>Mensagem</div>

          <div className="flex align-items-center gap-2 flex-wrap" style={{ marginBottom: 12 }}>
            <span className="text-secondary" style={{ fontSize: 13 }}>
              Variáveis:
            </span>

            {VARIABLE_BUTTONS.map((v) => (
              <Button
                key={v.token}
                type="button"
                label={v.label}
                icon="pi pi-plus"
                className="p-button-sm p-button-outlined"
                onClick={() => insertVariable(v.token)}
              />
            ))}

            <span className="text-secondary" style={{ fontSize: 12 }}>
              (CLIQUE PARA ENTRAR NO TEXTO)
            </span>
          </div>

          <label style={{ fontWeight: 600, fontSize: 13 }}>Objetivo da campanha (para IA)</label>

          <div className="flex flex-column md:flex-row gap-2">
            <InputText
              value={objetivo}
              onChange={(e) => setObjetivo(e.currentTarget.value)}
              placeholder="Ex: Vender 20% a mais de cimento este mês para clientes de obras residenciais."
              className="w-full"
            />
            <Button
              icon="pi pi-sparkles"
              label="Gerar com IA"
              onClick={handleGenerate}
              loading={generating}
              type="button"
              className="w-full md:w-auto"
            />
          </div>

          <label style={{ fontWeight: 600, fontSize: 13, marginTop: 12, display: "block" }}>
            Texto da Mensagem
          </label>

          <InputTextarea
            value={messageText}
            onChange={(e: any) => {
              syncCursorFromEvent(e)
              setMessageText(e.target.value)
            }}
            onFocus={syncCursorFromEvent}
            onClick={syncCursorFromEvent}
            onKeyUp={syncCursorFromEvent}
            onSelect={syncCursorFromEvent}
            rows={5}
            className="w-full"
            autoResize
          />

          <div style={{ marginTop: 12 }}>
            <div style={{ marginBottom: 6, fontWeight: 600, fontSize: 13 }}>Pré-visualização</div>
            <div
              className="p-3 border-round-xl bg-white"
              style={{
                border: "1px solid rgba(0,0,0,0.08)",
                minHeight: 80,
                whiteSpace: "pre-wrap",
                wordBreak: "break-word",
              }}
            >
              {previewText}
            </div>
          </div>
        </div>
      </div>
    </Dialog>
  )
}