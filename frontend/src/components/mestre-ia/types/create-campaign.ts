export type NotifyFn = (
  severity: "success" | "info" | "warn" | "error",
  summary: string,
  detail: string
) => void

export type LeadOption = { label: string; value: string }

export type CreateCampaignDialogProps = {
  visible: boolean
  onClose: () => void
  leadsOptions: LeadOption[]
  onCreate: (created: any) => void
  notify: NotifyFn
  conexaoWhatsAppId?: string
  conexaoEmailId?: string
  teamId: string
  userId: string
}

export type ObrasPlusRecord = {
  obra_id: string
  owner?: string
  professional?: string
  has_professional_phone?: boolean
  has_professional_email?: boolean
  has_owner_phone?: boolean
  has_owner_email?: boolean
  address?: string
  bairro?: string
  city?: string
  state?: string
}

export type ContactRecord = {
  telephone?: string | number
  email?: string
  city?: string
  state?: string
}

export type LeadComms = {
  telefone_e164: string
  email: string
  nome_contato: string
}