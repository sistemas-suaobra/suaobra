import React from "react";
import { Dropdown } from 'primereact/dropdown';
import { InputText } from 'primereact/inputtext';
import { Paginator } from 'primereact/paginator';
import { Button } from 'primereact/button';
import { Calendar } from 'primereact/calendar';
import { obrasPlusCity, obrasPlusState, isWaiting, isLoaded, user, User, loadUserState, userS, Contact, obrasPlusNeighborhood, type Templates, TeamProperties } from '../../store/store';
import { allCities, type City, makeCity } from '../../store/cities.js';
import { MultiSelect } from 'primereact/multiselect';
import { api, baseURL, Response, makeURL, userTrackProps } from '../../store/api';
import { doToast, jsonClone, serialize } from "../../utils/methods.js";
import { Dialog } from 'primereact/dialog';
import { atom, } from "nanostores";
import { useStore } from '@nanostores/react'
import { DataTable } from 'primereact/datatable';
import { Column, type ColumnBodyOptions } from 'primereact/column';
import { Tooltip } from "primereact/tooltip";
import { useVariable, type ObjectAny, type ObjectBoolean, type ObjectNumber, type ObjectString } from "../../utils/interfaces.js";
import { OverlayPanel } from 'primereact/overlaypanel';
import { InputTextarea } from 'primereact/inputtextarea';
import { useHookstate, type State } from "@hookstate/core";
import { LogOut } from "../NavBar.js";
import { SelectButton } from 'primereact/selectbutton';
import { Toast } from "primereact/toast";
import { addLocale, locale } from 'primereact/api';
import { pt_BR } from "primelocale/js/pt_BR.js";

declare global {
  interface Window {
    rudderAnalytics: any
  }
}



interface Props {
  Restricted?: boolean
}


interface Result {
  total: number;
  records: ResultRecord[];
}

export const isPhoneContact = (c: string) => c.startsWith('cont_phone_')
export const isEmailContact = (c: string) => c.startsWith('cont_email_')
export const isPhoneOrEmailContact = (c: string) => isPhoneContact(c) || isEmailContact(c)

export interface ResultRecord {
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

export interface ResultNeighborhoodRecord {
  bairro: string;
}

const selectedContact = atom<Contact[]>([]);

export default function ObraPlusPage(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  addLocale('pt_BR', pt_BR);
  locale("pt_BR");

  const sizes = [
    { label: 'Todos', code: '0-9999999' },
    { label: 'Até 100m²', code: '0-100' },
    { label: 'De 100m² até 250m²', code: '100-250' },
    { label: 'De 250m² até 500m²', code: '250-500' },
    { label: 'De 500m² até 1.000m²', code: '500-1000' },
    { label: 'De 1.000mm² até 5.000m²', code: '1000-5000' },
    { label: 'Acima de 5.000m²', code: '5000-9999999' },
  ];

  const orders = [
    { label: 'Mais recente', code: 'first_listing_date-desc,start_date-desc' },
    { label: 'Mais antiga', code: 'first_listing_date-asc,start_date-asc' },
    { label: 'Maior tamanho primeiro', code: 'size-desc' },
    { label: 'Menor tamanho primeiro', code: 'size-asc' },
  ];

  const types = [
    { label: 'Todos', code: 'todos' },
    { label: '1 - Projeto', code: '1 - PROJETO' },
    { label: '2 - Execucao', code: '2 - EXECUCAO' },
    { label: '3 - Gestao', code: '3 - GESTAO' },
    { label: '4 - Meio Ambiente E Planejamento Regional E Urbano', code: '4 - MEIO AMBIENTE E PLANEJAMENTO REGIONAL E URBANO' },
    { label: '5 - Atividades Especiais Em Arquitetura E Urbanismo', code: '5 - ATIVIDADES ESPECIAIS EM ARQUITETURA E URBANISMO' },
    { label: '6 - Outras', code: 'outros' },
  ];

  const statuses = [
    { label: 'Todos', code: 'todos' },
    { label: 'Em Andamento', code: 'em-andamento' },
    { label: 'Com Telefone', code: 'com-telefone' },
    { label: 'Com Telefone (Proprietário)', code: 'com-telefone-proprietario' },
    { label: 'Com Telefone (Profissional)', code: 'com-telefone-profissional' },
    { label: 'Com Email', code: 'com-email' },
    { label: 'Com Observação', code: 'com-observacao' },
    { label: 'Já Visualizadas', code: 'ja-visitada' },
    { label: 'Não Visualizadas', code: 'nao-visitada' },
    { label: 'Contactado', code: 'contactado' },
    { label: 'Não Contactado', code: 'nao-contactado' },
    { label: 'Contato Pendente', code: 'contato-pendente' },
    { label: 'Excluída', code: 'excluida' },
    { label: 'Leads', code: 'favorita' },
    // {label: 'Excluídas', code: 'excluida'},
  ];

  ///////////////////////////  HOOKS  ///////////////////////////
  const $obrasPlusCity = useStore(obrasPlusCity)
  const $obrasPlusNeighborhood = useStore(obrasPlusNeighborhood)
  const $obrasPlusState = useStore(obrasPlusState)
  const [locked, setLocked] = React.useState(props.Restricted || false)
  const [selectedState, setSelectedState] = React.useState($obrasPlusState || 'SP')
  const [selectedCity, setSelectedCity] = React.useState<City>($obrasPlusCity || makeCity('SP-SAO_PAULO'))
  const [selectedNeighborhood, setSelectedNeighborhood] = React.useState<ResultNeighborhoodRecord[]>([])
  const [selectedSize, setSelectedSize] = React.useState(sizes[0].code)
  const [filterValue, setFilterValue] = React.useState('')
  const [selectedOrder, setSelectedOrder] = React.useState(orders[0].code)
  const [selectedStatuses, setSelectedStatuses] = React.useState([statuses[0].code])
  const [startDateFrom, setStartDateFrom] = React.useState<Date | null>(null)
  const [startDateTo, setStartDateTo] = React.useState<Date | null>(null)
  const [endDateFrom, setEndDateFrom] = React.useState<Date | null>(null)
  const [endDateTo, setEndDateTo] = React.useState<Date | null>(null)
  const [records, setRecords] = React.useState<ResultRecord[]>([])
  const [total, setTotal] = React.useState(0)
  const [refresh, setRefresh] = React.useState(0)
  const [citiesOptions, setCitiesOptions] = React.useState<City[]>([])
  const [neighborhoodsOptions, setNeighborhoodsOptions] = React.useState<ResultNeighborhoodRecord[]>([])

  const [allowContact, setAllowContact] = React.useState(20);
  const [allowExport, setAllowExport] = React.useState(0);
  const [loading, setLoading] = React.useState(false);
  const [pageNumber, setPageNumber] = React.useState(1);
  const [offset, setOffset] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(10);
  const [batchContacts, setBatchContacts] = React.useState<ResultRecord[]>([]);
  const messengerBatchDialogVisible = useHookstate(false);

  const [neighborhoodRequestId, setNeighborhoodRequestId] = React.useState(0);
  const latestRequestIdRef = React.useRef(0);

  const batchOverlayPanel = React.useRef(null);

  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    if (locked || citiesOptions.length === 0) return

    let val = localStorage.getItem('obrasPlusCity')
    if (!citiesOptions.map(c => c.id).includes(val)) {
      val = citiesOptions[0].id
      localStorage.setItem('obrasPlusCity', val)
    }

    if (val && val !== selectedCity.id) {
      setSelectedCity(makeCity(val))
      obrasPlusCity.set(makeCity(val))
      return
    }

    if (!selectedCity || !selectedOrder) return
    getRecords(rowsPerPage, selectedStatuses).then(
      result => {
        setTotal(result.total)
        setRecords(result.recs)
      }
    )

  }, [selectedCity, selectedNeighborhood, selectedSize, selectedOrder, selectedStatuses, offset, rowsPerPage, refresh, locked, citiesOptions, startDateFrom, startDateTo, endDateFrom, endDateTo])

  /*
  In the useEffect hook, we increment the neighborhoodRequestId and update the latestRequestIdRef before making the API call.
  We wrap the getCityNeighborhoodRecords call in an async function to handle the Promise properly.
  After getting the result, we check if the current request is still the latest one by comparing latestRequestIdRef.current with the currentRequestId. If they match, we update the state.
  We add a cleanup function that resets the latestRequestIdRef when the component unmounts or when the effect is about to run again.
  This approach ensures that only the result of the latest request is used to update the state, effectively handling race conditions that might occur due to rapid changes in selectedCity. The calls will still be made in the order they're triggered, but only the latest result will be applied to the state.
  */
  React.useEffect(() => {
    if (!selectedCity) return;

    const currentRequestId = neighborhoodRequestId + 1;
    setNeighborhoodRequestId(currentRequestId);
    latestRequestIdRef.current = currentRequestId;

    const fetchNeighborhoods = async () => {
      try {
        const result = await getCityNeighborhoodRecords();
        if (latestRequestIdRef.current === currentRequestId) {
          setNeighborhoodsOptions(result.recs);
        }
      } catch (error) {
        console.error("Error fetching neighborhoods:", error);
      }
    };

    fetchNeighborhoods();

    return () => {
      // Cleanup function
      latestRequestIdRef.current = 0;
    };
  }, [selectedCity]);

  React.useEffect(() => {
    isLoaded.set(false)
    loadUserState().then(
      (user) => {
        if (user.team.blocked) LogOut()

        let cities = user.team?.cities?.sort().map(id => makeCity(id))
        if (cities.length) {
          setCitiesOptions(cities || [])
          setAllowContact(user.team.allow_contact || 0)
          setAllowExport(user.team?.allow_export || 0)
          setLocked(false)
          isLoaded.set(true)
        } else {
          setLocked(true)
          isLoaded.set(false)
        }
      }
    )

    window.rudderAnalytics?.page({
      userId: user.get().id,
      category: "obras-plus",
      name: window.document.title,
    })

  }, [])
  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const doRefresh = () => setRefresh(refresh + 1)

  const initBatchSelection = async () => {
    // 1. Ensure SuaObra WhatsApp Connector Chrome Extension is installed, if not, show steps and links, and a GIF
    // 2. Ensure Whatsapp number matches extension (auto fill-in) 
    // 3. Ensure details in inputted: Company name, Description, Sender Name
    // 4. Generate templates if missing, 3 templates for owners, 3 for professionals.
    // 5. Get records whether number has been contacted already
    // 6. Ready for batching

    let result = await getRecords(allowContact, selectedStatuses.concat([]))
    setBatchContacts(result.recs)
    messengerBatchDialogVisible.set(true)
  }

  // Função auxiliar para formatar data para o backend
  const formatDateForAPI = (date: Date | null): string => {
    if (!date) return ''
    return date.toISOString().split('T')[0] // formato YYYY-MM-DD
  }

  // Função para resetar todos os filtros
  const resetAllFilters = () => {
    setFilterValue('')
    setStartDateFrom(null)
    setStartDateTo(null)
    setEndDateFrom(null)
    setEndDateTo(null)
    doRefresh()
  }

  const getRecords = async (rowsPerPage: number, statuses: string[]) => {
    let recs: ResultRecord[] = [];
    let total = 0
    let payload = {
      city: selectedCity.city || '',
      bairro: (selectedNeighborhood || []).map(r => r.bairro).join('|'),
      order: selectedOrder,
      filter: filterValue,
      statuses: statuses.join(','),
      sizeMin: selectedSize.split('-')[0],
      sizeMax: selectedSize.split('-')[1],
      offset: offset.toString(),
      itemsPerPage: rowsPerPage.toString(),
      enriched: `false`,
      startDateFrom: formatDateForAPI(startDateFrom),
      startDateTo: formatDateForAPI(startDateTo),
      endDateFrom: formatDateForAPI(endDateFrom),
      endDateTo: formatDateForAPI(endDateTo),
      // legacy_id: user.get().legacy_id,
    }

    try {
      isWaiting.set(true)
      let resp = await api().get(`${baseURL()}/query/obras-plus`, payload)
      if (resp.error) throw new Error(resp.error)
      let data = (await resp.response.json()) as Result
      recs = data.records
      total = data.total
    } catch (error) {
      console.log(error)
      recs = []
    } finally {
      isWaiting.set(false)
      window.rudderAnalytics?.track(
        'obras-plus-get-records',
        { user: userTrackProps(), page_number: pageNumber, rows_per_page: rowsPerPage, state: selectedState, city: selectedCity, neighborhood: (selectedNeighborhood || []).map(r => r.bairro), size: selectedSize, order: selectedOrder, statuses: selectedStatuses, filter: filterValue, startDateFrom: formatDateForAPI(startDateFrom), startDateTo: formatDateForAPI(startDateTo), endDateFrom: formatDateForAPI(endDateFrom), endDateTo: formatDateForAPI(endDateTo), itemsPerPage: rowsPerPage, total: total, records: recs.length }
      )
    }

    return { total, recs }
  }

  const getCityNeighborhoodRecords = async () => {
    let barrios: string[] = [];
    let recs: ResultNeighborhoodRecord[] = [];
    let total = 0
    let payload = {
      city: selectedCity.city || '',
    }

    if (!payload.city) return

    try {
      isWaiting.set(true)
      let resp = await api().get(`${baseURL()}/query/obras-plus-neighborhood`, payload)
      if (resp.error) throw new Error(resp.error)
      let data = await resp.response.json()
      barrios = data.barrios as string[]
      total = barrios.length
      recs = barrios.map(b => { return { bairro: b } })
    } catch (error) {
      console.log(error)
      recs = []
    } finally {
      isWaiting.set(false)
      window.rudderAnalytics?.track(
        'obras-neighborhood-get-records',
        { user: userTrackProps(), city: selectedCity, total: total }
      )
    }

    return { recs }
  }

  const getExcelExport = async (numItems: number) => {
    let recs: ResultRecord[] = [];
    let total = 0
    let payload = {
      city: selectedCity.city,
      bairro: (selectedNeighborhood || []).map(r => r.bairro).join('|'),
      order: selectedOrder,
      filter: filterValue,
      statuses: selectedStatuses.join(','),
      sizeMin: selectedSize.split('-')[0],
      sizeMax: selectedSize.split('-')[1],
      offset: offset.toString(),
      itemsPerPage: user.get().team?.allow_export?.toString(),
      export: 'true',
      startDateFrom: formatDateForAPI(startDateFrom),
      startDateTo: formatDateForAPI(startDateTo),
      endDateFrom: formatDateForAPI(endDateFrom),
      endDateTo: formatDateForAPI(endDateTo),
      // legacy_id: user.get().legacy_id,
    }

    try {
      isWaiting.set(true)
      let resp = await api().get(`${baseURL()}/query/obras-plus-export`, payload)
      if (resp.error) throw new Error(resp.error)
      let blob = await resp.response.blob()

      var a = document.createElement('a');
      a.href = URL.createObjectURL(blob);
      a.download = `${selectedCity.city}.${new Date().getTime()}.xlsx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
    } catch (error) {
      console.log(error)
      recs = []
    } finally {
      isWaiting.set(false)
      window.rudderAnalytics?.track(
        'obras-plus-export',
        { user: userTrackProps(), items: user.get().team?.allow_export, state: selectedState, city: selectedCity, size: selectedSize, order: selectedOrder, statuses: selectedStatuses, filter: filterValue, startDateFrom: formatDateForAPI(startDateFrom), startDateTo: formatDateForAPI(startDateTo), endDateFrom: formatDateForAPI(endDateFrom), endDateTo: formatDateForAPI(endDateTo) }
      )
    }

    return { total, recs }
  }


  const toggle = async (obra_id: string, type: 'visit' | 'favorite' | 'exclude') => {
    let i = records.map(r => r.obra_id).indexOf(obra_id)
    if (i === -1) return console.log(`could not find '${obra_id}' for toggle`)

    let data: ObjectString = {
      team_id: user.get().team?.id,
      obra_id: records[i].obra_id,
    }

    if (type === 'visit') data.toggle_col = 'visited_at'
    if (type === 'favorite') data.toggle_col = 'favorited_at'
    if (type === 'exclude') data.toggle_col = 'excluded_at'

    api().patch(makeURL('/patch/lead-toggle'), data)
      .then((resp) => {
        if (resp.error) return
        const newRecords = [...records]
        let date_val = (new Date()).toISOString()
        if (type === 'visit') newRecords[i].visited_at = date_val
        if (type === 'favorite')
          newRecords[i].favorited_at = newRecords[i].favorited_at ? null : date_val
        if (type === 'exclude')
          newRecords[i].excluded_at = newRecords[i].excluded_at ? null : date_val
        setRecords(newRecords)
        if (newRecords[i].excluded_at) setTotal(total - 1)
      })
  }
  ///////////////////////////  JSX  ///////////////////////////

  const Loader = <div className="flex justify-content-center" style={{ position: "relative", paddingTop: '40px', paddingBottom: '120px' }} >
    <div className="result-loader"></div>
  </div>

  if (locked) {
    return <div className="text-center">
      <div>Área Restrita.</div>
      <br />
      <div>
        Entre em contato conosco para configurar um plano: <a href="mailto:contato@suaobra.com.br">contato@suaobra.com.br</a>
      </div>
    </div>
  }

  return (
    <div>
      <div className="formgrid grid">
        <div className="field md:col-3 col-12">
          <label htmlFor="city-dropdown">Cidade</label>
          <Dropdown
            id='city-dropdown'
            value={selectedCity}
            onChange={(e) => {
              localStorage.setItem('obrasPlusCity', e.value.id)
              obrasPlusCity.set(e.value)
              setOffset(0)
              setPageNumber(1)
              setSelectedCity(e.value)
            }}
            options={citiesOptions}
            filter
            optionLabel="city"
            // optionValue="id"
            placeholder="Selecione uma Cidade"
            emptyMessage="Nenhuma cidade encontrada"
            className="w-full"
          />
        </div>

        <div className="field md:col-3 col-12">
          <label htmlFor="neighborhood-dropdown">Bairro</label>
          <MultiSelect
            id='neighborhood-dropdown'
            value={selectedNeighborhood}
            onChange={(e) => {
              let barrios = (e.value as ResultNeighborhoodRecord[]).map(v => v.bairro)
              localStorage.setItem('obrasPlusNeighborhood', barrios.join('|'))
              obrasPlusNeighborhood.set(barrios)
              setOffset(0)
              setPageNumber(1)
              setSelectedNeighborhood(e.value)
            }}
            virtualScrollerOptions={{ itemSize: 43 }}
            options={neighborhoodsOptions}
            filter multiple
            optionLabel="bairro"
            // optionValue="id"
            placeholder="Filtrar os bairros"
            className="w-full"
          />
        </div>

        <div className="field md:col-6 col-12">
          <label htmlFor="search-filter">Filtro de Pesquisa</label>
          <div className="p-inputgroup">
            <InputText
              id='search-filter'
              placeholder="Digite o que procura..."
              aria-describedby="filter-help"
              className="w-full"
              value={filterValue}
              onChange={(e) => {
                setFilterValue(e.target.value)
                if (!e.target.value) doRefresh() // auto-refreshed when cleared
              }}
              onKeyDown={(e) => {
                if (e.key === 'Escape') { setFilterValue(''); doRefresh() }
                if (e.key === 'Enter') { doRefresh() }
              }}
            />
            <Button
              icon="pi pi-times"
              className="p-button-warning"
              tooltip="Limpar todos os filtros"
              tooltipOptions={{ position: 'top' }}
              onClick={() => { resetAllFilters() }}
            />
            <Button
              icon="pi pi-search"
              className="p-button-primary"
              tooltip="Pesquisar"
              tooltipOptions={{ position: 'top' }}
              onClick={() => {
                doRefresh()
              }}
            />
          </div>
        </div>

        {/* <div className="field md:col-3 col-6">
            <label htmlFor="type-dropdown">Tipo</label>
            <Dropdown
              id='type-dropdown'
              value={selectedType}
              onChange={(e) => {
                setSelectedType(e.value)
              }}
              options={types}
              optionLabel="label" 
              optionValue="code"
              placeholder="Selecione um tipo"
              className="w-full"
            />
          </div> */}

        <div className="field md:col-6 col-12">
          <label htmlFor="size-order">Obra Status Filtros</label>
          <MultiSelect
            id='status-dropdown'
            value={selectedStatuses}
            onChange={(e) => {
              if (e.value.includes('todos') && selectedStatuses.length > 0 && !selectedStatuses.includes('todos'))
                e.value = ['todos']  // select only todos
              else if (e.value.includes('todos') && e.value.length > 1)
                e.value = e.value.filter(v => v !== 'todos') // unselect todos
              else if (e.value.includes('ja-visitada') && selectedStatuses.includes('nao-visitada'))
                e.value = e.value.filter(v => v !== 'nao-visitada')
              else if (e.value.includes('nao-visitada') && selectedStatuses.includes('ja-visitada'))
                e.value = e.value.filter(v => v !== 'ja-visitada')
              else if (e.value.length === 0)
                e.value = ['todos']  // select todos
              setSelectedStatuses(e.value)
            }}
            tooltip={
              [
                'EM ANDAMENTO:     Mostrar as obras já iniciadas',
                'COM TELEFONE:     Mostrar as obras com telefone',
                'COM EMAIL:        Mostrar as obras com email',
                'COM OBSERVAÇÃO:   Mostrar as obras com observação',
                'CONTACTADO:       Mostrar as obras já contactada',
                'NÃO CONTACTADO:   Mostrar as obras não contactada',
                'CONTATO PENDENTE: Mostrar as obras com contato pendente',
                'JÁ VISITADAS:     Mostrar as obras já visitadas',
                'NÃO VISITADAS:    Mostrar as obras não visitadas',
                'LEADS:            Mostrar as obras salvas como Lead',
                'EXCLUÍDAS:        Mostrar as obras excluídas',
              ].join('\n')
            }
            virtualScrollerOptions={{ itemSize: 43 }}
            tooltipOptions={{ position: 'top', style: { lineHeight: 1.5 } }}
            options={statuses}
            optionLabel="label"
            optionValue="code"
            display="chip"
            placeholder="Select Status"
            // maxSelectedLabels={3}
            className="w-full"
          />
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="size-dropdown">Tamanho M²</label>
          <Dropdown
            id='size-dropdown'
            value={selectedSize}
            onChange={(e) => {
              setSelectedSize(e.value)
            }}
            options={sizes}
            optionLabel="label"
            optionValue="code"
            placeholder="Selecione um tamanho"
            className="w-full"
          />
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="size-order">Ordenação</label>
          <Dropdown
            id='order-dropdown'
            value={selectedOrder}
            onChange={(e) => {
              setSelectedOrder(e.value)
            }}
            options={orders}
            optionLabel="label"
            optionValue="code"
            className="w-full"
          />
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="start-date-from">Data de Início (De)</label>
          <div className="p-inputgroup">
            <Calendar
              id='start-date-from'
              value={startDateFrom}
              onChange={(e) => {
                setStartDateFrom(e.value as Date | null)
              }}
              placeholder="Selecione uma data"
              dateFormat="dd/mm/yy"
              showIcon
              className="w-full"
              tooltip="Filtrar obras que iniciaram a partir desta data"
              tooltipOptions={{ position: 'top' }}
            />
            {startDateFrom && (
              <Button
                icon="pi pi-times"
                className="p-button-text p-button-sm"
                tooltip="Limpar"
                onClick={() => {
                  setStartDateFrom(null)
                }}
              />
            )}
          </div>
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="start-date-to">Data de Início (Até)</label>
          <div className="p-inputgroup">
            <Calendar
              id='start-date-to'
              value={startDateTo}
              onChange={(e) => {
                setStartDateTo(e.value as Date | null)
              }}
              placeholder="Selecione uma data"
              dateFormat="dd/mm/yy"
              showIcon
              className="w-full"
              tooltip="Filtrar obras que iniciaram até esta data"
              tooltipOptions={{ position: 'top' }}
            />
            {startDateTo && (
              <Button
                icon="pi pi-times"
                className="p-button-text p-button-sm"
                tooltip="Limpar"
                onClick={() => {
                  setStartDateTo(null)
                }}
              />
            )}
          </div>
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="end-date-from">Data de Fim (De)</label>
          <div className="p-inputgroup">
            <Calendar
              id='end-date-from'
              value={endDateFrom}
              onChange={(e) => {
                setEndDateFrom(e.value as Date | null)
              }}
              placeholder="Selecione uma data"
              dateFormat="dd/mm/yy"
              showIcon
              className="w-full"
              tooltip="Filtrar obras que terminam a partir desta data"
              tooltipOptions={{ position: 'top' }}
            />
            {endDateFrom && (
              <Button
                icon="pi pi-times"
                className="p-button-text p-button-sm"
                tooltip="Limpar"
                onClick={() => {
                  setEndDateFrom(null)
                }}
              />
            )}
          </div>
        </div>

        <div className="field md:col-3 col-6">
          <label htmlFor="end-date-to">Data de Fim (Até)</label>
          <div className="p-inputgroup">
            <Calendar
              id='end-date-to'
              value={endDateTo}
              onChange={(e) => {
                setEndDateTo(e.value as Date | null)
              }}
              placeholder="Selecione uma data"
              dateFormat="dd/mm/yy"
              showIcon
              className="w-full"
              tooltip="Filtrar obras que terminam até esta data"
              tooltipOptions={{ position: 'top' }}
            />
            {endDateTo && (
              <Button
                icon="pi pi-times"
                className="p-button-text p-button-sm"
                tooltip="Limpar"
                onClick={() => {
                  setEndDateTo(null)
                }}
              />
            )}
          </div>
        </div>
      </div>

      <div id="top-results" className='my-3 border-bottom-1 border-solid border-gray-300' />

      <div>
        {
          loading ?
            <>{Loader}</>
            :
            <>
              <div className="grid">
                <div className="flex md:col-6 col-12">
                  <h2>{total.toLocaleString('pt-br')} Obras Encontradas</h2>
                </div>
                <div className="flex my-3 md:col-6 col-12 justify-content-end">
                  {
                    allowContact > 0 &&
                    <Button
                      label="Disparo em Massa"
                      tooltip={`Enviar mensagens do WhatsApp em lote`}
                      tooltipOptions={{ position: 'top' }}
                      onClick={(e) => batchOverlayPanel.current.toggle(e)}
                      severity="info"
                      icon='pi pi-whatsapp'
                      className="mr-2"
                    >
                      <style>{`
                        div.p-overlaypanel-content {
                          padding: 7px !important;
                        }
                      `}</style>
                      <OverlayPanel ref={batchOverlayPanel} className="p-0">
                        <Button
                          className="button mr-1"
                          label="Novo Lote"
                          severity="info"
                          onClick={initBatchSelection}
                          tooltip={`Selecione os primeiros ${allowContact} resultados para contato via WhatsApp`}
                          tooltipOptions={{ position: 'bottom', style: { width: '300px' } }}
                        />
                        <Button
                          className="button"
                          label="Lotes Antigos"
                          severity="info"
                          tooltip="Veja os leads que foram contatados"
                          tooltipOptions={{ position: 'bottom' }}
                          onClick={() => { setSelectedStatuses(selectedStatuses.concat('contactado').filter(s => s !== 'todos')) }}
                        />
                      </OverlayPanel>
                    </Button>
                  }
                  {
                    allowExport > 0 ?
                      <Button
                        label="Exportar Leads"
                        tooltip={`Exportar os primeiros ${allowExport} resultados`}
                        tooltipOptions={{ position: 'top' }}
                        icon='pi pi-file-excel'
                        onClick={() => getExcelExport(allowExport)}
                      />
                      :
                      null
                  }
                </div>
              </div>
              {
                records
                  .filter(r => !(r.excluded_at && !selectedStatuses.includes('excluida')))
                  .map((record, i) =>
                    <React.Fragment key={i}>
                      {/* {RecordCard(record, (type: 'visit'|'favorite'|'exclude') => toggle(i, type))} */}
                      <RecordCard record={record} toggle={(type: 'visit' | 'favorite' | 'exclude') => toggle(record.obra_id, type)} />
                    </React.Fragment>
                  )
              }
            </>
        }
      </div>

      <ContactModal />
      <MessengerBatchDialog records={batchContacts} visible={messengerBatchDialogVisible} />

      <div className="card">
        <Paginator
          first={offset}
          rows={rowsPerPage}
          totalRecords={total}
          rowsPerPageOptions={[10, 25, 50]}
          onPageChange={(event) => {
            setPageNumber(event.page + 1);
            setOffset(event.first);
            setRowsPerPage(event.rows);
            document.getElementById('top-results').scrollIntoView()
          }}
        />
      </div>

    </div>
  );
}

const googleSearch = (query: string) => `https://www.google.com/search?q=${query}`
export const whatsAppURL = (number: string) => {
  if (navigator.userAgent.match(/Android|iPhone|iPad|iPod/i)) {
    return `whatsapp://send?phone=${number}&text=Ol%C3%A1%20%20tudo%20bem?`;
  } else {
    return `https://web.whatsapp.com/send?phone=${number}&text=Ol%C3%A1%20%20tudo%20bem?`;
  }
};
export const telephoneURL = (number: string) => `tel:${number}`
export const emailURL = (email: string) => `mailto:${email}`

const parseDate = (date: string, add_days = 0) => {
  if (!date) return '';
  let parsed = Date.parse(date);
  if (isNaN(parsed)) return '';
  let d = new Date(parsed)
  if (add_days !== 0) {
    // Convert to epoch milliseconds, add days in milliseconds, then create new date
    const epochMs = d.getTime()
    const daysInMs = add_days * 24 * 60 * 60 * 1000
    d = new Date(epochMs + daysInMs)
  }
  return `${d.getUTCDate().toString().padStart(2, '0')}/${(d.getUTCMonth() + 1).toString().padStart(2, '0')}/${d.getUTCFullYear()}`
}

const calculateObraStage = (startDateStr: string, endDateStr: string): string => {
  const startDate = new Date(Date.parse(startDateStr))
  const endDate = new Date(Date.parse(endDateStr))
  const currentDate = new Date()

  // Se a data atual é posterior à data de término, a obra está finalizada
  if (currentDate > endDate) {
    return 'FINALIZADA'
  }

  // Calcular tempo total da obra em milissegundos
  const totalTime = endDate.getTime() - startDate.getTime()

  // Calcular tempo decorrido desde o início até agora
  const elapsedTime = currentDate.getTime() - startDate.getTime()

  // Se ainda não iniciou (data atual anterior à data de início)
  if (elapsedTime < 0) {
    return 'INICIO'
  }

  // Calcular porcentagem de progresso
  const percentage = (elapsedTime / totalTime) * 100

  // Determinar etapa baseada na porcentagem
  if (percentage <= 33) {
    return 'INICIO'
  } else if (percentage <= 66) {
    return 'ESTRUTURA'
  } else {
    return 'ACABAMENTO'
  }
}

interface RecordCardParams {
  record: ResultRecord;
  toggle: (type: 'visit' | 'favorite' | 'exclude') => void;
}

// const RecordCard = (record: ResultRecord, toggle: (type: 'visit'|'favorite'|'exclude') => void) => {
const RecordCard = (props: RecordCardParams) => {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const iconStyle = (hasInfo = true, isPending = false, isContacted = false) => {
    return {
      fontSize: '1.25rem',
      color: 'white',
      background: hasInfo ? (isContacted ? 'green' : (isPending ? 'orange' : 'navy')) : 'grey',
      cursor: 'pointer'
    }
  }

  const record = props.record
  const toggle = props.toggle

  const googleMapsURL = `https://www.google.com/maps/search/${record.address}`
  const hasOwnerPhone = record.has_owner_phone
  const hasOwnerEmail = record.has_owner_email
  const hasProfessionalPhone = record.has_professional_phone
  const hasProfessionalEmail = record.has_professional_email
  const ownerContactPending = !!record.owner_contact_pending_at
  const professionalContactPending = !!record.professional_contact_pending_at
  const ownerIsContacted = !!record.owner_contacted_at
  const professionalIsContacted = !!record.professional_contacted_at

  ///////////////////////////  HOOKS  ///////////////////////////
  const op = React.useRef(null);
  const noteRec = useHookstate({ id: '', obra_id: '', note: '', user_id: '' })

  ///////////////////////////  EFFECTS  ///////////////////////////

  React.useEffect(() => {
  }, [])

  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const loadNoteRec = async () => {
    let resp = await api().collection('obra_note').getFirstListItem(`obra_id = "${record.obra_id}"`)
    if (resp.error) return console.log(resp.error)
    let rec = await resp.record()
    noteRec.set(
      r => {
        r.id = rec.id
        r.obra_id = rec.obra_id
        r.note = rec.note
        r.user_id = rec.user_id
        return r
      }
    )
  }

  const saveNoteRec = async () => {
    let resp: Response
    if (!noteRec.note.get()) return

    let data = jsonClone(noteRec.get())
    data.user_id = user.get().id
    data.obra_id = record.obra_id

    if (noteRec.id.get())
      resp = await api().collection('obra_note').update(data.id, data)
    else
      resp = await api().collection('obra_note').create(data)

    if (resp.error) return console.log(resp.error)

    let rec = await resp.record()
    noteRec.set(
      r => {
        r.id = rec.id
        r.obra_id = rec.obra_id
        r.note = rec.note
        r.user_id = rec.user_id
        return r
      }
    )
  }

  const getContactRecords = async (record: ResultRecord, role: 'owner' | 'professional') => {
    let recs: Contact[] = [];
    let err = undefined
    let nome = role === 'owner' ? record.owner : record.professional
    let payload = {
      nome: nome,
      cidade: record.city,
      uf: record.state,
    }
    let url = `${baseURL()}/query/obras-plus-contacts?${serialize(payload)}`

    try {
      isWaiting.set(true)
      let response = await fetch(url)
      if (response.status >= 400) throw new Error(response.statusText)
      let data = (await response.json())
      recs = data.records.map(
        (rec: Contact) => {
          // add name
          rec.name = nome
          return new Contact(rec)
        }
      )
      // mark as visited
      toggle('visit')
    } catch (error) {
      console.log(error)
      err = error
      recs = []
    } finally {
      isWaiting.set(false)

      // set obra as read
      window.rudderAnalytics?.track(
        'obras-plus-vew-contact',
        { user: userTrackProps(), state: record.state, city: record.city, obra_id: record.obra_id, contact_name: nome, contact_role: role, error: err })
    }
    selectedContact.set(recs)
  }

  ///////////////////////////  JSX  ///////////////////////////
  return <>
    <div
      style={{
        paddingRight: '10px',
        backgroundColor: record.visited_at ? "#c8c9f7" : "white"
      }}
      className="border-round-3xl"
    >
      <div
        key={record.obra_number}
        className="formgrid grid border-round-left-3xl mr-1 p-4 mb-4 bg-white"
        style={{
          position: 'relative',
          borderTopRightRadius: '0',
          borderBottomRightRadius: '0',
        }}
      >

        {/* Column 1 */}
        <div className="field md:col-4 col-12">
          <p className="my-0"><strong>Proprietário</strong></p>
          <p className="mt-2"> {record.owner?.toUpperCase()}</p>

          <p>
            <a href={googleSearch(record.owner)} target="_blank">
              <i className="pi pi-google p-2 border-round-lg" style={iconStyle()}></i>
            </a>
            <i
              className="pi pi-phone p-2 border-round-lg ml-2"
              style={iconStyle(hasOwnerPhone)}
              onClick={() => hasOwnerPhone ? getContactRecords(record, 'owner') : null}
            />
            <i
              className="pi pi-whatsapp p-2 border-round-lg ml-2"
              style={iconStyle(hasOwnerPhone, ownerContactPending, ownerIsContacted)}
              onClick={() => hasOwnerPhone ? getContactRecords(record, 'owner') : null}
              title={ownerIsContacted ? 'Já Contactado' : ownerContactPending ? 'Contato Pendente' : null}
            />
            <i
              className="pi pi-envelope p-2 border-round-lg ml-2"
              style={iconStyle(hasOwnerEmail)}
              onClick={() => hasOwnerEmail ? getContactRecords(record, 'owner') : null}
            />
          </p>

          <br />
          <p className="my-0"><strong>Profissional</strong></p>
          <p className="mt-2"> {record.professional?.toUpperCase()}</p>

          <p className="mb-0">
            <a href={googleSearch(record.professional)} target="_blank">
              <i className="pi pi-google p-2 border-round-lg" style={iconStyle()}></i>
            </a>
            <i
              className="pi pi-phone p-2 border-round-lg ml-2"
              style={iconStyle(hasProfessionalPhone)}
              onClick={() => hasProfessionalPhone ? getContactRecords(record, 'professional') : null}
            />
            <i
              className="pi pi-whatsapp p-2 border-round-lg ml-2"
              style={iconStyle(hasProfessionalPhone, professionalContactPending, professionalIsContacted)}
              onClick={() => hasProfessionalPhone ? getContactRecords(record, 'professional') : null}
              title={professionalIsContacted ? 'Já Contactado' : professionalContactPending ? 'Contato Pendente' : null}
            />
            <i
              className="pi pi-envelope p-2 border-round-lg ml-2"
              style={iconStyle(hasProfessionalEmail)}
              onClick={() => hasProfessionalEmail ? getContactRecords(record, 'professional') : null}
            />
          </p>
        </div>

        {/* Column 2 */}
        <div className="field md:col-4 col-12">
          <p className="my-0"><strong>Endereço</strong></p>
          <p className="mt-2">
            <a target="_blank" href={googleMapsURL}>{record.address?.toUpperCase()}</a>
          </p>

          <br />
          <p className="my-0"><strong>Tamanho</strong></p>
          <p className="mt-2"> {record.size.toLocaleString('pt-br')} M²</p>

          <br />
          <p className="my-0"><strong>Etapa</strong></p>
          <p className="mt-2"> {calculateObraStage(record.start_date, record.end_date)} </p>
        </div>

        {/* Column 3 */}
        <div className="field md:col-4 col-12">
          <p className="my-0"><strong>Tipo</strong></p>
          <p className="mt-2"> {record.type} </p>

          {!!parseDate(record.start_date) && (
            <>
              <br />
              <br />
              <p className="my-0"><strong>Data de Início</strong></p>
              <p className="mt-2"> {parseDate(record.start_date)} </p>
            </>
          )}

          {!!parseDate(record.end_date) && (
            <>
              <br />
              {!parseDate(record.start_date) && <br />}
              <p className="my-0"><strong>Previsão de Término</strong></p>
              <p className="mt-2"> {parseDate(record.end_date)} </p>
            </>
          )}
        </div>

        <span
          style={{
            position: 'absolute',
            right: '1%',
            marginTop: '-15px',
          }}
        >
          <Button
            icon={record.excluded_at ? "pi pi-plus-circle" : "pi pi-times-circle"}
            tooltip={record.excluded_at ? "Incluir Obra nos Resultados" : "Excluir Obra dos Resultados"}
            tooltipOptions={{ position: 'top' }}
            className={"p-button-rounded p-button-text " + (record.excluded_at ? " p-button-info" : " p-button-secondary")}
            onClick={() => toggle('exclude')}
          />

          <Button
            icon={record.favorited_at ? "pi pi-heart-fill" : "pi pi-heart"}
            tooltip={record.favorited_at ? "Remover Lead" : "Adicionar Lead no VendaMais"}
            tooltipOptions={{ position: 'top' }}
            className="p-button-rounded p-button-text p-button-danger"
            onClick={() => toggle('favorite')}
          />
        </span>

        <span
          style={{
            position: 'absolute',
            right: '1%',
            bottom: '1%',
            marginTop: '-15px',
          }}
        >
          {
            true ? // used to be enables only for admins
              <>
                <Button
                  icon="pi pi-book"
                  tooltip={"Observações"}
                  tooltipOptions={{ position: 'top' }}
                  className={
                    "p-button-rounded p-button-text " +
                    (record.has_note || record.obra_id === noteRec.obra_id.get() ? 'p-button-info' : 'p-button-secondary')
                  }
                  onClick={(e) => op.current.toggle(e)}
                />
                <OverlayPanel
                  ref={op}
                  onHide={() => saveNoteRec()}
                  onShow={() => loadNoteRec()}
                >
                  <InputTextarea
                    className="w-full"
                    cols={50}
                    rows={10}
                    placeholder='Digite Observações'
                    value={noteRec.note.get()}
                    // disabled={!userS.properties.get()?.is_admin}
                    onChange={(e) => {
                      let value = e.target.value
                      if (noteRec.note.get() === '') {
                        let date = new Date().toLocaleDateString('pt-BR')
                        value = `${date} - ${e.target.value}`
                      }
                      noteRec.note.set(value)
                    }
                    }
                  />

                </OverlayPanel>
              </>
              :
              null
          }
        </span>
      </div>
    </div>
  </>
}

const ContactModal = () => {
  const [visible, setVisible] = React.useState(false)
  const paneSelection = useVariable<'Telefone' | 'Email'>('Telefone')
  const $selectedContact = useStore(selectedContact)
  const phoneRecords = $selectedContact.filter(r => r.telephone)
  const emailRecords = $selectedContact.filter(r => r.email)



  React.useEffect(() => {
    if ($selectedContact.length > 0) {
      setVisible(true)
      if (phoneRecords.length === 0) paneSelection.set('Email')
    }
  }, [$selectedContact])

  const paneItem = (val) => {
    return (
      <div style={{ width: '170px' }}>
        <strong>{val}</strong>
      </div>
    );
  };


  const phoneBody = (row: Contact, column: ColumnBodyOptions) => {
    let phone = row.telephone.toString()

    // Remove DDI +55 do Brasil se presente
    if (phone.startsWith('55') && phone.length >= 12) {
      phone = phone.slice(2)
    }

    let ddd = phone.slice(0, 2)
    let prefix = phone.slice(2, 7)
    let suffix = phone.slice(7, 11)

    // Telefone fixo (10 dígitos)
    if (phone.length === 10) {
      prefix = phone.slice(2, 6)
      suffix = phone.slice(6, 10)
    }

    return <div style={{}}>
      <span>({ddd}) </span>
      <span>{prefix} </span>
      <span> - </span>
      <span>{suffix} </span>
    </div>
  }

  const emailBody = (row: Contact, column: ColumnBodyOptions) => {
    return <div style={{}}>
      <a href={`mailto:${row.email}`}>{row.email}</a>
    </div>
  }

  return <>
    <Dialog
      header={$selectedContact.length > 0 ? $selectedContact[0].name : ''}
      visible={visible}
      style={{ maxWidth: '700px', height: ' 500px' }}
      className="w-9"
      onHide={() => {
        selectedContact.set([])
        setVisible(false)
      }}
      closeOnEscape
      dismissableMask
      closable
    >

      <div className="card flex justify-content-center pb-2">
        <SelectButton
          value={paneSelection.get()}
          onChange={(e) => paneSelection.set(e.value)}
          options={['Telefone', 'Email']}
          itemTemplate={paneItem}
        />
      </div>

      <div className="m-0">
        {
          paneSelection.get() === 'Telefone' ?
            <DataTable
              value={phoneRecords}
              scrollable
              scrollHeight='300px'
              emptyMessage="Nada encontrado"
            >
              <Column body={phoneBody} header="Número de Telefone"></Column>
              <Column field="city" header="Cidade"></Column>
              <Column field="state" header="UF"></Column>
              <Column body={contactActionsBody} header="Ações"
                headerStyle={{ width: '9em' }}
                headerClassName="justify-content-center text-center"
                bodyStyle={{ width: '9em' }} />
            </DataTable>
            :
            <DataTable
              value={emailRecords}
              scrollable
              scrollHeight='300px'
              emptyMessage="Nada encontrado"
            >
              <Column body={emailBody} header="Email"></Column>
              <Column field="city" header="Cidade"></Column>
              <Column field="state" header="UF"></Column>
              <Column body={contactActionsBody} header="Ações"
                headerStyle={{ width: '9em' }}
                headerClassName="justify-content-center text-center"
                bodyStyle={{ width: '9em' }} />
            </DataTable>
        }
      </div>
    </Dialog>
  </>
}


export const contactActionsBody = (row: Contact, column: ColumnBodyOptions) => {
  const n = Math.trunc(Math.random() * 10000000000)
  const iconStyle = { fontSize: '1.25rem', color: 'white', background: 'navy', cursor: 'pointer' }
  const whatsapp_id = `whatsapp-${row.contact_id}-${n}`
  const phone_id = `phone-${row.contact_id}-${n}`
  const email_id = `email-${row.contact_id}-${n}`

  return <div className="text-left justify-left" style={{}}>
    {/* Disabled due to Weird bug where tooltip lingers */}
    {/* <Tooltip target={`#${whatsapp_id}`} position="top">Abrir WhatsApp</Tooltip>
      <Tooltip target={`#${phone_id}`} position="top">Ligue para o telefone</Tooltip>
      <Tooltip target={`#${email_id}`} position="right">Enviar Email</Tooltip> */}

    {
      row.telephone ?
        <>
          <a id={phone_id} href={telephoneURL(row.telephone)} target="_blank">
            <i
              className="pi pi-phone p-2 border-round-lg ml-2"
              style={iconStyle}
            />
          </a>

          <a id={whatsapp_id} href={whatsAppURL(row.telephone)} target="_blank">
            <i
              className="pi pi-whatsapp p-2 border-round-lg ml-2"
              style={iconStyle}
            />
          </a>
        </>
        :
        <>
          <a id={email_id} href={emailURL(row.email)} target="_blank">
            <i
              className="pi pi-envelope p-2 border-round-lg ml-2"
              style={iconStyle}
            />
          </a>
        </>
    }
  </div>
}

interface MessengerBatchDialogParams {
  records: ResultRecord[];
  visible: State<boolean, {}>
}


/*
Columns:
- Obra Shortened address
- Contact name
- Contact role
- Contacted status [New (no chat/message found), Existing (chat/message found)]
- Actions: exclude/remove, Template to use & Message preview
*/

type statusType = 'Duplicado' | 'Novo' | 'Ja Contactato'
type roleType = 'Proprietário' | 'Profissional'

interface MessageQueueRecord {
  obra_id: string
  obra: string
  contact_name: string
  contact_role: roleType
  status: statusType
  template_id: number
  text: string
}

const MessengerBatchDialog = (props: MessengerBatchDialogParams) => {
  const records = useVariable([] as MessageQueueRecord[])
  const visible = useHookstate(props.visible)
  const loading = useHookstate(false)
  const viewSelection = useHookstate('Novo Lote')

  ///////////////////////////  VARIABLES  ///////////////////////////
  ///////////////////////////  HOOKS  ///////////////////////////
  const toast = React.useRef<Toast>(null)

  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    if (visible.get()) {
      prepareData()
    }
    console.log(window.token)
  }, [visible.get()])
  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const prepareData = async () => {

    let table_data: MessageQueueRecord[] = []

    // find out if recipient has been contacted before
    let name_contacted_map = await getRecipientStatusMap()

    let name_map: ObjectAny = {}
    for (let record of props.records) {
      let contact_name = record.has_owner_phone ? record.owner : record.professional
      let suffix = `${record.bairro} - ${record.city}, ${record.state}`
      let template_id = (Math.floor(Math.random() * 10000000000) % 2) + 1

      let m_record = {
        obra_id: record.obra_id,
        contact_name: contact_name,
        obra: record.address?.replaceAll(', ' + suffix, '').trim(),
        contact_role: (record.has_owner_phone ? 'Proprietário' : 'Profissional') as roleType,
        status: (contact_name in name_contacted_map ? 'Ja Contactato' : contact_name in name_map ? 'Duplicado' : 'Novo') as statusType,
        template_id: template_id,
        text: '',
      }
      // generate text
      m_record.text = makeText(m_record)

      table_data.push(m_record)
      name_map[contact_name] = null
    }

    // Sort records by contact_name in ascending order
    table_data.sort((a, b) => a.contact_name.localeCompare(b.contact_name));

    records.set(table_data)
  }

  const makeText = (record: MessageQueueRecord) => {
    let text = ''
    let team_properties = userS.team.properties.get()
    let templates = team_properties?.templates
    let sender_name = templates.sender
    let store = team_properties?.name
    let contact_first_name = record.contact_name.split(' ')[0].toLowerCase()
    contact_first_name = contact_first_name.charAt(0).toUpperCase() + contact_first_name.slice(1)

    if (record.contact_role == 'Proprietário') {
      if (record.template_id === 1) text = templates.owner1
      if (record.template_id === 2) text = templates.owner2
      text = text?.replaceAll('[nome]', contact_first_name)
      text = text?.replaceAll('[remetente]', sender_name)
      text = text?.replaceAll('[loja]', store)
    }

    if (record.contact_role == 'Profissional') {
      if (record.template_id === 1) text = templates.professional1
      if (record.template_id === 2) text = templates.professional2
      text = text?.replaceAll('[nome]', contact_first_name)
      text = text?.replaceAll('[remetente]', sender_name)
      text = text?.replaceAll('[loja]', store)
    }

    return text
  }

  const makeRecipientNames = () => props.records.map(r => r.owner).concat(props.records.map(r => r.professional))

  // this is get whether a number has been contacted before
  const getRecipientStatusMap = async () => {
    // makeRecipientNames will be mapped to their number
    let data = { names: makeRecipientNames() }
    loading.set(true)
    let resp = await api().post(`${baseURL()}/messenger/existing`, data)
    loading.set(false)
    if (resp.error) throw new Error(resp.error)
    let json_data = await resp.json()

    let name_contacted_map = (json_data.name_contacted_map || {}) as ObjectBoolean

    return name_contacted_map
  }

  const submit = async () => {
    let new_records = records.get().filter((r) => !shouldExclude(r.status))
    loading.set(true)
    let resp = await api().post(`${baseURL()}/messenger/queue/submit`, { records: new_records })
    loading.set(false)
    if (resp.error) return doToast(toast, {
      severity: 'error',
      summary: 'Erro',
      detail: 'Não foi possível enviar o lote. Ocorreu um erro.\n' + resp.error,
    }, 4000)
    visible.set(false)
  }

  ///////////////////////////  JSX  ///////////////////////////

  const paneItem = (val) => {
    return (
      <div style={{ width: '170px' }}>
        <strong>{val}</strong>
      </div>
    );
  };

  const shouldExclude = (status: statusType) => {
    return status === 'Duplicado' || status === 'Ja Contactato'
  }

  const statusCode = (status: statusType) => {
    if (status === 'Duplicado') return 'duplicado'
    if (status === 'Ja Contactato') return 'ja-contactato'
    if (status === 'Novo') return 'novo'
    return ''
  }

  const roleCode = (status: roleType) => {
    if (status === 'Proprietário') return 'proprietario'
    if (status === 'Profissional') return 'profissional'
    return ''
  }


  const recContactName = (row: MessageQueueRecord, column: ColumnBodyOptions) => {
    return <div>
      <span style={{ textDecoration: shouldExclude(row.status) ? 'line-through' : null }}>{row.contact_name}</span>
    </div>
  }

  const recObraBody = (row: MessageQueueRecord, column: ColumnBodyOptions) => {
    const obra = row.obra.length > 40 ? row.obra.substring(1, 40) + '...' : row.obra
    return <div>
      <span style={{ textDecoration: shouldExclude(row.status) ? 'line-through' : null }}>{obra}</span>
    </div>
  }

  const recContactedRole = (row: MessageQueueRecord, column: ColumnBodyOptions) => {
    return <div style={{ fontSize: '10px', fontWeight: 700, letterSpacing: '.3px' }} className={'text-center border-round-md contact-role ' + roleCode(row.contact_role)}>
      {row.contact_role}
    </div>
  }

  const recContactedBody = (row: MessageQueueRecord, column: ColumnBodyOptions) => {
    return <div style={{ fontSize: '10px', fontWeight: 700, letterSpacing: '.3px' }} className={'text-center border-round-md contact-status ' + statusCode(row.status)}>
      {row.status}
    </div>
  }

  const recActionsBody = (row: MessageQueueRecord, column: ColumnBodyOptions) => {
    const removeRow = () => {
      let new_records = [] as MessageQueueRecord[]
      for (let record of records.get()) {
        if (row.obra_id !== record.obra_id)
          new_records.push(record)
      }
      records.set(new_records)
    }

    const changetext = () => {
      let new_records = [] as MessageQueueRecord[]
      for (let record of records.get()) {
        if (row.obra_id === record.obra_id) {
          record.template_id = record.template_id === 1 ? 2 : 1
          record.text = makeText(record)
        }
        new_records.push(record)
      }
      records.set(new_records)
    }

    return <div>
      <Button
        className="button mr-1"
        icon='pi pi-pencil'
        size="small"
        tooltip={row.text}
        tooltipOptions={{ position: 'top', style: { width: '400px' } }}
        disabled={shouldExclude(row.status)}
        onClick={() => {
          // regenerate text
          changetext()
        }}
      />

      <Button
        className="button"
        icon='pi pi-times'
        size="small"
        tooltip={"Excluír" + (shouldExclude(row.status) ? " (será automaticamente excluído)" : '')}
        severity="warning"
        tooltipOptions={{ position: 'top' }}
        onClick={() => removeRow()} />
    </div>
  }

  return <div>
    <Toast ref={toast} />
    <Dialog
      visible={visible.get()}
      header='Contactar Leads'
      footer={<div className="text-center">
        <Button
          label="Submeter"
          severity="success"
          onClick={() => submit()}
          tooltip="Envie o lote para a fila. Cada mensagem será processada em ordem a cada poucos minutos."
          tooltipOptions={{ position: 'left', style: { width: '300px' } }}
        />
        <Button label="Cancelar" severity="secondary" onClick={() => visible.set(false)} />
      </div>}
      onHide={() => {
        records.set([])
        visible.set(false)
      }}
      style={{ height: '600px' }}
      draggable={false}
    >
      {/* <div className="card w-full flex justify-content-center pb-2">
        <SelectButton
          value={viewSelection.get()}
          onChange={(e) => viewSelection.set(e.value)}
          options={['Novo Lote', 'Lotes Antigos']}
          itemTemplate={paneItem}
        />
      </div> */}

      <DataTable
        value={records.get()}
        scrollable
        scrollHeight='400px'
        width='90hw'
        emptyMessage="Nada encontrado"
        className="text-sm"
        loading={loading.get()}
      >
        <Column body={recContactName} header="Nome"></Column>
        <Column body={recContactedRole} header="Título"></Column>
        <Column body={recContactedBody} header="Status"></Column>
        <Column body={recObraBody} header="Obra Enderco"></Column>
        <Column body={recActionsBody} header="Ações"
          headerStyle={{ width: '11em' }}
          headerClassName="justify-content-center text-center"
          bodyStyle={{ width: '11em' }} />
      </DataTable>
    </Dialog>
  </div>

}