import * as React from 'react';
import { api, baseURL, makeURL, userTrackProps } from '../../store/api';
import { Card } from 'primereact/card';
import { Lead, user, Stage, type List, $vendaMaisPageRefresh, $vendaMaisStageHeight, userS } from '../../store/store';
import { useVariable, type ObjectAny } from '../../utils/interfaces';
import { useStore } from '@nanostores/react'
import { useHookstate, hookstate, type State } from '@hookstate/core';
import { InputText } from 'primereact/inputtext';
import { Button } from 'primereact/button';
import { Rating } from 'primereact/rating';
import { Dropdown } from 'primereact/dropdown';
import { StageDialogPanel, type StageDialogParams } from './StageDialogPanel';
import { LeadDialogPanel, deleteLead, type LeadDialogParams } from './LeadDialogPanel';
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import { doToast, filterAndMatched, jsonClone, parseFilterString, uuidv4 } from '../../utils/methods';
import { SelectButton } from 'primereact/selectbutton';
import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { Tooltip } from 'primereact/tooltip';
import { Toast } from 'primereact/toast';
import type { ResultRecord } from '../obras-plus/ObraPlusPage';
import { Dialog } from 'primereact/dialog';
import { makeCity, type City, allCities } from '../../store/cities';
import { User } from '../../store/store';
        

interface Props {}

const $currentLead = hookstate({} as Lead)
export const doRefreshVendaMaisPage = () => $vendaMaisPageRefresh.set($vendaMaisPageRefresh.get() + 1)

export default function VendaMaisPage(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const viewModeOptions = [
      {icon: 'pi pi-th-large', value: 'kanban'},
      {icon: 'pi pi-list', value: 'table'},
  ];
  ///////////////////////////  HOOKS  ///////////////////////////
  const vendaMaisStageHeight = useStore($vendaMaisStageHeight)
  const refreshVendaMaisPage = useStore($vendaMaisPageRefresh)
  const currentLead = useHookstate($currentLead)
  const search = useHookstate<SearchProps>({show: false, query: '', cityOptions: [], records: [], selectedLead: new Lead()} as SearchProps)
  const leadDialog = useHookstate<LeadDialogParams>({lead: new Lead({}), stages:[]})
  const [selectedMember, setSelectedMember] = React.useState<User>(new User({ id: 'all', email: 'Todos' }));
  const [members, setMembers] = React.useState<User[]>([]);
  const leads = useVariable<Lead[]>([]);
  const stages = useHookstate<Stage[]>([]);
  const stageOptions = useVariable<Stage[]>([]);
  const selectedLeads = useVariable<Lead[]>([]);
  const warn = useVariable(false);
  const filter = useHookstate('')
  const $user = useStore(user)
  const selectedViewMode = useHookstate<'kanban' | 'table'>($user?.state?.venda_mais_mode || 'kanban')
  const toast = React.useRef<Toast>(null)

  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    window.rudderAnalytics?.page({
      userId: user.get().id,
      category: "venda-mais",
      name: window.document.title,
    })
  }, []);

  React.useEffect(() => {
    let cities = userS?.team?.cities?.get()
    // if(!(cities?.length && cities[0])) warn.set(true)
  }, [userS.get()]);

  React.useEffect(() => {
    doRefresh()
  }, [selectedMember]);


  React.useEffect(() => {
    doRefresh()
  }, [refreshVendaMaisPage])

  React.useEffect(() => {
    if(!search.show.get()) {
      if(search.makeNew.get()) createNewLead()
      search.makeNew.set(false)
    } else {
      doRefresh()
    }
  }, [search.show.get()])

  React.useEffect(() => {
    if(!search.selectedLead.get()?.list_lead_id) return

    // sync single lead
    syncSingleListLead(search.selectedLead.get().list_lead_id)
      .then(
        (lead) => currentLead.set(lead)
      )
  }, [search.selectedLead.get()])

  React.useEffect(() => {
    if(!leadDialog.show.get()) {
      let lead = getLeadByID(leadDialog.lead.list_lead_id.get())
      let orig_stage_id = lead.stage_id.get()
      lead.set(
        l => {
          l = new Lead(jsonClone(leadDialog.lead.get()))
          return l
        }
      )
      if(leadDialog.deleted?.get()) setLeadStage(lead.list_lead_id.get(), 'deleted')
      else setLeadStage(lead.list_lead_id.get(), lead.stage_id.get())
      syncLeads()
      $currentLead.set(new Lead())

      // having issues with not refresh, so pull data if stage_id has changes
      if(orig_stage_id !== lead.stage_id.get()) doRefresh()
      else syncSingleListLead(leadDialog.lead.list_lead_id.get())
    } else {
      window.rudderAnalytics?.track(
        'venda-mais-lead-dialog',
        {
          user: userTrackProps(),
          lead: {
            list_lead_id: leadDialog.lead.list_lead_id.get(),
            lead_id: leadDialog.lead.lead_id.get(),
            obra_id: leadDialog.lead.obra_id.get(),
          }
        }
      )
    }
  }, [leadDialog.show.get()]);

  React.useEffect(() => {
    if(!currentLead?.get()?.list_lead_id) return

    leadDialog.set(
      val => {
        val.lead = new Lead(jsonClone(currentLead.get()))
        val.stages = jsonClone(stages.get())
        val.deleted = false
        val.show = true
        return val
      }
    )
  }, [currentLead.get()]);

  React.useEffect(() => {
    if(selectedViewMode.get() === 'table' && (leads.get().length || 0) === 0) syncLeads()
  }, [selectedViewMode.get(), stages.get()]);

  React.useEffect(() => {
    let height = window.document.body.scrollHeight - 380
    if(height !== vendaMaisStageHeight) $vendaMaisStageHeight.set(height)
  }, []);


  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const doRefresh = async() => {
    getMembers()
    await getStages()
  }

  const doFilter = async() => {
    getMembers()
    await getStages()
  }

  const filteredLeads = () => {
    let results : Lead[] = []

    // Text filter
    if(filter.get()) {
      let filters = parseFilterString(filter.get())
      results = leads.get().filter(lead => filterAndMatched(Object.values(lead), filters))      
    } else {
      results = leads.get()
    }
    
    // User filter
    if (selectedMember?.id && selectedMember.id !== 'all') {
      results = results.filter(lead => lead.owner_email === selectedMember.email)
    }
    
    return results
  }

  const getMembers = async () => {
    // don't get members if not manager
    if(!userS?.is_manager.get()) {
      let member = new User(userS.get())
      setMembers([member]);
      return
    }

    // Add a "Todos" option as the first item
    const allOption = new User({ id: 'all', email: 'Todos' });
    
    // Fetch team members
    const resp = await api().get(makeURL('/team/members'), { team_id: userS?.team?.id?.get() });
    if (resp.error) {
      setMembers([allOption]);
      return;
    }
    
    // Transform team members into list format
    const memberLists = (await resp.records()).map(member => (new User(member)));
    
    // Set the lists with "Todos" as the first option
    setMembers([allOption, ...memberLists]);
  }

  const getStageByID = (id: string) => {
    for (let i = 0; i < stages.length; i++) {
      const stage = stages[i];
      if(stage.id.get() === id) return stage
    }
    return hookstate({} as Stage)
  }

  const getLeadByID = (list_lead_id: string) => {
    for (let i = 0; i < stages.length; i++) {
      const stage = stages[i];
      for (let j = 0; j < stage.leads.length; j++) {
        const lead = stage.leads[j];
        if(lead.list_lead_id.get() === list_lead_id) return lead
      }
    }
    return hookstate(new Lead())
  }

  const setDefaultSelectedList = (list_id: string) => {
    $user.state.selected_list = list_id
    $user.saveState()
    user.set($user)
  }

  const getStages = async () => {
    // Instead of filtering by list_id, we get all stages
    const resp = await api().collection('list_stage').getFullList({
      sort: 'order',
    })
    if(resp.error) return
    let new_stages = (await resp.records() as any[]).map(s => new Stage(s))
    for(let stage of new_stages) {
      stage.leads = await getStageLeads(stage.id)
      stage.leads = sortLeads(stage.leads, stage.properties.order_by)
    }
    stageOptions.set(new_stages)
    stages.set(new_stages)
  }

  const syncLeads = async () => {
    let new_leads : Lead[] = []
    for(let stage of stages.get()) {
      if(stage.leads.length === 0) continue
      new_leads = new_leads.concat((jsonClone(stage.leads) as any[]).map(l => new Lead(l)))
    }
    if(new_leads.length) leads.set(new_leads)
  }

  const getStageLeads = async (stage_id: string) => {
    let filter_user_id = selectedMember.id === 'all' ? '' : selectedMember.id
    let resp = await api().get(makeURL('/query/crm/leads'), {stage_id, filter_user_id })
    if(resp.error) return
    let new_leads = (await resp.records() as any[]).map(v => new Lead(v))
    return new_leads
  }

  const getSingleListLead = async (list_lead_id: string) => {
    let resp = await api().get(makeURL('/query/crm/leads'), { list_lead_id })
    if(resp.error) return
    return new Lead(await resp.record())
  }

  const syncSingleListLead = async (list_lead_id: string) => {
    let lead = await getSingleListLead(list_lead_id)
    let leadS = getLeadByID(list_lead_id)
    if(!leadS.lead_id.get()) {
      // is not in any stage, add
      setLeadStage(lead.list_lead_id, lead.stage_id)
      leadS = getLeadByID(list_lead_id)
    }
    leadS.set(lead)
    return lead
  }

  const setLeadStage = (src_list_lead_id: string, tgt_stage_id: string, tgt_index = 0) => {
    let lead = getLeadByID(src_list_lead_id)
    let src_stage = getStageByID(lead.stage_id.get())
    let tgt_stage = getStageByID(tgt_stage_id)

    let removed : Lead

    if(!src_stage.leads.get() || src_stage.leads.length === 0) return

    src_stage.leads.set(
      leads => {
        let src_index = leads.map(l => l.list_lead_id).indexOf(lead.list_lead_id.get()) as number
        if(src_index !== -1)
          [removed] = leads.splice(src_index, 1)
        return sortLeads(leads, src_stage.properties.order_by.get())
      }
    )

    if(tgt_stage_id === 'deleted') return

    tgt_stage.leads.set(
      leads => {
        if(removed) {
          removed.stage_id = tgt_stage_id
          leads.splice(tgt_index, 0, removed)
        }
        return sortLeads(leads, tgt_stage.properties.order_by.get())
      }
    )

  }

  const handleDragEnd = async (event) => {

    let draggable_id = event.draggableId  as string
    let src_stage_id = draggable_id.split('--')[0]
    let src_list_lead_id = draggable_id.split('--')[1]
    let tgt_stage_id = event.destination.droppableId
    let tgt_index =  event.destination.index

    if(tgt_stage_id && src_stage_id && tgt_stage_id !== src_stage_id) {
      // if stages are different, need to remove and add
      setLeadStage(src_list_lead_id, tgt_stage_id, tgt_index)

      await api().collection('list_lead').update(src_list_lead_id, {stage_id: tgt_stage_id})
      
      syncLeads()
    }
  }

  const createNewLead = async () => {
      // search.show.set(true)
      let id = `oppo_${uuidv4().replaceAll('-', '')}`
      let lead = new Lead({obra_id: id})
      let err = undefined

      try {
        let data = { obra_id: lead.obra_id, toggle_col: 'favorited_at' }
        let resp = await api().patch(makeURL('/patch/lead-toggle'), data)
        if(resp.error) throw new Error(resp.error)

        let rec = await resp.record()
        lead.lead_id = rec.lead_id
        lead.list_lead_id = rec.list_lead_id

        syncSingleListLead(lead.list_lead_id) // get latest list
        
        $currentLead.set(lead)
      } catch (error) {
        err = error
        return doToast(toast, {
          severity: 'error',
          summary: 'Algo deu errado'
        })
      } finally {
        window.rudderAnalytics?.track(
          'venda-mais-add-lead',
          { user: userTrackProps(), error: err }
        )
      }
    }

  ///////////////////////////  JSX  ///////////////////////////


  const viewModeIconTemplate = (option) => {
      return <i className={option.icon} style={{ fontSize: '1.25rem' }}></i>;
  }

  const obraBody = (lead: Lead) => {  
      const divStyle = useHookstate({textDecoration: 'none'})

      return <> 
        { makeLeadTooltip(lead) }
        <div 
          id={lead.obra_id}
          className='cursor-pointer'
          onClick={() => $currentLead.set(lead)}
          onPointerEnter={() => divStyle.textDecoration.set('underline')}
          onPointerLeave={() => divStyle.textDecoration.set('none')}
          style={jsonClone(divStyle.get())}
        >
          {lead.title}
        </div>
      </>
  }

  const stageBody = (lead_: Lead) => {

      return <div>
        <Dropdown
          className="w-full step-dropdown"
          style={{padding: 0, paddingLeft: 3, paddingRight: 3}}
          value={lead_.stage_id}
          onChange={async (e) => {
            let lead = getLeadByID(lead_.list_lead_id)
            lead.stage_id.set(e.value)

            let resp = await api().collection('list_lead')
              .update(lead.list_lead_id.get(), {stage_id: lead.stage_id.get()})

            if(resp.error) doToast(toast, {
                severity: 'error',
                summary: 'Não foi possível salvar'
              })

            syncLeads()
          }}
          options={stageOptions.get()}
          optionLabel="name" 
          optionValue="id"
        />
      </div>
  }

  const ratingBody = (lead_: Lead) => {        
      const divStyle = useHookstate({textDecoration: 'none'})

      return <div>
        <Rating
          className="ml-3"
          value={lead_.lead_properties.rating}
          cancel={false}
          onChange={async (e) => {
            let lead = getLeadByID(lead_.list_lead_id)
            lead.lead_properties.rating.set(e.value)

            let resp = await api().collection('lead')
              .update(lead.lead_id.get(), {properties: lead.lead_properties.get()})

            if(resp.error)
              doToast(toast, {
                severity: 'error',
                summary: 'Não foi possível salvar'
              })
              
            lead.lead_properties.set((await resp.record()).properties)
            
            syncLeads()
          }}
        />
      </div>
  }
  
  return (
  <div>
    <Toast ref={toast} />
    <SearchDialog search={search}/>

    {
      warn.get()?
      <div className="text-center mb-4">
        <div style={{color: 'blue'}}>
          Entre em contato conosco para configurar um plano: <a href="mailto:contato@suaobra.com.br">contato@suaobra.com.br</a>
        </div>
      </div>
      :
      null
    }

    <div className="formgrid grid px-2">
      <div className="field md:col-3 col-12">
        <label htmlFor="city-dropdown">Filtrar por Responsável</label>
        <Dropdown
          id='list-dropdown'
          value={selectedMember}
          onChange={(e) => {
            console.log(e.value)
            setSelectedMember(e.value)
          }}
          options={members}
          optionLabel="email" 
          // optionValue="id"
          placeholder="Selecione um Membro"
          className="w-full"
        />
      </div>

      <div className="field md:col-6 col-12">
        <label htmlFor="city-dropdown">Filtro</label>
          <div className="p-inputgroup">
              <InputText
                id='search-filter'
                placeholder="Digite o que procura..."
                aria-describedby="filter-help"
                className="w-full"
                value={filter.get()}
                onChange={(e) => {
                  filter.set(e.target.value)
                  if(!e.target.value) doFilter() // auto-refreshed when cleared
                }}
                onKeyDown={(e) => {
                  if(e.key === 'Escape') { filter.set(''); doFilter() } 
                  if(e.key === 'Enter') { doFilter() } 
                }}
              />
              <Button
                icon="pi pi-times"
                className="p-button-warning"
                tooltip="Limpar"
                tooltipOptions={{position: 'top'}}
                onClick={() => { filter.set(''); doFilter() }}
              />
              <Button
                icon="pi pi-search"
                className="p-button-primary"
                tooltip="Pesquisar"
                tooltipOptions={{position: 'top'}}
                onClick={() => {
                  doFilter()
                }}
              />
          </div>
      </div>

      <div className="md:col-3 col-12">
        <label htmlFor="city-dropdown">Ações</label>
        <div className="w-full flex flex-wrap mt-2">
            <Button
              tooltip='Adicionar Novo Lead'
              tooltipOptions={{position: 'bottom'}}
              className='mr-2 success'
              icon='pi pi-plus'
              outlined
              onClick={async () => {
                search.show.set(true)
              }}
            />

            <SelectButton 
              tooltip='Modo: "Funil de Vendas" o "Planilha"'
              className='mr-2'
              tooltipOptions={{position: 'bottom'}}
              value={selectedViewMode.get()}
              onChange={async (e) => {
                selectedViewMode.set(e.value)
                window.rudderAnalytics?.track(
                  'venda-mais-switch-mode',
                  { user: userTrackProps(), mode: e.value }
                )
              }}
              itemTemplate={viewModeIconTemplate}
              optionLabel="value"
              options={viewModeOptions}
            />

            <Button
              tooltip='Actualizar'
              tooltipOptions={{position: 'bottom'}}
              className='mr-2'
              icon='pi pi-refresh'
              outlined
              onClick={() => {
                doRefresh()
              }}
            />

            {
              selectedViewMode.get() === 'table' && selectedLeads?.get()?.length > 0 ?
              <Button
                tooltip={`Remover ${selectedLeads?.get()?.length} Leads` }
                tooltipOptions={{position: 'top'}}
                className='mr-2'
                icon='pi pi-trash'
                outlined
                severity="danger"
                onClick={async () => {
                  let promises : Promise<{success: boolean, error: string}>[] = []
                  for(let lead of selectedLeads.get()) {
                    let promise = deleteLead(lead.list_lead_id, lead.lead_id)
                    promises.push(promise)
                  }

                  for(let promise of promises) {
                    let resp = await promise
                    if(resp.error) doToast(toast, {
                        severity: 'error',
                        summary: 'Não foi possível remover',
                        detail: resp.error,
                      })
                  }

                  await doRefresh()
                  syncLeads()

                  window.rudderAnalytics?.track(
                    'venda-mais-bulk-delete',
                    { user: userTrackProps(), leads: selectedLeads.get() }
                  )
                }}
              />

              :

              null
            }

            {/* <Button
              tooltip='Exportar Planilha'
              tooltipOptions={{position: 'bottom'}}
              icon='pi pi-file-excel'
              // severity="success"
              outlined
              size='small'
            /> */}

        </div>
      </div>

    </div>
    <div className="flex flex-wrap">
      {
        selectedViewMode.get() === 'kanban' ?
          <DragDropContext onDragEnd={handleDragEnd}>
            {
              stages.get()
                .filter(
                  s => selectedMember.id === 'all' || true // Show all stages regardless of filter
                )
                .map(
                s => <div
                        key={s.id}
                        className="col-12 md:col-6 lg:col-3"
                        style={{height: $vendaMaisStageHeight.get()}}
                      >
                      <StageElem 
                        stage={getStageByID(s.id)} 
                        filter={filter} 
                        userFilter={selectedMember.id} 
                      />
                </div>
              )
            }
          </DragDropContext>
        :
          <DataTable
            id='obras-table'
            value={filteredLeads()}
            scrollable
            showGridlines
            scrollHeight={`${$vendaMaisStageHeight.get()}px`}
            style={{ width: '100%' }}
            sortMode="multiple"
            emptyMessage="Nenhum resultado encontrado"
            removableSort
            selectionMode={'checkbox'}
            selection={selectedLeads.get()}
            onSelectionChange={(e) => { selectedLeads.set(e.value) }}
          >
            <Column selectionMode="multiple" headerStyle={{ width: '3rem' }}></Column>
            <Column header="Obra" body={obraBody}/>
            <Column sortable sortField='owner_str' field="owner_str" header="Responsável"/>
            <Column sortable sortField='size' field="size_str" header="Tamanho"/>
            <Column sortable sortField='lead_properties.valor' field="valor" header="Valor"/>
            <Column sortable sortField='stage_id' body={stageBody} header="Etapa de Vendas" className='table-steps'/>
            <Column sortable sortField='lead_properties.rating' header="Qualificação" body={ratingBody}/>
          </DataTable>
      }
      <LeadDialogPanel state={leadDialog} />
    </div>
  </div>
  )
}

interface SearchProps {
  show: boolean
  city: City
  cityOptions: City[]
  query: string
  loading: boolean
  records: ResultRecord[]
  selectedLead: Lead
  makeNew?: boolean
}


function SearchDialog(props: { search: State<SearchProps> }) {
  const state = useHookstate(props.search)
  const errorMsg = useVariable('')
  const searched = useVariable(false)

  React.useEffect(() => {
    
    let cities = jsonClone<string[]>(userS.team?.cities?.get() || [])
    if(cities.length && cities[0]) {
      if(cities.length == 1 && cities[0].endsWith('-*')) {
        let prefix = cities[0].replaceAll('*', '')
        state.cityOptions.set(allCities.filter(c => c.id.startsWith(prefix)).map(c => c.id).sort().map(id => makeCity(id)))
      } else
        state.cityOptions.set(cities.sort().filter(id => id).map(id => makeCity(id)))
    } else {
      state.show.set(false)
      return
    }

    if(!state.city.get()?.id)
      state.city.set(makeCity(localStorage.getItem('obrasPlusCity')))

    searched.set(false)
  }, [state.show.get()])

  const reset = (makeNew=false) => {
    errorMsg.set('')
    searched.set(false)
    state.set( 
      s => {
        s.show = false
        s.makeNew = makeNew
        s.query = ''
        s.records = []
        return s
      })
  }

  const doSearch = async () => {
    let payload = {
      state: state.city.get()?.state,
      city: state.city.get()?.city,
      query: state.query.get(),
    }
    let records : ResultRecord[] = []
    let error = undefined
    try {
      state.loading.set(true)
      let resp = await api().get(`${baseURL()}/query/crm/search`, payload)
      if(resp.error) throw new Error(resp.error)
      records = (await resp.response.json()) as ResultRecord[]
      if(records.length === 0) errorMsg.set('Não encontrou nenhuma obra...')
      else errorMsg.set('')
    } catch (err) {
      error = err
      console.log(err)
      records = []
    } finally {
      searched.set(true)
      state.set(
        s => {
          s.loading = false
          s.records = records
          return s
        }
      )

      window.rudderAnalytics?.track(
        'venda-mais-crm-search',
        { user: userTrackProps(), payload, error }
      )
    }
  }
  
  return <>
    <Dialog
      header='Adicionar Nova Obra'
      visible={state.show.get()}
      onHide={() => { reset() }}
      onShow={() => {
        setTimeout(() => {
          window.document.getElementById('obra-search-filter')?.focus()
        }, 30);
      }}
      closable
      closeOnEscape
      dismissableMask
      modal={true}
      // showHeader={false}
      style={{width: '700px'}}
    >
      <div className='grid'>
        <div className="col-4">
            <Dropdown
              id='obra-city-dropdown'
              value={state.city.get()}
              onChange={(e) => {
                localStorage.setItem('obrasPlusCity', e.value.id)
                state.city.set(e.value)
              }}
              options={jsonClone<City[]>(state.cityOptions.get() || [])}
              filter
              optionLabel="city" 
              // optionValue="id"
              placeholder="Cidade"
              emptyMessage="Nenhuma cidade encontrada"
              className="w-full"
            />
        </div>

        <div className="col-8">
          <div className="p-inputgroup">
            <InputText
                id='obra-search-filter'
                placeholder="Digite enderco, bairro, numero..."
                className="w-full"
                value={state.query.get()}
                onChange={(e) => {
                  state.query.set(e.target.value)
                }}
                onKeyDown={(e) => {
                  if(e.key === 'Escape') { 
                    state.query.set(''); 
                    state.records.set([])
                  } 
                  if(e.key === 'Enter') { doSearch() } 
                }}
            />

            <Button
              icon="pi pi-times"
              className="p-button-warning"
              tooltip="Limpar"
              tooltipOptions={{position: 'top'}}
              onClick={() => { state.query.set(''); state.records.set([]) }}
            />

            <Button
              icon="pi pi-search"
              className="p-button-primary"
              tooltip="Pesquisar"
              tooltipOptions={{position: 'top'}}
              onClick={() => {
                doSearch()
              }}
            />
          </div>
        </div>
        <div className="col-12">
          {
            state.records.get().map(
              (rec, i) => {
                let lead = new Lead(rec)
                return <React.Fragment key={`search-record-${i}`}>
                  { makeLeadTooltip(lead, 'right', false) }
                  <p
                    id={lead.obra_id}
                    className='cursor-pointer hover:surface-100 border-round-xs'
                    style={{
                      color: lead.favorited_at ? 'blue' : ''
                    }}
                    onClick={async () => {
                      if(lead.favorited_at) {
                        // open the lead since it's already
                        state.show.set(false)
                        return state.selectedLead.set(lead)
                      }
                      
                      let data = { obra_id: lead.obra_id, toggle_col: 'favorited_at' }
                      let resp = await api().patch(makeURL('/patch/lead-toggle'), data)
                      if(resp.error) return

                      state.records.set(
                        recs => {
                          for (let i = 0; i < recs.length; i++) {
                            const rec = recs[i]
                            if(rec.obra_id === lead.obra_id) {
                              recs[i].favorited_at = '1' // set blue
                            }
                          }
                          return recs
                        }
                      )

                      let rec = await resp.record()
                      lead.lead_id = rec.lead_id
                      lead.list_lead_id = rec.list_lead_id

                      state.show.set(false)
                      state.selectedLead.set(lead)
                    }}
                  >
                    {rec.address}
                  </p>
                </React.Fragment>
              }
            )
          }
        </div>
        <div style={{color: 'red'}} className='pb-2' >{errorMsg.get()}</div>
      </div>

      {
          searched.get() && (errorMsg.get() || state.query.get()?.length > 4) ?
          <div className="flex justify-content-center">
            <Button
              label="Adicionar Obra Personalizada"
              icon="pi pi-plus"
              severity="success"
              size="small"
              tooltip='Se você não encontrou o endereço, você pode criar uma Obra Personalizada'
              tooltipOptions={{position: 'bottom'}}
              onClick={(e) => { reset(true) }}
            />
          </div>
          :
          null
      }
    </Dialog>
  </>
}

interface StageProps {
  stage: State<Stage>
  filter: State<string>
  userFilter: string
}

function StageElem(props: StageProps) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  ///////////////////////////  HOOKS  ///////////////////////////
  const stage = useHookstate(props.stage)
  const filter = useHookstate(props.filter)
  const userFilter = useHookstate(props.userFilter)
  const leads = useHookstate(stage.leads)
  const stageDialog = useHookstate<StageDialogParams>({stage: jsonClone(stage.get())})

  ///////////////////////////  EFFECTS  ///////////////////////////
  
  React.useEffect(() => {
    if(!stageDialog.show.get()) doRefreshVendaMaisPage()
  }, [stageDialog.show.get()]);

  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const filterMatched = (lead: ObjectAny) => {
    // First check text filter
    if(filter.get()) {
      let filters = parseFilterString(filter.get())
      if (!filterAndMatched(Object.values(lead), filters)) {
        return false;
      }
    }
    
    // Then check user filter
    if (userFilter.get() !== 'all') {
      if (lead.owner_email !== userFilter.get()) {
        return false;
      }
    }
    
    return true;
  }

  const getListStyle = isDraggingOver => ({
    height: $vendaMaisStageHeight.get(),
    background: isDraggingOver ? "lightblue" : "lightgrey",
  });

  const toggleSort = async () => {
    let properties = jsonClone(stage.properties.get()) as ObjectAny
    if(properties.order_by === 'list_lead_updated-desc') properties.order_by = 'list_lead_updated-asc'
    else if(properties.order_by === 'list_lead_updated-asc') properties.order_by = 'list_lead_updated-desc'
    let record = await (
      await api().collection('list_stage').update(stage.id.get(), { properties })
    ).record()
    
    stage.set(
      s => {
        s.properties = (new Stage(record)).properties
        s.leads = sortLeads(jsonClone(leads.get()).map(l => new Lead(l)), properties.order_by)
        return s
      }
    )
  }
  
  ///////////////////////////  JSX  ///////////////////////////

  return <Droppable key={stage.id.get()} droppableId={stage.id.get()}>
    {(provided, snapshot) => (
    <div
      className="p-3 pt-1 border-round-lg"
      ref={provided.innerRef}
      {...provided.draggableProps}
      {...provided.dragHandleProps}
      style={getListStyle(snapshot.isDraggingOver)}
    >
        
      <div className="flex justify-content-between flex-wrap">
        <h3 className="mb-0">{ stage.name.get() }</h3>
        <h3
          className="mb-0 cursor-pointer"
        >
          <i className="pi pi-sort-alt mr-2" onClick={() => toggleSort()}></i>
          <i className="pi pi-cog" onClick={() => stageDialog.show.set(true)}></i>
        </h3>
      </div>

      <p className="mt-1">{ leads.get()?.length || 0 } obras</p>

      {/* <DataScroller
        id='stage-scroller'
        value={leads.get()}
        itemTemplate={(data) => <OneLead lead={data} index={1}/>}
        rows={3}
        inline
        scrollHeight={`${$vendaMaisStageHeight.get() - 100}px`}
        // header="Scroll Down to Load More"
        style={{backgroundColor: '#E0E0E0'}}
      /> */}
        
      <div
        className="overflow-y-scroll stage-leads"
        style={{
          height: $vendaMaisStageHeight.get() - 100,
          scrollbarWidth: 'none'
        }}
      >
        {
          leads.filter(l => filterMatched(jsonClone(l.get()))).map((lead, index) => {
            return <LeadElem key={index} lead={lead} index={index} tooltipPosition={stage.order.get() >= 3 ? 'left' : 'right'}/>
          })
        }
      </div>
      
      {provided.placeholder}
      <StageDialogPanel state={stageDialog} />
    </div>

    )}
  </Droppable>
}

interface CardProps {
  lead: State<Lead>
  index: number
  tooltipPosition: 'left' | 'right'
}

function LeadElem(props: CardProps) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const lead = useHookstate(props.lead)
  const titleStyle = useHookstate({textDecoration: 'none'})
  const index = props.index
  const tooltipPosition = props.tooltipPosition
  ///////////////////////////  HOOKS  ///////////////////////////
  ///////////////////////////  EFFECTS  ///////////////////////////
  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const getItemStyle = (isDragging, draggableStyle) => ({
    // some basic styles to make the items look a bit nicer
    userSelect: "none",

    // change background colour if dragging
    background: isDragging ? "lightgreen" : "grey",

    // styles we need to apply on draggables
    ...draggableStyle
  });

  ///////////////////////////  JSX  ///////////////////////////

  return <Draggable
    key={lead.list_lead_id.get()}
    draggableId={`${lead.stage_id.get()}--${lead.list_lead_id.get()}`}
    index={index}
  >
    {(provided, snapshot) => (
      <div
        className="stage-lead mb-2 border-round-lg cursor-pointer"
        ref={provided.innerRef}
        {...provided.draggableProps}
        {...provided.dragHandleProps}
        style={getItemStyle(
          snapshot.isDragging,
          provided.draggableProps.style
        )}
      >
        { makeLeadTooltip(new Lead(jsonClone(lead.get())), tooltipPosition) }
        <Card
          id={lead.obra_id.get()}
          className='lead-card'
          style={{paddingTop: 0, paddingBottom: 0, minHeight: 75, zIndex: 1}}
        >
          <h5
            className='my-0'
            onClick={() => $currentLead.set(new Lead(jsonClone(lead.get())))}
            onPointerEnter={() => titleStyle.textDecoration.set('underline')}
            onPointerLeave={() => titleStyle.textDecoration.set('none')}
            style={jsonClone(titleStyle.get())}
          >
            {lead.title.get()}
          </h5>
          <div className='flex justify-content-between align-content-center flex-wrap pt-2'>

            <div>
              <Rating
                value={lead.lead_properties.rating.get()}
                cancel={false}
                onChange={async (e) => {
                  lead.lead_properties.rating.set(e.value)
                  let resp = await api().collection('lead')
                    .update(lead.lead_id.get(), {properties: lead.lead_properties.get()})

                  lead.lead_properties.set((await resp.record()).properties)
                }}
              />
            </div>

            <div>
              {
                lead.lead_properties.has_alert.get() ?
                <i
                  className="pi pi-clock mr-2"
                  onClick={() => {}}
                  style={{color: lead.lead_properties.alert_reached.get() ? 'red' : 'blue'}}
                />
                :
                null
              }
              <span>{lead.get().size} m²</span> 
            </div>
          </div>
        </Card>
      </div>
    )}
  </Draggable>
}


const sortLeads = (leads: Lead[], order_by: string) => {
  let [key, sort] = order_by.split('-')
  return leads.sort(function(a, b) { 
    if(sort === 'asc') return a[key] - b[key]
    if(sort === 'desc') return b[key] - a[key]
  })
}

// Helper function to filter stages by user ID
const filterStagesByUserID = (stage: State<Stage>, userId: string) => {
  // For Kanban view, we need to show all stages but filter the leads within each stage
  return true; // Always show all stages, we'll filter the leads instead
}

// Helper function to filter leads by user ID
const filterLeadsByUserID = (lead: Lead, userId: string) => {
  // If viewing all leads, don't filter
  if (userId === 'all') return true;
  
  // Filter by the lead's owner (user who owns this lead)
  return lead.owner_email === userId;
}

export const makeLeadTooltip = (lead: Lead, position : "top" | "bottom" | "left" | "right" | "mouse" = 'top', show_valor=true) => {
  return <Tooltip
      target={`#${lead.obra_id}`}
      style={{fontSize: '15px', minWidth: '230px', maxWidth: '700px', maxHeight: '400px', overflowY: 'hidden'}}
      position={position}
    >
      <span><strong>Proprietário:</strong> {lead.owner}</span>
      <br/><span><strong>Profissional:</strong> {lead.professional}</span>
      <br/>
      <br/><span><strong>Enderco:</strong> {lead.address_short}</span>
      <br/><span><strong>Bairro:</strong> {lead.bairro}</span>
      <br/><span><strong>Cidade:</strong> {lead.cidade}</span>
      {
        show_valor ?
        <>
          <br/>
          <br/><span><strong>Valor:</strong> {lead.valor}</span>
        </>
        :
        null
      }
      <br/>
      <br/><span><strong>Responsável:</strong> {lead.owner_str}</span>
      
    </Tooltip>
}