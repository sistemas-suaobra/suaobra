import * as React from "react";
import { baseURL } from "../../../store/api";
import { requestWithLog } from "../utils/requestWithLog";
import { normalizeQr } from "../utils/normalizeQr";

type NotifyFn = (
  severity: "success" | "info" | "warn" | "error",
  summary: string,
  detail: string
) => void;

type ConexaoWhatsappResponse = {
  conexao?: any;
  whatsapp?: any;
  alreadyExists?: boolean;
};

type GetConexaoWhatsappResp = {
  exists?: boolean;
  conexao?: any;
  whatsapp?: any;
};

type WhatsMeowQrResp = {
  success?: boolean;
  code?: number;
  data?: { QRCode?: string };
};

type WhatsMeowConnectResp = {
  success?: boolean;
  raw?: any;
};

export function useWhatsappConnection(notify: NotifyFn) {
  const [waConnected, setWaConnected] = React.useState(false);
  const [waSessionOk, setWaSessionOk] = React.useState(false);
  /** JID do dispositivo (ex: 5511999998888@s.whatsapp.net) — preenchido pelo webhook Connected */
  const [waJid, setWaJid] = React.useState<string>("");
  /** ID do registro conexoes_whatsapp no PocketBase — necessário para o realtime */
  const [waRecordId, setWaRecordId] = React.useState<string>("");

  const [waDialogVisible, setWaDialogVisible] = React.useState(false);

  const [waCreating, setWaCreating] = React.useState(false);
  const [waSessionLoading, setWaSessionLoading] = React.useState(false);

  const [waQrLoading, setWaQrLoading] = React.useState(false);
  const [waQr, setWaQr] = React.useState<string>("");
  const [waQrError, setWaQrError] = React.useState<string>("");

  const criarConexaoWhatsApp = React.useCallback(async () => {
    try {
      setWaCreating(true);

      const url = `https://api-hml.suaobra.com.br/conexoes/whatsapp`;
      const { data } = await requestWithLog<ConexaoWhatsappResponse>({
        label: "Criar Conexão WhatsApp",
        method: "POST",
        url,
        body: {},
      });

      setWaConnected(true);
      setWaSessionOk(false);

      if (data?.whatsapp?.id) {
        setWaRecordId(data.whatsapp.id);
      }

      notify(
        "success",
        "WhatsApp",
        data?.alreadyExists
          ? "Instância já existia. Reutilizando."
          : "Instância criada. Agora clique em 'Conectar Sessão'."
      );

      console.log("conexao_whatsapp (parsed):", data);
    } catch (e: any) {
      console.error(e);
      notify("error", "WhatsApp", e?.message || "Falha ao criar conexão WhatsApp.");
    } finally {
      setWaCreating(false);
    }
  }, [notify]);

  const conectarSessaoWhatsApp = React.useCallback(async () => {
    try {
      setWaSessionLoading(true);

      const url = `${baseURL()}/conexoes/whatsapp/connect`;
      const { data } = await requestWithLog<WhatsMeowConnectResp>({
        label: "Conectar Sessão WhatsApp",
        method: "POST",
        url,
        body: {},
      });

      setWaSessionOk(true);
      notify("success", "WhatsApp", "Sessão conectando. Agora clique em 'Ver QR Code'.");
      console.log("whatsapp_connect_raw (parsed):", data?.raw);
    } catch (e: any) {
      console.error(e);
      setWaSessionOk(false);
      notify("error", "WhatsApp", e?.message || "Falha ao conectar sessão.");
    } finally {
      setWaSessionLoading(false);
    }
  }, [notify]);

  const carregarQRCode = React.useCallback(async () => {
    setWaQrLoading(true);
    setWaQr("");
    setWaQrError("");

    try {
      const url = `${baseURL()}/conexoes/whatsapp/qr`;
      const { data } = await requestWithLog<WhatsMeowQrResp>({
        label: "Obter QRCode WhatsApp",
        method: "GET",
        url,
      });

      const qr = normalizeQr(data?.data?.QRCode);

      if (!qr) {
        setWaQr("");
        setWaQrError(
          "QR vazio. Se já estiver logado, é normal. Se não, clique em 'Conectar Sessão' e tente novamente."
        );
        notify("warn", "WhatsApp", "QR Code vazio.");
        return;
      }

      setWaQr(qr);
      notify("success", "WhatsApp", "QR Code carregado. Escaneie no WhatsApp.");
    } catch (e: any) {
      console.error(e);
      setWaQrError(e?.message || "Falha ao buscar QR Code.");
      notify("error", "WhatsApp", e?.message || "Falha ao buscar QR Code.");
    } finally {
      setWaQrLoading(false);
    }
  }, [notify]);

  const abrirECarregarQRCode = React.useCallback(async () => {
    setWaDialogVisible(true);
    await carregarQRCode();
  }, [carregarQRCode]);

  const disconnectWhatsApp = React.useCallback(async () => {
    try {
      const url = `${baseURL()}/conexoes/whatsapp/disconnect`;
      await requestWithLog({ label: "Disconnect WhatsApp", method: "POST", url, body: {} });
    } catch (e) {
      console.error("disconnect error:", e);
    }
    setWaConnected(false);
    setWaSessionOk(false);
    setWaJid("");
    notify("info", "WhatsApp", "Desconectado.");
  }, [notify]);

  const sendTestMessage = React.useCallback(async (phone: string, body: string): Promise<void> => {
    try {
      const url = `${baseURL()}/conexoes/whatsapp/send-test`;
      const { data } = await requestWithLog<{ success?: boolean; error?: string; data?: any }>({
        label: "Enviar Mensagem Teste",
        method: "POST",
        url,
        body: { phone, body },
      });

      if (data?.success) {
        notify("success", "WhatsApp", "Mensagem enviada com sucesso!");
      } else {
        notify("error", "WhatsApp", data?.error || "Falha ao enviar mensagem.");
      }
    } catch (e: any) {
      console.error(e);
      notify("error", "WhatsApp", e?.message || "Falha ao enviar mensagem.");
    }
  }, [notify]);

  /** Verifica se já conectou — retorna true se conectado (apenas PB, NÃO consulta wuzapi) */
  const checkConnected = React.useCallback(async (): Promise<boolean> => {
    try {
      const url = `${baseURL()}/conexoes/whatsapp`;
      const { data } = await requestWithLog<GetConexaoWhatsappResp>({
        label: "Get Conexão WhatsApp",
        method: "GET",
        url,
      });

      const exists = !!data?.exists && !!data?.conexao && !!data?.whatsapp;
      setWaConnected(exists);

      if (data?.whatsapp?.id) {
        setWaRecordId(data.whatsapp.id);
      }

      const jid: string = data?.whatsapp?.device_jid ?? "";
      const conectadoEm: string = data?.whatsapp?.conectado_em ?? "";
      // Só mostra conectado se AMBOS existirem (device_jid preenchido E conectado_em preenchido)
      if (conectadoEm && jid) {
        setWaSessionOk(true);
        setWaJid(jid);
        return true;
      }

      setWaSessionOk(false);
      setWaJid("");
      return false;
    } catch (e: any) {
      console.error(e);
      return false;
    }
  }, []);

  const loadWhatsapp = checkConnected;

  // ── Polling: enquanto o dialog de QR estiver aberto, verifica a cada 3s ──
  React.useEffect(() => {
    if (!waDialogVisible) return;

    console.log("[WA Poll] iniciando polling (dialog aberto)");
    const interval = setInterval(async () => {
      console.log("[WA Poll] verificando conexão via /status...");
      try {
        const url = `${baseURL()}/conexoes/whatsapp/status`;
        const { data } = await requestWithLog<{ connected?: boolean; jid?: string }>({
          label: "Check WA Status",
          method: "GET",
          url,
        });

        if (data?.connected) {
          console.log("[WA Poll] CONECTADO! jid=", data.jid, " Fechando dialog...");
          const jid = data.jid ?? "";
          setWaSessionOk(true);
          setWaJid(jid);
          setWaDialogVisible(false);
          const phone = jid ? `+${jid.split("@")[0]}` : "";
          notify("success", "WhatsApp conectado!", phone ? `Número: ${phone}` : "Conexão estabelecida!");
          clearInterval(interval);
        }
      } catch (e) {
        console.error("[WA Poll] erro:", e);
      }
    }, 3000);

    return () => {
      console.log("[WA Poll] parando polling");
      clearInterval(interval);
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [waDialogVisible]);

  return {
    waConnected,
    waSessionOk,
    waJid,
    waDialogVisible,
    setWaDialogVisible,


    waCreating,
    waSessionLoading,

    waQrLoading,
    waQr,
    waQrError,

    loadWhatsapp,
    criarConexaoWhatsApp,
    conectarSessaoWhatsApp,
    abrirECarregarQRCode,
    carregarQRCode,
    disconnectWhatsApp,
    sendTestMessage,
  };
}