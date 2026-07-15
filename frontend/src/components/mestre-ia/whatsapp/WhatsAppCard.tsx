import React from "react";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";
import { InputText } from "primereact/inputtext";
import { InputTextarea } from "primereact/inputtextarea";
import { WhatsAppQrDialog } from "./WhatsAppQrDialog";
import { MestreIaTransitionLoader } from "../MestreIaTransitionLoader";
import { formatWhatsappJid } from "../utils/whatsappJid";
import { baseURL } from "../../../store/api";
import { user } from "../../../store/store";

type Props = {
  waConnected: boolean;
  waOwned: boolean;
  waSessionOk: boolean;
  waJid?: string;
  waDialogVisible: boolean;
  setWaDialogVisible: (v: boolean) => void;
  waCreating: boolean;
  waDeleting: boolean;
  waSessionLoading: boolean;
  waQrLoading: boolean;
  waQr: string;
  waQrError: string;

  criarConexaoWhatsApp: () => void;
  abrirECarregarQRCode: () => void;
  carregarQRCode: () => Promise<void>;
  verificarStatusWhatsapp: () => Promise<boolean>;
  disconnectWhatsApp: () => Promise<void>;
  removeWhatsAppInstance: () => Promise<void>;
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
          Authorization: user.get().token || "",
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

  const phoneDisplay = React.useMemo(
    () => formatWhatsappJid(props.waJid ?? ""),
    [props.waJid]
  );

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
          style={{
            background: "#25d366",
            border: "none",
            maxWidth: "100%",
          }}
        />
      );
    }

    return (
      <Tag
        value="Sessão não autenticada"
        severity="warning"
        icon="pi pi-exclamation-triangle"
      />
    );
  };

  const fullyConnected = props.waSessionOk && !!props.waJid;
  // Está "emprestando" a conexão legada/compartilhada do time (não é a própria).
  const isBorrowed = props.waConnected && !props.waOwned;

  const handleSendTest = async () => {
    if (!testPhone || !testMessage) return;

    setSending(true);
    try {
      await props.sendTestMessage(testPhone, testMessage);
    } finally {
      setSending(false);
    }
  };

  const primaryActionLabel =
    props.waSessionLoading || props.waQrLoading ? "Gerando QR..." : "Ver QR Code";

  const primaryActionDisabled = props.waSessionLoading || props.waQrLoading;
  // Excluir/gerenciar só a conexão PRÓPRIA — nunca a compartilhada emprestada.
  const canRemoveInstance = props.waConnected && props.waOwned && !props.waCreating;

  const handleRemoveInstance = async () => {
    const confirmed = window.confirm(
      "Tem certeza que deseja excluir a instância do WhatsApp? Essa ação remove a conexão atual."
    );
    if (!confirmed) return;

    await props.removeWhatsAppInstance();
  };

  return (
    <div className="col-12 lg:col-6">
      <div
        className="bg-white border-round-3xl p-3 md:p-4 border-1 surface-border h-full whatsapp-card"
        style={{ position: "relative" }}
      >
        {props.waCreating ? (
          <MestreIaTransitionLoader overlay caption="Criando instância WhatsApp…" />
        ) : null}
        <div className="flex flex-column gap-3 mb-3">
          <div className="flex flex-column md:flex-row md:align-items-start md:justify-content-between gap-3">
            <div className="flex align-items-center gap-2 min-w-0">
              <i className="pi pi-whatsapp text-2xl" />
              <h3 className="m-0">WhatsApp</h3>
            </div>

            <div className="flex flex-wrap gap-2 whatsapp-status-wrap">
              <WhatsAppStatus />
              <WhatsAppSessionStatus />
            </div>
          </div>

          <div className="text-secondary" style={{ lineHeight: 1.6, wordBreak: "break-word" }}>
            {isBorrowed ? (
              <>
                Você está usando o número compartilhado do time
                {fullyConnected ? <> ({phoneDisplay})</> : null}. Conecte o seu
                próprio número para ter uma conexão isolada dos colegas.
              </>
            ) : fullyConnected ? (
              <span style={{ color: "#25d366", fontWeight: 600 }}>
                WhatsApp conectado! Número: {phoneDisplay}
              </span>
            ) : props.waConnected ? (
              <>Conecte ao seu WhatsApp e dispare mensagens para seus leads.</>
            ) : (
              <>Crie a instância para começar a conexão com o WhatsApp.</>
            )}
          </div>
        </div>

        <div className="formgrid grid">
          <div className="field col-12">
            {isBorrowed ? (
              <Button
                label={props.waCreating ? "Criando..." : "Conectar meu próprio número"}
                icon="pi pi-plus"
                onClick={props.criarConexaoWhatsApp}
                className="w-full"
                disabled={props.waCreating}
              />
            ) : !props.waConnected ? (
              <Button
                label={props.waCreating ? "Criando..." : "Conectar (criar instância)"}
                icon="pi pi-plus"
                onClick={props.criarConexaoWhatsApp}
                className="w-full"
                disabled={props.waCreating}
              />
            ) : fullyConnected ? (
              <Button
                label="Desconectar"
                icon="pi pi-power-off"
                onClick={props.disconnectWhatsApp}
                className="w-full"
                severity="secondary"
              />
            ) : (
              <Button
                label={primaryActionLabel}
                icon="pi pi-qrcode"
                onClick={props.abrirECarregarQRCode}
                className="w-full"
                severity="help"
                disabled={primaryActionDisabled}
              />
            )}
          </div>

          {canRemoveInstance ? (
            <div className="field col-12">
              <Button
                label={props.waDeleting ? "Excluindo..." : "Excluir instância"}
                icon="pi pi-trash"
                onClick={handleRemoveInstance}
                className="w-full"
                severity="danger"
                outlined
                disabled={props.waDeleting}
              />
            </div>
          ) : null}

          {fullyConnected && !isBorrowed && (
            <>
              <div className="col-12">
                <hr
                  className="my-2 md:my-3"
                  style={{ border: "none", borderTop: "1px solid #dee2e6" }}
                />
                <h4 className="m-0 mb-2 text-sm font-semibold">
                  Testar Envio de Mensagem
                </h4>
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
                  Formato: DDI + DDD + número, sem espaços ou caracteres especiais
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
                  rows={4}
                  autoResize
                  className="w-full"
                  disabled={sending}
                />
              </div>

              <div className="field col-12">
                <div className="flex flex-column sm:flex-row gap-2">
                  <Button
                    label={sending ? "Enviando..." : "Enviar Teste"}
                    icon="pi pi-send"
                    onClick={handleSendTest}
                    className="w-full"
                    severity="success"
                    disabled={sending || !testPhone || !testMessage}
                  />
                </div>
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
          onCheckConnected={props.verificarStatusWhatsapp}
        />
      </div>

      <style>{`
        .whatsapp-card .p-tag {
          white-space: normal !important;
          word-break: break-word;
          max-width: 100%;
        }

        .whatsapp-status-wrap .p-tag {
          max-width: 100%;
        }

        @media screen and (max-width: 768px) {
          .whatsapp-card {
            border-radius: 1.25rem !important;
          }

          .whatsapp-card h3 {
            font-size: 1.1rem;
          }

          .whatsapp-card .p-button {
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
}