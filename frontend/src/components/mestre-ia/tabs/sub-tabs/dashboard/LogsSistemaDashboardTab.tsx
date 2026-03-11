import React from "react";
import { DataTable } from "primereact/datatable";
import { Column, type ColumnBodyOptions } from "primereact/column";
import { InputText } from "primereact/inputtext";
import { Dropdown } from "primereact/dropdown";
import { Dialog } from "primereact/dialog";
import { Toast } from "primereact/toast";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";

import type { LogRow, LogTipo, LogSortMode, LogTipoFilter } from "../../../types/dashboard";

export default function LogsSistemaDashboardTab() {
  const toast = React.useRef<Toast>(null);

  const notify = (severity: "success" | "info" | "warn" | "error", summary: string, detail: string) => {
    toast.current?.show({ severity, summary, detail, life: 3000 });
  };

  const [logs, setLogs] = React.useState<LogRow[]>(() => [
    {
      id: "1",
      dataHora: "2026-02-25 10:12",
      createdAt: "2026-02-25T10:12:00-03:00",
      tipo: "CAMPANHA",
      acao: "Campanha criada",
      detalhe: "Campanha A • 50 leads",
    },
    {
      id: "2",
      dataHora: "2026-02-25 10:15",
      createdAt: "2026-02-25T10:15:00-03:00",
      tipo: "IA",
      acao: "Intenção detectada",
      detalhe: "“orçamento” → Resposta sugerida",
    },
    {
      id: "3",
      dataHora: "2026-02-25 10:18",
      createdAt: "2026-02-25T10:18:00-03:00",
      tipo: "SISTEMA",
      acao: "Conexão WhatsApp",
      detalhe: "Sessão iniciada (mock)",
    },
  ]);

  const [logSearch, setLogSearch] = React.useState("");
  const [logTipoFilter, setLogTipoFilter] = React.useState<LogTipoFilter>("ALL");
  const [logSortMode, setLogSortMode] = React.useState<LogSortMode>("CREATED_DESC");

  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false);
  const [deleteTarget, setDeleteTarget] = React.useState<LogRow | null>(null);

  const [detailDialogOpen, setDetailDialogOpen] = React.useState(false);
  const [detailTarget, setDetailTarget] = React.useState<LogRow | null>(null);

  const logTipoOptions = [
    { label: "Todos", value: "ALL" as LogTipoFilter },
    { label: "Campanha", value: "CAMPANHA" as LogTipoFilter },
    { label: "IA", value: "IA" as LogTipoFilter },
    { label: "Sistema", value: "SISTEMA" as LogTipoFilter },
  ];

  const logSortOptions = [
    { label: "Mais recentes", value: "CREATED_DESC" as LogSortMode },
    { label: "Mais antigos", value: "CREATED_ASC" as LogSortMode },
  ];

  const filteredSortedLogs = React.useMemo(() => {
    const q = logSearch.trim().toLowerCase();
    let data = [...logs];

    if (logTipoFilter !== "ALL") data = data.filter((l) => l.tipo === logTipoFilter);

    if (q) {
      data = data.filter((l) => [l.dataHora, l.tipo, l.acao, l.detalhe].join(" ").toLowerCase().includes(q));
    }

    const byCreatedDesc = (a: LogRow, b: LogRow) => +new Date(b.createdAt) - +new Date(a.createdAt);
    const byCreatedAsc = (a: LogRow, b: LogRow) => +new Date(a.createdAt) - +new Date(b.createdAt);

    if (logSortMode === "CREATED_DESC") data.sort(byCreatedDesc);
    if (logSortMode === "CREATED_ASC") data.sort(byCreatedAsc);

    return data;
  }, [logs, logSearch, logTipoFilter, logSortMode]);

  const tipoTagTemplate = (row: LogRow) => {
    const map: Record<LogTipo, { label: string; severity: "info" | "warning" | "success" }> = {
      CAMPANHA: { label: "Campanha", severity: "info" },
      IA: { label: "IA", severity: "warning" },
      SISTEMA: { label: "Sistema", severity: "success" },
    };
    const conf = map[row.tipo];
    return <Tag value={conf.label} severity={conf.severity} className="border-round-xl" />;
  };

  const acaoTemplate = (row: LogRow) => (
    <div>
      <div className="flex align-items-center gap-2">
        <strong style={{ fontSize: 14 }}>{row.acao}</strong>
      </div>
      <div className="text-secondary" style={{ marginTop: 4, lineHeight: 1.4 }}>
        {row.detalhe}
      </div>
    </div>
  );

  const openDelete = (row: LogRow) => {
    setDeleteTarget(row);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (!deleteTarget) return;
    setLogs((prev) => prev.filter((l) => l.id !== deleteTarget.id));
    notify("success", "Log", "Removido com sucesso.");
    setDeleteDialogOpen(false);
    setDeleteTarget(null);
  };

  const openDetails = (row: LogRow) => {
    setDetailTarget(row);
    setDetailDialogOpen(true);
  };

  const copyLog = async (row: LogRow) => {
    const text = `[${row.dataHora}] (${row.tipo}) ${row.acao} - ${row.detalhe}`;
    try {
      await navigator.clipboard.writeText(text);
      notify("success", "Copiado", "Log copiado para a área de transferência.");
    } catch {
      notify("warn", "Ops", "Não consegui copiar automaticamente.");
    }
  };

  const logAcoesTemplate = (row: LogRow, _opts: ColumnBodyOptions) => (
    <div className="flex align-items-center gap-2 justify-content-end">
      <Button
        icon="pi pi-eye"
        className="p-button-text p-button-rounded"
        severity="secondary"
        tooltip="Ver detalhes"
        tooltipOptions={{ position: "top" }}
        onClick={() => openDetails(row)}
      />

      <Button
        icon="pi pi-copy"
        className="p-button-text p-button-rounded"
        severity="secondary"
        tooltip="Copiar"
        tooltipOptions={{ position: "top" }}
        onClick={() => copyLog(row)}
      />

      <Button
        icon="pi pi-trash"
        className="p-button-text p-button-rounded"
        severity="danger"
        tooltip="Excluir"
        tooltipOptions={{ position: "top" }}
        onClick={() => openDelete(row)}
      />
    </div>
  );

  const clearLogFilters = () => {
    setLogSearch("");
    setLogTipoFilter("ALL");
    setLogSortMode("CREATED_DESC");
  };

  return (
    <div>
      <Toast ref={toast} />

      <div className="flex align-items-center justify-content-between mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Logs do Sistema</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Registre e acompanhe todos os eventos do sistema
          </div>
        </div>
      </div>

      {/* Top bar (igual Leads) */}
      <div className="bg-white border-round-3xl p-3 mb-3 border-1 surface-border">
        <div className="formgrid grid align-items-end">
          <div className="field col-12 md:col-4 mb-0">
            <label>Pesquisar</label>
            <span className="p-input-icon-left w-full">
              <i className="pi pi-search" />
              <InputText
                className="w-full"
                placeholder="Ação, detalhe, tipo, data..."
                value={logSearch}
                onChange={(e) => setLogSearch(e.target.value)}
              />
            </span>
          </div>

          <div className="field col-12 md:col-3 mb-0">
            <label>Tipo</label>
            <Dropdown className="w-full" value={logTipoFilter} options={logTipoOptions} onChange={(e) => setLogTipoFilter(e.value)} />
          </div>

          <div className="field col-12 md:col-3 mb-0">
            <label>Ordenação</label>
            <Dropdown className="w-full" value={logSortMode} options={logSortOptions} onChange={(e) => setLogSortMode(e.value)} />
          </div>

          <div className="field col-12 md:col-2 mb-0">
            <Button className="w-full" icon="pi pi-filter-slash" label="Limpar" severity="secondary" onClick={clearLogFilters} />
          </div>
        </div>
      </div>

      {/* Table container (igual Leads) */}
      <div className="bg-white border-round-3xl p-2 border-1 surface-border">
        <DataTable value={filteredSortedLogs} dataKey="id" rowHover className="p-datatable-sm" emptyMessage="Nenhum log encontrado.">
          <Column field="dataHora" header="Data/Hora" style={{ width: "16%" }} />
          <Column header="Tipo" body={tipoTagTemplate} style={{ width: "12%" }} />
          <Column header="Ação" body={acaoTemplate} />
          <Column header="Ações" body={logAcoesTemplate} style={{ width: "16%" }} />
        </DataTable>

        <style>{`
          .p-datatable .p-datatable-thead > tr > th {
            background: #fff !important;
            border: 0 !important;
            color: #6B7280 !important;
            font-weight: 600 !important;
            padding: 1.1rem 1rem !important;
          }
          .p-datatable .p-datatable-tbody > tr > td {
            border: 0 !important;
            border-top: 1px solid rgba(0,0,0,0.04) !important;
            padding: 1.1rem 1rem !important;
            vertical-align: middle !important;
          }
        `}</style>
      </div>

      {/* Dialog de detalhes */}
      <Dialog
        header="Detalhes do log"
        visible={detailDialogOpen}
        style={{ width: "640px", maxWidth: "95vw" }}
        onHide={() => setDetailDialogOpen(false)}
        draggable={false}
        dismissableMask
        footer={
          <div className="flex justify-content-end gap-2">
            <Button label="Copiar" icon="pi pi-copy" severity="secondary" onClick={() => detailTarget && copyLog(detailTarget)} />
            <Button label="Fechar" icon="pi pi-times" severity="secondary" onClick={() => setDetailDialogOpen(false)} />
          </div>
        }
      >
        <div style={{ lineHeight: 1.8 }}>
          <div className="text-secondary"><strong>Data/Hora:</strong> {detailTarget?.dataHora || "-"}</div>
          <div className="text-secondary"><strong>Tipo:</strong> {detailTarget?.tipo || "-"}</div>
          <div className="text-secondary"><strong>Ação:</strong> {detailTarget?.acao || "-"}</div>
          <div className="text-secondary"><strong>Detalhe:</strong> {detailTarget?.detalhe || "-"}</div>
        </div>
      </Dialog>

      {/* Dialog de excluir */}
      <Dialog
        header="Excluir log"
        visible={deleteDialogOpen}
        style={{ width: "520px", maxWidth: "95vw" }}
        onHide={() => setDeleteDialogOpen(false)}
        draggable={false}
        dismissableMask
        footer={
          <div className="flex justify-content-end gap-2">
            <Button label="Excluir" icon="pi pi-trash" severity="danger" onClick={confirmDelete} />
            <Button label="Cancelar" icon="pi pi-times" severity="secondary" onClick={() => setDeleteDialogOpen(false)} />
          </div>
        }
      >
        <div className="text-secondary">
          Tem certeza que deseja excluir este log?
          <div style={{ marginTop: 10 }}>
            <strong>{deleteTarget?.acao}</strong>
            <div className="text-secondary" style={{ marginTop: 6 }}>
              {deleteTarget?.detalhe}
            </div>
          </div>
        </div>
      </Dialog>
    </div>
  );
}