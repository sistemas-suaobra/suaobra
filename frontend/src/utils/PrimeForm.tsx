import * as React from "react";
import { InputText } from "primereact/inputtext";
import { InputTextarea } from 'primereact/inputtextarea';
import { Calendar } from 'primereact/calendar';
import { useVariable, type ObjectAny } from "./interfaces";
import { InputMask } from 'primereact/inputmask';
import { InputNumber } from 'primereact/inputnumber'
import { Dropdown } from 'primereact/dropdown';
import { Chips } from 'primereact/chips'
import { useHookstate, type State } from "@hookstate/core";
import type { CSSProperties } from "react";
import { AutoComplete } from 'primereact/autocomplete';

export interface PrimeFields { [key: string]: PrimeField; }; 

export interface PrimeField {
  label?: string
  size?: number
  type?: 'string' | 'mask' | 'text-area' | 'date' | 'timestamp' | 'boolean' | 'number' | 'dropdown' | 'chips' | 'value'
  placeholder?: string
  options?: ObjectAny;
  className?: string;
}

export interface PrimeFormProps {
  fields: { [key: string]: PrimeField}
  defaults?: PrimeField
  getter: (key:string) => any;
  setter: (key:string, value: any) => void;
  buttons?: () => JSX.Element
}

export default function PrimeForm(props: PrimeFormProps) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const fields = props.fields
  const defaults = props.defaults || {} as PrimeField
  const set = props.setter
  const get = props.getter

  ///////////////////////////  HOOKS  ///////////////////////////
  ///////////////////////////  EFFECTS  ///////////////////////////
  ///////////////////////////  FUNCTIONS  ///////////////////////////
  ///////////////////////////  JSX  ///////////////////////////

  const Field = (key: string, field: PrimeField) => {
    const id = 'input-' + key
    const type = field.type || defaults.type || 'string'
    const placeholder = field.placeholder || defaults.placeholder || field.label
    const options = {
      ...defaults,
      ...(field.options || {} as ObjectAny), // this overwrites
    } || ({} as ObjectAny)

    if(type === 'text-area')
      return <InputTextarea
        id={id}
        className="w-full"
        placeholder={placeholder}
        value={ get(key) as string }
        onChange={(e) => set(key, e.target.value)}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'date')
      return <Calendar
        id={id}
        className="w-full"
        placeholder={placeholder}
        value={ get(key) as string }
        onChange={(e) => set(key, e.target.value)}
        dateFormat={options.dateFormat}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'timestamp')
      return <Calendar
        id={id}
        className="w-full"
        placeholder={placeholder}
        value={ get(key) as string }
        onChange={(e) => set(key, e.target.value)}
        dateFormat={options.dateFormat}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'number')
      return <InputNumber
        id={id}
        className="w-full"
        placeholder={placeholder}
        showButtons={options.showButtons}
        step={options.step}
        value={ get(key) as number }
        onChange={(e) => set(key, e.value)}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'chips')
      return <Chips
        id={id}
        className="w-full"
        value={ (get(key) || []) as string[] }
        onChange={(e) => set(key, e.value)}
        separator=","
        addOnBlur={true}
        allowDuplicate={false}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'mask')
      return <InputMask
        id={id}
        className="w-full"
        autoClear={options.autoClear || false}
        placeholder={placeholder}
        value={ get(key) as string }
        onChange={(e) => set(key, e.target.value)}
        mask={options.mask}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'dropdown')
      return <Dropdown
        id={id}
        className="w-full"
        placeholder={placeholder}
        value={ get(key) as string }
        onChange={(e) => set(key, e.value)}
        options={options.options}
        optionLabel={options.optionLabel}
        optionValue={options.optionValue}
        filter={options.filter}
        tooltip={options.tooltip}
        tooltipOptions={options.tooltipOptions}
      />

    if(type === 'value')
      return <div className={options.className} style={options.style}> { get(key) } </div>

    return <InputText
      id={id}
      className="w-full"
      placeholder={placeholder}
      value={ get(key) as string }
      onChange={(e) => set(key, e.target.value)}
      tooltip={options.tooltip}
      tooltipOptions={options.tooltipOptions}
    />
  }


  return (
    <div className="formgrid grid">
      {
        Object.keys(fields).map(
          key => {
            let field = fields[key]
            let size = field.size || defaults?.size || 6
            let className = field.className || defaults?.className || ''

            return <div
              key={key}
              className={`field md:col-${size} col-12 ` + className}>
              <label htmlFor={key}>{field.label || ''}</label>
              { Field(key, field) }
            </div>
          }
        )
      }
      { props.buttons() }
    </div>
  );
};



export interface EditableTextProps {
  value: State<string>
  autoComplete?: string[]
  editable?: boolean
  size?: number
  className?: string
  style?: CSSProperties
  transform?: (input: string) => string;
}

export function EditableText(props: EditableTextProps) {

  const editable = props.editable === undefined ? true : props.editable
  const autoComplete = props.autoComplete || []
  const value = useHookstate(props.value)
  const editing = useHookstate(false)
  const origText = useVariable(props.value.get())
  const ref = React.useRef(null)

  React.useEffect(() => {
    if(editing.get()) ref.current?.focus() // focus on editing
    else {
      if(props.transform !== undefined) {
        props.value.set(props.transform(props.value.get()))
      }
      origText.set(props.value.get())
    }
  }, [editing.get()]);


  const search = (event) => {
    let val = (event.query as string).toUpperCase()
    return autoComplete.filter(item => item.includes(val));
  }

  return <span className={props.className}>
    
    {
      editable && editing.get() ?
        autoComplete.length > 0 ?
          // <AutoComplete
          //   id="owner-input"
          //   ref={ref}
          //   className="ml-2 lead-form-input"
          //   size={props.size || 20}
          //   style={props.style}
          //   value={ value.get() }
          //   completeMethod={search}
          //   suggestions={autoComplete}
          //   dropdown={true}
          //   // forceSelection
          //   onChange={async (e) => {
          //     value.set(e.value)
          //   }}
          // />
          <span style={{width: props.size || 200}}>
            <Dropdown
              className="editable-dropdown"
              style={{padding: 0, paddingLeft: 3, paddingRight: 3}}
              value={ value.get()}
              onChange={(e) => value.set(e.value)}
              options={autoComplete}
              filter
              // optionLabel="name" 
              // optionValue="id"
            />
          </span>
          :
          <InputText
            id="owner-input"
            ref={ref}
            className="ml-2 lead-form-input"
            size={props.size || 40}
            style={props.style}
            value={ value.get() }
            onChange={async (e) => {
              value.set(e.target.value)
            }}
            onKeyDown={(e) => {
              if(e.key === 'Enter') { editing.set(v => v ? false : true) } 
            }}
          />
      :
      <span> { value.get() }</span>
    }
    {
      editable ?
      <i
        className="pi pi-pencil ml-1 cursor-pointer"
        style={{ color: 'grey', fontSize: '0.7rem' }}
        onClick={() => editing.set(v => v ? false : true)}
      />
      :
      null
    }

    {
      editing.get() ?
        <i
          className="pi pi-times ml-1 cursor-pointer"
          style={{ color: 'orange', fontSize: '0.7rem' }}
          onClick={() => {
            value.set(origText.get())
            editing.set(v => v ? false : true)
          }}
        />
      :
      null
    }
  </span>
}