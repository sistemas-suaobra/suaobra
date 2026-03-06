import { useState, useEffect, useRef } from "react";
import { Button } from "primereact/button";
import { Tag } from "primereact/tag";
import { InputText } from "primereact/inputtext";
import { Password } from "primereact/password";
import { Dropdown } from "primereact/dropdown";
import { InputTextarea } from "primereact/inputtextarea";
import { Dialog } from "primereact/dialog";
import { Toast } from "primereact/toast";
import { useEmailConnection } from "../hooks/useEmailConnection";

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

  // Formulário de configuração
  const [formNome, setFormNome] = useState("");
  const [formRemetenteEmail, setFormRemetenteEmail] = useState("");
  const [formReplyTo, setFormReplyTo] = useState("");
  const [formSmtpHost, setFormSmtpHost] = useState("");
  const [formSmtpPort, setFormSmtpPort] = useState(587);
  const [formSmtpUsuario, setFormSmtpUsuario] = useState("");
  const [formSmtpSenha, setFormSmtpSenha] = useState("");
  const [formCriptografia, setFormCriptografia] = useState("STARTTLS");

  // Formulário de teste
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
    // Preenche o formulário com os valores atuais
    setFormNome(emailNome);
    setFormRemetenteEmail(remetenteEmail);
    setFormReplyTo(replyTo);
    setFormSmtpHost(smtpHost);
    setFormSmtpPort(smtpPort);
    setFormSmtpUsuario(smtpUsuario);
    setFormSmtpSenha(""); // Não preenche senha por segurança
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

  return (
    <>
      <Toast ref={toast} />

      <div className="col-12 md:col-6">
        <div className="bg-white border-round-3xl p-4 border-1 surface-border h-full">
          <div className="flex align-items-center justify-content-between mb-3">
            <div className="flex align-items-center gap-2">
              <i className="pi pi-envelope text-2xl" />
              <h3 className="m-0">Email</h3>
            </div>

            <div className="flex gap-2">
              <Tag
                value={configured ? "Configurado" : "Não configurado"}
                severity={configured ? "success" : "danger"}
                icon={configured ? "pi pi-check" : "pi pi-times"}
              />
            </div>
          </div>

          <div className="text-secondary mb-3" style={{ lineHeight: 1.5 }}>
            {loading ? (
              <span>Carregando...</span>
            ) : configured ? (
              <span style={{ color: "#6366f1", fontWeight: 600 }}>
                Email configurado: {remetenteEmail}
              </span>
            ) : (
              <>
                Configure as credenciais SMTP para enviar emails automáticos.
              </>
            )}
          </div>

          {configured && !loading && (
            <div className="text-sm text-600 mb-3" style={{ lineHeight: 1.6 }}>
              <div><strong>Nome:</strong> {emailNome}</div>
              <div><strong>Servidor:</strong> {smtpHost}:{smtpPort}</div>
              <div><strong>Criptografia:</strong> {criptografia}</div>
            </div>
          )}

          <div className="formgrid grid">
            <div className="field col-12 flex gap-2">
              {!configured ? (
                <Button
                  label="Configurar Email"
                  icon="pi pi-cog"
                  onClick={handleConfigure}
                  className="w-full"
                />
              ) : (
                <>
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
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Dialog de Configuração */}
      <Dialog
        header="Configurar Email"
        visible={showConfigDialog}
        style={{ width: "600px" }}
        onHide={() => setShowConfigDialog(false)}
        footer={
          <div>
            <Button label="Cancelar" icon="pi pi-times" onClick={() => setShowConfigDialog(false)} className="p-button-text" />
            <Button label="Salvar" icon="pi pi-check" onClick={handleSaveConfig} loading={saving} />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <div className="field">
            <label htmlFor="nome">Nome da Conexão *</label>
            <InputText id="nome" value={formNome} onChange={(e) => setFormNome(e.target.value)} className="w-full" placeholder="Ex: Email Principal" />
          </div>

          <div className="field">
            <label htmlFor="remetente_email">Email Remetente *</label>
            <InputText
              id="remetente_email"
              value={formRemetenteEmail}
              onChange={(e) => setFormRemetenteEmail(e.target.value)}
              className="w-full"
              placeholder="contato@suaobra.com.br"
            />
          </div>

          <div className="field">
            <label htmlFor="reply_to">Reply-To (opcional)</label>
            <InputText id="reply_to" value={formReplyTo} onChange={(e) => setFormReplyTo(e.target.value)} className="w-full" placeholder="suporte@suaobra.com.br" />
          </div>

          <div className="field">
            <label htmlFor="smtp_host">Servidor SMTP *</label>
            <InputText id="smtp_host" value={formSmtpHost} onChange={(e) => setFormSmtpHost(e.target.value)} className="w-full" placeholder="smtp.gmail.com" />
          </div>

          <div className="field">
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

          <div className="field">
            <label htmlFor="criptografia">Criptografia *</label>
            <Dropdown id="criptografia" value={formCriptografia} options={criptografiaOptions} onChange={(e) => setFormCriptografia(e.value)} className="w-full" />
          </div>

          <div className="field">
            <label htmlFor="smtp_usuario">Usuário SMTP *</label>
            <InputText id="smtp_usuario" value={formSmtpUsuario} onChange={(e) => setFormSmtpUsuario(e.target.value)} className="w-full" placeholder="usuario@gmail.com" />
          </div>

          <div className="field">
            <label htmlFor="smtp_senha">Senha SMTP {configured ? "(deixe vazio para não alterar)" : "*"}</label>
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

      {/* Dialog de Teste */}
      <Dialog
        header="Enviar Email de Teste"
        visible={showTestDialog}
        style={{ width: "500px" }}
        onHide={() => setShowTestDialog(false)}
        footer={
          <div>
            <Button label="Cancelar" icon="pi pi-times" onClick={() => setShowTestDialog(false)} className="p-button-text" />
            <Button label="Enviar" icon="pi pi-send" onClick={handleSendTest} loading={sending} />
          </div>
        }
      >
        <div className="flex flex-column gap-3">
          <div className="field">
            <label htmlFor="test_to">Para (email) *</label>
            <InputText id="test_to" value={testTo} onChange={(e) => setTestTo(e.target.value)} className="w-full" placeholder="destinatario@exemplo.com" />
          </div>

          <div className="field">
            <label htmlFor="test_subject">Assunto *</label>
            <InputText id="test_subject" value={testSubject} onChange={(e) => setTestSubject(e.target.value)} className="w-full" />
          </div>

          <div className="field">
            <label htmlFor="test_body">Mensagem *</label>
            <InputTextarea id="test_body" value={testBody} onChange={(e) => setTestBody(e.target.value)} className="w-full" rows={5} />
          </div>
        </div>
      </Dialog>
    </>
  );
}
