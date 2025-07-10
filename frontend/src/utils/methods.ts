
import type { Toast, ToastMessage } from "primereact/toast";
import { type ObjectAny, type ObjectString } from "./interfaces";
import { md5 as MD5 } from 'js-md5';

export const serialize = function(obj: ObjectString) : string {
  var str = [];
  for (var p in obj)
    if (obj.hasOwnProperty(p)) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
    }
  return str.join("&");
}

export function IsValid(obj: any) {
  return Object.keys(obj).length !== 0
}

export function jsonClone<T = any>(val: any) { 
  if (IsValid(val)) {
    return JSON.parse(JSON.stringify(val)) as T
  }
  return {} as T
}

export const TryParseJSON = (data: any) => {
    if(typeof data === 'string' || data instanceof String) data = JSON.parse(data.toString())
    if(!data) data = {}
    return data
}


// parseFilterString returns the filter string split into tokens
// where words inside quotes represent one token
// filterStr = 'hello dear brother' => ['hello', 'dear', 'brother']
// filterStr = 'hello "dear brother"' => ['hello', '"dear brother"']
export let parseFilterString = (filterStr: string) => { 
  let filters : string[] = [];
  let inQuote = false;
  let token = ""
  for (let i = 0; i < filterStr.length; i++) {
    const char = filterStr[i];
    if (char === '"') inQuote = !inQuote 
    token = token + char
    if ((!inQuote && char === ' ') || i === filterStr.length - 1) {
      token = token.trim()
      if(token.length > 0) filters.push(token)
      token = ''
    }
  }
  return filters
}

export const filterAndMatched = (row: any[], filters: string[]) => { 
  filters = filters.filter(v => v.trim() !== '')
  if(filters.length === 0) return true
  let include = filters.map(v => false)
  for(let val of row) {
    for (let i = 0; i < filters.length; i++) {
      const filter = filters[i].toLowerCase().trim()
      if (filter.startsWith('"') && filter.endsWith('"')) {
        if (`${val}`.toLowerCase() === filter.replaceAll('"', '')) include[i] = true
      } else {
        if (`${val}`.toLowerCase().includes(filter)) include[i] = true
      }
    }
    if(include.every(v => v === true)) { return true }
  }
  return include.every(v => v === true)
} 


export const doToast = (toast: React.MutableRefObject<Toast>, msg: ToastMessage, duration=3000) => {
  if(toast === null || toast.current === null) return console.error('toast is null')
  msg.life = duration;
  (toast.current as Toast).show(msg);
}

export function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, (c: any) =>
    (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
  );
}

export const md5 = (text: string) : string => {
  return MD5(text)
}