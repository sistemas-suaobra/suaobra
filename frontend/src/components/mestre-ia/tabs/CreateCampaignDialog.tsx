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
import { useCreateCampaign } from "../hooks/useCreateCampaign"
import { api, baseURL } from "../../../store/api"

type CursorPos = { start: number; end: number }
type RecipientKind = "OWNER" | "PROFISSIONAL"
type RecipientFilter = "ALL" | RecipientKind

type RecipientOption = {
  value: string
  obraId: string
  contatoTipo: RecipientKind
  nomeContato: string
  label: string
  cidade: string
  bairro: string
  uf: string
  address: string
  hasPhone: boolean
  hasEmail: boolean
}

const normalizeRecipientKind = (value: unknown): RecipientKind | null => {
  const kind = String(value ?? "")
    .toUpperCase()
    .trim()

  if (kind === "OWNER") return "OWNER"
  if (kind === "PROFISSIONAL" || kind === "PROFESSIONAL") return "PROFISSIONAL"

  return null
}

const makeRecipientValue = (obraId: string, contatoTipo: RecipientKind) =>
  `${obraId}::${contatoTipo}`

const parseRecipientValue = (
  value: string
): { obra_id: string; contato_tipo: RecipientKind } | null => {
  if (!value) return null

  const [obraId, contatoTipoRaw] = String(value).split("::")
  if (!obraId) return null

  const contatoTipo = normalizeRecipientKind(contatoTipoRaw)
  if (!contatoTipo) return null

  return {
    obra_id: obraId,
    contato_tipo: contatoTipo,
  }
}

function formatLocation(bairro?: string, cidade?: string, uf?: string) {
  const parts: string[] = []

  if (safeStr(bairro)) parts.push(safeStr(bairro))
  if (safeStr(cidade)) parts.push(safeStr(cidade))

  const base = parts.join(" - ")
  const state = safeStr(uf)

  if (!base && !state) return ""
  if (!base) return state
  if (!state) return base

  return `${base}/${state}`
}

function getRecipientTypeLabel(kind: RecipientKind) {
  return kind === "OWNER" ? "Proprietário" : "Profissional"
}

function getRecipientChannelsLabel(hasPhone: boolean, hasEmail: boolean) {
  const channels = [hasPhone ? "telefone" : "", hasEmail ? "email" : ""].filter(Boolean)

  if (!channels.length) return "contato a resolver"

  return channels.join(" + ")
}

function getPlaceholderByFilter(filter: RecipientFilter) {
  if (filter === "OWNER") return "Selecione proprietários para esta campanha"
  if (filter === "PROFISSIONAL") return "Selecione profissionais para esta campanha"
  return "Selecione destinatários para esta campanha"
}

function recipientHasAnyContact(
  option: Pick<RecipientOption, "hasPhone" | "hasEmail">
): boolean {
  return option.hasPhone || option.hasEmail
}

function recipientMatchesChannels(
  option: Pick<RecipientOption, "hasPhone" | "hasEmail">,
  channelWa: boolean,
  channelEmail: boolean
): boolean {
  if (!option.hasPhone && !option.hasEmail) return false
  if (channelWa && !channelEmail) return option.hasPhone
  if (channelEmail && !channelWa) return option.hasEmail
  return option.hasPhone || option.hasEmail
}

function describeRecipientContactGap(
  option: Pick<RecipientOption, "nomeContato" | "hasPhone" | "hasEmail">,
  channelWa: boolean,
  channelEmail: boolean
): string {
  if (!option.hasPhone && !option.hasEmail) {
    return `${option.nomeContato} (sem telefone e sem e-mail)`
  }
  if (channelWa && channelEmail) {
    if (!option.hasPhone && !option.hasEmail) {
      return `${option.nomeContato} (sem telefone e sem e-mail)`
    }
    if (!option.hasPhone) return `${option.nomeContato} (sem telefone)`
    if (!option.hasEmail) return `${option.nomeContato} (sem e-mail)`
  }
  if (channelWa && !option.hasPhone) {
    return `${option.nomeContato} (sem telefone para WhatsApp)`
  }
  if (channelEmail && !option.hasEmail) {
    return `${option.nomeContato} (sem e-mail)`
  }
  return option.nomeContato
}

export default function CreateCampaignDialog(props: CreateCampaignDialogProps) {
  const {
    visible,
    onClose,
    onCreate,
    notify,
    conexaoWhatsAppId,
    conexaoEmailId,
    teamId,
    userId,
  } = props

  const [selectedRecipients, setSelectedRecipients] = React.useState<string[]>([])
  const [recipientFilter, setRecipientFilter] = React.useState<RecipientFilter>("ALL")
  const [channelWa, setChannelWa] = React.useState(true)
  const [channelEmail, setChannelEmail] = React.useState(false)
  const [iaContinuar, setIaContinuar] = React.useState(true)
  const [emailSubject, setEmailSubject] = React.useState("")
  const [messageText, setMessageText] = React.useState(
    "Olá, {{primeiroNome}}, tudo bem? Podemos conversar sobre sua obra no {{bairro}} em {{cidade}}?"
  )
  const [objetivo, setObjetivo] = React.useState("")
  const [generating, setGenerating] = React.useState(false)
  const [ocultarJaContactados, setOcultarJaContactados] = React.useState(true)

  const cursorRef = React.useRef<CursorPos>({ start: 0, end: 0 })
  const textareaElRef = React.useRef<HTMLTextAreaElement | null>(null)

  const syncCursorFromEvent = (e: any) => {
    const el = (e?.target || e?.currentTarget) as HTMLTextAreaElement | undefined
    if (!el) return

    textareaElRef.current = el

    const start =
      typeof el.selectionStart === "number" ? el.selectionStart : el.value.length
    const end =
      typeof el.selectionEnd === "number" ? el.selectionEnd : el.value.length

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
    { label: "1º Nome", token: "{{primeiroNome}}" },
    { label: "Cidade", token: "{{cidade}}" },
    { label: "Bairro", token: "{{bairro}}" },
  ]

  const handleGenerate = async () => {
    if (!objetivo.trim()) {
      notify("warn", "Objetivo em branco", "Descreva o objetivo da campanha para a IA.")
      return
    }

    setGenerating(true)

    try {
      const resp = await api().post(`${baseURL()}/campanhas/gerar-mensagem-ia`, {
        objetivo: objetivo.trim(),
      })

      if (resp.error) throw new Error(resp.error)

      const data = await resp.response.json()
      setMessageText(data?.mensagem || "")
      notify("success", "Mensagem gerada", "A IA gerou uma nova sugestão de mensagem.")
    } catch (error: any) {
      const detail =
        error?.response?.data?.message ||
        error?.message ||
        "Não foi possível gerar a mensagem."

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

  const { obraRecords, contactedRecipientSet, loading: loadingLeads, fetchError } =
    useCampaignLeadOptions({
    visible,
    teamId,
    selectedCity,
    selectedNeighborhood,
    filterValue,
    startDateFrom,
    startDateTo,
    endDateFrom,
    endDateTo,
    ocultarJaContactados,
  })

  const { saving, createCampaign } = useCreateCampaign({ teamId, userId, notify })

  React.useEffect(() => {
    if (!visible) return

    setSelectedRecipients([])
    setRecipientFilter("ALL")
    setChannelWa(true)
    setChannelEmail(false)
    setIaContinuar(true)
    setEmailSubject("")
    setMessageText(
      "Olá, {{primeiroNome}}, tudo bem? Podemos conversar sobre sua obra no {{bairro}} em {{cidade}}?"
    )
    setObjetivo("")
    setOcultarJaContactados(false)
    cursorRef.current = { start: 0, end: 0 }
    textareaElRef.current = null
  }, [visible])

  const recipientOptionsLocal = React.useMemo<RecipientOption[]>(() => {
    const list: RecipientOption[] = []
    const seen = new Set<string>()

    for (const rec of obraRecords ?? []) {
      const obraId = String((rec as any)?.obra_id ?? "").trim()
      if (!obraId) continue

      const cidade = safeStr((rec as any)?.city)
      const bairro = safeStr((rec as any)?.bairro)
      const uf = safeStr((rec as any)?.state)
      const address = safeStr((rec as any)?.address)

      const ownerName = safeStr((rec as any)?.owner)
      const professionalName = safeStr((rec as any)?.professional)

      const ownerHasPhone = !!(rec as any)?.has_owner_phone
      const ownerHasEmail = !!(rec as any)?.has_owner_email
      const professionalHasPhone = !!(rec as any)?.has_professional_phone
      const professionalHasEmail = !!(rec as any)?.has_professional_email

      if (ownerName) {
        const value = makeRecipientValue(obraId, "OWNER")
        const ownerJaContactado = contactedRecipientSet.has(value)

        if (!seen.has(value) && (!ocultarJaContactados || !ownerJaContactado)) {
          const location = formatLocation(bairro, cidade, uf)
          const channels = getRecipientChannelsLabel(ownerHasPhone, ownerHasEmail)

          const option: RecipientOption = {
            value,
            obraId,
            contatoTipo: "OWNER",
            nomeContato: ownerName,
            cidade,
            bairro,
            uf,
            address,
            hasPhone: ownerHasPhone,
            hasEmail: ownerHasEmail,
            label: `${ownerName} (Proprietário)${location ? ` — ${location}` : ""}${
              address ? ` • ${address}` : ""
            } • ${channels}`,
          }

          list.push(option)
          seen.add(value)
        }
      }

      if (professionalName) {
        const value = makeRecipientValue(obraId, "PROFISSIONAL")
        const professionalJaContactado = contactedRecipientSet.has(value)

        if (!seen.has(value) && (!ocultarJaContactados || !professionalJaContactado)) {
          const location = formatLocation(bairro, cidade, uf)
          const channels = getRecipientChannelsLabel(
            professionalHasPhone,
            professionalHasEmail
          )

          const option: RecipientOption = {
            value,
            obraId,
            contatoTipo: "PROFISSIONAL",
            nomeContato: professionalName,
            cidade,
            bairro,
            uf,
            address,
            hasPhone: professionalHasPhone,
            hasEmail: professionalHasEmail,
            label: `${professionalName} (Profissional)${
              location ? ` — ${location}` : ""
            }${address ? ` • ${address}` : ""} • ${channels}`,
          }

          list.push(option)
          seen.add(value)
        }
      }
    }

    return list
  }, [obraRecords, contactedRecipientSet, ocultarJaContactados])

  const recipientOptionMap = React.useMemo(() => {
    const map = new Map<string, RecipientOption>()

    for (const item of recipientOptionsLocal) {
      map.set(item.value, item)
    }

    return map
  }, [recipientOptionsLocal])

  React.useEffect(() => {
    setSelectedRecipients((prev) => {
      const next = prev.filter((value) => {
        const item = recipientOptionMap.get(value)
        if (!item) return false
        return recipientMatchesChannels(item, channelWa, channelEmail)
      })
      if (next.length > 50) return next.slice(0, 50)
      return next
    })
  }, [recipientOptionMap, channelWa, channelEmail])

  const recipientStats = React.useMemo(() => {
    let owners = 0
    let professionals = 0

    for (const item of recipientOptionsLocal) {
      if (item.contatoTipo === "OWNER") owners++
      if (item.contatoTipo === "PROFISSIONAL") professionals++
    }

    return {
      owners,
      professionals,
      total: recipientOptionsLocal.length,
    }
  }, [recipientOptionsLocal])

  const filteredRecipientOptions = React.useMemo(() => {
    if (recipientFilter === "ALL") return recipientOptionsLocal
    return recipientOptionsLocal.filter((item) => item.contatoTipo === recipientFilter)
  }, [recipientFilter, recipientOptionsLocal])

  const selectedRecipientStats = React.useMemo(() => {
    let owners = 0
    let professionals = 0

    for (const value of selectedRecipients) {
      const item = recipientOptionMap.get(value)
      if (!item) continue

      if (item.contatoTipo === "OWNER") owners++
      if (item.contatoTipo === "PROFISSIONAL") professionals++
    }

    return {
      owners,
      professionals,
      total: selectedRecipients.length,
    }
  }, [selectedRecipients, recipientOptionMap])

  const handleRecipientFilterChange = React.useCallback(
    (filter: RecipientFilter) => {
      setRecipientFilter(filter)

      if (filter === "ALL") return

      setSelectedRecipients((prev) =>
        prev.filter((value) => recipientOptionMap.get(value)?.contatoTipo === filter)
      )
    },
    [recipientOptionMap]
  )

  const clearRecipients = React.useCallback(() => {
    setSelectedRecipients([])
  }, [])

  const getCidadeBairroStr = React.useCallback(() => {
    const firstRecipientValue = selectedRecipients[0]
    const selectedRecipient = firstRecipientValue
      ? recipientOptionMap.get(firstRecipientValue)
      : null

    const cidadeStr =
      safeStr(selectedRecipient?.cidade) || safeStr(selectedCity?.city) || ""

    const bairroStr =
      safeStr(selectedRecipient?.bairro) ||
      safeStr(selectedNeighborhood?.[0]?.bairro) ||
      ""

    return { cidadeStr, bairroStr, selectedRecipient }
  }, [selectedRecipients, recipientOptionMap, selectedCity, selectedNeighborhood])

  const previewText = React.useMemo(() => {
    const firstRecipientValue = selectedRecipients[0]
    const selectedRecipient = firstRecipientValue
      ? recipientOptionMap.get(firstRecipientValue)
      : null

    if (!selectedRecipient) {
      return applyLeadVariables(messageText, {
        nome: "Cliente",
        cidade: "",
        bairro: "",
        nome_contato: "Cliente",
        city: "",
      })
    }

    return applyLeadVariables(messageText, {
      nome: safeStr(selectedRecipient.nomeContato) || "Cliente",
      cidade: safeStr(selectedRecipient.cidade),
      bairro: safeStr(selectedRecipient.bairro),
      nome_contato: safeStr(selectedRecipient.nomeContato) || "Cliente",
      city: safeStr(selectedRecipient.cidade),
    })
  }, [messageText, selectedRecipients, recipientOptionMap])

  const previewContactName = React.useMemo(() => {
    const firstRecipientValue = selectedRecipients[0]
    const selectedRecipient = firstRecipientValue
      ? recipientOptionMap.get(firstRecipientValue)
      : null

    return safeStr(selectedRecipient?.nomeContato) || "Cliente"
  }, [selectedRecipients, recipientOptionMap])

  const recipientItemTemplate = (option: RecipientOption) => {
    const channels = getRecipientChannelsLabel(option.hasPhone, option.hasEmail)
    const location = formatLocation(option.bairro, option.cidade, option.uf)

    return (
      <div className="flex flex-column gap-1 py-1">
        <div className="flex align-items-center gap-2 flex-wrap">
          <span style={{ fontWeight: 600 }}>{option.nomeContato}</span>
          <Tag
            value={getRecipientTypeLabel(option.contatoTipo)}
            severity={option.contatoTipo === "OWNER" ? "info" : "success"}
          />
          <span className="text-secondary" style={{ fontSize: 12 }}>
            {channels}
          </span>
        </div>

        <div className="text-secondary" style={{ fontSize: 12 }}>
          {location}
          {option.address ? ` • ${option.address}` : ""}
        </div>
      </div>
    )
  }

  const selectedRecipientTemplate = (value: string) => {
    const item = recipientOptionMap.get(value)
    if (!item) return value

    return `${item.nomeContato} (${item.contatoTipo === "OWNER" ? "Prop." : "Prof."})`
  }

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
          if (!selectedRecipients.length) {
            notify(
              "warn",
              "Selecione os destinatários",
              "Escolha pelo menos um proprietário ou profissional."
            )
            return
          }

          if (!channelWa && !channelEmail) {
            notify(
              "warn",
              "Selecione um canal",
              "Marque pelo menos WhatsApp ou E-mail."
            )
            return
          }

          if (channelEmail && !emailSubject.trim()) {
            notify(
              "warn",
              "Assunto obrigatório",
              "Preencha o assunto do e-mail para continuar."
            )
            return
          }

          const semContato = selectedRecipients
            .map((value) => recipientOptionMap.get(value))
            .filter((item): item is RecipientOption => !!item)
            .filter((item) => !recipientHasAnyContact(item))

          if (semContato.length) {
            const nomes = semContato
              .map((item) => `${item.nomeContato} (sem telefone e sem e-mail)`)
              .join("; ")

            notify(
              "warn",
              "Destinatários sem contato",
              `Remova da seleção: ${nomes}. Essas pessoas não têm telefone nem e-mail cadastrado.`
            )
            return
          }

          const incompativelCanal = selectedRecipients
            .map((value) => recipientOptionMap.get(value))
            .filter((item): item is RecipientOption => !!item)
            .filter(
              (item) =>
                recipientHasAnyContact(item) &&
                !recipientMatchesChannels(item, channelWa, channelEmail)
            )

          if (incompativelCanal.length) {
            const nomes = incompativelCanal
              .map((item) => describeRecipientContactGap(item, channelWa, channelEmail))
              .join("; ")

            notify(
              "warn",
              "Destinatários sem contato",
              `Remova da seleção: ${nomes}. Eles não têm telefone/e-mail para os canais marcados.`
            )
            return
          }

          const { cidadeStr, bairroStr } = getCidadeBairroStr()

          const destinatarios = selectedRecipients
            .map(parseRecipientValue)
            .filter(Boolean) as Array<{
            obra_id: string
            contato_tipo: RecipientKind
          }>

          createCampaign({
            destinatarios,
            channelWa,
            channelEmail,
            iaContinuar,
            emailSubject: emailSubject.trim(),
            messageText,
            selectedCity,
            cidade: cidadeStr,
            bairro: bairroStr,
            conexaoWhatsAppId,
            conexaoEmailId,
            ocultarJaContactados,
            onCreate,
            onClose,
          })
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

        <div className="field col-12 md:col-6 xl:col-6 mb-0">
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

        <div className="field col-12" style={{ marginTop: "1rem" }}>
          <div className="flex align-items-center justify-content-between flex-wrap gap-3">
            <div className="flex align-items-center gap-2">
              <InputSwitch
                checked={ocultarJaContactados}
                onChange={(e) => setOcultarJaContactados(!!e.value)}
              />
              <label style={{ margin: 0 }}>Ocultar já contactados</label>
            </div>

            <div className="text-secondary" style={{ fontSize: 12 }}>
              Quando ativo, a lista esconde leads já contactados, mas sua seleção manual sempre será respeitada.
            </div>
          </div>
        </div>

        <div className="field col-12">
          <div className="flex align-items-center justify-content-between flex-wrap gap-2 mb-2">
            <label style={{ fontWeight: 700, marginBottom: 0 }}>
              Destinatários ({loadingLeads ? "carregando…" : `${recipientStats.total} disponíveis`})
            </label>

            <div className="flex gap-2 flex-wrap">
              <Button
                type="button"
                label={`Proprietários (${recipientStats.owners})`}
                className={
                  recipientFilter === "OWNER" ? "p-button-sm" : "p-button-sm p-button-outlined"
                }
                onClick={() => handleRecipientFilterChange("OWNER")}
              />
              <Button
                type="button"
                label={`Profissionais (${recipientStats.professionals})`}
                className={
                  recipientFilter === "PROFISSIONAL"
                    ? "p-button-sm"
                    : "p-button-sm p-button-outlined"
                }
                onClick={() => handleRecipientFilterChange("PROFISSIONAL")}
              />
              <Button
                type="button"
                label="Todos"
                className={
                  recipientFilter === "ALL" ? "p-button-sm" : "p-button-sm p-button-outlined"
                }
                onClick={() => handleRecipientFilterChange("ALL")}
              />
              <Button
                type="button"
                label="Limpar"
                className="p-button-sm p-button-text"
                onClick={clearRecipients}
              />
            </div>
          </div>

          <MultiSelect
            value={selectedRecipients}
            options={filteredRecipientOptions}
            disabled={loadingLeads}
            onChange={(e) => {
              const val = (e.value || []) as string[]
              if (val.length > 50) {
                notify("warn", "Limite excedido", "Você pode selecionar no máximo 50 leads por disparo.")
                setSelectedRecipients(val.slice(0, 50))
              } else {
                setSelectedRecipients(val)
              }
            }}
            placeholder={getPlaceholderByFilter(recipientFilter)}
            className="w-full"
            filter
            maxSelectedLabels={3}
            optionLabel="label"
            optionValue="value"
            itemTemplate={recipientItemTemplate}
            selectedItemTemplate={selectedRecipientTemplate}
            emptyFilterMessage="Nenhum destinatário encontrado."
            display="chip"
          />

          <div className="text-secondary" style={{ fontSize: 12, marginTop: 8 }}>
            {fetchError ? (
              <span style={{ color: "var(--red-500)" }}>{fetchError}</span>
            ) : (
              <>
                Selecionados: {selectedRecipientStats.total} • Proprietários:{" "}
                {selectedRecipientStats.owners} • Profissionais:{" "}
                {selectedRecipientStats.professionals}
              </>
            )}
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
            <InputSwitch
              checked={iaContinuar}
              onChange={(e) => setIaContinuar(!!e.value)}
            />
            <div className="text-secondary" style={{ flex: 1, minWidth: 220 }}>
              {iaContinuar
                ? "IA continua a conversa após resposta"
                : "IA não continuará"}
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
              placeholder="Assunto do e-mail"
            />
          </div>
        ) : null}

        <div className="field col-12">
          <div style={{ fontSize: 16, fontWeight: 700, marginBottom: 8 }}>
            Mensagem
          </div>

          <div
            className="flex align-items-center gap-2 flex-wrap"
            style={{ marginBottom: 12 }}
          >
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
              clique para inserir no texto
            </span>
          </div>

          <label style={{ fontWeight: 600, fontSize: 13 }}>
            Objetivo da campanha (para IA)
          </label>

          <div className="flex flex-column md:flex-row gap-2">
            <InputText
              value={objetivo}
              onChange={(e) => setObjetivo(e.currentTarget.value)}
              placeholder="Ex: Reativar clientes de obras residenciais em andamento."
              className="w-full"
            />
            <Button
              icon="pi pi-bolt"
              label="Gerar IA"
              onClick={handleGenerate}
              loading={generating}
              type="button"
              className="w-full md:w-auto"
            />
          </div>

          <label
            style={{
              fontWeight: 600,
              fontSize: 13,
              marginTop: 12,
              display: "block",
            }}
          >
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
            <div style={{ marginBottom: 6, fontWeight: 600, fontSize: 13 }}>
              Pré-visualização no WhatsApp.
            </div>

            <div
              className="wa-preview-phone"
              style={{
                maxWidth: 360,
              }}
            >
              <div className="wa-preview-notch" />

              <div className="wa-preview-screen">
                <div className="wa-preview-topbar">
                  <div className="wa-preview-avatar">{previewContactName.charAt(0)}</div>
                  <div className="wa-preview-contact-meta">
                    <div className="wa-preview-contact-name">{previewContactName}</div>
                    <div className="wa-preview-contact-status">online</div>
                  </div>
                </div>

                <div className="wa-preview-chat-bg">
                  <div className="wa-preview-bubble-wrap">
                    <div className="wa-preview-bubble">{previewText}</div>
                    <div className="wa-preview-bubble-time">14:47</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <style>{`
        .wa-preview-phone {
          background: #111b21;
          border: 1px solid rgba(0, 0, 0, 0.25);
          border-radius: 28px;
          padding: 10px;
          box-shadow: 0 18px 38px rgba(0, 0, 0, 0.25);
          position: relative;
          margin: 0 auto;
        }

        .wa-preview-notch {
          width: 44%;
          height: 18px;
          border-radius: 0 0 12px 12px;
          background: #000;
          position: absolute;
          top: 10px;
          left: 50%;
          transform: translateX(-50%);
          z-index: 3;
        }

        .wa-preview-screen {
          border-radius: 22px;
          overflow: hidden;
          background: #0b141a;
          border: 1px solid rgba(255, 255, 255, 0.08);
          min-height: 310px;
          display: flex;
          flex-direction: column;
        }

        .wa-preview-topbar {
          background: #202c33;
          color: #e9edef;
          min-height: 54px;
          padding: 8px 12px;
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .wa-preview-avatar {
          width: 32px;
          height: 32px;
          border-radius: 999px;
          background: #6b7c85;
          color: #fff;
          font-weight: 700;
          display: flex;
          align-items: center;
          justify-content: center;
          text-transform: uppercase;
          flex-shrink: 0;
        }

        .wa-preview-contact-meta {
          min-width: 0;
        }

        .wa-preview-contact-name {
          font-size: 13px;
          font-weight: 600;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          max-width: 220px;
        }

        .wa-preview-contact-status {
          font-size: 11px;
          color: #aebac1;
        }

        .wa-preview-chat-bg {
          flex: 1;
          padding: 12px;
          background-image: url("/fundo-whatsapp.jpg");
          background-size: cover;
          background-position: center;
          display: flex;
          align-items: flex-start;
          justify-content: flex-end;
        }

        .wa-preview-bubble-wrap {
          max-width: 88%;
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 4px;
        }

        .wa-preview-bubble {
          background: #d9fdd3;
          color: #111b21;
          border-radius: 7.5px;
          border-top-right-radius: 2px;
          padding: 8px 10px;
          font-size: 13px;
          line-height: 1.45;
          white-space: pre-wrap;
          word-break: break-word;
          box-shadow: 0 1px 0 rgba(0, 0, 0, 0.12);
        }

        .wa-preview-bubble-time {
          font-size: 10px;
          color: rgba(17, 27, 33, 0.72);
          padding-right: 2px;
        }
      `}</style>
    </Dialog>
  )
}