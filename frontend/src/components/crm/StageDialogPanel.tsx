import * as React from "react";
import type { Stage } from "../../store/store";
import { type State, useHookstate } from "@hookstate/core";
import { Button } from "primereact/button";
import { Dialog } from "primereact/dialog";
import { InputText } from "primereact/inputtext";
import { SelectButton } from "primereact/selectbutton";
import { api } from "../../store/api";


export interface StageDialogParams {
  show?: boolean;
  stage: Stage;
}

export function StageDialogPanel(props: { state: State<StageDialogParams> }) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  ///////////////////////////  HOOKS  ///////////////////////////
  const settings = props.state
  const show = useHookstate(props.state.show)
  const stage = useHookstate(props.state.stage)
  const orderModified = useHookstate(false)

  ///////////////////////////  EFFECTS  ///////////////////////////
  ///////////////////////////  FUNCTIONS  ///////////////////////////
  ///////////////////////////  JSX  ///////////////////////////
  return (
    <div>
        <Dialog
          header={stage.name.get()}
          onHide={() => show.set(false)}
          visible={show.get()}
          style={{maxWidth: '300px'}}
        >
          <div className="formgrid grid">
            <div className="field col-12">
              <label htmlFor="stage-name">Nome</label>
              <InputText
                id='stage-name'
                value={stage.name.get()}
                onChange={(e) => {
                  stage.name.set(e.target.value)
                }}
                placeholder="Digite o Nome"
                className="w-full"
              />
            </div>

            <div className="field col-12 hidden">
              <label htmlFor="stage-order">Ordem</label>
              <SelectButton
                id='stage-order'
                value={stage.order.get()}
                onChange={(e) => {
                  if(e.value === stage.order.get()) return
                  stage.order.set(e.value)
                  orderModified.set(true)
                }}
                options={[1,2,3]}
                className='w-full'
              />
            </div>

            <div className="field col-12">
              <Button
                label="Salvar"
                className="p-button-success w-full"
                onClick={async () => { 
                  let data = { name: stage.name.get()}
                  if(orderModified.get()) {
                    // TODO:
                    let resp = await api().collection('list_stage').getList(1, 50, {
                      filter: `list_id = '${stage.list_id.get()}'`,
                      skipTotal: true,
                    })
                    if(resp.error) return           
                  }
                  let resp = await api().collection('list_stage').update(stage.id.get(), data)
                  if(resp.error) return
                  show.set(false)
                }}
              />
            </div>

          </div>
        </Dialog>
    </div>
  )
}