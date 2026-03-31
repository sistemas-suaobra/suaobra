import React from "react";
import { TabView, TabPanel } from "primereact/tabview";
import { Calendar } from "primereact/calendar";
import { Button } from "primereact/button";

import CamapanhasDashboardTab from "./sub-tabs/dashboard/CamapanhasDashboardTab";
import { MestreIaTransitionLoader } from "../MestreIaTransitionLoader";

export default function DashboardTab() {
  const subTabTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);
  const [dashSubIndex, setDashSubIndex] = React.useState(0);
  const [dashSubTransitioning, setDashSubTransitioning] = React.useState(false);

  const [dataInicio, setDataInicio] = React.useState<Date | null>(null);
  const [dataFim, setDataFim] = React.useState<Date | null>(null);

  React.useEffect(() => {
    return () => {
      if (subTabTimer.current) clearTimeout(subTabTimer.current);
    };
  }, []);

  const handleDashSubTabChange = React.useCallback((e: { index: number }) => {
    setDashSubIndex(e.index);
    setDashSubTransitioning(true);
    if (subTabTimer.current) clearTimeout(subTabTimer.current);
    subTabTimer.current = setTimeout(() => {
      setDashSubTransitioning(false);
      subTabTimer.current = null;
    }, 400);
  }, []);

  const clearFilters = () => {
    setDataInicio(null);
    setDataFim(null);
  };

  return (
    <div className="w-full">
      {/* Header */}
      <div className="mb-3">
        <div style={{ fontSize: 18, fontWeight: 700 }}>Relatórios e Logs</div>
        <div className="text-secondary" style={{ marginTop: 4 }}>
          Acompanhe o desempenho e histórico do sistema
        </div>
      </div>

      {/* Filtros (gerais do dashboard) */}
      <div className="bg-white border-round-2xl p-4 mb-3" style={{ border: "1px solid rgba(0,0,0,0.06)" }}>
        <div className="formgrid grid align-items-end">
          <div className="field col-12 md:col-3">
            <label style={{ fontWeight: 700 }}>Data Início</label>
            <Calendar
              value={dataInicio}
              onChange={(e) => setDataInicio(e.value as Date | null)}
              dateFormat="dd/mm/yy"
              placeholder="dd/mm/aaaa"
              showIcon
              className="w-full"
            />
          </div>

          <div className="field col-12 md:col-3">
            <label style={{ fontWeight: 700 }}>Data Fim</label>
            <Calendar
              value={dataFim}
              onChange={(e) => setDataFim(e.value as Date | null)}
              dateFormat="dd/mm/yy"
              placeholder="dd/mm/aaaa"
              showIcon
              className="w-full"
            />
          </div>

          <div className="field col-12 md:col-3">
            <Button label="Limpar Filtros" icon="pi pi-filter-slash" severity="secondary" onClick={clearFilters} />
          </div>
        </div>
      </div>

      {/* Tabs internas */}
      <div className="border-round-2xl overflow-hidden" style={{ position: "relative" }}>
        {dashSubTransitioning ? (
          <MestreIaTransitionLoader overlay caption="Carregando…" />
        ) : null}
        <TabView activeIndex={dashSubIndex} onTabChange={handleDashSubTabChange}>
          <TabPanel header="Campanhas" leftIcon="pi pi-fw pi-megaphone">
            <CamapanhasDashboardTab />
          </TabPanel>
        </TabView>
      </div>
    </div>
  );
}