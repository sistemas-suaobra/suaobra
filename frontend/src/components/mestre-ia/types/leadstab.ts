export type SortMode = "NAME_ASC" | "NAME_DESC" | "CREATED_DESC" | "CREATED_ASC";
export type FavFilter = "ALL" | "FAVORITES";

export type LeadRow = {
  id: string;
  nome: string;
  email?: string;
  telefone?: string;
  cidade: string;
  bairro: string;
  obraTipo: string;
  areaM2: number;
  status: "Novo" | "Em andamento" | "Fechado";
  jaContactado?: boolean;
  favorito?: boolean;
  createdAt: string;
};
