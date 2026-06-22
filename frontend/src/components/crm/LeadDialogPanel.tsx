import { type State, useHookstate } from "@hookstate/core";
import { Dialog } from "primereact/dialog";
import * as React from "react";
import { Contact, Lead, LeadProperties, ObraProperties, Stage, userS } from "../../store/store";
import { TabView, TabPanel } from 'primereact/tabview';
import { Editor } from 'primereact/editor';
import { Rating } from "primereact/rating";
import { api, makeURL } from "../../store/api";
import { InputNumber } from "primereact/inputnumber";
import { DataTable } from "primereact/datatable";
import { Column, type ColumnBodyOptions } from "primereact/column";
import { telephoneURL, whatsAppURL, emailURL, isPhoneOrEmailContact, isPhoneContact, isEmailContact } from "../obras-plus/ObraPlusPage";
import { Tooltip } from "primereact/tooltip";
import { useVariable, type ObjectBoolean, type ObjectAny } from "../../utils/interfaces";
import { Dropdown } from "primereact/dropdown";
import { jsonClone } from "../../utils/methods";
import { Button } from "primereact/button";
import PrimeForm, { EditableText, type PrimeFields } from "../../utils/PrimeForm";
import { allCities } from "../../store/cities";
import { OverlayPanel } from "primereact/overlaypanel";

export interface LeadDialogParams {
  show?: boolean;
  lead: Lead;
  stages: Stage[];
  deleted?: boolean
}

export function LeadDialogPanel(props: { state: State<LeadDialogParams> }) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const keyStyle = {width: 100}

  const email_fields : PrimeFields = {
    name: { label: 'Nome', size: 12, type: 'string'},
    email: { label: 'Email', size: 12, type: 'string'},
    // city: { label: 'Cidade', type: 'dropdown', options: {options: allCities.map(c => c.id), filter: true}},
  }
  
  const telephone_fields : PrimeFields = {
    name: { label: 'Nome', size: 12, type: 'string'},
    telephone: { label: 'Telefone', type: 'mask', options: {mask: '(99) 99999 - 9999'}},
    city: { label: 'Cidade', type: 'dropdown', options: {options: allCities.map(c => c.id), filter: true}},
  }

  type TimeUnit = 'Horas'|'Dias'|'Semanas'|'Meses'
  const schedule_fields : PrimeFields = {
    unit: { label: 'Unidade de tempo', type: 'dropdown', options: {options: ['Horas','Dias','Semanas','Meses']}},
    number: { label: 'Número de unidade', type: 'number', options: { showButtons: true, step: 1}},
    date_str: { label: 'Hora do lembrete', type: 'value', options: { style: {color: 'blue'} }},
  }
  const schedule_fields_readonly : PrimeFields = {
    date_str: { label: 'Hora do lembrete', type: 'value', options: { style: {color: 'blue'} }},
  }
  
  ///////////////////////////  HOOKS  ///////////////////////////
  const show = useHookstate(props.state.show)
  const deleted = useHookstate(props.state.deleted)
  const error = useHookstate('')
  const contact_phones = useVariable<Contact[]>([])
  const contact_emails = useVariable<Contact[]>([])
  const lead = useHookstate(props.state.lead)
  const stages = jsonClone(props.state.stages.get())
  const leadV = props.state.lead.get()
  const customCityID = useHookstate('')

  const telephonePanel = {
    ref: React.useRef(null),
    state: useHookstate<Contact>({name: '', telephone: '', city: '', state: ''} as Contact)
  }
  const emailPanel = {
    ref: React.useRef(null),
    state: useHookstate<Contact>({name: '', telephone: '', city: '', state: ''} as Contact)
  }
  const schedulePanel = {
    ref: React.useRef(null),
    state: useHookstate({unit: 'Dias' as TimeUnit, number: 0, date_str: ''})
  }

  // Add team members variable
  const teamMembers = useVariable<{id: string, email: string, name?: string}[]>([])
  const selectedOwnerId = useHookstate('')

  ///////////////////////////  EFFECTS  ///////////////////////////

  React.useEffect(() => {
    if(show.get()) {
      getContacts()

      // build obra properties
      buildObraProps()
      
      // reset schedule
      schedulePanel.state.set({unit: 'Dias', number: 0, date_str: ''})
      
      // Fetch team members if user is a manager
      if (userS?.is_manager?.get()) {
        getTeamMembers()
      }
    }
  }, [show.get()]);

  // React.useEffect(() => {
  //   getContacts()
  // }, [lead.lead_properties.contacts.get()]);

  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const getTeamMembers = async () => {
    let resp = await api().get(makeURL('/team/members'), {})
    if (resp.error) {
      error.set(resp.error)
      return
    }
    const records = await resp.records()
    teamMembers.set(records.map(r => ({
      id: r.id,
      email: r.email,
      name: r.properties?.name
    })))

    // Set the current owner ID if available
    if (lead.owner_email.get()) {
      const matchingMember = records.find(r => r.email === lead.owner_email.get())
      if (matchingMember) {
        selectedOwnerId.set(matchingMember.id)
      }
    } else {
      selectedOwnerId.set('')
    }
  }

  const getContacts = async () => {
    let records : Contact[] = []
    if(lead.is_obra.get()) {
      let resp = await api().get(makeURL('/query/crm/contacts'), {obra_id: lead.obra_id.get()})
      if(resp.error) return
      records = await resp.records() as Contact[]
    }
    
    contact_phones.set(
      phones => {
        phones = []
        for(let contact of lead.lead_properties.contacts.get()) {
          if(!contact.telephone) continue
          phones.push(new Contact(contact))
        }
        
        for(let rec of records.filter(r => r.telephone))
          phones.push(rec)

        return phones
      }
    )
    
    contact_emails.set(
      emails => {
        emails = []
        for(let contact of lead.lead_properties.contacts.get()) {
          if(!contact.email) continue
          emails.push(new Contact(contact))
        }
        
        for(let rec of records.filter(r => r.email))
          emails.push(rec)
          
        return emails
      }
    )
  }

  const buildObraProps = () => {
      let obra_properties = new ObraProperties(jsonClone(lead.lead_properties.obra.get()))

      if(lead.is_obra.get()) {
        obra_properties.address = obra_properties.address || lead.address_short.get() || ''
        obra_properties.bairro = obra_properties.bairro || lead.bairro.get() || ''
        obra_properties.city = obra_properties.city || lead.city.get() || ''
        obra_properties.state = obra_properties.state || lead.state.get() || ''
        obra_properties.owner = obra_properties.owner || lead.owner.get() || ''
        obra_properties.professional = obra_properties.professional || lead.professional.get() || ''
        obra_properties.size = obra_properties.size || lead.size.get()
      }

      for(let city of allCities) {
        if(city.city.toUpperCase() === obra_properties.city?.toUpperCase()) customCityID.set(city.id)
      }


      lead.lead_properties.obra.set(obra_properties)
  }

  const setFromCustomCity = () => {
      lead.lead_properties.obra.set(
        obra => {
          for(let city of allCities) {
            if(city.id === customCityID.get()) {
              obra.city = city.city
              obra.state = city.state
            }
          }
          return obra
        }
      )
  }

  const saveLead = async () => {

    if (deleted.get()) {
      let resp = await deleteLead(lead.list_lead_id.get(), lead.lead_id.get())
      if(resp.error) error.set(resp.error)
      return resp.success
    }

    let lead_properties = new LeadProperties(jsonClone(lead.lead_properties.get()))
    if(lead.is_opportunity.get()) setFromCustomCity()
    lead_properties.obra = lead.get().obra_properties_payload

    let promise_list_lead = api().collection('list_lead')
      .update(lead.list_lead_id.get(), {stage_id: lead.stage_id.get()})

    let promise_lead = api().collection('lead')
      .update(lead.lead_id.get(), {properties: lead_properties})

    // Update owner ID if changed and user is manager
    let promise_owner_result = { success: true, error: '', data: null }
    if (userS?.is_manager?.get() && selectedOwnerId.get() && selectedOwnerId.get() !== '') {
      const resp_owner = await api().patch(makeURL('/patch/lead-owner'), {
        lead_id: lead.lead_id.get(),
        owner_id: selectedOwnerId.get()
      })
      promise_owner_result = {
        success: !resp_owner.error,
        error: resp_owner.error,
        data: await resp_owner.json()
      }
    }

    let resp_list_lead = await promise_list_lead
    let resp_lead = await promise_lead
    error.set(resp_lead.error || resp_list_lead.error || promise_owner_result.error)

    let resp_lead_record = await resp_lead.record()
    lead.set(
      l => {
        l.lead_properties = new LeadProperties(resp_lead_record.properties)
        l.sync_obra_properties()
        return l
      }
    )

    // Update owner info if owner was changed
    if (promise_owner_result.success && promise_owner_result.data?.owner_email) {
      lead.owner_email.set(promise_owner_result.data.owner_email)
      lead.owner_name.set(promise_owner_result.data.owner_name)
    }

    if(!error.get()) return true
    return false
  }

  const inputTransform = (val: string) => {
    return val?.toUpperCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "").trim()
  }

  const computeAlertAt = () => {
    let alert_at = new Date().getTime()
    let num  = schedulePanel.state.number.get()
    let unit = schedulePanel.state.unit.get()
    if(!num || `${num}` === '' || num < 0) return undefined

    if(unit === 'Horas') alert_at = alert_at + (num * 1000 * 3600 )
    if(unit === 'Dias') alert_at = alert_at + (num * 1000 * 3600 * 24 )
    if(unit === 'Semanas') alert_at = alert_at + (num * 1000 * 3600 * 24 * 7 )
    if(unit === 'Meses') alert_at = alert_at + (num * 1000 * 3600 * 24 * 30 )
    return new Date(alert_at)
  }
  ///////////////////////////  JSX  ///////////////////////////

  const starContact = (row: Contact, column: ColumnBodyOptions) => {
    let starred_contacts = lead.lead_properties.starred_contacts.get().includes(row.contact_id)
    let className = starred_contacts ? "pi pi-star-fill" : "pi pi-star"
    let color = starred_contacts ? "gold" : "lightgrey"
    return <div>
      <span
        className="cursor-pointer"
        onClick={async () => {
          lead.lead_properties.starred_contacts.set(
            sn => {
              if(sn.includes(row.contact_id)) sn = sn.filter(v => v !== row.contact_id)
              else sn.push(row.contact_id)
              return sn
            }
          )
        }}
      >
        <i className={className} style={{ fontSize: '1rem', color: color }}/>
      </span>
    </div>
  }

  const phoneBody = (row: Contact, column: ColumnBodyOptions) => {
    let ddd = row.telephone.toString().slice(0, 2)
    let prefix = row.telephone.toString().slice(2, 7)
    let suffix = row.telephone.toString().slice(7, 11)
    if (row.telephone.toString().length === 10) {
      prefix = row.telephone.toString().slice(2, 6)
      suffix = row.telephone.toString().slice(6, 11)
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

  const actionsBody = (row: Contact, column: ColumnBodyOptions) => {
    const n = Math.trunc(Math.random() * 10000000000)
    const iconStyle = { fontSize: '1.25rem', color: 'white', background: 'navy', cursor: 'pointer'}
    const whatsapp_id = `whatsapp-${row.contact_id}-${n}`
    const phone_id = `phone-${row.contact_id}-${n}`
    const email_id = `email-${row.contact_id}-${n}`
    const edit_id = `edit-${row.contact_id}-${n}`

    return <div className="flex justify-content-center text-center" style={{}}>
      {/* Disabled due to Weird bug where tooltip lingers */}
      {/* <Tooltip target={`#${whatsapp_id}`} position="top">Abrir WhatsApp</Tooltip>
      <Tooltip target={`#${phone_id}`} position="top">Ligue para o telefone</Tooltip>
      <Tooltip target={`#${email_id}`} position="right">Enviar Email</Tooltip> */}
  
      {
        isPhoneOrEmailContact(row.contact_id) ?
        <i
          id={edit_id}
          className="pi pi-pencil p-2 border-round-lg ml-2"
          style={iconStyle}
          onClick={(e) => {
            if(isPhoneContact(row.contact_id)) {
              telephonePanel.state.set(row)
              telephonePanel.ref.current.show(e)
            }
            if(isEmailContact(row.contact_id)) {
              emailPanel.state.set(row)
              emailPanel.ref.current.show(e)
            }
          }}
        />
        :
        null
      }

      {
        row.telephone ?
          <>
            <a id={phone_id} href={telephoneURL(row.telephone)} target="_blank">
              <i
                className="pi pi-phone p-2 border-round-lg ml-1"
                style={iconStyle}
              />
            </a>

            <a id={whatsapp_id} href={whatsAppURL(row.telephone)} target="_blank">
              <i
                className="pi pi-whatsapp p-2 border-round-lg ml-1"
                style={iconStyle}
              />
            </a>
          </>
          :
          <>
            <a id={email_id} href={emailURL(row.email)} target="_blank">
              <i
                className="pi pi-envelope p-2 border-round-lg ml-1"
                style={iconStyle}
              />
            </a>
          </>
      }
    </div> 
  }

  const ScheduleOverlay = (
    <OverlayPanel
      ref={schedulePanel.ref}
      style={{width: '300px'}}
      onShow={() => {
        if(lead.lead_properties.has_alert.get())
          schedulePanel.state.date_str.set(lead.lead_properties.alert_date_str.get())
      }}
    >
      <PrimeForm
        fields={lead.lead_properties.has_alert.get() ? schedule_fields_readonly : schedule_fields }
        getter={(key:string) => schedulePanel.state[key]?.get()}
        setter={(key:string, value: any) => { 
          schedulePanel.state[key].set(value)
          schedulePanel.state.date_str.set(computeAlertAt()?.toLocaleString('pt-BR') || '-')
        }}
        defaults={{size: 12}}
        buttons={() => {
          return <>
            <div className='field col-12 flex'>
              {
                !lead.lead_properties.has_alert.get() ?
                <>
                  <Button
                    label='Salvar'
                    size="small"
                    onClick={async () => {
                      // hide
                      schedulePanel.ref.current.hide()

                      let alert_at = computeAlertAt()

                      lead.lead_properties.alert_at.set(alert_at?.getTime())
                      lead.lead_properties.alerted.set(false)
                    }}
                  />

                  <Button
                    label='Cancelar'
                    className="ml-1"
                    severity='secondary'
                    size="small"
                    onClick={async () => {                  
                      // hide
                      schedulePanel.ref.current.hide()
                    }}
                  />
                </>
              :
              <Button
                label='Remover Lembrete'
                className="ml-1"
                severity='danger'
                icon='pi pi-trash'
                size="small"
                onClick={async () => {         
                  // hide
                  schedulePanel.ref.current.hide()

                  lead.lead_properties.alert_at.set(undefined)
                  lead.lead_properties.alerted.set(undefined)
      
                  // reset schedule
                  schedulePanel.state.set({unit: 'Dias', number: 0, date_str: ''})

                }}
              />
        }
            </div>
          </>
        }}
      />
    </OverlayPanel>
  )

  return (
    <div>
        <Dialog
          header={lead.title.get()}
          onHide={async () => {
            let success = await saveLead()
            if(success) show.set(false)
          }}
          visible={show.get()}
          style={{minWidth: '900px', maxWidth: '900px', minHeight: '700px', maxHeight: '700px'}}
          dismissableMask
        >
          <span style={{color: 'red'}}> {error.get()} </span>
          <span className="cursor-pointer" style={{position: 'absolute', top:90, right:30, zIndex:9999}}>
            <Button
              icon="pi pi-clock"
              className="mr-2"
              size="small"
              tooltip='Agende um lembrete para entrar em contato com o cliente'
              tooltipOptions={{position: 'top'}}
              onClick={(e) => { 
                schedulePanel.state.date_str.set(computeAlertAt()?.toLocaleString('pt-BR') || '-')
                schedulePanel.ref.current.toggle(e)
              }}
              outlined={!lead.lead_properties.has_alert.get()}
            />
            { ScheduleOverlay }

            <Button
              tooltip={deleted.get() ? 'Cancelar remover lead' :"Remover lead"}
              icon='pi pi-trash'
              severity="danger"
              size="small"
              onClick={async () => {
                props.state.deleted.set(v => !v)
              }}
              outlined={!deleted.get()}
            />
          </span>

          <TabView>
            <TabPanel header="Obra">

              <div className="grid overflow-scroll">
                <div className="md:col-7 col-12">
                  <div className="mb-1 mt-1">
                    <strong>Proprietário:</strong> 
                    <EditableText
                      value={lead.lead_properties.obra.owner}
                      transform={inputTransform}
                      size={35}
                    />
                  </div>
                  <div className="mb-3">
                    <strong>Profissional:</strong> 
                    <EditableText
                      value={lead.lead_properties.obra.professional}
                      transform={inputTransform}
                      size={35}
                    />
                  </div>
                  <div className="mb-1">
                    <strong>Enderco:</strong> 
                    <EditableText
                      value={lead.lead_properties.obra.address}
                      transform={inputTransform}
                      editable={lead.is_opportunity.get()}
                      size={40}
                    />
                  </div>
                  {
                    lead.is_obra.get() ?
                    <>
                      <div className="mb-1"><strong>Bairro:</strong> {leadV.bairro}</div>
                      <div className="mb-3"><strong>Cidade:</strong> {leadV.city}, {leadV.state}</div>
                      <div className="mb-1"><strong>Data Inicio:</strong> <span style={{marginLeft: 20}}>{leadV.start_date_str}</span></div>
                      <div className="mb-1"><strong>Data Termino:</strong> {leadV.end_date_str}</div>
                    </>
                    :
                    <div className="mb-3">
                      <strong>Cidade:</strong> 
                      <EditableText
                        value={customCityID}
                        transform={inputTransform}
                        autoComplete={allCities.map(c => c.id)}
                      />
                    </div>
                  }
                </div>

                <div className="md:col-5 col-12">
                    {
                      lead.is_obra.get() ?
                      <div className="mb-2 flex flex-wrap">
                        <strong  style={{width: 130}}>Tamanho:</strong>
                        <span> {leadV.size} m² </span>
                      </div>
                      :
                      <div className="mb-2 flex flex-wrap align-content-center">
                        <div className="flex flex-wrap align-content-center">
                          <strong style={keyStyle}>Tamanho:</strong> 
                        </div>
                        <InputNumber
                          id="size-input"
                          className="ml-2"
                          size={12}
                          suffix=" m²"
                          showButtons
                          step={10}
                          maxFractionDigits={0}
                          style={{padding: 0}}
                          value={lead.lead_properties.obra.size.get()}
                          onChange={async (e) => lead.lead_properties.obra.size.set(e.value)}
                        />
                      </div>
                    }
                    
                  <div className="mb-3 flex flex-wrap">
                    <strong style={keyStyle}>Qualificação:</strong>
                    <Rating
                      className="ml-3"
                      value={lead.lead_properties.rating.get()}
                      cancel={false}
                      onChange={async (e) => lead.lead_properties.rating.set(e.value)}
                    />
                  </div>

                  <div className="mb-2 flex flex-wrap align-content-center">
                    <div className="flex flex-wrap align-content-center">
                      <strong style={keyStyle}>Valor:</strong> 
                    </div>
                    <InputNumber
                      id="valor-input"
                      className="ml-2"
                      size={12}
                      locale="pt-BR"
                      mode="currency"
                      currency="BRL"
                      showButtons
                      step={1000}
                      maxFractionDigits={0}
                      style={{padding: 0}}
                      value={lead.lead_properties.valor.get()}
                      onChange={async (e) => lead.lead_properties.valor.set(e.value)}
                    />
                  </div>

                  <div className="mb-4 flex flex-wrap">
                    <div className="flex flex-wrap align-content-center">
                      <strong style={keyStyle}>Etapa de vendas:</strong> 
                    </div>
                    <div style={{width: 170}}>
                      <Dropdown
                        className="ml-2 w-full step-dropdown"
                        style={{padding: 0, paddingLeft: 3, paddingRight: 3}}
                        value={lead.stage_id.get()}
                        onChange={(e) => lead.stage_id.set(e.value)}
                        options={stages}
                        optionLabel="name" 
                        optionValue="id"
                      />
                    </div>
                  </div>

                  {
                    userS?.is_manager?.get() &&
                    <div className="mb-4 flex flex-wrap">
                      <div className="flex flex-wrap align-content-center">
                        <strong style={keyStyle}>Responsável:</strong> 
                      </div>
                      <div style={{width: 170}}>
                        <Dropdown
                          className="ml-2 w-full step-dropdown"
                          style={{padding: 0, paddingLeft: 3, paddingRight: 3}}
                          value={selectedOwnerId.get()}
                          onChange={(e) => selectedOwnerId.set(e.value)}
                          options={teamMembers.get()}
                          optionLabel="email" 
                          optionValue="id"
                          placeholder="Selecione um usuário"
                        />
                      </div>
                    </div>
                  }
                </div>
              </div>

              <div className="mb-2 flex flex-wrap">
                <strong>Observações:</strong> 
              </div>
              <Editor
                value={lead.lead_properties.observations.get()}
                onTextChange={(e) => lead.lead_properties.observations.set(e.htmlValue)}
                style={{ height: '200px' }}
              />
            </TabPanel>
            
            <TabPanel header="Telefones">
              <div className="m-0">
                <DataTable
                  value={contact_phones.get()}
                  scrollable
                  scrollHeight='400px'
                  emptyMessage="Nenhum contacto encontrado"
                >
                  <Column body={starContact} header='Fav.'></Column>
                  <Column field='name' header="Nome"></Column>
                  <Column body={phoneBody} header="Número de Telefone"></Column>
                  <Column field="city" header="Cidade"></Column>
                  {/* <Column field="state" header="UF"></Column> */}
                  <Column field='lead_contact' body={actionsBody} header="Ações" 
                    headerStyle={{width: '11em'}} 
                    headerClassName="flex justify-content-center text-center"
                    bodyStyle={{width: '11em'}}/>
                </DataTable>
              </div>

              <div className="flex justify-content-center pt-2">
                <Button
                  label="Adicionar Telefone"
                  icon="pi pi-plus"
                  severity="success"
                  size="small"
                  onClick={(e) => telephonePanel.ref.current.toggle(e)}
                />
                <OverlayPanel
                  ref={telephonePanel.ref}
                  style={{width: '300px'}}
                >
                  <PrimeForm
                    fields={telephone_fields}
                    getter={(key:string) => telephonePanel.state[key]?.get()}
                    setter={(key:string, value: string) => {
                      value = (value || '').toUpperCase()
                      telephonePanel.state[key].set(value)
                    }}
                    defaults={{size: 12}}
                    buttons={() => {
                      return <>
                        <div className='field col-12 flex'>
                          <Button
                            label='Salvar'
                            onClick={async () => { 
                              lead.lead_properties.contacts.set(
                                contacts => {
                                  let contact = new Contact(jsonClone(telephonePanel.state.get()))
                                  for (let i = 0; i < contacts.length; i++) {
                                    if(contacts[i].contact_id === contact.contact_id) {
                                      contacts[i] = contact
                                      return contacts
                                    }
                                  }
                                  contacts.push(contact)
                                  return contacts
                                }
                              )
                              telephonePanel.state.set({name: '', telephone: '', city: '', state: ''} as Contact)
                              
                              // refresh && hide
                              getContacts()
                              telephonePanel.ref.current.hide()
                            }}
                          />
                        </div>
                      </>
                    }}
                  />
                </OverlayPanel>
              </div>
            </TabPanel>

            <TabPanel header="Emails">
              <div className="m-0">
                <DataTable
                  value={contact_emails.get()}
                  scrollable
                  scrollHeight='400px'
                  emptyMessage="Nenhum contacto encontrado"
                >
                  <Column body={starContact} header='Fav.'></Column>
                  <Column field='name' header="Nome"></Column>
                  <Column body={emailBody} header="Email"></Column>
                  {/* <Column field="city" header="Cidade"></Column> */}
                  {/* <Column field="state" header="UF"></Column> */}
                  <Column field='lead_contact' body={actionsBody} header="Ações" 
                    headerStyle={{width: '11em'}} 
                    headerClassName="flex justify-content-center text-center"
                    bodyStyle={{width: '11em'}}/>
                </DataTable>
              </div>

              <div className="flex justify-content-center pt-2">
                <Button
                  label="Adicionar Email"
                  icon="pi pi-plus"
                  severity="success"
                  size="small"
                  onClick={(e) => emailPanel.ref.current.toggle(e)}
                />
                <OverlayPanel
                  ref={emailPanel.ref}
                  style={{width: '300px'}}
                >
                  <PrimeForm
                    fields={email_fields}
                    getter={(key:string) => emailPanel.state[key]?.get()}
                    setter={(key:string, value: string) => {
                      value = (value || '').toUpperCase()
                      if(key === 'email') value = value.toLowerCase()
                      emailPanel.state[key].set(value)
                    }}
                    defaults={{size: 12}}
                    buttons={() => {
                      return <>
                        <div className='field col-12 flex'>
                          <Button
                            label='Salvar'
                            onClick={async () => {
                              lead.lead_properties.contacts.set(
                                contacts => {
                                  let contact = new Contact(jsonClone(emailPanel.state.get()))
                                  for (let i = 0; i < contacts.length; i++) {
                                    if(contacts[i].contact_id === contact.contact_id) {
                                      contacts[i] = contact
                                      return contacts
                                    }
                                  }
                                  contacts.push(contact)
                                  return contacts
                                }
                              )
                              emailPanel.state.set({name: '', telephone: '', city: '', state: ''} as Contact)
                              
                              // refresh && hide
                              getContacts()
                              emailPanel.ref.current.hide()
                            }}
                          />
                        </div>
                      </>
                    }}
                  />
                </OverlayPanel>
              </div>
            </TabPanel>
          </TabView>

        </Dialog>
    </div>
  )
}

export const deleteLead = async (list_lead_id: string, lead_id: string) => {
    let error = ''
    let resp = await api().collection('list_lead').delete(list_lead_id)
    if(resp.error) {
      error = resp.error
      return { success: false, error }
    }

    // remove favorite_at
    resp = await api().collection('lead').update(lead_id, {favorited_at: null, excluded_at: (new Date()).toISOString().replaceAll('T', ' ')})
    if(resp.error) {
      error = resp.error
      return { success: false, error }
    }

    return { success: true, error }
}