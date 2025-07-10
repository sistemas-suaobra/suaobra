import * as React from "react";
import { isLoaded, User, user, UserProperties, userS } from '../../store/store.js';
import { InputText } from "primereact/inputtext";
import { Password } from 'primereact/password';
import { useVariable } from "../../utils/interfaces.js";
import { Button } from "primereact/button";
import { api, PB } from "../../store/api.js";
import { Toast } from 'primereact/toast';
import { doToast, jsonClone } from "../../utils/methods.js";
import { Checkbox } from 'primereact/checkbox';
import { Dialog } from "primereact/dialog";
import { LogOut } from "../NavBar.js";
import type { PrimeFields } from "../../utils/PrimeForm.js";
import PrimeForm from "../../utils/PrimeForm.js";
import { useHookstate, type State } from "@hookstate/core";

declare global {
  interface Window {
    rudderAnalytics: any
  }
}   

interface Props {}

const authIsValid = () => PB().authStore.isValid && PB().authStore.token != ''
const authIsVerified = () =>  PB().authStore.model?.verified as boolean
const teamIsBlocked = () =>  (PB().authStore.model?.expand?.team_id?.blocked || false) as boolean
const navigateToDashboard = () => window.location.assign('/dashboard')
const navigateToIndex = () => window.location.assign('/')

export function LoginPage(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  ///////////////////////////  HOOKS  ///////////////////////////
  const email = useVariable<string>('')
  const isNewUser = useVariable(false)
  const emailIsInvalid = useVariable<boolean>(false)
  const password = useVariable<string>('')
  const passwordIsInvalid = useVariable<boolean>(false)
  const toast = React.useRef<Toast>(null)

  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    isLoaded.set(true)
    if(authIsValid()) navigateToDashboard()
    
    window.rudderAnalytics?.page({
      userId: user.get().id,
      category: "login",
      name: window.document.title,
    })
  }, [])

  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const authWithPassword = async () => {
    const pb = PB()
    let newUser : User
    let error = undefined

    try {
      let userData = await pb.collection('user').authWithPassword(email.get(), password.get(), { expand: 'team_id'})
      newUser = new User({
        id: userData.record.id,
        email: email.get(),
        legacy_id: userData.record.legacy_id,
        team: userData.record.expand?.team_id,
        token: PB().authStore.token,
        state: userData.record.state,
      })

      if(authIsValid()) {

        if(teamIsBlocked() || !userData.record.expand?.team_id) {
          PB().authStore.clear()
          return doToast(toast, {
            severity: 'info',
            summary: 'Problema com sua conta',
            detail: 'Por favor entre em contato conosco: contato@suaobra.com.br'
          }, 4000)
        }

        emailIsInvalid.set(false)
        passwordIsInvalid.set(false)
        user.set(newUser)

        // rudderstack identify
        window.rudderAnalytics?.identify(user.get().id, {
          user: {
            id: newUser.id,
            email: newUser.email,
            team_id: newUser.team.id,
          }
        })

        navigateToDashboard()
      } else {
        throw new Error('could not login')
      }

    } catch (err) {
      console.log(err)
      error = err

      doToast(toast, {
        severity: 'error',
        summary: 'Erro',
        detail: 'Usuário ou senha incorretos',
      }, 4000)

      passwordIsInvalid.set(false)

    } finally {
      window.rudderAnalytics?.track(
        'login-attempt',
        {
          email: email.get(),
          error,
          auth_is_verified: authIsVerified(),
          team_is_blocked: teamIsBlocked(),
          auth_is_valid: authIsValid(),
        }
      )
    }
  }

  const sendPasswordResetToken = async () => {
    let success = await PB().collection('user').requestPasswordReset(email.get())

    window.rudderAnalytics?.track(
      'login-request-password-reset',
      {
        email: email.get(),
        success,
      }
    )

    if(success)
      doToast(toast, {
        severity: 'info',
        summary: 'Sucesso',
        detail: 'Por favor verifique seu email'
      }, 4000)
    else
      doToast(toast, {
        severity: 'error',
        summary: 'Erro',
        detail: 'E-mail inválido ou algo está errado'
      }, 4000)
  }

  const createUserAndValidate = async () => {
    let user_id = undefined
    let error = undefined
    const data = {
      "email": email.get(),
      "emailVisibility": true,
      "password": password.get(),
      "passwordConfirm": password.get(),
    }

    try {
      let record = await PB().collection('user').create(data)
      user_id = record.id
    } catch (err) {
      error = err
    } finally {
      window.rudderAnalytics?.track(
        'login-create-user',
        { email: email.get(), id: user_id, success: user_id !== undefined, error }
      )
    }

    if(error) {
      doToast(toast, {
        severity: 'error',
        summary: 'Erro',
        detail: 'E-mail inválido ou algo está errado'
      }, 4000)

      return
    }

    // Request email verification
    let success = false
    error = undefined
    try {
      success = await PB().collection('user').requestVerification(email.get());
    } catch (err) {
      error = err
    } finally {
      window.rudderAnalytics?.track(
        'login-request-verification',
        { email: email.get(), success, error }
      )
    }
    
    // login
    authWithPassword()
  }

  ///////////////////////////  JSX  ///////////////////////////
  return (
    <div className="grid h-screen bg-white">
      <Toast ref={toast} />
      <div className="lg:col-6 col-12 h-full flex align-items-stretch flex-wrap align-content-center justify-content-center">
          <div className="formgrid px-7 border-round-sm bg-white font-bold">

              <h2 style={{color: '#F69731', fontSize: 33, fontFamily: 'Myriad Pro', fontWeight: '700', wordWrap: 'break-word'}}>Crie seu perfil grátis</h2>

              <p style={{width: 432, height: 68, color: '#6C70C5', fontSize: 16, fontFamily: 'Myriad Pro', fontWeight: '400', wordWrap: 'break-word'}}>Com um perfil você poderá encontrar as melhores obras residenciais para seu time comercial agir rapidamente e vender mais!</p>

              <div className="field col-12">
                <label htmlFor="email-input">Email</label>
                <InputText
                  id='login-email-input'
                  autoComplete="email"      
                  placeholder="Digite seu email"
                  className={emailIsInvalid.get()? 'w-full p-invalid':'w-full'}
                  value={ email.get() }
                  onChange={(e) => {
                    let val = e.target.value?.toLowerCase()
                    email.set(val)
                    if(val.length >= 4) {
                      if(!val.includes('.')) emailIsInvalid.set(true)
                      else if(!val.includes('@')) emailIsInvalid.set(true)
                      else emailIsInvalid.set(false)
                    }
                  }}
                />
              </div>

              <div className="field col-12">
                <label htmlFor="password-input">Senha</label>
                <Password
                  id='login-password-input'            
                  placeholder="Digite sua senha"
                  autoComplete="current-password"
                  className={passwordIsInvalid.get()? 'w-full p-invalid':'w-full'}
                  inputClassName="w-full"
                  feedback={false}
                  value={ password.get() }
                  onChange={(e) => { password.set(e.target.value) }}
                  toggleMask
                />
                {
                  password.get().length > 0 &&  password.get().length < 5 ?
                  <div className="font-light text-red-600 text-sm mt-1">* Sua senha precisa ter no minimo 5 caracteres</div>
                  :
                  null
                }
              </div>

              <div className="field col-12 m-0">
                <div className="flex align-items-center">
                    <Checkbox
                      inputId="new-client"
                      name="new-client"
                      // value="Cheese"
                      onChange={(e) => isNewUser.set(e.checked)}
                      checked={isNewUser.get()}
                    />
                    <label htmlFor="new-client" className="ml-2">
                    Não tenho uma conta, crie uma para mim
                    </label>
                </div>
              </div>

              <div className="field col-12 m-0">
                <Button
                  label={isNewUser.get() ? 'Criar conta' : "Login"}
                  className={"w-full " + (isNewUser.get() ? 'p-button-success' : "p-button-primary")}
                  onClick={async () => {
                    if(email.get().length < 4) {
                      emailIsInvalid.set(true)
                      return
                    }

                    if(password.get().length < 5) {
                      passwordIsInvalid.set(true)
                      return
                    }

                    if(isNewUser.get()) {
                      createUserAndValidate()
                    } else {
                      await authWithPassword()
                    }
                  }}
                  rounded
                />
              </div>

              <div className="field col-12 m-0">
                <Button
                  label="Esqueci minha senha"
                  className="w-full p-0 m-0 p-button-secondary"
                  disabled={isNewUser.get()}
                  onClick={async () => {
                    if(email.get().length < 4) {
                      emailIsInvalid.set(true)
                      return
                    }
                    await sendPasswordResetToken()
                  }}
                  text
                />
              </div>

              <div className="field col-12 text-center">
                <br/>
                <img id='nav-bar-logo' src="/logo-text-2.svg" alt="" style={{maxHeight: '50px'}}/>
              </div>
              
          </div>
      </div>
      <div className="hidden lg:col-6 col-0 h-full lg:flex flex-wrap align-content-center justify-content-center" style={{backgroundColor: '#322782'}}>
          <div className="text-center">
              <img
                className="w-10"
                src="/login-graphic.svg"
                style={{ maxHeight: '500px'}}/>
          </div>
      </div>
    </div>
  )
};



export function VerifyPanel(props: { lus: State<LoginUserState, {}>}) {

  const visible = useVariable(false)
  const email = useVariable('')
  const message = useVariable({error: false, text: ''})
  const state = useHookstate(props.lus)
  
  // leads load
  React.useEffect(() => {
    email.set(PB().authStore.model.email)
    if(!state.loaded.get()) return
    if(!state.verified.get()) visible.set(true)
  }, [state.loaded]);

  const requestVerification = async () => {

    // Request email verification
    let success = false
    let error = undefined
    try {
      success = await PB().collection('user').requestVerification(email.get());
    } catch (err) {
      error = err
    } finally {
      window?.rudderAnalytics?.track(
        'login-request-verification',
        { email: email.get(), success, error }
      )
    }

    if(success) {
      message.set({error: false, text: `Uma mensagem de verificação foi enviada. Por favor verifique seu email.`})
    } else
      message.set({error: true, text: 'E-mail inválido ou algo está errado. Entre em contato conosco pelo e-mail contato@suaobra.com.br'})
  }

  return <>
    <Dialog
      header='Confirme seu email'
      visible={visible.get()}
      style={{ maxWidth: '500px' }}
      className="w-7"
      onHide={() => {
        // selectedContactInfo.set([])
        // setVisible(false)
      }}
    >
      <div className="px-3">
        <div>Antes de usar esta ferramenta, precisa confirmar seu email.</div>
        <p>Por favor verifique seu email (<i>{email.get()}</i>) e confirme sua conta.</p>
        <p>É possível que o email esteja na caixa de spam.</p>
        <Button
          label="Enviar verificação novamente"
          onClick={() => requestVerification()}
        />
        <Button
          label="Sair"
          className="ml-2 p-button-secondary"
          onClick={() => LogOut()}
        />
        <p style={{color: message.get().error ? 'red' : 'blue' }}>{message.get().text}</p>
      </div>
    </Dialog>
  </>
}


export function InfoInputPanel(props: { lus: State<LoginUserState, {}> }) {

  const info_fields : PrimeFields = {
    name: { label: 'Nome', type: 'string', options: { tooltip: 'Seu nome'}},
    phone: { label: 'Telefone', type: 'mask', options: {mask: '(99) 99999 - 9999', tooltip: 'Seu número de telefone'}},
  }

  const state = useHookstate(props.lus)
  const error = useVariable('')
  const visible = useVariable(false)
  
  // leads load
  React.useEffect(() => {
    if(!state.loaded.get()) return

    if(state.verified.get() && !state.properties.on_boarded.get())
      visible.set(true)

  }, [state.loaded]);

  const saveUser = async () => {
    // clean up phone
    state.properties.phone.set(v => v.replace(/[^0-9]/g, ''))

    if(state.properties.name.get().length < 5) return error.set('Por favor insira um nome válido')
    if(state.properties.phone.get().length < 10) return error.set('Por favor insira um telefone válido')
    
    state.properties.on_boarded.set(true)

    let resp = await api().collection('user').update(
      state.id.get(), 
      {properties: jsonClone(state.properties.get())}, 
    )
    error.set(resp.error)
    if(resp.error) return false
    return true
  }

  return <>
    <Dialog
      header='Informação básica'
      visible={visible.get()}
      style={{ maxWidth: '400px' }}
      className="w-7"
      onHide={() => {
        // selectedContactInfo.set([])
        // setVisible(false)
      }}
    >
      <div className="px-3">
        <PrimeForm
          fields={info_fields}
          getter={(key:string) => state.properties[key]?.get()}
          setter={(key:string, value: any) => {
            state.properties[key].set(value)
          }}
          defaults={{size: 12}}
          buttons={() => {
            return <>
              <div className='field col-12'>
                {/* <Button
                  label='Cancelar'
                  className="p-button-warning mr-2"
                  onClick={() => { show.set(false) }}
                /> */}
                <Button
                  label='Salvar'
                  className="w-full"
                  onClick={async () => { 
                    let success = await saveUser() 
                    if(success)
                      visible.set(false)
                  }}
                />
              </div>
            </>
          }}
        />
        <p style={{color: 'red'}}> { error.get() }</p>
      </div>
    </Dialog>
  </>
}

export function VerifyEmail(props: {}) {

  React.useEffect(() => {

    let token = window.location.hash.substr(1)
    console.log('token = ' + token)
    
    confirm(token)

  }, []);

  const confirm = async (token: string) => {

    let success = false
    let error = undefined

    try {
      success = await PB().collection('user').confirmVerification(token)
    } catch (err) {
      error = err
    } finally {
      window.rudderAnalytics?.track(
        'login-confirm-verification',
        { email: PB().authStore.model.email, success, error }
      )
    }

    // back to index
    navigateToIndex()
  }

  return <></>

}

export interface LoginUserState {
   id: string
   loaded: boolean
   verified: boolean
   is_manager: boolean
   properties: UserProperties
}