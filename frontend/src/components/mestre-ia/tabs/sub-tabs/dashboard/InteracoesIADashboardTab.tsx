import React from "react";
import { Chart } from "primereact/chart";
import { percent } from "../../../utils/dashboard";

const StatCard = (props: {
  icon: string;
  iconBg: string;
  title: string;
  value: string | number;
  subtitle: string;
}) => {
  return (
    <div className="col-12 md:col-3">
      <div className="border-round-2xl p-4 bg-white" style={{ border: "1px solid rgba(0,0,0,0.06)" }}>
        <div className="flex align-items-center justify-content-between">
          <div>
            <div style={{ fontSize: 12, color: "#6B7280", fontWeight: 600 }}>{props.title}</div>
            <div style={{ fontSize: 28, fontWeight: 800, marginTop: 8 }}>{props.value}</div>
            <div style={{ fontSize: 12, color: "#6B7280", marginTop: 2 }}>{props.subtitle}</div>
          </div>

          <div
            className="flex align-items-center justify-content-center border-round-xl"
            style={{ width: 46, height: 46, background: props.iconBg }}
          >
            <i className={props.icon} style={{ fontSize: 18 }} />
          </div>
        </div>
      </div>
    </div>
  );
};

const SectionCard = (props: { title: string; subtitle?: string; children: React.ReactNode }) => {
  return (
    <div className="bg-white border-round-2xl p-4" style={{ border: "1px solid rgba(0,0,0,0.06)" }}>
      <div className="mb-3">
        <div style={{ fontSize: 16, fontWeight: 800 }}>{props.title}</div>
        {props.subtitle ? (
          <div className="text-secondary" style={{ marginTop: 4 }}>
            {props.subtitle}
          </div>
        ) : null}
      </div>
      {props.children}
    </div>
  );
};

export default function InteracoesIADashboardTab() {
  const campanhaStats = React.useMemo(() => {
    const enviadas = 7;
    const lidas = 0;
    const respostas = 0;
    const taxa = percent(respostas, enviadas);
    return { enviadas, lidas, respostas, taxa };
  }, []);

  const leadStats = React.useMemo(() => {
    return {
      total: 5,
      contatados: 2,
      naoContatados: 3,
      vips: 2,
    };
  }, []);

  const statusLeads = React.useMemo(() => {
    const labels = ["Já Contatados", "Não Contatados", "VIPs"];
    const data = [leadStats.contatados, leadStats.naoContatados, leadStats.vips];

    const options = {
      plugins: { legend: { position: "bottom" as const } },
      cutout: "70%",
    };

    return {
      data: { labels, datasets: [{ data }] },
      options,
    };
  }, [leadStats]);

  return (
    <div>
      <div className="flex align-items-center justify-content-between mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Interações IA</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Visualize as interações e automações com inteligência artificial
          </div>
        </div>
      </div>

      <div className="mb-3">
        <div className="grid">
          <StatCard
            icon="pi pi-users"
            iconBg="rgba(99,102,241,0.12)"
            title="Total de Leads"
            value={leadStats.total}
            subtitle="Leads cadastrados"
          />
          <StatCard
            icon="pi pi-user-plus"
            iconBg="rgba(249,115,22,0.12)"
            title="Já Contatados"
            value={leadStats.contatados}
            subtitle="Leads com contato"
          />
          <StatCard
            icon="pi pi-user-minus"
            iconBg="rgba(249,115,22,0.12)"
            title="Não Contatados"
            value={leadStats.naoContatados}
            subtitle="Aguardando contato"
          />
          <StatCard
            icon="pi pi-star-fill"
            iconBg="rgba(168,85,247,0.12)"
            title="VIPs"
            value={leadStats.vips}
            subtitle="Leads prioritários"
          />
        </div>
      </div>

      <div className="grid">
        <div className="col-12 lg:col-8">
          <SectionCard title="Resumo das Automações" subtitle="Métricas de envio (mock)">
            <div className="grid">
              <StatCard
                icon="pi pi-send"
                iconBg="rgba(59,130,246,0.12)"
                title="Mensagens Enviadas"
                value={campanhaStats.enviadas}
                subtitle="Total no período"
              />
              <StatCard
                icon="pi pi-eye"
                iconBg="rgba(34,197,94,0.12)"
                title="Mensagens Lidas"
                value={campanhaStats.lidas}
                subtitle="Aberturas registradas"
              />
              <StatCard
                icon="pi pi-comment"
                iconBg="rgba(168,85,247,0.12)"
                title="Respostas"
                value={campanhaStats.respostas}
                subtitle="Respostas do lead"
              />
              <StatCard
                icon="pi pi-chart-line"
                iconBg="rgba(245,158,11,0.12)"
                title="Taxa de Resposta"
                value={`${campanhaStats.taxa}%`}
                subtitle="Respostas / envios"
              />
            </div>
          </SectionCard>
        </div>

        <div className="col-12 lg:col-4">
          <SectionCard title="Status dos Leads" subtitle="Distribuição (mock)">
            <Chart type="doughnut" data={statusLeads.data} options={statusLeads.options} />
          </SectionCard>
        </div>
      </div>
    </div>
  );
}