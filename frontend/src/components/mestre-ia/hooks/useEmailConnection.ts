import { useState, useCallback } from "react";
import { useStore } from "@nanostores/react";
import { user } from "../../../store/store";
import { baseURL, PB } from "../../../store/api";

interface EmailConfig {
  nome: string;
  remetente_email: string;
  reply_to: string;
  smtp_host: string;
  smtp_port: number;
  smtp_usuario: string;
  smtp_senha: string;
  criptografia: string;
}

interface UseEmailConnectionReturn {
  loading: boolean;
  configured: boolean;
  saving: boolean;
  sending: boolean;
  emailRecordId: string;
  emailNome: string;
  remetenteEmail: string;
  replyTo: string;
  smtpHost: string;
  smtpPort: number;
  smtpUsuario: string;
  criptografia: string;
  loadConfig: () => Promise<void>;
  saveConfig: (config: EmailConfig) => Promise<boolean>;
  sendTestEmail: (to: string, subject: string, body: string) => Promise<{ success: boolean; error?: string }>;
}

export function useEmailConnection(): UseEmailConnectionReturn {
  const currentUser = useStore(user);

  const [loading, setLoading] = useState(false);
  const [configured, setConfigured] = useState(false);
  const [saving, setSaving] = useState(false);
  const [sending, setSending] = useState(false);

  const [emailRecordId, setEmailRecordId] = useState("");
  const [emailNome, setEmailNome] = useState("");
  const [remetenteEmail, setRemetenteEmail] = useState("");
  const [replyTo, setReplyTo] = useState("");
  const [smtpHost, setSmtpHost] = useState("");
  const [smtpPort, setSmtpPort] = useState(587);
  const [smtpUsuario, setSmtpUsuario] = useState("");
  const [criptografia, setCriptografia] = useState("STARTTLS");

  const loadConfig = useCallback(async () => {
    if (!currentUser?.id) return;

    setLoading(true);
    try {
      const token = currentUser.token;
      const response = await fetch(`${baseURL()}/conexoes/email`, {
        headers: {
          Authorization: token,
        },
      });

      const data = await response.json();

      if (data.exists && data.email) {
        setConfigured(true);
        setEmailRecordId(data.email.id || "");
        setEmailNome(data.email.conexoes_email || "");
        setRemetenteEmail(data.email.remetente_email || "");
        setReplyTo(data.email.reply_to || "");
        setSmtpHost(data.email.smtp_host || "");
        setSmtpPort(data.email.smtp_port || 587);
        setSmtpUsuario(data.email.smtp_usuario || "");
        setCriptografia(data.email.criptografia || "STARTTLS");
      } else {
        setConfigured(false);
      }
    } catch (error) {
      console.error("Erro ao carregar configuração de email:", error);
    } finally {
      setLoading(false);
    }
  }, [currentUser?.id]);

  const saveConfig = useCallback(
    async (config: EmailConfig): Promise<boolean> => {
      if (!currentUser?.id) return false;

      setSaving(true);
      try {
        const token = currentUser.token;
        const response = await fetch(`${baseURL()}/conexoes/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
          body: JSON.stringify(config),
        });

        const data = await response.json();

        if (data.success && data.email) {
          setConfigured(true);
          setEmailRecordId(data.email.id || "");
          setEmailNome(data.email.conexoes_email || "");
          setRemetenteEmail(data.email.remetente_email || "");
          setReplyTo(data.email.reply_to || "");
          setSmtpHost(data.email.smtp_host || "");
          setSmtpPort(data.email.smtp_port || 587);
          setSmtpUsuario(data.email.smtp_usuario || "");
          setCriptografia(data.email.criptografia || "STARTTLS");
          return true;
        }

        return false;
      } catch (error) {
        console.error("Erro ao salvar configuração de email!:", error);
        return false;
      } finally {
        setSaving(false);
      }
    },
    [currentUser?.id]
  );

  const sendTestEmail = useCallback(
    async (to: string, subject: string, body: string): Promise<{ success: boolean; error?: string }> => {
      if (!currentUser?.id) return { success: false, error: "Usuário não autenticado" };

      setSending(true);
      try {
        const token = currentUser.token;
        const response = await fetch(`${baseURL()}/conexoes/email/send-test`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
          body: JSON.stringify({ to, subject, body }),
        });

        const data = await response.json();

        if (data.success) {
          return { success: true };
        } else {
          return { success: false, error: data.error || "Erro ao enviar email" };
        }
      } catch (error) {
        console.error("Erro ao enviar email de teste:", error);
        return { success: false, error: String(error) };
      } finally {
        setSending(false);
      }
    },
    [currentUser?.id]
  );

  return {
    loading,
    configured,
    saving,
    sending,
    emailRecordId,
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
  };
}
