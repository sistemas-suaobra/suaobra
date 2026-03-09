import React from "react";
import { TabView, TabPanel } from "primereact/tabview";
import { Toast } from "primereact/toast";
import ConexaoTab from "./tabs/ConexaoTab";
import LeadsTab from "./tabs/LeadsTab";
import AgenteIaTab from "./tabs/AgenteIATab";
import CampanhasTab from "./tabs/CampanhasTab";
import { IntentDialog, type Intent } from "./tabs/IntentDialog";
import DashboardTab from "./tabs/DashboardTab";
import { api, baseURL } from "../../store/api";
import { loadUserState, user } from "../../store/store";

export default function MestreIaPage() {
  const toast = React.useRef<Toast>(null);
  const [intents, setIntents] = React.useState<Intent[]>([]);
  const [loadingIntents, setLoadingIntents] = React.useState(false);
  const [intentDialogOpen, setIntentDialogOpen] = React.useState(false);
  const [editingIntent, setEditingIntent] = React.useState<Intent | null>(null);

  const notify = (severity: "success" | "info" | "warn" | "error", summary: string, detail: string) => {
    toast.current?.show({ severity, summary, detail, life: 3000 });
  };

  // Carrega intenções do backend
  const loadIntents = React.useCallback(async () => {
    setLoadingIntents(true);
    try {
      const resp = await api().get(baseURL() + "/agente-ia/intencoes");
      const data = await resp.json();
      
      // Mapeia do formato backend para frontend
      const mapped: Intent[] = (data?.intencoes || []).map((item: any) => ({
        id: item.id,
        titulo: item.nome || "",
        ativo: item.ativa || false,
        keywords: item.palavras_chave || [],
        respostaBase: item.resposta || "",
      }));
      
      setIntents(mapped);
    } catch (err: any) {
      console.error("Erro ao carregar intenções1:", err);
      notify("error", "Erro", err?.response?.data?.message || "Erro ao carregar intenções");
    } finally {
      setLoadingIntents(false);
    }
  }, []);

  // Carrega user e intenções na montagem
  React.useEffect(() => {
    loadUserState().then(() => {
      loadIntents();
    });
  }, [loadIntents]);

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

      // Mapeia de frontend para backend
      const backendPayload = {
        nome: payload.titulo,
        palavras_chave: payload.keywords, // Array de strings, não string separada por vírgulas
        resposta: payload.respostaBase,
        ativa: payload.ativo, // Boolean, não string
        prioridade: 50, // Number, não string
        team_id: teamId,
      };

      if (editingIntent) {
        // Atualiza intenção existente
        await api().patch(
          baseURL() + `/agente-ia/intencoes/${editingIntent.id}`,
          backendPayload as any
        );
        notify("success", "Atualizado", "Intenção atualizada com sucesso");
      } else {
        // Cria nova intenção
        await api().post(baseURL() + "/agente-ia/intencoes", backendPayload);
        notify("success", "Criado", "Nova intenção criada com sucesso");
      }

      // Recarrega lista
      await loadIntents();
      closeIntentDialog();
    } catch (err: any) {
      console.error("Erro ao salvar intenção:", err);
      notify("error", "Erro", err?.response?.data?.message || "Erro ao salvar intenção");
    }
  };

  const deleteIntent = async (id: string) => {
    try {
      await fetch(baseURL() + `/agente-ia/intencoes/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': (user.get() as any)?.token || '' },
      });
      notify("success", "Excluído", "Intenção removida com sucesso");
      await loadIntents();
    } catch (err: any) {
      console.error("Erro ao deletar intenção:", err);
      notify("error", "Erro", err?.response?.data?.message || "Erro ao deletar intenção");
    }
  };

  const toggleActive = async (id: string, next: boolean) => {
    try {
      await api().patch(baseURL() + `/agente-ia/intencoes/${id}`, {
        ativa: next, // Boolean, não string
      } as any);
      
      // Atualiza localmente
      setIntents((prev) => prev.map((i) => (i.id === id ? { ...i, ativo: next } : i)));
    } catch (err: any) {
      console.error("Erro ao alterar status:", err);
      notify("error", "Erro", err?.response?.data?.message || "Erro ao alterar status");
    }
  };

  return (
    <div className="bg-white border-round-3xl p-4">
      <Toast ref={toast} />

      <TabView>
        <TabPanel header="Campanhas" leftIcon="pi pi-fw pi-megaphone">
          <CampanhasTab/>
        </TabPanel>

        <TabPanel header="Agente IA" leftIcon="pi pi-fw pi-bolt">
          <AgenteIaTab
            intents={intents}
            onNew={openNewIntent}
            onEdit={openEditIntent}
            onDelete={deleteIntent}
            onToggleActive={toggleActive}
          />
        </TabPanel>

        <TabPanel header="Leads" leftIcon="pi pi-fw pi-users">
          <LeadsTab />
        </TabPanel>

        <TabPanel header="Relatórios" leftIcon="pi pi-fw pi-chart-line">
          <DashboardTab />
        </TabPanel>

        <TabPanel header="Conexão" leftIcon="pi pi-fw pi-qrcode">
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
    </div>
  );
}