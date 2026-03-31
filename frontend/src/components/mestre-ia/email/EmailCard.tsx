import React, { useEffect, useRef, useState } from "react";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";
import { InputText } from "primereact/inputtext";
import { Password } from "primereact/password";
import { Dropdown } from "primereact/dropdown";
import { InputTextarea } from "primereact/inputtextarea";
import { Dialog } from "primereact/dialog";
import { Toast } from "primereact/toast";
import { useEmailConnection } from "../hooks/useEmailConnection";
import { MestreIaTransitionLoader } from "../MestreIaTransitionLoader";

export function EmailCard() {
  const toast = useRef<Toast>(null);

  const {
    loading,
    configured,
    saving,
    sending,
    emailNome,
    remetenteEmail,
    replyTo,
    smtpHost,
    smtpPort,
    smtpUsuario,
    criptografia,
    loadConfig,
    saveConfig,
    sendTestEmail,
  } = useEmailConnection();

  const [showConfigDialog, setShowConfigDialog] = useState(false);
  const [showTestDialog, setShowTestDialog] = useState(false);

  const [formNome, setFormNome] = useState("");
  const [formRemetenteEmail, setFormRemetenteEmail] = useState("");
  const [formReplyTo, setFormReplyTo] = useState("");
  const [formSmtpHost, setFormSmtpHost] = useState("");
  const [formSmtpPort, setFormSmtpPort] = useState(587);
  const [formSmtpUsuario, setFormSmtpUsuario] = useState("");
  const [formSmtpSenha, setFormSmtpSenha] = useState("");
  const [formCriptografia, setFormCriptografia] = useState("STARTTLS");

  const [testTo, setTestTo] = useState("");
  const [testSubject, setTestSubject] = useState("Email de Teste - Sua Obra");
  const [testBody, setTestBody] = useState("Este é um email de teste da plataforma Sua Obra.");

  const criptografiaOptions = [
    { label: "Nenhuma", value: "NONE" },
    { label: "STARTTLS (porta 587)", value: "STARTTLS" },
  ];

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const handleConfigure = () => {
    setFormNome(emailNome);
    setFormRemetenteEmail(remetenteEmail);
    setFormReplyTo(replyTo);
    setFormSmtpHost(smtpHost);
    setFormSmtpPort(smtpPort);
    setFormSmtpUsuario(smtpUsuario);
    setFormSmtpSenha("");
    setFormCriptografia(criptografia);
    setShowConfigDialog(true);
  };

  const handleSaveConfig = async () => {
    if (!formNome || !formRemetenteEmail || !formSmtpHost || !formSmtpPort || !formSmtpUsuario) {
      toast.current?.show({
        severity: "warn",
        summary: "Campos obrigatórios",
        detail: "Preencha todos os campos obrigatórios",
        life: 3000,
      });
      return;
    }

    const success = await saveConfig({
      nome: formNome,
      remetente_email: formRemetenteEmail,
      reply_to: formReplyTo,
      smtp_host: formSmtpHost,
      smtp_port: formSmtpPort,
      smtp_usuario: formSmtpUsuario,
      smtp_senha: formSmtpSenha,
      criptografia: formCriptografia,
    });

    if (success) {
      toast.current?.show({
        severity: "success",
        summary: "Sucesso",
        detail: "Configuração salva com sucesso",
        life: 3000,
      });
      setShowConfigDialog(false);
    } else {
      toast.current?.show({
        severity: "error",
        summary: "Erro",
        detail: "Erro ao salvar configuração",
        life: 3000,
      });
    }
  };

  const handleOpenTestDialog = () => {
    setTestTo("");
    setTestSubject("Email de Teste - Sua Obra");
    setTestBody("Este é um email de teste da plataforma Sua Obra.");
    setShowTestDialog(true);
  };

  const handleSendTest = async () => {
    if (!testTo || !testSubject || !testBody) {
      toast.current?.show({
        severity: "warn",
        summary: "Campos obrigatórios",
        detail: "Preencha todos os campos",
        life: 3000,
      });
      return;
    }

    const result = await sendTestEmail(testTo, testSubject, testBody);

    if (result.success) {
      toast.current?.show({
        severity: "success",
        summary: "Sucesso",
        detail: "Email enviado com sucesso",
        life: 3000,
      });
      setShowTestDialog(false);
    } else {
      toast.current?.show({
        severity: "error",
        summary: "Erro ao enviar",
        detail: result.error || "Erro ao enviar email",
        life: 5000,
      });
    }
  };

  const configFooter = (
    <div className="flex flex-column sm:flex-row justify-content-end gap-2 w-full">
      <Button
        label="Cancelar"
        icon="pi pi-times"
        onClick={() => setShowConfigDialog(false)}
        className="p-button-text w-full sm:w-auto"
      />
      <Button
        label="Salvar"
        icon="pi pi-check"
        onClick={handleSaveConfig}
        loading={saving}
        className="w-full sm:w-auto"
      />
    </div>
  );

  const testFooter = (
    <div className="flex flex-column sm:flex-row justify-content-end gap-2 w-full">
      <Button
        label="Cancelar"
        icon="pi pi-times"
        onClick={() => setShowTestDialog(false)}
        className="p-button-text w-full sm:w-auto"
      />
      <Button
        label="Enviar"
        icon="pi pi-send"
        onClick={handleSendTest}
        loading={sending}
        className="w-full sm:w-auto"
      />
    </div>
  );

  return (
    <>
      <Toast ref={toast} />

      <div className="col-12 lg:col-6">
        <div className="bg-white border-round-3xl p-3 md:p-4 border-1 surface-border h-full email-card">
          <div className="flex flex-column gap-3 mb-3">
            <div className="flex flex-column md:flex-row md:align-items-start md:justify-content-between gap-3">
              <div className="flex align-items-center gap-2 min-w-0">
                <i className="pi pi-envelope text-2xl" />
                <h3 className="m-0">Email</h3>
              </div>

              <div className="flex flex-wrap gap-2">
                <Tag
                  value={configured ? "Configurado" : "Não configurado"}
                  severity={configured ? "success" : "danger"}
                  icon={configured ? "pi pi-check" : "pi pi-times"}
                />
              </div>
            </div>

            <div className="text-secondary" style={{ lineHeight: 1.6, wordBreak: "break-word" }}>
              {loading ? (
                <MestreIaTransitionLoader row size="sm" minHeight={0} caption="Carregando configuração…" />
              ) : configured ? (
                <span style={{ color: "#6366f1", fontWeight: 600 }}>
                  Email configurado: {remetenteEmail}
                </span>
              ) : (
                <>Configure as credenciais SMTP para enviar emails automáticos.</>
              )}
            </div>

            {configured && !loading && (
              <div className="text-sm text-600" style={{ lineHeight: 1.7, wordBreak: "break-word" }}>
                <div>
                  <strong>Nome:</strong> {emailNome}
                </div>
                <div>
                  <strong>Servidor:</strong> {smtpHost}:{smtpPort}
                </div>
                <div>
                  <strong>Criptografia:</strong> {criptografia}
                </div>
              </div>
            )}
          </div>

          <div className="formgrid grid">
            <div className="field col-12">
              {!configured ? (
                <Button
                  label="Configurar Email"
                  icon="pi pi-cog"
                  onClick={handleConfigure}
                  className="w-full"
                />
              ) : (
                <div className="flex flex-column sm:flex-row gap-2">
                  <Button
                    label="Editar Configuração"
                    icon="pi pi-pencil"
                    onClick={handleConfigure}
                    className="w-full"
                    severity="secondary"
                  />
                  <Button
                    label="Enviar Email de Teste"
                    icon="pi pi-send"
                    onClick={handleOpenTestDialog}
                    className="w-full"
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      <Dialog
        header="Configurar Email"
        visible={showConfigDialog}
        style={{ width: "700px", maxWidth: "96vw" }}
        breakpoints={{ "960px": "90vw", "640px": "96vw" }}
        onHide={() => setShowConfigDialog(false)}
        footer={configFooter}
      >
        <div className="formgrid grid">
          <div className="field col-12 md:col-6">
            <label htmlFor="nome">Nome da Conexão *</label>
            <InputText
              id="nome"
              value={formNome}
              onChange={(e) => setFormNome(e.target.value)}
              className="w-full"
              placeholder="Ex: Email Principal"
            />
          </div>

          <div className="field col-12 md:col-6">
            <label htmlFor="remetente_email">Email Remetente *</label>
            <InputText
              id="remetente_email"
              value={formRemetenteEmail}
              onChange={(e) => setFormRemetenteEmail(e.target.value)}
              className="w-full"
              placeholder="contato@suaobra.com.br"
            />
          </div>

          <div className="field col-12 md:col-6">
            <label htmlFor="reply_to">Reply-To (opcional)</label>
            <InputText
              id="reply_to"
              value={formReplyTo}
              onChange={(e) => setFormReplyTo(e.target.value)}
              className="w-full"
              placeholder="suporte@suaobra.com.br"
            />
          </div>

          <div className="field col-12 md:col-6">
            <label htmlFor="smtp_host">Servidor SMTP *</label>
            <InputText
              id="smtp_host"
              value={formSmtpHost}
              onChange={(e) => setFormSmtpHost(e.target.value)}
              className="w-full"
              placeholder="smtp.gmail.com"
            />
          </div>

          <div className="field col-12 md:col-4">
            <label htmlFor="smtp_port">Porta SMTP *</label>
            <InputText
              id="smtp_port"
              type="number"
              value={String(formSmtpPort)}
              onChange={(e) => setFormSmtpPort(parseInt(e.target.value) || 587)}
              className="w-full"
              placeholder="587"
            />
          </div>

          <div className="field col-12 md:col-8">
            <label htmlFor="criptografia">Criptografia *</label>
            <Dropdown
              id="criptografia"
              value={formCriptografia}
              options={criptografiaOptions}
              onChange={(e) => setFormCriptografia(e.value)}
              className="w-full"
            />
          </div>

          <div className="field col-12 md:col-6">
            <label htmlFor="smtp_usuario">Usuário SMTP *</label>
            <InputText
              id="smtp_usuario"
              value={formSmtpUsuario}
              onChange={(e) => setFormSmtpUsuario(e.target.value)}
              className="w-full"
              placeholder="usuario@gmail.com"
            />
          </div>

          <div className="field col-12 md:col-6">
            <label htmlFor="smtp_senha">
              Senha SMTP {configured ? "(deixe vazio para não alterar)" : "*"}
            </label>
            <Password
              id="smtp_senha"
              value={formSmtpSenha}
              onChange={(e) => setFormSmtpSenha(e.target.value)}
              className="w-full"
              inputClassName="w-full"
              toggleMask
              feedback={false}
            />
          </div>
        </div>
      </Dialog>

      <Dialog
        header="Enviar Email de Teste"
        visible={showTestDialog}
        style={{ width: "560px", maxWidth: "96vw" }}
        breakpoints={{ "960px": "90vw", "640px": "96vw" }}
        onHide={() => setShowTestDialog(false)}
        footer={testFooter}
      >
        <div className="flex flex-column gap-3">
          <div className="field">
            <label htmlFor="test_to">Para (email) *</label>
            <InputText
              id="test_to"
              value={testTo}
              onChange={(e) => setTestTo(e.target.value)}
              className="w-full"
              placeholder="destinatario@exemplo.com"
            />
          </div>

          <div className="field">
            <label htmlFor="test_subject">Assunto *</label>
            <InputText
              id="test_subject"
              value={testSubject}
              onChange={(e) => setTestSubject(e.target.value)}
              className="w-full"
            />
          </div>

          <div className="field">
            <label htmlFor="test_body">Mensagem *</label>
            <InputTextarea
              id="test_body"
              value={testBody}
              onChange={(e) => setTestBody(e.target.value)}
              className="w-full"
              rows={5}
              autoResize
            />
          </div>
        </div>
      </Dialog>

      <style>{`
        .email-card .p-tag {
          white-space: normal !important;
          word-break: break-word;
          max-width: 100%;
        }

        @media screen and (max-width: 768px) {
          .email-card {
            border-radius: 1.25rem !important;
          }

          .email-card h3 {
            font-size: 1.1rem;
          }

          .email-card .p-button {
            width: 100%;
          }
        }
      `}</style>
    </>
  );
}