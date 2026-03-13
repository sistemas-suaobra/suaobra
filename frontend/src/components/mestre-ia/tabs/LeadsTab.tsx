import React from "react";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";
import { Toast } from "primereact/toast";
import { Paginator } from "primereact/paginator";
import { InputSwitch } from "primereact/inputswitch";
import { useStore } from "@nanostores/react";
import { obrasPlusCity, isWaiting, user, loadUserState } from "../../../store/store";
import { type City, makeCity } from "../../../store/cities";
import { api, baseURL } from "../../../store/api";

interface ResultRecord {
  obra_id: string;
  owner: string;
  professional: string;
  has_professional_phone: boolean;
  has_professional_email: boolean;
  has_owner_phone: boolean;
  has_owner_email: boolean;
  address: string;
  bairro: string;
  city: string;
  state: string;
  size: number;
  obra_number: number;
  type: string;
  start_date: string;
  end_date: string;
  visited_at: string;
  favorited_at: string;
  excluded_at: string;
  owner_contact_pending_at: string;
  professional_contact_pending_at: string;
  owner_contacted_at: string;
  professional_contacted_at: string;
  has_note: boolean;
}

interface ResultNeighborhoodRecord {
  bairro: string;
}

interface Result {
  total: number;
  records: ResultRecord[];
}

const orders = [
  { label: "Mais recente", value: "first_listing_date-desc,start_date-desc" },
  { label: "Mais antiga", value: "first_listing_date-asc,start_date-asc" },
  { label: "Maior tamanho", value: "size-desc" },
  { label: "Menor tamanho", value: "size-asc" },
];

export default function LeadsTab() {
  const toast = React.useRef<Toast>(null);
  const $obrasPlusCity = useStore(obrasPlusCity);

  const [loading, setLoading] = React.useState(false);
  const [records, setRecords] = React.useState<ResultRecord[]>([]);
  const [total, setTotal] = React.useState(0);
  const [citiesOptions, setCitiesOptions] = React.useState<City[]>([]);
  const [neighborhoodsOptions, setNeighborhoodsOptions] = React.useState<ResultNeighborhoodRecord[]>([]);

  const [selectedCity, setSelectedCity] = React.useState<City | null>($obrasPlusCity || null);
  const [selectedNeighborhood, setSelectedNeighborhood] = React.useState<ResultNeighborhoodRecord[]>([]);
  const [selectedOrder, setSelectedOrder] = React.useState(orders[0].value);
  const [filterValue, setFilterValue] = React.useState("");

  const [offset, setOffset] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [refresh, setRefresh] = React.useState(0);

  const [ocultarJaContactados, setOcultarJaContactados] = React.useState(false);

  const notify = (severity: "success" | "info" | "warn" | "error", summary: string, detail: string) => {
    toast.current?.show({ severity, summary, detail, life: 3000 });
  };

  React.useEffect(() => {
    loadUserState().then((userData) => {
      const cities = userData.team?.cities?.sort().map((id: string) => makeCity(id)) || [];
      if (cities.length) {
        setCitiesOptions(cities);

        let savedCity = localStorage.getItem("obrasPlusCity");
        if (!savedCity || !cities.map((c: City) => c.id).includes(savedCity)) {
          savedCity = cities[0].id;
        }

        const city = makeCity(savedCity);
        setSelectedCity(city);
        obrasPlusCity.set(city);
      }
    });
  }, []);

  React.useEffect(() => {
    if (!selectedCity) return;

    const fetchNeighborhoods = async () => {
      try {
        const resp = await api().get(`${baseURL()}/query/obras-plus-neighborhood`, {
          city: selectedCity.city || "",
        });
        if (resp.error) throw new Error(resp.error);

        const data = await resp.response.json();
        const barrios = (data.barrios as string[]) || [];
        setNeighborhoodsOptions(barrios.map((b) => ({ bairro: b })));
      } catch (error) {
        console.error("Erro ao buscar bairros:", error);
      }
    };

    fetchNeighborhoods();
  }, [selectedCity]);

  React.useEffect(() => {
    if (!selectedCity) return;

    const fetchLeads = async () => {
      setLoading(true);
      isWaiting.set(true);

      try {
        const payload = {
          city: selectedCity.city || "",
          bairro: (selectedNeighborhood || []).map((r) => r.bairro).join("|"),
          order: selectedOrder,
          filter: filterValue,
          sizeMin: "0",
          sizeMax: "9999999",
          offset: offset.toString(),
          itemsPerPage: rowsPerPage.toString(),
          enriched: "false",
          startDateFrom: "",
          startDateTo: "",
          endDateFrom: "",
          endDateTo: "",
          ocultarJaContactados: ocultarJaContactados ? "true" : "false",
        };

        const resp = await api().get(`${baseURL()}/query/leads-plus`, payload);
        if (resp.error) throw new Error(resp.error);

        const data = (await resp.response.json()) as Result;
        setRecords(data.records || []);
        setTotal(data.total || 0);
      } catch (error) {
        console.error("Erro ao buscar leads:", error);
        setRecords([]);
        setTotal(0);
      } finally {
        setLoading(false);
        isWaiting.set(false);
      }
    };

    fetchLeads();
  }, [
    selectedCity,
    selectedNeighborhood,
    selectedOrder,
    offset,
    rowsPerPage,
    refresh,
    filterValue,
    ocultarJaContactados,
  ]);

  const doRefresh = () => setRefresh((r) => r + 1);

  const toggleFavorite = async (obra_id: string) => {
    try {
      const data = {
        team_id: user.get().team?.id || "",
        obra_id,
        toggle_col: "favorited_at",
      };

      const resp = await api().patch(`${baseURL()}/patch/lead-toggle`, data);
      if (resp.error) throw new Error(resp.error);

      setRecords((prev) => prev.filter((r) => r.obra_id !== obra_id));
      setTotal((prev) => Math.max(prev - 1, 0));
      notify("success", "Removido", "Lead removido dos favoritos");
    } catch (error) {
      notify("error", "Erro", "Erro ao remover lead");
    }
  };

  const onPageChange = (e: { first: number; rows: number }) => {
    setOffset(e.first);
    setRowsPerPage(e.rows);
  };

  return (
    <div className="w-full">
      <Toast ref={toast} />

      <div className="flex align-items-center justify-content-between mb-3 flex-wrap gap-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Leads</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Todas as obras possíveis para contato na sua região ({total} encontrados)
          </div>
        </div>

        <div className="flex align-items-center gap-2">
          <span className="text-secondary" style={{ fontSize: 14 }}>
            Ocultar já contactados
          </span>
          <InputSwitch
            checked={ocultarJaContactados}
            onChange={(e) => {
              setOffset(0);
              setOcultarJaContactados(!!e.value);
            }}
          />
          <Button
            icon="pi pi-refresh"
            className="p-button-text p-button-rounded"
            severity="secondary"
            tooltip="Atualizar"
            tooltipOptions={{ position: "top" }}
            onClick={doRefresh}
          />
        </div>
      </div>

      <div className="border-round-3xl p-2 surface-border">
        {loading ? (
          <div className="flex justify-content-center align-items-center" style={{ minHeight: 120 }}>
            <i className="pi pi-spin pi-spinner" style={{ fontSize: 32 }} />
          </div>
        ) : records.length === 0 ? (
          <div className="text-center py-5 text-secondary">
            Nenhum lead encontrado.
          </div>
        ) : (
          <div className="grid">
            {records.map((row) => {
              const contactado = !!(row.owner_contacted_at || row.professional_contacted_at);
              const pendente = !!(row.owner_contact_pending_at || row.professional_contact_pending_at);

              return (
                <div key={row.obra_id} className="col-12 md:col-6 lg:col-4 xl:col-3 p-2">
                  <div className="border-round-xl border-1 surface-border p-3 h-full flex flex-column justify-content-between">
                    <div>
                      <div className="flex align-items-center gap-2 mb-2">
                        <strong style={{ fontSize: 16 }}>
                          {row.owner || row.professional || "Sem nome"}
                        </strong>

                        {contactado ? (
                          <i
                            className="pi pi-check-circle"
                            style={{ color: "#22C55E", fontSize: 16 }}
                            title="Já contactado"
                          />
                        ) : null}

                        <i
                          className="pi pi-star-fill"
                          style={{ color: "#F59E0B", fontSize: 16 }}
                          title="Lead"
                        />
                      </div>

                      {row.professional && row.owner ? (
                        <div className="mb-2 text-secondary">{row.professional}</div>
                      ) : null}

                      <div className="mb-2 text-secondary">
                        <i className="pi pi-map-marker" /> {row.city}, {row.bairro}
                      </div>

                      <div className="mb-2 text-secondary">
                        <i className="pi pi-home" /> {row.type?.split(" - ")[1] || row.type || "N/A"} |{" "}
                        {row.size?.toFixed(0) || 0}m²
                      </div>

                      <div className="mb-3" style={{ lineHeight: 1.7 }}>
                        {(row.has_owner_phone || row.has_professional_phone) && (
                          <div className="flex align-items-center gap-2 text-secondary">
                            <i className="pi pi-phone" />
                            <span>
                              {row.has_owner_phone ? "Proprietário" : ""}
                              {row.has_owner_phone && row.has_professional_phone ? " / " : ""}
                              {row.has_professional_phone ? "Profissional" : ""}
                            </span>
                          </div>
                        )}

                        {(row.has_owner_email || row.has_professional_email) && (
                          <div className="flex align-items-center gap-2 text-secondary">
                            <i className="pi pi-envelope" />
                            <span>
                              {row.has_owner_email ? "Proprietário" : ""}
                              {row.has_owner_email && row.has_professional_email ? " / " : ""}
                              {row.has_professional_email ? "Profissional" : ""}
                            </span>
                          </div>
                        )}

                        {!row.has_owner_phone &&
                          !row.has_professional_phone &&
                          !row.has_owner_email &&
                          !row.has_professional_email && (
                            <span className="text-secondary">Sem contato</span>
                          )}
                      </div>

                      <div className="mb-2">
                        {contactado ? (
                          <Tag value="Contactado" severity="success" className="border-round-xl" />
                        ) : pendente ? (
                          <Tag value="Pendente" severity="warning" className="border-round-xl" />
                        ) : (
                          <Tag value="Novo" severity="info" className="border-round-xl" />
                        )}
                      </div>
                    </div>

                    <div className="flex align-items-center gap-2 justify-content-end mt-2">
                      <Button
                        icon="pi pi-star-fill"
                        className="p-button-text p-button-rounded"
                        severity="warning"
                        tooltip="Remover dos leads"
                        tooltipOptions={{ position: "top" }}
                        onClick={() => toggleFavorite(row.obra_id)}
                      />
                      <Button
                        icon="pi pi-external-link"
                        className="p-button-text p-button-rounded"
                        severity="secondary"
                        tooltip="Ver em Obras+"
                        tooltipOptions={{ position: "top" }}
                        onClick={() => window.open(`/obras-plus?obra=${row.obra_id}`, "_blank")}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <Paginator
          first={offset}
          rows={rowsPerPage}
          totalRecords={total}
          rowsPerPageOptions={[10, 25, 50]}
          onPageChange={onPageChange}
          className="border-top-1 surface-border mt-2 pt-2"
        />
      </div>
    </div>
  );
}