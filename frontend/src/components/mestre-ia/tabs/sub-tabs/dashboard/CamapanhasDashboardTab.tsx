import React from "react";
import { Chart } from "primereact/chart";

import { percent } from "../../../utils/dashboard";
import { api, makeURL } from "../../../../../store/api";

const StatCard = (props: {
  icon: string;
  iconBg: string;
  title: string;
  value: string | number;
  subtitle: string;
}) => {
  return (
    <div className="col-12 md:col-4">
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

export default function CamapanhasDashboardTab() {
  const [stats, setStats] = React.useState({
    enviadas: 0,
    lidas: 0,
    respostas: 0,
    taxa: 0,
    campanhas: [] as any[]
  });

  React.useEffect(() => {
    api().get(makeURL('/query/dashboard/campanhas'), {}).then(async resp => {
      const data = await resp.json();
      if (data && data.stats) {
        setStats(data.stats);
      }
    });
  }, []);

  const campanhaStats = React.useMemo(() => {
    const enviadas = stats.enviadas;
    const lidas = stats.lidas;
    const respostas = stats.respostas;
    // taxa already calculated in backend but we can use percent helper too if we want string format
    // backend returns float like 0.2, frontend expects formatted string or number?
    // Looking at usage below...
    
    // The original code was: const taxa = percent(respostas, enviadas);
    // percent helper likely returns a formatted string or number.
    
    const taxa = percent(respostas, enviadas);
    return { enviadas, lidas, respostas, taxa };
  }, [stats]);

  const campanhasBar = React.useMemo(() => {
    // Se não tiver dados, mostra array vazio para não quebrar
    const labels = (stats.campanhas || []).map(c => c.nome);
    const data = (stats.campanhas || []).map(c => c.enviados);

    const options = {
      plugins: { legend: { display: false } },
      scales: {
        x: { grid: { display: false } },
        y: { beginAtZero: true },
      },
    };

    return {
      data: {
        labels,
        datasets: [{ label: "Envios", data, backgroundColor: "#3B82F6" }],
      },
      options,
    };
  }, [stats]);

  const tiposCampanha = React.useMemo(() => {
    const labels = ["WhatsApp", "E-mail"];
    // Calculando percentual baseado no que temos (por enquanto tudo whatsapp)
    const whatsapp = stats.enviadas || 0;
    const email = 0; // Backend needs to separate this if mixed
    const data = [whatsapp, email];

    const options = {
      plugins: { legend: { position: "bottom" as const } },
      cutout: "70%",
    };

    return {
      data: { 
        labels, 
        datasets: [{ 
          data,
          backgroundColor: ["#3B82F6", "#EF4444"]
        }] 
      },
      options,
    };
  }, [stats]);

  return (
    <div>
      <div className="flex align-items-center justify-content-between mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Campanhas</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Acompanhe o desempenho de suas campanhas de marketing
          </div>
        </div>
      </div>

      <div className="grid">
        <StatCard
          icon="pi pi-send"
          iconBg="rgba(59,130,246,0.12)"
          title="Mensagens Enviadas"
          value={campanhaStats.enviadas}
          subtitle="Total no período"
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
          iconBg="rgba(249,115,22,0.12)"
          title="Taxa de Resposta"
          value={campanhaStats.taxa}
          subtitle="Respostas / envios"
        />

        <div className="col-12 md:col-8">
          <SectionCard title="Desempenho por Campanha" subtitle="Envios por campanha">
            <Chart type="bar" data={campanhasBar.data} options={campanhasBar.options} />
          </SectionCard>
        </div>

        <div className="col-12 md:col-4">
          <SectionCard title="Tipos de Campanha" subtitle="Distribuição por canal">
            <div className="flex justify-content-center">
              <Chart type="doughnut" data={tiposCampanha.data} options={tiposCampanha.options} className="w-full" />
            </div>
          </SectionCard>
        </div>
      </div>
    </div>
  );
}