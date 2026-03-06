import React from "react";
import { Button } from "primereact/button";
import { InputSwitch } from "primereact/inputswitch";
import type { Intent } from "./IntentDialog";
import type { AgenteIaTabProps } from "../types/agenteiatab";

export default function AgenteIaTab(props: AgenteIaTabProps) {
  const { intents, onNew, onEdit, onDelete, onToggleActive } = props;

  return (
    <div className="w-full">
      <div className="flex align-items-center justify-content-between mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Intenções Configuradas</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Defina as intenções que o agente deve identificar
          </div>
        </div>

        <Button icon="pi pi-plus" label="Nova Intenção" onClick={onNew} />
      </div>

      <div className="flex flex-column gap-3">
        {intents.map((intent) => (
          <div key={intent.id} className="bg-white border-round-2xl p-4 border-1 surface-border">
            <div className="flex align-items-center justify-content-between">
              <div className="flex align-items-center gap-2">
                <i className="pi pi-robot" style={{ color: "#2563EB" }} />
                <div style={{ fontWeight: 700, fontSize: 16 }}>{intent.titulo}</div>

                <span
                  className="border-round-xl px-2 py-1"
                  style={{
                    background: intent.ativo ? "rgba(37,99,235,0.12)" : "rgba(107,114,128,0.12)",
                    color: intent.ativo ? "#1D4ED8" : "#6B7280",
                    fontSize: 12,
                    fontWeight: 700,
                  }}
                >
                  {intent.ativo ? "Ativo" : "Inativo"}
                </span>
              </div>

              <div className="flex align-items-center gap-2">
                <InputSwitch checked={intent.ativo} onChange={(e) => onToggleActive(intent.id, !!e.value)} />

                <Button
                  icon="pi pi-pencil"
                  className="p-button-text p-button-rounded"
                  severity="secondary"
                  tooltip="Editar"
                  tooltipOptions={{ position: "top" }}
                  onClick={() => onEdit(intent)}
                />

                <Button
                  icon="pi pi-trash"
                  className="p-button-text p-button-rounded"
                  severity="danger"
                  tooltip="Excluir"
                  tooltipOptions={{ position: "top" }}
                  onClick={() => onDelete(intent.id)}
                />
              </div>
            </div>

            <div className="mt-3 flex flex-wrap gap-2">
              {intent.keywords.map((k, idx) => (
                <span
                  key={`${intent.id}-kw-${idx}`}
                  className="border-round-xl px-2 py-1"
                  style={{
                    border: "1px solid rgba(0,0,0,0.08)",
                    background: "#fff",
                    fontSize: 12,
                    fontWeight: 600,
                    color: "#374151",
                  }}
                >
                  {k}
                </span>
              ))}
            </div>

            <div
              className="mt-3 border-round-xl p-3"
              style={{
                background: "rgba(243,244,246,0.8)",
                border: "1px solid rgba(0,0,0,0.04)",
                color: "#374151",
              }}
            >
              {intent.respostaBase}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}