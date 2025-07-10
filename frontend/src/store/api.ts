import { LogOut } from "../components/NavBar";
import type { ObjectAny, ObjectString } from "../utils/interfaces"
import { serialize } from "../utils/methods"
import { allCities } from "./cities"
import { isWaiting, loadUser, user, type User } from "./store"
import PocketBase, { RecordService, type RecordModel, type ListResult, type CommonOptions, type RecordListOptions, type RecordOptions, type RecordFullListOptions } from 'pocketbase';

export const isProd = () => window.location.hostname.includes('suaobra.com.br')
export const isStage = () => window.location.hostname.includes('suaobra.ocral.app.br')
export const isDev = () => !(isStage() || isProd())

export const baseURL =  () => isProd() ? 'https://api.suaobra.com.br' :
                              isStage() ? 'https://api.suaobra.ocral.app.br' :
                              window.location.hostname.includes('suaobra.test') ? `https://api.suaobra.test` :
                              `http://${window.location.hostname}:8090`

export const PB = () => new PocketBase(baseURL())


export const makeURL = (route: string) => `${baseURL()}${route}`

const loadUserOrLogout = async () => {
  if(window.location.href?.includes('obras-plus-iframe')) return // no need to login
  if(! (await loadUser())) LogOut()
}

export const api = () => {
  return {

    collection: (collection: string) => {
      return new PocketBaseCollection(collection)
    },

    get: async (url: string, data: ObjectString, loggedIn = true) => {
      let response = new Response()
      try {
        isWaiting.set(true)
        if(loggedIn && !user.get().token) await loadUserOrLogout()
        let resp = await fetch(`${url}?${serialize(data)}`, {
          method: 'GET',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Authorization': user.get().token,
          },
        })
        response = new Response(resp)
        await response.init()
        
      } catch (error) {
        response.setError(error) 
      } finally {
        isWaiting.set(false)
      }
      return response
    },

    post: async (url: string, data: ObjectAny, loggedIn = true) => {
      let response = new Response()
      try {
        isWaiting.set(true)
        if(loggedIn && !user.get().token) await loadUserOrLogout()
        let resp = await fetch(url, {
          method: 'POST',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Authorization': user.get().token,
          },
          body: JSON.stringify(data),
        })
        response = new Response(resp)
        await response.init()
        
      } catch (error) {
        response.setError(error) 
      } finally {
        isWaiting.set(false)
      }
      return response
    },

    patch: async (url: string, data: ObjectString, loggedIn = true) => {
      let response = new Response()
      try {
        isWaiting.set(true)
        if(loggedIn && !user.get().token) await loadUserOrLogout()
        let resp = await fetch(url, {
          method: 'PATCH',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Authorization': user.get().token,
          },
          body: JSON.stringify(data),
        })
        response = new Response(resp)
        await response.init()
        
      } catch (error) {
        response.setError(error) 
      } finally {
        isWaiting.set(false)
      }
      return response
    },

    put: async (url: string, data: ObjectString, loggedIn = true) => {
      let response = new Response()
      try {
        isWaiting.set(true)
        if(loggedIn && !user.get().token) await loadUserOrLogout()
        let resp = await fetch(url, {
          method: 'PUT',
          headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            'Authorization': user.get().token,
          },
          body: JSON.stringify(data),
        })
        response = new Response(resp)
        await response.init()
        
      } catch (error) {
        response.setError(error) 
      } finally {
        isWaiting.set(false)
      }
      return response
    },

  }
}

export class Response {
  response: globalThis.Response
  data: any
  error: string
  _json_read: boolean
  _json_payload: any
  _pb_data: any

  constructor(response?: globalThis.Response){
    this.response = response
  }

  get type() {
    return this.response?.headers.get('content-type') || ''
  }

  get status() {
    return this.response?.status
  }

  async init() {
    this._json_read = false
    if(this.response?.status >= 400) {
      let data = await this.json()
      if(data?.error) this.setError(data?.error)
      else if(data?.message) this.setError(data?.message)
      else this.setError(this.response.statusText)
    }
  }

  setError(error: string) {
    console.error(error)
    this.error = error.toString()
  }

  async json() {
    if(this._json_payload) return this._json_payload

    try {
      if(this._json_read === false) {
        this._json_read = true
        this.data = await this.response?.json()
      }
    } catch (error) {
      this.setError(error)
      return undefined
    } finally {
    }
    return this.data
  }

  async record() {
    let recs = await this.records()
    if(recs.length > 0) return recs[0]
    if(this.data) return this.data
    return {}
  }
  
  async records() {
    let recs :  ObjectAny[] = []

    if(this._json_payload) {
      if(Array.isArray(this._json_payload)) recs = this._json_payload
      else recs = [this._json_payload]
      return recs
    }

    try {
      recs = await this.json()
      if(!recs) return []
    } catch (error) {
      this.setError(error)
    }

    return recs
  }

  setPbRecordModel(data: RecordModel) {
    this._pb_data = data
    this._json_payload = data
  }

  setPbRecordModels(data: RecordModel[]) {
    this._pb_data = data
    this._json_payload = data
  }

  setPbListResult(data: ListResult<RecordModel>) {
    this._pb_data = data
    this._json_payload = data.items
  }

}

export class PocketBaseCollection {
  pb: PocketBase
  api: RecordService

  constructor(collection: string){
    this.pb = PB()
    this.api = this.pb.collection(collection)
  }

  async getFullList(options?: RecordFullListOptions) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let data = await this.api.getFullList(options)
      resp.setPbRecordModels(data)
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
  async update(id: string, bodyParams?: { [key: string]: any;} | FormData, options?: RecordOptions) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let data = await this.api.update(id, bodyParams, options)
      resp.setPbRecordModel(data)
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
  async getFirstListItem(filter: string, options?: RecordListOptions) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let data = await this.api.getFirstListItem(filter, options)
      resp.setPbRecordModel(data)
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
  async getList(page?: number, perPage?: number, options?: RecordListOptions) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let data = await this.api.getList(page, perPage, options)
      resp.setPbListResult(data)
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
  async delete(id: string, options?: CommonOptions) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let success = await this.api.delete(id, options)
      resp._json_payload = {success: success}
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
  async create(bodyParams?: { [key: string]: any;} | FormData) {
    let resp = new Response()
    try {
      isWaiting.set(true)
      let data = await this.api.create(bodyParams)
      resp.setPbRecordModel(data)
    } catch (error) {
      resp.setError(error)
    } finally {
      isWaiting.set(false)
    }
    return resp
  }
  
}

export const userTrackProps = () => { return { id: user.get().id, legacy_id: user.get().legacy_id, email: user.get().email, team_id: user.get().team.id } }