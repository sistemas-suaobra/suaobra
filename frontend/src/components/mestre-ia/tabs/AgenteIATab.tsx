import React from "react"
import { Button } from "primereact/button"
import { InputSwitch } from "primereact/inputswitch"
import type { AgenteIaTabProps } from "../types/agenteiatab"
import { MestreIaTransitionLoader } from "../MestreIaTransitionLoader"

export default function AgenteIaTab(props: AgenteIaTabProps) {
  const { intents, loading, onNew, onEdit, onDelete, onToggleActive } = props

  if (loading) {
    return (
      <MestreIaTransitionLoader minHeight={320} caption="Carregando intenções…" />
    )
  }

  return (
    <div className="w-full">
      <div className="flex flex-column md:flex-row md:align-items-center md:justify-content-between gap-3 mb-3">
        <div className="min-w-0">
          <div style={{ fontSize: 18, fontWeight: 700 }}>Intenções Configuradas</div>
          <div className="text-secondary" style={{ marginTop: 4, lineHeight: 1.5 }}>
            Defina as intenções que o agente deve identificar
          </div>
        </div>

        <Button
          icon="pi pi-plus"
          label="Nova Intenção"
          onClick={onNew}
          className="w-full md:w-auto"
        />
      </div>

      <div className="flex flex-column gap-3">
        {intents.map((intent) => (
          <div
            key={intent.id}
            className="bg-white border-round-2xl p-3 md:p-4 border-1 surface-border"
          >
            <div className="flex flex-column gap-3">
              <div className="flex flex-column lg:flex-row lg:align-items-start lg:justify-content-between gap-3">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap align-items-center gap-2">
                    <i className="pi pi-robot" style={{ color: "#2563EB", fontSize: 16 }} />

                    <div
                      style={{
                        fontWeight: 700,
                        fontSize: 16,
                        wordBreak: "break-word",
                      }}
                    >
                      {intent.titulo}
                    </div>

                    <span
                      className="border-round-xl px-2 py-1"
                      style={{
                        background: intent.ativo
                          ? "rgba(37,99,235,0.12)"
                          : "rgba(107,114,128,0.12)",
                        color: intent.ativo ? "#1D4ED8" : "#6B7280",
                        fontSize: 12,
                        fontWeight: 700,
                        whiteSpace: "nowrap",
                      }}
                    >
                      {intent.ativo ? "Ativo" : "Inativo"}
                    </span>
                  </div>
                </div>

                <div className="flex flex-wrap align-items-center gap-2 w-full lg:w-auto">
                  <div
                    className="flex align-items-center justify-content-between border-round-xl px-3 py-2 w-full sm:w-auto"
                    style={{
                      border: "1px solid rgba(0,0,0,0.06)",
                      background: "#FAFAFA",
                      minWidth: 150,
                    }}
                  >
                    <span
                      style={{
                        fontSize: 13,
                        fontWeight: 600,
                        color: "#4B5563",
                      }}
                    >
                      {intent.ativo ? "Ativo" : "Inativo"}
                    </span>

                    <InputSwitch
                      checked={intent.ativo}
                      onChange={(e) => onToggleActive(intent.id, !!e.value)}
                    />
                  </div>

                  <Button
                    icon="pi pi-pencil"
                    className="p-button-text p-button-rounded"
                    severity="secondary"
                    tooltip="Editar"
                    tooltipOptions={{ position: "top" }}
                    onClick={() => onEdit(intent)}
                    aria-label={`Editar intenção ${intent.titulo}`}
                  />

                  <Button
                    icon="pi pi-trash"
                    className="p-button-text p-button-rounded"
                    severity="danger"
                    tooltip="Excluir"
                    tooltipOptions={{ position: "top" }}
                    onClick={() => onDelete(intent.id)}
                    aria-label={`Excluir intenção ${intent.titulo}`}
                  />
                </div>
              </div>

              <div className="flex flex-wrap gap-2">
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
                      wordBreak: "break-word",
                      maxWidth: "100%",
                    }}
                  >
                    {k}
                  </span>
                ))}
              </div>

              <div
                className="border-round-xl p-3"
                style={{
                  background: "rgba(243,244,246,0.8)",
                  border: "1px solid rgba(0,0,0,0.04)",
                  color: "#374151",
                  lineHeight: 1.6,
                  wordBreak: "break-word",
                }}
              >
                {intent.respostaBase}
              </div>
            </div>
          </div>
        ))}

        {intents.length === 0 && (
          <div
            className="border-round-2xl p-4 md:p-5 text-center"
            style={{
              background: "#fff",
              border: "1px solid rgba(0,0,0,0.06)",
              color: "#6B7280",
            }}
          >
            Nenhuma intenção cadastrada ainda.
          </div>
        )}
      </div>
    </div>
  )
}