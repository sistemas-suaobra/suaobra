import { atom } from 'nanostores';
import { allCities, type City } from './cities';
import { PB } from './api';
import type { ObjectAny } from '../utils/interfaces';
import { TryParseJSON, jsonClone, md5, uuidv4 } from '../utils/methods';
import { hookstate, type State } from '@hookstate/core';

export const isLoaded = atom(false);
export const isWaiting = atom(false);
export const smallBool = atom(false);
export const obrasPlusCity = atom<City>();
export const obrasPlusNeighborhood = atom<string[]>();
export const obrasPlusState = atom<string>('SP');
export const $vendaMaisPageRefresh = atom(0);
export const $vendaMaisStageHeight = atom(500);

export const user = atom<User>({} as User);
export const userS = hookstate<User>({} as User);

declare global {
  interface Window {
    token: string;
  }
}

export const loadUser = () => {
  const pb = PB()
  if (!pb.authStore.isValid || !pb.authStore.token) return false


  let userData = pb.authStore.model

  let u = new User({
    id: userData.id,
    email: userData.email,
    legacy_id: userData.legacy_id,
    team: userData.expand?.team_id,
    token: pb.authStore.token,
  })
  user.set(u)
  
  if(!userS?.id?.get()) userS.set(u)

  console.log(`loaded user ${userData.id}`)
  return true
}

export const loadUserState = async () => {
  const pb = PB()
  if (!pb.authStore.isValid || !pb.authStore.token) return

  let userData = pb.authStore.model
  // load user state
  let data = await PB().collection('user').getOne(userData.id, { expand: 'team_id' })

  userS.set(u => {
    data.team = data.expand.team_id
    u = new User(data)
    if (u.team.cities.includes('SP-*')) {
      u.team.cities = allCities.filter(c => c.id.startsWith('SP-')).map(c => c.id).sort()
    }
    if (u.team.cities.includes('GO-*')) {
      u.team.cities = allCities.filter(c => c.id.startsWith('GO-')).map(c => c.id).sort()
    }
    if (u.team.cities.includes('SC-*')) {
      u.team.cities = allCities.filter(c => c.id.startsWith('SC-')).map(c => c.id).sort()
    }
    if (u.team.cities.includes('DF-*')) {
      u.team.cities = allCities.filter(c => c.id.startsWith('DF-')).map(c => c.id).sort()
    }
    return u
  })

  return new User(jsonClone(userS.get()))
}

export class User {
  id?: string;
  legacy_id?: string;
  email?: string;
  manager?: boolean;
  team?: Team;
  state?: UserState;
  token?: string;
  properties?: UserProperties;
  verified: boolean
  expand: any

  loaded: boolean

  constructor(data : ObjectAny = {}) {
    this.id = data.id
    this.legacy_id = data.legacy_id
    this.email = data.email
    this.manager = data.manager || false
    this.team = new Team(data.team)
    this.state = data.state || {}
    this.token = data.token
    this.verified = data.verified
    this.properties = new UserProperties(data.properties || {})

    this.loaded = !!this.id
  }

  async saveState() {
    await PB().collection('user').update(this.id, {state: this.state}, { expand: 'team_id'})
  }

  get is_manager() {
    return this.manager
  }
}

export interface UserWhatsappProperties {
  on_boarded: boolean
  connected: boolean
}

export class UserProperties {
  name: string
  phone: string
  is_admin: boolean
  on_boarded: boolean
  whatsapp: UserWhatsappProperties
  constructor(data : ObjectAny = {}){
    if(data === null || (data as any) == 'null' || !data) data = {}

    // If data is a string, parse it into an object
    if (typeof data === 'string') {
      try {
        data = JSON.parse(data);
      } catch (e) {
        console.error('Failed to parse UserProperties data:', e);
        data = {};
      }
    }

    this.name = data.name || ''
    this.phone = data.phone || ''
    this.is_admin = data.is_admin || false
    this.on_boarded = data.on_boarded || false
    this.whatsapp = data.whatsapp || { on_boarded: false, connected: false }
  }
}

export class Team {
  id: string;
  name?: string;
  active?: boolean;
  blocked?: boolean;
  cities?: string[];
  state?: TeamState;
  properties: TeamProperties;
  entitlements?: Entitlement;
  export: number;

  constructor(data : ObjectAny = {}) {
    this.id = data.id
    this.name = data.name
    this.active = data.active
    this.blocked = data.blocked
    this.cities = data.cities
    this.export = data.export
    this.state = data.state || {}
    this.properties = new TeamProperties(TryParseJSON(data.properties))
    this.entitlements = data.entitlements
  }

  async saveState() {
    await PB().collection('team').update(this.id, {state: this.state})
  }

  get allow_export() {
    return this.export || 0
  }

  get allow_contact() {
    return this.entitlements?.allow_contact || 0
  }
}

export class TeamProperties {
  name: string
  description: string
  founded_date: string
  type: string // PF or PJ
  cpf: string
  cnpj: string
  telephone: string
  whatsapp: string
  address: Address
  templates: Templates
  website: string
  maps_url: string
  industry: string
  keywords: string
  lead_introduction_text: string
  constructor(data : ObjectAny = {}){
    this.name = data.name
    this.description = data.description
    this.founded_date = data.founded_date
    this.type = data.type
    this.cpf = data.cpf
    this.cnpj = data.cnpj
    this.telephone = data.telephone
    this.whatsapp = data.whatsapp
    this.address = data.address || {}
    this.templates = data.templates || {}
    this.website = data.website
    this.maps_url = data.maps_url
    this.industry = data.industry
    this.keywords = data.keywords
    this.lead_introduction_text = data.lead_introduction_text || ''
  }
}

export interface Address {
  enderco: string; // street name
  numero: string; // number on street
  complemento: string; // house or apartment number
  bairro: string; // neighborhood
  cidade: string; // cityi
  uf: string; // state
  cep: string; // postal code
}

export interface Templates {
  sender: string
  context: string
  owner1: string
  owner2: string
  // owner3: string
  professional1: string
  professional2: string
  // professional3: string
}

export const brasilStates = ['AC','AL','AM','AP','BA','CE','DF','ES','GO','MA','MG','MS','MT','PA','PB','PE','PI','PR','RJ','RN','RO','RR','RS','SC','SE','SP','TO']

export interface Entitlement {
  allow_contact?: number;
}

export interface UserState {
  selected_list: string
  venda_mais_mode: 'kanban' | 'table'
}

export interface TeamState {
  selected_list: string
  info_obras_password: string
}

export class Stage {
  id: string;
  list_id: string;
  name: string;
  order: number;
  expand: { list_id: List }
  leads: Lead[]
  properties: StageProperties;

  constructor(data : ObjectAny = {}){
    this.id = data.id
    this.list_id = data.list_id
    this.name = data.name
    this.order = data.order
    this.expand = data.expand
    this.leads = data.leads || []
    this.leads = this.leads.map(l => new Lead(l))
    this.properties = new StageProperties(TryParseJSON(data.properties))
  }
}

export class StageProperties {
  order_by: string
  constructor(data : ObjectAny = {}){
    this.order_by = data.order_by || 'rank1-desc' || 'list_lead_updated-desc'
  }
}

export interface List {
  id: string;
  name?: string;
}

export class Lead {
  lead_id: string;
  list_lead_id: string;
  stage_id: string;
  obra_id: string;
  address: string;
  bairro: string;
  owner_email: string;
  owner_name: string;
  city: string;
  state: string;
  owner: string;
  professional: string;
  size: number;
  start_date: number;
  end_date: number;
  favorited_at: string;
  lead_updated: string;
  list_lead_updated: string;
  lead_properties: LeadProperties;

  constructor(data : ObjectAny = {}){
    this.lead_id = data.lead_id
    this.list_lead_id = data.list_lead_id
    this.stage_id = data.stage_id
    this.obra_id = data.obra_id
    this.address = data.address
    this.bairro = data.bairro
    this.owner_email = data.owner_email
    this.owner_name = data.owner_name
    this.city = data.city
    this.state = data.state
    this.owner = data.owner
    this.professional = data.professional
    this.size = data.size
    this.start_date = data.start_date
    this.end_date = data.end_date
    this.favorited_at = data.favorited_at
    this.lead_updated = data.lead_updated
    this.list_lead_updated = data.list_lead_updated

    let properties = TryParseJSON(data.lead_properties)
    this.lead_properties = new LeadProperties(properties)

    this.sync_obra_properties()
  }

  sync_obra_properties() {
    if(this.is_obra) {
      this.lead_properties.obra.owner = this.lead_properties.obra.owner || this.owner
      this.lead_properties.obra.professional = this.lead_properties.obra.professional || this.professional
      this.lead_properties.obra.address = this.lead_properties.obra.address || this.address_short
      this.lead_properties.obra.bairro = this.lead_properties.obra.bairro || this.bairro
      this.lead_properties.obra.city = this.lead_properties.obra.city || this.city
      this.lead_properties.obra.state = this.lead_properties.obra.state || this.state
      this.lead_properties.obra.size = this.lead_properties.obra.size || this.size
    }

    if(this.is_opportunity) {
      this.address = this.lead_properties.obra.address
      this.bairro = this.lead_properties.obra.bairro
      this.city = this.lead_properties.obra.city
      this.state = this.lead_properties.obra.state
      this.owner = this.lead_properties.obra.owner
      this.professional = this.lead_properties.obra.professional
      this.size = this.lead_properties.obra.size
      this.start_date = this.lead_properties.obra.start_date
      this.end_date = this.lead_properties.obra.end_date
    }
  }

  get owner_str() {
    return this.owner_email
    return this.owner_name?.trim() || this.owner_email
  }

  get is_obra() {
    // Obras do Obras+ usam qualquer obra_id que nao seja oportunidade manual (oppo_*).
    // Importacoes CAU geram ids sem prefixo fixo; o prefixo obra_ nao e obrigatorio.
    return !!this.obra_id && !this.is_opportunity
  }

  get is_opportunity() {
    return this.obra_id?.startsWith('oppo_') || false
  }

  get title() {
    // return `${this.bairro} - ${this.city}, ${this.state}`
    let owner = this.owner || this.lead_properties.obra.owner
    let city = this.city || this.lead_properties.obra.city
    let state = this.state || this.lead_properties.obra.state
    if(!owner && !city && !state) return 'Indefinido'
    return `${owner || ''} - ${city || ''}, ${state || ''}`
  }

  get valor() {
    if(this.lead_properties?.valor) return `R$ ${this.lead_properties.valor.toLocaleString('pt-BR')}`
    return '-'
  }

  get cidade() {
    return `${this.city}, ${this.state}`
  }

  get address_short() {
    let suffix = `${this.bairro} - ${this.city}, ${this.state}`
    return this.address?.replaceAll(', ' + suffix, '').trim()
  }

  get start_date_str() {
    if(this.start_date) return (new Date(this.start_date*1000)).toLocaleDateString('pt-BR', { timeZone: 'UTC' })
    return '-'
  }

  get end_date_str() {
    if(this.end_date) return (new Date(this.end_date*1000)).toLocaleDateString('pt-BR', { timeZone: 'UTC' })
    return '-'
  }

  get size_str() {
    return `${this.size || '-'} m²`
  }

  get obra_properties_payload() {
    let data = {} as ObraProperties
    
    if(this.lead_properties.obra.address)
      data.address = this.lead_properties.obra.address
    
    if(this.lead_properties.obra.bairro)
      data.bairro = this.lead_properties.obra.bairro
    
    if(this.lead_properties.obra.city)
      data.city = this.lead_properties.obra.city
    
    if(this.lead_properties.obra.state)
      data.state = this.lead_properties.obra.state
    
    if(this.lead_properties.obra.owner)
      data.owner = this.lead_properties.obra.owner
    
    if(this.lead_properties.obra.professional)
      data.professional = this.lead_properties.obra.professional

    if(this.lead_properties.obra.size)
      data.size = this.lead_properties.obra.size

    if(this.is_obra) {
      if(this.lead_properties.obra.owner === this.owner)
        delete data.owner
      if(this.lead_properties.obra.professional === this.professional)
        delete data.professional
    }
    
    return data
  }
}

export class LeadProperties {
  rating: number
  observations: string
  valor: number
  starred_contacts: string[]
  contacts: Contact[]
  obra: ObraProperties
  alert_at: number
  alerted: boolean
  constructor(data : ObjectAny = {}){
    this.rating = data.rating || 0
    this.valor = data.valor || 0
    this.observations = data.observations || ''
    this.alert_at = data.alert_at
    this.alerted = data.alerted
    this.starred_contacts = data.starred_contacts || []
    this.contacts = ((data.contacts || []) as any[]).map(c => new Contact(c))
    this.obra = new ObraProperties(data.obra || {})
  }

  get alert_date() {
    if(!this.alert_at) return undefined
    return new Date(this.alert_at)
  }
  get alert_date_str() {
    let alert_date = this.alert_date
    if(!alert_date) return '-'
    return alert_date.toLocaleString('pt-BR')
  }

  get has_alert() {
    let alert_date = this.alert_date
    if(!alert_date) return false
    return true
  }

  get alert_reached() {
    let alert_date = this.alert_date
    if(!alert_date) return false
    return alert_date.getTime() < (new Date()).getTime()
  }
}

export class ObraProperties {
  obra_id: string;
  address: string;
  telefone: string;
  bairro: string;
  city: string;
  state: string;
  owner: string;
  professional: string;
  size: number;
  start_date: number;
  end_date: number;

  constructor(data : ObjectAny = {}){
    this.obra_id = data.obra_id
    this.address = data.address || ''
    this.telefone = data.telefone || ''
    this.bairro = data.bairro || ''
    this.city = data.city || ''
    this.state = data.state || ''
    this.owner = data.owner || ''
    this.professional = data.professional || ''
    this.size = data.size
  }
}

export class Contact {
  name: string
  telephone?: string
  email?: string
  city?: string
  state?: string
  contact_id?: string
  person_id?: string;
  company_id?: string;

  constructor(data : ObjectAny = {}){
    this.name = data.name || ''
    this.telephone = data.telephone
    this.email = data.email
    this.person_id = data.person_id
    this.company_id = data.company_id
    this.city = data.city || ''
    this.state = data.state || ''
    
    if(this.email)
      this.email = this.email.toLowerCase()

    if(this.telephone)
      this.telephone = `${this.telephone}`.replaceAll(/\D/g, '')
    
    this.contact_id = data.contact_id || this.make_contact_id()
  }

  make_contact_id() {
    let hash = uuidv4().replaceAll('-', '')
    let type = this.telephone ? 'phone' : this.email ? 'email' : ''
    return `cont_${type}_${hash}`
  }

  get as_contact_info() {
    return {
      nome: this.name,
      telephone: this.telephone,
      email: this.email,
      city: this.city,
      state: this.state,
      contact_id: this.contact_id,
    }
  }
}