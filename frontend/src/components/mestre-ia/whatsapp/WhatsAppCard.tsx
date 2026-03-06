import React from "react";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";
import { InputText } from "primereact/inputtext";
import { InputTextarea } from "primereact/inputtextarea";
import { WhatsAppQrDialog } from "./WhatsAppQrDialog";
import { baseURL } from "../../../store/api";
import { user } from "../../../store/store";

type Props = {
  waConnected: boolean;
  waSessionOk: boolean;
  waJid?: string;
  waDialogVisible: boolean;
  setWaDialogVisible: (v: boolean) => void;
  waCreating: boolean;
  waSessionLoading: boolean;
  waQrLoading: boolean;
  waQr: string;
  waQrError: string;

  criarConexaoWhatsApp: () => void;
  conectarSessaoWhatsApp: () => void;
  abrirECarregarQRCode: () => void;
  carregarQRCode: () => void;
  disconnectWhatsApp: () => void;
  sendTestMessage: (phone: string, body: string) => Promise<void>;
};

export function WhatsAppCard(props: Props) {
  const [testPhone, setTestPhone] = React.useState("");
  const [testMessage, setTestMessage] = React.useState("Olá! Esta é uma mensagem de teste.");
  const [sending, setSending] = React.useState(false);
  const [fixingWebhook, setFixingWebhook] = React.useState(false);

  const handleFixWebhook = async () => {
    setFixingWebhook(true);
    try {
      const resp = await fetch(`${baseURL()}/conexoes/whatsapp/fix-webhook`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": user.get().token || "",
        },
      });
      const data = await resp.json();
      if (resp.ok && data.ok) {
        alert(`Webhook atualizado!\nURL: ${data.webhook_url}`);
      } else {
        alert(`Erro: ${data.message || JSON.stringify(data)}`);
      }
    } catch (e) {
      alert(`Erro ao corrigir webhook: ${e}`);
    } finally {
      setFixingWebhook(false);
    }
  };

  // Formata o JID (5511999998888@s.whatsapp.net) em número legível
  const phoneDisplay = React.useMemo(() => {
    if (!props.waJid) return "";
    const raw = props.waJid.split("@")[0];
    // +55 11 99999-8888
    if (raw.length >= 12) {
      return `+${raw.slice(0, 2)} ${raw.slice(2, 4)} ${raw.slice(4, 9)}-${raw.slice(9)}`;
    }
    return `+${raw}`;
  }, [props.waJid]);

  const WhatsAppStatus = () => (
    <Tag
      value={props.waConnected ? "Instância criada" : "Sem instância"}
      severity={props.waConnected ? "success" : "danger"}
      icon={props.waConnected ? "pi pi-check" : "pi pi-times"}
    />
  );

  const WhatsAppSessionStatus = () => {
    if (props.waSessionOk && props.waJid) {
      return (
        <Tag
          value={phoneDisplay}
          severity="success"
          icon="pi pi-whatsapp"
          style={{ background: "#25d366", border: "none" }}
        />
      );
    }
    return (
      <Tag
        value={props.waSessionOk ? "Sessão iniciada" : "Sessão não iniciada"}
        severity={props.waSessionOk ? "success" : "warning"}
        icon={props.waSessionOk ? "pi pi-check" : "pi pi-exclamation-triangle"}
      />
    );
  };

  const fullyConnected = props.waSessionOk && !!props.waJid;

  const handleSendTest = async () => {
    if (!testPhone || !testMessage) return;
    setSending(true);
    try {
      await props.sendTestMessage(testPhone, testMessage);
    } finally {
      setSending(false);
    }
  };

  // Quando já está conectado via QR (device_jid preenchido), não exibe mais o botão de QR
  const primaryActionLabel = props.waSessionOk
    ? "Ver QR Code"
    : props.waSessionLoading
    ? "Iniciando..."
    : "Iniciar Sessão";

  const primaryActionIcon = props.waSessionOk ? "pi pi-qrcode" : "pi pi-link";
  const primaryActionSeverity = props.waSessionOk ? ("info" as const) : ("help" as const);

  const onPrimaryAction = () => {
    if (!props.waSessionOk) {
      props.conectarSessaoWhatsApp();
      return;
    }
    props.abrirECarregarQRCode();
  };

  const primaryDisabled = props.waSessionLoading;

  return (
    <div className="col-12 md:col-6">
      <div className="bg-white border-round-3xl p-4 border-1 surface-border h-full">
        <div className="flex align-items-center justify-content-between mb-3">
          <div className="flex align-items-center gap-2">
            <i className="pi pi-whatsapp text-2xl" />
            <h3 className="m-0">WhatsApp</h3>
          </div>

          <div className="flex gap-2">
            <WhatsAppStatus />
            <WhatsAppSessionStatus />
          </div>
        </div>

        <div className="text-secondary mb-3" style={{ lineHeight: 1.5 }}>
          {fullyConnected ? (
            <span style={{ color: "#25d366", fontWeight: 600 }}>
              WhatsApp conectado! Número: {phoneDisplay}
            </span>
          ) : (
            <>
              Fluxo certo: 1) Criar instância → 2) Iniciar sessão → 3) Ver QR.
            </>
          )}
        </div>

        <div className="formgrid grid">
          <div className="field col-12 flex gap-2">
            {!props.waConnected ? (
              <Button
                label={props.waCreating ? "Criando..." : "Conectar (criar instância)"}
                icon="pi pi-plus"
                onClick={props.criarConexaoWhatsApp}
                className="w-full"
                disabled={props.waCreating}
              />
            ) : fullyConnected ? (
              // ── Estado: conectado via QR ──────────────────────────────────
              <Button
                label="Desconectar"
                icon="pi pi-power-off"
                onClick={props.disconnectWhatsApp}
                className="w-full"
                severity="secondary"
              />
            ) : (
              // ── Estado: instância criada, sessão pendente / QR pendente ──
              <>
                <Button
                  label={primaryActionLabel}
                  icon={primaryActionIcon}
                  onClick={onPrimaryAction}
                  className="w-full"
                  severity={primaryActionSeverity}
                  disabled={primaryDisabled}
                />

                <Button
                  label="Desconectar"
                  icon="pi pi-power-off"
                  onClick={props.disconnectWhatsApp}
                  className="w-full"
                  severity="secondary"
                />
              </>
            )}
          </div>

          {/* ── Form de Teste (só aparece quando conectado) ── */}
          {fullyConnected && (
            <>
              <div className="col-12">
                <hr className="my-3" style={{ border: "none", borderTop: "1px solid #dee2e6" }} />
                <h4 className="m-0 mb-2 text-sm font-semibold">Testar Envio de Mensagem</h4>
              </div>
              
              <div className="field col-12">
                <label htmlFor="test-phone" className="block text-sm font-medium mb-2">
                  Número de Destino (com DDI)
                </label>
                <InputText
                  id="test-phone"
                  value={testPhone}
                  onChange={(e) => setTestPhone(e.target.value)}
                  placeholder="Ex: 5511999998888"
                  className="w-full"
                  disabled={sending}
                />
                <small className="block mt-1 text-500">
                  Formato: DDI + DDD + número (sem espaços ou caracteres especiais)
                </small>
              </div>

              <div className="field col-12">
                <label htmlFor="test-message" className="block text-sm font-medium mb-2">
                  Mensagem
                </label>
                <InputTextarea
                  id="test-message"
                  value={testMessage}
                  onChange={(e) => setTestMessage(e.target.value)}
                  rows={3}
                  className="w-full"
                  disabled={sending}
                />
              </div>

              <div className="field col-12">
                <Button
                  label={sending ? "Enviando..." : "Enviar Mensagem de Teste"}
                  icon="pi pi-send"
                  onClick={handleSendTest}
                  className="w-full"
                  severity="success"
                  disabled={sending || !testPhone || !testMessage}
                />
              </div>
            </>
          )}
        </div>

        <WhatsAppQrDialog
          visible={props.waDialogVisible}
          onHide={() => props.setWaDialogVisible(false)}
          waQrLoading={props.waQrLoading}
          waQr={props.waQr}
          waQrError={props.waQrError}
          onRefresh={props.carregarQRCode}
        />
      </div>
    </div>
  );
}