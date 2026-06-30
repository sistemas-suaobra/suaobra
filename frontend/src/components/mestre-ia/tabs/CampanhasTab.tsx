import React from "react";
import { InputText } from "primereact/inputtext";
import { Button } from "primereact/button";
import { Toast } from "primereact/toast";
import { DataTable } from "primereact/datatable";
import { Column, type ColumnBodyOptions } from "primereact/column";
import { Dropdown } from "primereact/dropdown";
import { Tag } from "primereact/tag";
import { useStore } from "@nanostores/react";

import { obrasPlusCity, user, loadUserState } from "../../../store/store";
import { type City, makeCity } from "../../../store/cities";
import { api, baseURL, PB } from "../../../store/api";
import type { LeadOption, Campaign, CampaignStatus } from "../types/campanhastab";
import CreateCampaignDialog from "./CreateCampaignDialog";
import { MestreIaTransitionLoader } from "../MestreIaTransitionLoader";

type SortMode = "CREATED_DESC" | "CREATED_ASC" | "NAME_ASC" | "NAME_DESC";
type StatusFilter = "ALL" | CampaignStatus;

const sortOptions = [
  { label: "Cadastro: mais recente", value: "CREATED_DESC" as SortMode },
  { label: "Cadastro: mais antigo", value: "CREATED_ASC" as SortMode },
  { label: "Nome A → Z", value: "NAME_ASC" as SortMode },
  { label: "Nome Z → A", value: "NAME_DESC" as SortMode },
];

const statusOptions = [
  { label: "Todos", value: "ALL" as StatusFilter },
  { label: "Rascunho", value: "RASCUNHO" as StatusFilter },
  { label: "Agendada", value: "AGENDADA" as StatusFilter },
  { label: "Em andamento", value: "EM_ANDAMENTO" as StatusFilter },
  { label: "Pausada", value: "PAUSADA" as StatusFilter },
  { label: "Concluída", value: "CONCLUIDA" as StatusFilter },
  { label: "Cancelada", value: "CANCELADA" as StatusFilter },
];

function fmtDate(iso: string) {
  try {
    const d = new Date(iso);
    return d.toLocaleString("pt-BR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

function useIsMobile(breakpoint = 768) {
  const [isMobile, setIsMobile] = React.useState(false);

  React.useEffect(() => {
    const media = window.matchMedia(`(max-width: ${breakpoint}px)`);

    const onChange = () => setIsMobile(media.matches);

    onChange();

    if (media.addEventListener) {
      media.addEventListener("change", onChange);
      return () => media.removeEventListener("change", onChange);
    }

    media.addListener(onChange);
    return () => media.removeListener(onChange);
  }, [breakpoint]);

  return isMobile;
}

// Interface de resultado da API obras-plus
interface ResultRecord {
  obra_id: string;
  owner: string;
  professional: string;
  has_professional_phone: boolean;
  has_professional_email: boolean;
  has_owner_phone: boolean;
  has_owner_email: boolean;
  bairro: string;
  city: string;
}

interface Result {
  total: number;
  records: ResultRecord[];
}

export default function CampanhasTab() {
  const isMobile = useIsMobile(768);

  const [refreshKey, setRefreshKey] = React.useState(0);
  const toast = React.useRef<Toast>(null);
  const $obrasPlusCity = useStore(obrasPlusCity);

  const notify = (
    severity: "success" | "info" | "warn" | "error",
    summary: string,
    detail: string
  ) => {
    toast.current?.show({ severity, summary, detail, life: 3000 });
  };

  const [leadsOptions, setLeadsOptions] = React.useState<LeadOption[]>([]);
  const [loadingLeads, setLoadingLeads] = React.useState(false);
  const [selectedCity, setSelectedCity] = React.useState<City | null>($obrasPlusCity || null);
  const [citiesOptions, setCitiesOptions] = React.useState<City[]>([]);

  const [conexaoWhatsAppId, setConexaoWhatsAppId] = React.useState<string>("");
  const [conexaoEmailId, setConexaoEmailId] = React.useState<string>("");

  const [teamId, setTeamId] = React.useState<string>("");
  const [userId, setUserId] = React.useState<string>("");

  const [campaigns, setCampaigns] = React.useState<Campaign[]>([]);
  const [loadingCampaigns, setLoadingCampaigns] = React.useState(false);

  const [search, setSearch] = React.useState("");
  const [statusFilter, setStatusFilter] = React.useState<StatusFilter>("ALL");
  const [sortMode, setSortMode] = React.useState<SortMode>("CREATED_DESC");
  const [createOpen, setCreateOpen] = React.useState(false);

  const onCreateCampaign = async (newCampaign: Campaign) => {
    try {
      setCampaigns((prev) => [newCampaign, ...prev]);
      setCreateOpen(false);
      setRefreshKey((k) => k + 1);
      notify("success", "Campanha criada", "A campanha foi criada com sucesso.");
    } catch (error: any) {
      notify("error", "Erro", error?.message || "Erro ao criar campanha");
    }
  };

  React.useEffect(() => {
    loadUserState().then((userData) => {
      if (!userData) return;

      setTeamId(userData.team?.id || "");
      setUserId(userData.id || "");

      const cities =
        userData.team?.cities?.sort().map((id: string) => makeCity(id)) || [];

      if (cities.length) {
        setCitiesOptions(cities);

        let savedCity = localStorage.getItem("obrasPlusCity");
        if (!cities.map((c: City) => c.id).includes(savedCity || "")) {
          savedCity = cities[0].id;
        }

        const city = makeCity(savedCity!);
        setSelectedCity(city);
        obrasPlusCity.set(city);
      }
    });
  }, []);

  React.useEffect(() => {
    if (!selectedCity) return;

    const fetchLeads = async () => {
      setLoadingLeads(true);

      try {
        const payload = {
          city: selectedCity.city || "",
          bairro: "",
          order: "first_listing_date-desc,start_date-desc",
          filter: "",
          sizeMin: "0",
          sizeMax: "9999999",
          offset: "0",
          itemsPerPage: "200",
          enriched: "false",
          startDateFrom: "",
          startDateTo: "",
          endDateFrom: "",
          endDateTo: "",
        };

        const resp = await api().get(`${baseURL()}/query/leads-plus`, payload);
        if (resp.error) throw new Error(resp.error);

        const data = (await resp.response.json()) as Result;

        const options: LeadOption[] = (data.records || []).map((r) => {
          const nome = r.owner || r.professional || "Sem nome";
          const contato =
            r.has_owner_phone || r.has_professional_phone
              ? "📞"
              : r.has_owner_email || r.has_professional_email
              ? "✉️"
              : "";

          return {
            label: `${nome} — ${r.city}, ${r.bairro} ${contato}`,
            value: r.obra_id,
            owner: r.owner,
            professional: r.professional,
            has_owner_phone: r.has_owner_phone,
            has_professional_phone: r.has_professional_phone,
            has_owner_email: r.has_owner_email,
            has_professional_email: r.has_professional_email,
            bairro: r.bairro,
            city: r.city,
          };
        });

        setLeadsOptions(options);
      } catch (error) {
        console.error("Erro ao buscar leads:", error);
        setLeadsOptions([]);
      } finally {
        setLoadingLeads(false);
      }
    };

    fetchLeads();
  }, [selectedCity]);

  const fetchCampaigns = React.useCallback(async () => {
    if (!teamId) return;

    setLoadingCampaigns(true);

    try {
      const pb = PB();
      pb.authStore.save(user.get().token, user.get());

      const records = await pb.collection("campanhas").getList(1, 100, {
        filter: `team_id = "${teamId}"`,
        sort: "-created",
      });

      const mapped: Campaign[] = records.items.map((r: any) => ({
        id: r.id,
        team_id: r.team_id,
        nome: r.nome,
        conexao_id: r.conexao_id,
        status: r.status || "RASCUNHO",
        mensagem_template: r.mensagem_template || "",
        criado_por: r.criado_por,
        iniciado_em: r.iniciado_em,
        finalizado_em: r.finalizado_em,
        created: r.created,
        updated: r.updated,
        leads: [],
        channelWa: true,
        channelEmail: false,
        iaContinuar: !!r.manter_ia,
      }));

      setCampaigns(mapped);
    } catch (error) {
      console.error("Erro ao buscar campanhas:", error);
    } finally {
      setLoadingCampaigns(false);
    }
  }, [teamId]);

  React.useEffect(() => {
    fetchCampaigns();
  }, [teamId, fetchCampaigns, refreshKey]);

  React.useEffect(() => {
    if (!teamId) return;

    const fetchConexoes = async () => {
      try {
        const pb = PB();
        pb.authStore.save(user.get().token, user.get());

        const conexoes = await pb.collection("conexoes").getFullList({
          filter: `team_id = "${teamId}" && ativo = true`,
        });

        for (const con of conexoes) {
          if (con.canal === "WHATSAPP") {
            setConexaoWhatsAppId(con.id);
          } else if (con.canal === "EMAIL") {
            setConexaoEmailId(con.id);
          }
        }
      } catch (error) {
        console.error("Erro ao buscar conexões:", error);
      }
    };

    fetchConexoes();
  }, [teamId]);

  const filteredSorted = React.useMemo(() => {
    const q = search.trim().toLowerCase();
    let data = [...campaigns];

    if (statusFilter !== "ALL") {
      data = data.filter((c) => c.status === statusFilter);
    }

    if (q) {
      data = data.filter((c) => {
        const blob = [c.nome, c.status, c.mensagem_template || ""]
          .join(" ")
          .toLowerCase();

        return blob.includes(q);
      });
    }

    const byCreatedDesc = (a: Campaign, b: Campaign) =>
      +new Date(b.created || "") - +new Date(a.created || "");
    const byCreatedAsc = (a: Campaign, b: Campaign) =>
      +new Date(a.created || "") - +new Date(b.created || "");
    const byNameAsc = (a: Campaign, b: Campaign) =>
      a.nome.localeCompare(b.nome, "pt-BR", { sensitivity: "base" });
    const byNameDesc = (a: Campaign, b: Campaign) =>
      b.nome.localeCompare(a.nome, "pt-BR", { sensitivity: "base" });

    if (sortMode === "CREATED_DESC") data.sort(byCreatedDesc);
    if (sortMode === "CREATED_ASC") data.sort(byCreatedAsc);
    if (sortMode === "NAME_ASC") data.sort(byNameAsc);
    if (sortMode === "NAME_DESC") data.sort(byNameDesc);

    return data;
  }, [campaigns, search, statusFilter, sortMode]);

  const startCampaign = async (id: string) => {
    try {
      await api().post(`${baseURL()}/campanhas/${id}/iniciar`, {});
      setCampaigns((prev) =>
        prev.map((c) =>
          c.id === id ? { ...c, status: "EM_ANDAMENTO" as CampaignStatus } : c
        )
      );
      notify("success", "Campanha iniciada", "O envio será processado em background.");
    } catch (error: any) {
      const detail =
        error?.response?.data?.message ||
        error?.message ||
        "Erro ao iniciar campanha";
      notify("error", "Erro", detail);
    }
  };

  const deleteCampaign = async (id: string) => {
    const campanha = campaigns.find((c) => c.id === id);
    if (!campanha) return;

    const statusPermitidos: CampaignStatus[] = ["CONCLUIDA", "RASCUNHO"];
    if (!statusPermitidos.includes(campanha.status)) {
      notify(
        "warn",
        "Não permitido",
        "Só é possível excluir campanhas com status Concluída ou Rascunho."
      );
      return;
    }

    if (!window.confirm("Tem certeza que deseja excluir esta campanha e todos os seus dados?")) {
      return;
    }

    try {
      const pb = PB();
      pb.authStore.save(user.get().token, user.get());

      // 1. Deletar conversas de IA vinculadas à campanha
      try {
        const conversas = await pb.collection("conversas_ia").getFullList({
          filter: `campanha_id = "${id}"`,
        });
        for (const conv of conversas) {
          await pb.collection("conversas_ia").delete(conv.id);
        }
      } catch {
        // conversas_ia pode não existir ou não ter registros — ok
      }

      // 2. Deletar destinatários da campanha
      try {
        const destinatarios = await pb.collection("campanha_destinatarios").getFullList({
          filter: `campanha_id = "${id}"`,
        });
        for (const d of destinatarios) {
          await pb.collection("campanha_destinatarios").delete(d.id);
        }
      } catch {
        // se não encontrar destinatários, continua
      }

      // 3. Deletar a campanha
      await pb.collection("campanhas").delete(id);

      setCampaigns((prev) => prev.filter((c) => c.id !== id));
      notify("success", "Campanha removida", "Campanha e todos os dados associados foram excluídos.");
    } catch (error: any) {
      console.error("Erro ao excluir campanha:", error);
      notify(
        "error",
        "Erro ao excluir",
        error?.message || "Não foi possível excluir a campanha. Verifique se não há dados vinculados."
      );
    }
  };

  const statusMeta = (status: CampaignStatus) => {
    const map: Record<
      CampaignStatus,
      { label: string; severity: "info" | "warning" | "success" | "danger" }
    > = {
      RASCUNHO: { label: "Rascunho", severity: "info" },
      AGENDADA: { label: "Agendada", severity: "info" },
      EM_ANDAMENTO: { label: "Em andamento", severity: "warning" },
      PAUSADA: { label: "Pausada", severity: "warning" },
      CONCLUIDA: { label: "Concluída", severity: "success" },
      CANCELADA: { label: "Cancelada", severity: "danger" },
    };

    return map[status] || { label: status, severity: "info" as const };
  };

  const statusTag = (row: Campaign) => {
    const conf = statusMeta(row.status);
    return <Tag value={conf.label} severity={conf.severity} className="border-round-xl" />;
  };

  const campanhaCell = (row: Campaign) => (
    <div>
      <div style={{ fontWeight: 700, fontSize: 14 }}>{row.nome}</div>
      <div className="text-secondary" style={{ fontSize: 12, marginTop: 4 }}>
        Criada em {fmtDate(row.created || "")}
      </div>
    </div>
  );

  const mensagemCell = (row: Campaign) => {
    const text = row.mensagem_template || "";
    const short = text.length > 80 ? `${text.slice(0, 80)}…` : text;
    return <span className="text-secondary">{short || "—"}</span>;
  };

  const canDelete = (status: CampaignStatus) =>
    status === "CONCLUIDA" || status === "RASCUNHO";

  const actionsCell = (row: Campaign, _opts: ColumnBodyOptions) => (
    <div className="flex align-items-center gap-2 justify-content-end flex-wrap">
      <Button
        label="INICIAR"
        icon="pi pi-play"
        className="p-button-sm"
        disabled={row.status === "EM_ANDAMENTO" || row.status === "CONCLUIDA"}
        onClick={() => startCampaign(row.id)}
      />
      <Button
        icon="pi pi-trash"
        className="p-button-text p-button-rounded p-button-sm"
        severity="danger"
        tooltip={canDelete(row.status) ? "Excluir" : "Só campanhas concluídas ou rascunhos podem ser excluídas"}
        tooltipOptions={{ position: "top" }}
        disabled={!canDelete(row.status)}
        onClick={() => deleteCampaign(row.id)}
      />
    </div>
  );

  return (
    <div className="w-full">
      <Toast ref={toast} />

      <div className="flex flex-column md:flex-row align-items-start md:align-items-center justify-content-between gap-3 mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Campanhas</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Crie campanhas de disparo para seus leads • {leadsOptions.length} lead(s) disponíveis
          </div>
        </div>

        <div className="flex flex-column sm:flex-row gap-2 w-full md:w-auto">
          <Button
            icon="pi pi-plus"
            label="Nova campanha"
            onClick={() => setCreateOpen(true)}
            disabled={!selectedCity}
            className="w-full sm:w-auto"
          />
          <Button
            icon="pi pi-refresh"
            label="Atualizar"
            severity="info"
            onClick={() => setRefreshKey((k) => k + 1)}
            className="w-full sm:w-auto"
          />
        </div>
      </div>

      <div
        className="bg-white border-round-3xl p-2 md:p-3 border-1 surface-border"
        style={{ position: "relative", minHeight: loadingCampaigns ? 240 : undefined }}
      >
        {loadingCampaigns ? (
          <MestreIaTransitionLoader overlay caption="Carregando campanhas…" />
        ) : null}
        {!isMobile ? (
          <DataTable
            value={filteredSorted}
            dataKey="id"
            rowHover
            loading={false}
            className="p-datatable-sm"
            emptyMessage="Nenhuma campanha encontrada. Crie uma nova campanha!"
            scrollable
            tableStyle={{ minWidth: "56rem" }}
          >
            <Column header="Campanha" body={campanhaCell} style={{ width: "30%" }} />
            <Column header="Mensagem" body={mensagemCell} />
            <Column header="Status" body={statusTag} style={{ width: "14%" }} />
            <Column header="Ações" body={actionsCell} style={{ width: "18%" }} />
          </DataTable>
        ) : (
          <div className="flex flex-column gap-3">
            {loadingCampaigns ? null : filteredSorted.length === 0 ? (
              <div className="text-center text-secondary py-4">
                Nenhuma campanha encontrada. Crie uma nova campanha!
              </div>
            ) : (
              filteredSorted.map((row) => {
                const conf = statusMeta(row.status);
                const text = row.mensagem_template || "";
                const short = text.length > 140 ? `${text.slice(0, 140)}…` : text;

                return (
                  <div
                    key={row.id}
                    className="border-1 surface-border border-round-2xl p-3"
                    style={{ background: "#fff" }}
                  >
                    <div className="flex align-items-start justify-content-between gap-2 mb-2">
                      <div style={{ minWidth: 0, flex: 1 }}>
                        <div
                          style={{
                            fontWeight: 700,
                            fontSize: 15,
                            lineHeight: 1.3,
                            wordBreak: "break-word",
                          }}
                        >
                          {row.nome}
                        </div>
                        <div
                          className="text-secondary"
                          style={{ fontSize: 12, marginTop: 4 }}
                        >
                          Criada em {fmtDate(row.created || "")}
                        </div>
                      </div>

                      <Tag
                        value={conf.label}
                        severity={conf.severity}
                        className="border-round-xl"
                      />
                    </div>

                    <div
                      className="text-secondary"
                      style={{
                        fontSize: 13,
                        lineHeight: 1.5,
                        marginTop: 8,
                        wordBreak: "break-word",
                      }}
                    >
                      {short || "—"}
                    </div>

                    <div className="flex flex-column gap-2 mt-3">
                      <Button
                        label="INICIAR"
                        icon="pi pi-play"
                        className="p-button-sm w-full"
                        disabled={
                          row.status === "EM_ANDAMENTO" || row.status === "CONCLUIDA"
                        }
                        onClick={() => startCampaign(row.id)}
                      />
                      <Button
                        label={canDelete(row.status) ? "Excluir" : "Excluir (indisponível)"}
                        icon="pi pi-trash"
                        className="p-button-sm p-button-outlined w-full"
                        severity="danger"
                        disabled={!canDelete(row.status)}
                        onClick={() => deleteCampaign(row.id)}
                      />
                    </div>
                  </div>
                );
              })
            )}
          </div>
        )}

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

          @media (max-width: 768px) {
            .p-dropdown,
            .p-inputtext,
            .p-button {
              width: 100%;
            }

            .p-card,
            .p-tag {
              max-width: 100%;
            }
          }
        `}</style>
      </div>

      <CreateCampaignDialog
        visible={createOpen}
        onClose={() => setCreateOpen(false)}
        leadsOptions={leadsOptions}
        onCreate={onCreateCampaign}
        notify={notify}
        conexaoWhatsAppId={conexaoWhatsAppId}
        conexaoEmailId={conexaoEmailId}
        teamId={teamId}
        userId={userId}
      />
    </div>
  );
}