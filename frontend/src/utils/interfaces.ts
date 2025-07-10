import * as React from "react";

export interface ObjectString { [key: string]: string; }; 
export interface ObjectBoolean { [key: string]: boolean; }; 
export interface ObjectNumber { [key: string]: number; }; 
export interface ObjectAny { [key: string]: any; }; 
export interface RecordsData { headers: string[], rows: any[] }
export interface Variable<S> {
  get: () => S;
  set: (val: S | ((prevState: S) => S)) => void;
  put: (doPut: ((prevState: S) => void)) => void;
}

export function useVariable<S>(initialState: S | (() => S)): Variable<S> {
  const [value, setValue] = React.useState<S>(initialState)
  let putValue = (doPut: ((prevState: S) => void)) => {
    setValue(
      S => {
        doPut(S)
        return S
      }
    )
  }

  return {
    get: () => value,
    set: setValue,
    put: putValue,
  }
}