import React from "react";
import { Dialog } from "primereact/dialog";
import { Button } from "primereact/button";
import { ProgressSpinner } from "primereact/progressspinner";

type Props = {
  visible: boolean;
  onHide: () => void;
  waQrLoading: boolean;
  waQr: string;
  waQrError: string;
  onRefresh: () => Promise<void> | void;
  onCheckConnected: () => Promise<boolean>;
};

export function WhatsAppQrDialog({
  visible,
  onHide,
  waQrLoading,
  waQr,
  waQrError,
  onRefresh,
  onCheckConnected,
}: Props) {
  const [checking, setChecking] = React.useState(false);

  const handleCheckConnected = async () => {
    try {
      setChecking(true);
      await onCheckConnected();
    } finally {
      setChecking(false);
    }
  };

  const footer = (
    <div className="flex flex-column sm:flex-row gap-2 justify-content-end">
      <Button
        type="button"
        label="Atualizar QR"
        icon="pi pi-refresh"
        severity="secondary"
        outlined
        onClick={onRefresh}
        disabled={waQrLoading || checking}
      />

      <Button
        type="button"
        label={checking ? "Verificando..." : "Já escaneei"}
        icon="pi pi-check"
        onClick={handleCheckConnected}
        loading={checking}
        disabled={waQrLoading}
      />

      <Button
        type="button"
        label="Fechar"
        icon="pi pi-times"
        text
        onClick={onHide}
        disabled={checking}
      />
    </div>
  );

  return (
    <Dialog
      header="Conectar WhatsApp"
      visible={visible}
      onHide={onHide}
      style={{ width: "100%", maxWidth: 520 }}
      modal
      closable={!checking}
      footer={footer}
      breakpoints={{ "960px": "75vw", "640px": "95vw" }}
    >
      <div className="flex flex-column align-items-center text-center gap-3 py-2">
        <div className="text-700" style={{ lineHeight: 1.6 }}>
          Abra o WhatsApp no celular, toque em <strong>Aparelhos conectados</strong> e escaneie o QR Code abaixo.
        </div>

        {waQrLoading ? (
          <div className="flex flex-column align-items-center gap-3 py-5">
            <ProgressSpinner style={{ width: 56, height: 56 }} strokeWidth="4" />
            <div className="text-600">Gerando QR Code...</div>
          </div>
        ) : waQr ? (
          <div className="flex flex-column align-items-center gap-3">
            <div
              className="surface-50 border-1 surface-border border-round-xl p-3"
              style={{ width: "100%", maxWidth: 320 }}
            >
              <img
                src={waQr}
                alt="QR Code do WhatsApp"
                style={{ width: "100%", display: "block", borderRadius: 12 }}
              />
            </div>

            <small className="text-600">
              Depois de escanear, clique em <strong>Já escaneei</strong>.
            </small>
          </div>
        ) : (
          <div className="flex flex-column align-items-center gap-2 py-4">
            <i className="pi pi-exclamation-triangle text-yellow-500 text-3xl" />
            <div className="text-700">
              {waQrError || "Não foi possível carregar o QR Code agora."}
            </div>
          </div>
        )}
      </div>
    </Dialog>
  );
}