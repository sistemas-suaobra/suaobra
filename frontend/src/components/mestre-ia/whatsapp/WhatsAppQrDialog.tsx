import React from "react";
import { Dialog } from "primereact/dialog";
import { Button } from "primereact/button";

type Props = {
  visible: boolean;
  onHide: () => void;

  waQrLoading: boolean;
  waQr: string;
  waQrError: string;

  onRefresh: () => void;
};

export function WhatsAppQrDialog({
  visible,
  onHide,
  waQrLoading,
  waQr,
  waQrError,
  onRefresh,
}: Props) {
  return (
    <Dialog
      header="Conectar WhatsApp"
      visible={visible}
      style={{ width: "520px", maxWidth: "95vw" }}
      onHide={onHide}
      draggable={false}
      dismissableMask
      closable
      footer={
        <div className="flex justify-content-end gap-2">
          <Button
            label={waQrLoading ? "Atualizando..." : "Atualizar QR"}
            icon="pi pi-refresh"
            onClick={onRefresh}
            severity="info"
            disabled={waQrLoading}
          />
          <Button
            label="Fechar"
            icon="pi pi-times"
            onClick={onHide}
            severity="secondary"
          />
        </div>
      }
    >
      <div className="text-secondary mb-3">
        Se o QR vier vazio, pode ser porque já está logado. Se não, conecte a sessão e tente de novo.
      </div>

      <div
        className="flex align-items-center justify-content-center border-1 surface-border border-round-2xl"
        style={{ height: 340, background: "var(--surface-50)" }}
      >
        {waQrLoading ? (
          <div className="text-center">
            <i className="pi pi-spin pi-spinner" style={{ fontSize: 28 }} />
            <div className="mt-3 text-secondary">Carregando QR Code...</div>
          </div>
        ) : waQr ? (
          <div className="text-center">
            <div
              className="border-1 surface-border border-round-xl"
              style={{
                width: 260,
                height: 260,
                background: "white",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                margin: "0 auto",
                overflow: "hidden",
                padding: 10,
              }}
            >
              <img
                src={waQr}
                alt="QR Code WhatsApp"
                style={{ width: "100%", height: "100%", objectFit: "contain" }}
              />
            </div>

            <div className="mt-3 text-secondary">Abra o WhatsApp no celular e escaneie.</div>
          </div>
        ) : (
          <div className="text-center">
            <i className="pi pi-exclamation-triangle" style={{ fontSize: 26, opacity: 0.75 }} />
            <div className="mt-3 text-secondary">
              {waQrError || "Nenhum QR carregado ainda. Clique em 'Atualizar QR'."}
            </div>
          </div>
        )}
      </div>
    </Dialog>
  );
}