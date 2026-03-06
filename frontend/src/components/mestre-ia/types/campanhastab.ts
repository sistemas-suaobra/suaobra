// Lead vem da API obras-plus (obras favoritadas)
export type LeadOption = {
  label: string;
  value: string; // obra_id
  owner?: string;
  professional?: string;
  has_owner_phone?: boolean;
  has_professional_phone?: boolean;
  has_owner_email?: boolean;
  has_professional_email?: boolean;
  bairro?: string;
  city?: string;
};

export type CampaignMessage = {
  id: string;
  text: string;
};

// Status conforme schema PocketBase
export type CampaignStatus = "RASCUNHO" | "AGENDADA" | "EM_ANDAMENTO" | "PAUSADA" | "CONCLUIDA" | "CANCELADA";

// Destinatário conforme schema PocketBase
export type CampanhaDestinatario = {
  id?: string;
  team_id: string;
  campanha_id: string;
  lead_id: string; // obra_id da lead
  telefone_e164?: string;
  email?: string;
  status: "PENDENTE" | "EM_FILA" | "ENVIADO" | "FALHOU" | "IGNORADO";
  tentativas: number;
  erro?: string;
  enviado_em?: string;
};

// Campanha conforme schema PocketBase
export type Campaign = {
  id: string;
  team_id?: string;
  nome: string;
  conexao_id?: string;
  status: CampaignStatus;
  mensagem_template: string;
  criado_por?: string;
  iniciado_em?: string;
  finalizado_em?: string;
  created?: string;
  updated?: string;

  // Campos locais para UI (não salvos diretamente)
  leads: string[]; // obra_ids selecionados
  channelWa: boolean;
  channelEmail: boolean;
  iaContinuar: boolean; // alias de manter_ia (campo salvo no PB)
  manter_ia?: boolean;  // campo real no PocketBase
  emailSubject?: string;
  destinatarios?: CampanhaDestinatario[];
};