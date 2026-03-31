import React from "react";
import { TabView, TabPanel } from "primereact/tabview";
import { Toast } from "primereact/toast";
import ConexaoTab from "./tabs/ConexaoTab";
import LeadsTab from "./tabs/LeadsTab";
import AgenteIaTab from "./tabs/AgenteIATab";
import CampanhasTab from "./tabs/CampanhasTab";
import { IntentDialog, type Intent } from "./tabs/IntentDialog";
import DashboardTab from "./tabs/DashboardTab";
import { MestreIaTransitionLoader } from "./MestreIaTransitionLoader";
import { api, baseURL } from "../../store/api";
import { loadUserState, user } from "../../store/store";

export default function MestreIaPage() {
  const toast = React.useRef<Toast>(null);
  const tabTransitionTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  const [activeTabIndex, setActiveTabIndex] = React.useState(0);
  const [tabTransitioning, setTabTransitioning] = React.useState(false);

  const [intents, setIntents] = React.useState<Intent[]>([]);
  const [loadingIntents, setLoadingIntents] = React.useState(false);
  const [intentDialogOpen, setIntentDialogOpen] = React.useState(false);
  const [editingIntent, setEditingIntent] = React.useState<Intent | null>(null);

  const notify = React.useCallback(
    (
      severity: "success" | "info" | "warn" | "error",
      summary: string,
      detail: string
    ) => {
      toast.current?.show({ severity, summary, detail, life: 3000 });
    },
    []
  );

  const loadIntents = React.useCallback(async () => {
    setLoadingIntents(true);

    try {
      const resp = await api().get(baseURL() + "/agente-ia/intencoes");
      const data = await resp.json();

      const mapped: Intent[] = (data?.intencoes || []).map((item: any) => ({
        id: item.id,
        titulo: item.nome || "",
        ativo: item.ativa || false,
        keywords: item.palavras_chave || [],
        respostaBase: item.resposta || "",
      }));

      setIntents(mapped);
    } catch (err: any) {
      console.error("Erro ao carregar intenções:", err);
      notify(
        "error",
        "Erro",
        err?.response?.data?.message || "Erro ao carregar intenções!"
      );
    } finally {
      setLoadingIntents(false);
    }
  }, [notify]);

  React.useEffect(() => {
    loadUserState().then(() => {
      loadIntents();
    });
  }, [loadIntents]);

  React.useEffect(() => {
    return () => {
      if (tabTransitionTimer.current) clearTimeout(tabTransitionTimer.current);
    };
  }, []);

  const handleMainTabChange = React.useCallback((e: { index: number }) => {
    setActiveTabIndex(e.index);
    setTabTransitioning(true);
    if (tabTransitionTimer.current) clearTimeout(tabTransitionTimer.current);
    tabTransitionTimer.current = setTimeout(() => {
      setTabTransitioning(false);
      tabTransitionTimer.current = null;
    }, 400);
  }, []);

  const openNewIntent = () => {
    setEditingIntent(null);
    setIntentDialogOpen(true);
  };

  const openEditIntent = (intent: Intent) => {
    setEditingIntent(intent);
    setIntentDialogOpen(true);
  };

  const closeIntentDialog = () => {
    setIntentDialogOpen(false);
    setEditingIntent(null);
  };

  const saveIntent = async (payload: Omit<Intent, "id">) => {
    try {
      const currentUser = user.get() as any;
      const teamId = currentUser?.team?.id || currentUser?.team_id || "";

      const backendPayload = {
        nome: payload.titulo,
        palavras_chave: payload.keywords,
        resposta: payload.respostaBase,
        ativa: payload.ativo,
        prioridade: 50,
        team_id: teamId,
      };

      if (editingIntent) {
        await api().patch(
          baseURL() + `/agente-ia/intencoes/${editingIntent.id}`,
          backendPayload as any
        );
        notify("success", "Atualizado", "Intenção atualizada com sucesso");
      } else {
        await api().post(baseURL() + "/agente-ia/intencoes", backendPayload);
        notify("success", "Criado", "Nova intenção criada com sucesso");
      }

      await loadIntents();
      closeIntentDialog();
    } catch (err: any) {
      console.error("Erro ao salvar intenção:", err);
      notify(
        "error",
        "Erro",
        err?.response?.data?.message || "Erro ao salvar intenção"
      );
    }
  };

  const deleteIntent = async (id: string) => {
    try {
      await fetch(baseURL() + `/agente-ia/intencoes/${id}`, {
        method: "DELETE",
        headers: {
          Authorization: (user.get() as any)?.token || "",
        },
      });

      notify("success", "Excluído", "Intenção removida com sucesso");
      await loadIntents();
    } catch (err: any) {
      console.error("Erro ao deletar intenção:", err);
      notify(
        "error",
        "Erro",
        err?.response?.data?.message || "Erro ao deletar intenção"
      );
    }
  };

  const toggleActive = async (id: string, next: boolean) => {
    try {
      await api().patch(baseURL() + `/agente-ia/intencoes/${id}`, {
        ativa: next,
      } as any);

      setIntents((prev) =>
        prev.map((i) => (i.id === id ? { ...i, ativo: next } : i))
      );
    } catch (err: any) {
      console.error("Erro ao alterar status:", err);
      notify(
        "error",
        "Erro",
        err?.response?.data?.message || "Erro ao alterar status"
      );
    }
  };

  return (
    <div
      className="mestre-ia-page bg-white border-round-3xl p-4"
      style={{ position: "relative" }}
    >
      <Toast ref={toast} />

      {tabTransitioning ? <MestreIaTransitionLoader overlay caption="Carregando…" /> : null}

      <TabView
        activeIndex={activeTabIndex}
        onTabChange={handleMainTabChange}
        scrollable
        className="mestre-ia-tabs"
      >
        <TabPanel header="Campanhas" leftIcon="pi pi-megaphone">
          <CampanhasTab />
        </TabPanel>

        <TabPanel header="Agente IA" leftIcon="pi pi-bolt">
          <AgenteIaTab
            intents={intents}
            loading={loadingIntents}
            onNew={openNewIntent}
            onEdit={openEditIntent}
            onDelete={deleteIntent}
            onToggleActive={toggleActive}
          />
        </TabPanel>

        <TabPanel header="Leads" leftIcon="pi pi-users">
          <LeadsTab />
        </TabPanel>

        <TabPanel header="Relatórios" leftIcon="pi pi-chart-line">
          <DashboardTab />
        </TabPanel>

        <TabPanel header="Conexão" leftIcon="pi pi-qrcode">
          <ConexaoTab />
        </TabPanel>
      </TabView>

      <IntentDialog
        visible={intentDialogOpen}
        editing={editingIntent}
        onClose={closeIntentDialog}
        onSave={saveIntent}
        toastRef={toast}
      />

      <style>{`
        .mestre-ia-page {
          width: 100%;
        }

        .mestre-ia-tabs .p-tabview-nav-container {
          background: #f7f7f8;
          border-radius: 1.2rem;
          padding: 0.35rem;
          border: 1px solid #ececec;
        }

        .mestre-ia-tabs .p-tabview-nav {
          background: transparent;
          border: 0 !important;
          gap: 0.35rem;
        }

        .mestre-ia-tabs .p-tabview-nav li {
          margin: 0 !important;
          flex: 0 0 auto;
        }

        .mestre-ia-tabs .p-tabview-nav li .p-tabview-nav-link {
          background: transparent !important;
          border: 1px solid transparent !important;
          border-radius: 0.9rem;
          padding: 0.95rem 1rem !important;
          min-height: 48px;
          display: flex;
          align-items: center;
          gap: 0.45rem;
          white-space: nowrap;
          color: #6b7280 !important;
          font-weight: 600;
          box-shadow: none !important;
          transition: all 0.2s ease;
        }

        .mestre-ia-tabs .p-tabview-nav li .p-tabview-nav-link:focus {
          box-shadow: none !important;
        }

        .mestre-ia-tabs .p-tabview-nav li.p-highlight .p-tabview-nav-link {
          background: #ffffff !important;
          border-color: #c7d2fe !important;
          color: #5b5ce2 !important;
        }

        .mestre-ia-tabs .p-tabview-left-icon {
          margin-right: 0 !important;
          font-size: 0.95rem;
          line-height: 1;
        }

        .mestre-ia-tabs .p-tabview-panels {
          background: transparent;
          padding: 1rem 0 0 0;
        }

        .mestre-ia-tabs .p-tabview-nav-btn {
          width: 2.25rem;
          color: #6b7280;
          background: transparent;
          border-radius: 0.8rem;
        }

        .mestre-ia-tabs .p-tabview-nav-btn:hover {
          background: rgba(0, 0, 0, 0.04);
        }

        @media screen and (max-width: 768px) {
          .mestre-ia-page {
            padding: 0.85rem;
            border-radius: 1rem;
          }

          .mestre-ia-tabs .p-tabview-nav-container {
            padding: 0.25rem;
            border-radius: 1rem;
          }

          .mestre-ia-tabs .p-tabview-nav {
            gap: 0.25rem;
          }

          .mestre-ia-tabs .p-tabview-nav li .p-tabview-nav-link {
            padding: 0.75rem 0.85rem !important;
            min-height: 42px;
            font-size: 0.92rem;
            border-radius: 0.8rem;
          }

          .mestre-ia-tabs .p-tabview-left-icon {
            font-size: 0.85rem;
          }

          .mestre-ia-tabs .p-tabview-panels {
            padding-top: 0.8rem;
          }

          .mestre-ia-tabs .p-tabview-nav-btn {
            width: 2rem;
          }
        }

        @media screen and (max-width: 480px) {
          .mestre-ia-page {
            padding: 0.65rem;
          }

          .mestre-ia-tabs .p-tabview-nav li .p-tabview-nav-link {
            padding: 0.7rem 0.75rem !important;
            font-size: 0.88rem;
            min-height: 40px;
          }

          .mestre-ia-tabs .p-tabview-left-icon {
            font-size: 0.8rem;
          }
        }
      `}</style>
    </div>
  );
}