import React from 'react'
import { Button } from 'primereact/button';
import { PB, api, baseURL } from '../store/api';
import { User, brasilStates, loadUserState, userS } from '../store/store';
import { useVariable } from '../utils/interfaces';
import { useHookstate, type State } from '@hookstate/core';
import { doToast, jsonClone } from '../utils/methods';
import { Dialog } from 'primereact/dialog';
import { Accordion, AccordionTab } from 'primereact/accordion';
import { Avatar } from 'primereact/avatar';
import { InputText } from 'primereact/inputtext';
import PrimeForm, { type PrimeFields } from '../utils/PrimeForm';
import { Password } from 'primereact/password';
import { Toast } from 'primereact/toast';
import { ConfirmDialog, confirmDialog } from 'primereact/confirmdialog';
import { DataTable } from 'primereact/datatable';
import { Column } from 'primereact/column';
import { InputSwitch } from 'primereact/inputswitch';

interface Props {}

export default function NavBar(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  ///////////////////////////  HOOKS  ///////////////////////////
  const name = useVariable('')
  const isManager = useVariable(false)
  const profileDialog = useHookstate<DialogParams>({user: new User(jsonClone(userS.get()))})
  const teamDialog = useHookstate<DialogParams>({user: new User(jsonClone(userS.get()))})
  
  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    name.set(PB().authStore.model.email)
    loadUser().then(u => {
      profileDialog.user.set(u)
      teamDialog.user.set(u)
    })
  }, [])


  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const loadUser = async () => {
    let user = await loadUserState()
    if(user.team.allow_contact && false) {
      // get token for extension
      let resp = await api().get(`${baseURL()}/messenger/user`, {})
      let record = await resp.json()
      window.token = record?.user?.token
    }
    isManager.set(user.is_manager)
    return user
  }

  ///////////////////////////  JSX  ///////////////////////////
  return (
  <>
    <div className='flex flex-wrap w-full align-items-center justify-content-between bg-white m-0 py-3 h-4 shadow-1 nav-bar'>
      {/* Logo */}
      <img id='nav-bar-logo' src="/logo-text.svg" alt="" style={{maxHeight: '50px'}}/>

      <div className='flex align-items-center justify-content-right'>
        {/* Client name */}
        <div id='nav-bar-client' className='mr-3 cursor-pointer' onClick={() => profileDialog.show.set(true)}>
          <strong> { name.get()}</strong>
        </div>
        
        {/* Client icon */}
        <div className='mr-3 cursor-pointer' onClick={() => profileDialog.show.set(true)}>

          <Avatar image="/logo.svg" size="xlarge" shape="circle" />
        </div>
        

        {
          isManager.get() &&
          <div className='team-button'>
            <Button
              id='team-button'
              style={{color: 'black'}}
              size='small'
              label="Equipe"
              icon="pi pi-users"
              className='mr-2'
              onClick={() => teamDialog.show.set(true)}
              rounded
              outlined/>
          </div>
        }

        <div className='logout-button'>
          <Button
            id='logout-button'
            style={{color: 'black'}}
            size='small'
            label="Sair"
            icon="pi pi-sign-out"
            iconPos="right"
            onClick={() => LogOut()}
            outlined/>
        </div>
      </div>
    </div>
    <ProfileDialog state={profileDialog} />
    <TeamDialog state={teamDialog} />
  </>
  );
};

export const LogOut = () => {
  PB().authStore.clear()
  window.location.assign('/')
}

export interface DialogParams {
  show?: boolean;
  user: User;
}

function ProfileDialog(props: { state: State<DialogParams> }) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const propertyStyle : React.CSSProperties = { paddingBottom: 10 }
  const propertyIconStyle : React.CSSProperties = { fontSize: '0.9rem' }

  const profile_fields : PrimeFields = {
    name: { label: 'Nome da Loja', type: 'string'},
    founded_date: { label: 'Data de Fundação', type: 'date', options: {dateFormat: "dd/mm/yy" }},
    description: { label: 'Descrição', size: 12, type: 'text-area'},
    cpf: { label: 'CPF', type: 'mask', options: {mask: '999.999.999-99'}},
    cnpj: { label: 'CNPJ', type: 'mask', options: {mask: '99.999.999/9999-99'}},
    telephone: { label: 'Telefone', type: 'mask', options: {mask: '(99) 99999 - 9999'}},
    whatsapp: { label: 'WhatsApp', type: 'mask', options: {mask: '(99) 99999 - 9999'}},
    website: { label: 'Site da Internet', type: 'string'},
    industry: { label: 'Indústria', type: 'string'},
    maps_url: { label: 'Google Maps Link', size: 12, type: 'string'},
    keywords: { label: 'Palavras-chave', size: 12, type: 'chips'},
  }

  const address_fields : PrimeFields = {
    enderco: { label: 'Enderco', size: 12, type: 'string'},
    numero: { label: 'Numero', type: 'string'},
    complemento: { label: 'Complemento', type: 'string'},
    bairro: { label: 'Bairro', type: 'string'},
    cidade: { label: 'Cidade', type: 'string'},
    uf: { label: 'UF', type: 'dropdown', options: {options: brasilStates}},
    cep: { label: 'CEP', type: 'mask', options: {mask: '99999-999'}},
  }

  const context_fields : PrimeFields = {
    sender: { label: 'Remetente', placeholder: 'O nome da pessoa de onde a mensagem mostra vem',  size: 12, type: 'string'},
    context: { label: 'Contexto Adicional', size: 12, type: 'text-area', placeholder: 'O que você quer que o modelo enfatize? Você pode colocar algo como "estamos vendendo na pós-construção" ou "podemos ir até você em 24 horas".'},
  }

  const template_fields : PrimeFields = {
    owner1: { label: 'Modelo - Proprietário (estilo 1)', size: 12, type: 'text-area', placeholder: 'Clique em "Gerar Modelos" ou escreva você mesmo o modelo.'},
    owner2: { label: 'Modelo - Proprietário (estilo 2)', size: 12, type: 'text-area', placeholder: 'Clique em "Gerar Modelos" ou escreva você mesmo o modelo.'},
    professional1: { label: 'Modelo - Profissional (estilo 1)', size: 12, type: 'text-area', placeholder: 'Clique em "Gerar Modelos" ou escreva você mesmo o modelo.'},
    professional2: { label: 'Modelo - Profissional (estilo 2)', size: 12, type: 'text-area', placeholder: 'Clique em "Gerar Modelos" ou escreva você mesmo o modelo.'},
  }

  ///////////////////////////  HOOKS  ///////////////////////////
  const show = useHookstate(props.state.show)
  const user = useHookstate(props.state.user)
  const team = useHookstate(user.team)
  const email = useVariable(user.email.get())
  const error = useVariable('')
  const isModified = useVariable(false)
  const passwordOld = useVariable('')
  const passwordNew = useVariable('')
  const passwordNewConfirm = useVariable('')
  const toast = React.useRef<Toast>(null)
  const dialogRef = React.useRef(null)
  const templatesLoading = useHookstate(false)

  ///////////////////////////  EFFECTS  ///////////////////////////
  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const saveTeam = async () => {
    // Get current values to compare with original for key fields
    const currentProperties = team.properties.get()
    
    // Save the team first
    let resp = await api().collection('team').update(
      team.id.get(), 
      {properties: jsonClone(currentProperties)}, 
    )
    error.set(resp.error)
    if(resp.error) return false
    isModified.set(false)
    
    // Check if key fields were modified that require lead introduction regeneration
    const originalUser = await loadUserState() // reload and get original
    const originalProperties = originalUser.team.properties
    
    const keyFields = ['name', 'description', 'industry', 'keywords', 'founded_date']
    const hasKeyFieldChanges = keyFields.some(field => 
      currentProperties[field] !== originalProperties[field]
    )
    
    if(hasKeyFieldChanges) {
      // Generate new lead introduction text
      try {
        const leadIntroResp = await api().post(`${baseURL()}/messenger/generate-lead-introduction`, {})
        
        if(!leadIntroResp.error) {
          const leadIntroData = await leadIntroResp.json()
          
          // Check if there's a warning (text generated but not saved)
          if(leadIntroData.warning) {
            team.properties.lead_introduction_text.set(leadIntroData.lead_introduction_text)
            doToast(toast, {
              severity: 'warn',
              summary: 'Parcialmente salvo',
              detail: 'Perfil salvo e texto gerado, mas houve problema ao salvar o texto no banco.',
            }, 5000)
          } else {
            // Update the team state with the new text
            team.properties.lead_introduction_text.set(leadIntroData.lead_introduction_text)
            
            doToast(toast, {
              severity: 'success',
              summary: 'Sucesso',
              detail: 'Perfil salvo e texto de apresentação atualizado automaticamente!',
            }, 4000)
          }
        } else {
          // If lead intro generation fails, still show success for the profile save
          console.warn('Failed to generate lead introduction:', leadIntroResp.error)
          doToast(toast, {
            severity: 'warn',
            summary: 'Parcialmente salvo',
            detail: 'Perfil salvo, mas houve um problema ao gerar o texto de apresentação.',
          }, 4000)
        }
      } catch (err) {
        console.warn('Error generating lead introduction:', err)
        // Profile was saved successfully, just warn about the lead intro
        doToast(toast, {
          severity: 'warn',
          summary: 'Parcialmente salvo',
          detail: 'Perfil salvo, mas o serviço de IA está temporariamente indisponível.',
        }, 4000)
      }
    } else {
      // No key fields changed, just show normal success
      doToast(toast, {
        severity: 'success',
        summary: 'Sucesso',
        detail: 'Perfil salvo com sucesso!',
      }, 3000)
    }
    
    return true
  }

  const generateTemplates = async () => {
    if(team.properties.name.get().trim() === ''  || team.properties.description.get().trim() === '') 
      return doToast(toast, {
        severity: 'error',
        summary: 'Erro',
        detail: 'Você precisa salvar o nome e a descrição da sua loja (em Alterar perfil)',
      }, 4000)

    templatesLoading.set(true)
    let resp = await api().post(`${baseURL()}/messenger/generate-templates`, {})
    templatesLoading.set(false)
    if(resp.error) throw new Error(resp.error)
    let json_data = await resp.json()
    let owner_templates = json_data.owner as string[]
    let professional_templates = json_data.professional as string[]
    team.properties.templates.set(t => {
      if(owner_templates?.length >= 2) {
        t.owner1 = owner_templates[0]
        t.owner2 = owner_templates[1]
      }

      if(professional_templates?.length >= 2) {
        t.professional1 = professional_templates[0]
        t.professional2 = professional_templates[1]
      }

      return t
    })

    return true
  }

  const requestEmailChange = async () => {
    let success = await PB().collection('user').requestEmailChange(email.get());
    return success
  }

  const updatePassword = async () => {
    let resp = await api().collection('user').update(
      user.id.get(),
      {
        password: passwordNew.get(),
        passwordConfirm: passwordNewConfirm.get(),
        oldPassword: passwordOld.get(),
      }, 
    )
    error.set(resp.error)
    if(resp.error) return false
    return true
  }
  ///////////////////////////  JSX  ///////////////////////////
  return (
    <div>
        <Toast ref={toast} />
        <ConfirmDialog />

        <Dialog
          ref={dialogRef}
          header={'Perfil'}
          onHide={async () => {
            if(isModified.get())
              return confirmDialog({
                  message: 'Existem alterações não salvas. Tem certeza de que deseja fechar?',
                  header: 'Confirmação',
                  icon: 'pi pi-info-circle',
                  acceptClassName: 'p-button-danger',
                  acceptLabel: 'Sim',
                  rejectLabel: 'Não',
                  accept: () => show.set(false),
                  reject: () => {},
                  style: { maxWidth: '400px' }
              })
            
            isModified.set(false)
            show.set(false)
          }}
          visible={show.get()}
          style={{width: 600}}
          dismissableMask
        >

          <span style={{color: 'red', position: 'absolute', top: 53, left: 24}}> { error.get() } </span>
          <div className='grid mx-5'>
            <div className='col'>
              <img className='border-circle w-12rem h-12rem' src="/logo.svg" alt="" />
            </div>
            <div className='col align-content-center'>
              <h3> {team.name.get()} </h3>
              <div style={propertyStyle}>
                <i className="pi pi-map-marker mr-1" style={propertyIconStyle}></i>
                {team.properties.address.cidade.get()}, {team.properties.address.uf.get()}
              </div>

              <div style={propertyStyle}>
                <i className="pi pi-globe mr-1" style={propertyIconStyle}></i>
                {team.properties.website?.get()}
              </div>

              <div style={propertyStyle}>
                <i className="pi pi-phone mr-1" style={propertyIconStyle}></i>
                {team.properties.telephone?.get() || team.properties.whatsapp?.get()}
              </div>

              {team.properties.lead_introduction_text?.get() && (
                <div style={{...propertyStyle, marginTop: '10px', padding: '10px', backgroundColor: '#f8f9fa', borderRadius: '5px', borderLeft: '4px solid #007bff'}}>
                  <i className="pi pi-comment mr-1" style={propertyIconStyle}></i>
                  <span style={{fontStyle: 'italic', color: '#495057'}}>
                    {team.properties.lead_introduction_text.get()}
                  </span>
                </div>
              )}

              {/* <div style={propertyStyle}>
                <i className="pi pi-whatsapp mr-1" style={propertyIconStyle}></i>
                {team.properties.whatsapp?.get()}
              </div> */}
              
            </div>
          </div>
          <br/>

          <Accordion >
              <AccordionTab header="Alterar Perfil">
                <PrimeForm
                  fields={profile_fields}
                  getter={(key:string) => team.properties[key]?.get()}
                  setter={(key:string, value: any) => {
                    if(!isModified.get()) isModified.set(true)
                    team.properties[key].set(value)
                  }}
                  defaults={{size: 6}}
                  buttons={() => {
                    return <>
                      <div className='field col-6'>
                        {/* <Button
                          label='Cancelar'
                          className="p-button-warning mr-2"
                          onClick={() => { show.set(false) }}
                        /> */}
                        <Button
                          label='Salvar'
                          onClick={async () => { 
                            let success = await saveTeam() 
                            if(success)
                              doToast(toast, {
                                severity: 'info',
                                summary: 'Sucesso'
                              }, 1000)
                          }}
                        />
                      </div>
                    </>
                  }}
                />
              </AccordionTab>

              <AccordionTab header="Alterar Enderco">
                <PrimeForm
                  fields={address_fields}
                  getter={(key:string) => team.properties?.address[key]?.get()}
                  setter={(key:string, value: any) => {
                    if(!isModified.get()) isModified.set(true)
                    team.properties.address[key].set(value)
                  }}
                  defaults={{size: 6}}
                  buttons={() => {
                    return <>
                      <div className='field col-12 flex'>
                        {/* <Button
                          label='Cancelar'
                          className="p-button-warning mr-2"
                          onClick={() => { show.set(false) }}
                        /> */}
                        <Button
                          label='Salvar'
                          onClick={async () => { 
                            let success = await saveTeam() 
                            if(success)
                              doToast(toast, {
                                severity: 'info',
                                summary: 'Sucesso'
                              }, 1000)
                          }}
                        />
                      </div>
                    </>
                  }}
                />
              </AccordionTab>

              { team.allow_contact?.get() && user.properties.whatsapp.on_boarded.get() &&
              <AccordionTab header="Alterar Modelos">
                <div className='pb-1'>Aqui você pode definir os modelos para as mensagens a serem enviadas aos prospects. Clique no botão "Gerar Modelos" abaixo e geraremos modelos para você com IA.</div>
                <div>Quando você enviar um lote de mensagens, selecionaremos aleatoriamente entre o estilo 1 e 2, dependendo se o destinatário é um proprietário ou um profissional.</div>
                <hr/>

                <PrimeForm
                  fields={context_fields}
                  getter={(key:string) => team.properties?.templates[key]?.get()}
                  setter={(key:string, value: any) => {
                    if(!isModified.get()) isModified.set(true)
                    team.properties.templates[key].set(value)
                  }}
                  defaults={{size: 6}}
                  buttons={() => {
                    return <>
                      <div className='field col-12 flex'>
                        <Button
                          label='Gerar Modelos'
                          className='ml-1'
                          loading={templatesLoading.get()}
                          severity='info'
                          onClick={async () => { 
                            if(!(await saveTeam())) return  // save first
                            let success = await generateTemplates() 
                            if(success)
                              doToast(toast, {
                                severity: 'info',
                                summary: 'Sucesso'
                              }, 1000)
                          }}
                        />
                      </div>
                    </>
                  }}
                />

                <hr/>

                <PrimeForm
                  fields={template_fields}
                  getter={(key:string) => team.properties?.templates[key]?.get()}
                  setter={(key:string, value: any) => {
                    if(!isModified.get()) isModified.set(true)
                    team.properties.templates[key].set(value)
                  }}
                  defaults={{size: 6}}
                  buttons={() => {
                    return <>
                      <div className='field col-12 flex'>
                        {/* <Button
                          label='Cancelar'
                          className="p-button-warning mr-2"
                          onClick={() => { show.set(false) }}
                        /> */}
                        <Button
                          label='Salvar'
                          onClick={async () => { 
                            let success = await saveTeam() 
                            if(success)
                              doToast(toast, {
                                severity: 'info',
                                summary: 'Sucesso'
                              }, 1000)
                          }}
                        />
                      </div>
                    </>
                  }}
                />
              </AccordionTab>
              }

              <AccordionTab header="Alterar Email">
                <div className="formgrid grid">
                  <div className="field col-12">
                    <label htmlFor="search-filter">Email Atual</label>
                    <InputText
                      disabled
                      aria-describedby="filter-help"
                      className="w-full"
                      value={user.email.get()}
                    />
                  </div>
                  <div className="field col-12">
                    <label htmlFor="search-filter">Novo Email</label>
                    <div className="p-inputgroup">
                      <InputText
                        placeholder="Digite o novo email"
                        className="w-full"
                        value={email.get()}
                        onChange={(e) => {
                          email.set(e.target.value)
                        }}
                      />
                      <Button
                        icon="pi pi-times"
                        className="p-button-warning"
                        tooltip="Cancelar"
                        tooltipOptions={{position: 'top'}}
                        onClick={() => { email.set(user.email.get()) }}
                      />
                      <Button
                        icon="pi pi-check"
                        className="p-button-primary"
                        tooltip="Salvar"
                        tooltipOptions={{position: 'top'}}
                        onClick={async () => {
                          let success = await requestEmailChange()
                          if(success)
                            doToast(toast, {
                              severity: 'info',
                              summary: 'Sucesso'
                            }, 1000)
                          }}
                      />
                    </div>
                  </div>
                </div>
              </AccordionTab>
              <AccordionTab header="Alterar Senha">
                <div className="formgrid grid">
                  <div className="field md:col-6 col-12">
                    <label htmlFor="search-filter">Senha Atual</label>
                    <Password
                      placeholder="Digite a senha actual"
                      className="w-full"
                      value={passwordOld.get()}
                      feedback={false} 
                      onChange={(e) => {
                        passwordOld.set(e.target.value)
                      }}
                    />
                  </div>
                  <div className="field md:col-6 col-12">
                    <label htmlFor="search-filter">Nova Senha</label>
                    <Password
                      placeholder="Digite a nova senha"
                      className="w-full"
                      feedback={false} 
                      value={passwordNew.get()}
                      onChange={(e) => {
                        passwordNew.set(e.target.value)
                      }}
                    />
                  </div>
                  <div className="field md:col-6 col-12">
                    <label htmlFor="search-filter">Confirma Nova Senha</label>
                    <Password
                      placeholder="A nova senha de novo"
                      className="w-full"
                      feedback={false} 
                      value={passwordNewConfirm.get()}
                      onChange={(e) => {
                        passwordNew.set(e.target.value)
                      }}
                    />
                  </div>
                    
                  <div className="field md:col-6 col-12 pt-4">
                    <Button
                      icon="pi pi-times"
                      className="p-button-warning mr-2"
                      tooltip="Cancelar"
                      tooltipOptions={{position: 'top'}}
                      onClick={() => { passwordOld.set(''); passwordNew.set('') }}
                    />
                    <Button
                      icon="pi pi-check"
                      className="p-button-primary"
                      tooltip="Salvar"
                      tooltipOptions={{position: 'top'}}
                      onClick={() => {
                        let success = updatePassword()
                        if(success)
                          doToast(toast, {
                            severity: 'info',
                            summary: 'Sucesso'
                          }, 1000)
                      }}
                    />
                  </div>
                </div>
              </AccordionTab>
          </Accordion>
        </Dialog>
    </div>
  );
};

function TeamDialog(props: { state: State<DialogParams> }) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const propertyStyle : React.CSSProperties = { paddingBottom: 10 }
  const propertyIconStyle : React.CSSProperties = { fontSize: '0.9rem' }

  ///////////////////////////  HOOKS  ///////////////////////////
  const show = useHookstate(props.state.show)
  const user = useHookstate(props.state.user)
  const team = useHookstate(user.team)
  const error = useVariable('')
  const isModified = useVariable(false)
  const toast = React.useRef<Toast>(null)
  const dialogRef = React.useRef(null)
  const teamMembers = useHookstate<User[]>([])
  const newUserEmail = useVariable('')
  const selectedUserId = useVariable('')

  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    if (show.get()) {
      // Load team members when dialog is opened
      loadTeamMembers()
    }
  }, [show.get()])

  ///////////////////////////  FUNCTIONS  ///////////////////////////
  const loadTeamMembers = async () => {
    try {
      let resp = await api().get(`${baseURL()}/team/members`, {
        team_id: team.id.get()
      })
      if (resp.error) {
        error.set(resp.error)
        return false
      }
      
      let members = (await resp.json()) as any[]
      teamMembers.set(members.map(m => new User(m)))
      error.set('')
      return true
    } catch (e) {
      error.set(e.toString())
      return false
    }
  }

  const inviteUser = async () => {
    if (!newUserEmail.get().trim()) {
      doToast(toast, {
        severity: 'error',
        summary: 'Erro',
        detail: 'Digite um email válido',
      }, 4000)
      return false
    }

    try {
      let resp = await api().post(`${baseURL()}/team/invite`, {
        team_id: team.id.get(),
        email: newUserEmail.get().trim().toLowerCase(),
      })
      
      if (resp.error) {
        error.set(resp.error)
        return false
      }
      
      doToast(toast, {
        severity: 'success',
        summary: 'Sucesso',
        detail: 'Convite enviado com sucesso',
      }, 4000)
      
      newUserEmail.set('')
      await loadTeamMembers()
      return true
    } catch (e) {
      error.set(e.toString())
      return false
    }
  }

  const removeUser = async (userId: string) => {
    confirmDialog({
      message: 'Tem certeza que deseja remover este usuário da equipe?',
      header: 'Confirmação',
      icon: 'pi pi-exclamation-triangle',
      acceptLabel: 'Sim',
      rejectLabel: 'Não',
      acceptClassName: 'p-button-danger',
      accept: async () => {
        try {
          let resp = await api().post(`${baseURL()}/team/remove-member`, {
            team_id: team.id.get(),
            user_id: userId
          })
          
          if (resp.error) {
            error.set(resp.error)
            return false
          }
          
          doToast(toast, {
            severity: 'success',
            summary: 'Sucesso',
            detail: 'Usuário removido com sucesso',
          }, 4000)
          
          await loadTeamMembers()
          return true
        } catch (e) {
          error.set(e.toString())
          return false
        }
      }
    })
  }

  const toggleManager = async (userId: string, isManager: boolean) => {
    try {
      let resp = await api().post(`${baseURL()}/team/set-manager`, {
        team_id: team.id.get(),
        user_id: userId,
        is_manager: !isManager
      })
      
      if (resp.error) {
        error.set(resp.error)
        return false
      }
      
      doToast(toast, {
        severity: 'success',
        summary: 'Sucesso',
        detail: !isManager ? 'Usuário agora é gerente' : 'Usuário não é mais gerente',
      }, 4000)
      
      await loadTeamMembers()
      return true
    } catch (e) {
      error.set(e.toString())
      return false
    }
  }

  ///////////////////////////  TEMPLATES  ///////////////////////////
  const emailTemplate = (rowData: User) => {
    return <span>{rowData.email}</span>
  }

  const managerTemplate = (rowData: User) => {
    return (
      <InputSwitch
        checked={rowData.is_manager || false}
        onChange={() => toggleManager(rowData.id, rowData.is_manager)}
        disabled={rowData.id === userS.get().id}
        tooltip={rowData.id === userS.get().id ? "Você não pode remover seu próprio acesso de gerente" : ""}
        tooltipOptions={{ position: 'top' }}
      />
    )
  }

  const actionTemplate = (rowData: User) => {
    return (
      <Button
        icon="pi pi-trash"
        className="p-button-danger p-button-text"
        tooltip="Remover Usuário"
        disabled={rowData.id === userS.get().id}
        tooltipOptions={{ position: 'top' }}
        onClick={() => removeUser(rowData.id)}
      />
    )
  }

  ///////////////////////////  JSX  ///////////////////////////
  return (
    <div>
      <Toast ref={toast} />
      <ConfirmDialog />

      <Dialog
        ref={dialogRef}
        header={'Gerenciar Equipe'}
        onHide={() => show.set(false)}
        visible={show.get()}
        style={{width: 800}}
        dismissableMask
        draggable={false}
      >
        <span style={{color: 'red', position: 'absolute', top: 53, left: 24}}>{error.get()}</span>

        <div className="field col-12">
          <DataTable value={teamMembers.get().map(m => m)} responsiveLayout="scroll">
            <Column field="email" header="Email" body={emailTemplate} style={{ width: '60%' }} />
            <Column field="properties.is_manager" header="Gerente" body={managerTemplate} style={{ width: '20%' }} />
            <Column header="Ações" body={actionTemplate} style={{ width: '20%' }} />
          </DataTable>
        </div>

        <hr style={{ borderColor: '#e0e0e0', margin: '1rem 0', opacity: 0.3 }}/>

        <div className="formgrid grid">
          <div className="field col-12">
            <h3>Convidar Novo Usuário</h3>
            <div className="p-inputgroup">
              <InputText
                placeholder="Digite o email do usuário"
                className="w-full"
                value={newUserEmail.get()}
                onChange={(e) => newUserEmail.set(e.target.value)}
              />
              <Button
                icon="pi pi-user-plus"
                className="p-button-primary"
                tooltip="Convidar Usuário"
                tooltipOptions={{position: 'top'}}
                onClick={inviteUser}
              />
            </div>
          </div>
        </div>
      </Dialog>
    </div>
  );
}