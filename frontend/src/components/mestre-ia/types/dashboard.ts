export type LogTipo = "CAMPANHA" | "IA" | "SISTEMA";

export type LogRow = {
  id: string;
  dataHora: string; // display
  createdAt: string; // ISO
  tipo: LogTipo;
  acao: string;
  detalhe: string;
};

export type LogSortMode = "CREATED_DESC" | "CREATED_ASC";
export type LogTipoFilter = "ALL" | LogTipo;