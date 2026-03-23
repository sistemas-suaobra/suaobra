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

type WhatsMeowStatusResp = {
  connected?: boolean;
  jid?: string;
};

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function useWhatsappConnection(notify: NotifyFn) {
  const [waConnected, setWaConnected] = React.useState(false);
  const [waSessionOk, setWaSessionOk] = React.useState(false);
  const [waJid, setWaJid] = React.useState<string>("");

  const [waDialogVisible, setWaDialogVisible] = React.useState(false);

  const [waCreating, setWaCreating] = React.useState(false);
  const [waSessionLoading, setWaSessionLoading] = React.useState(false);

  const [waQrLoading, setWaQrLoading] = React.useState(false);
  const [waQr, setWaQr] = React.useState<string>("");
  const [waQrError, setWaQrError] = React.useState<string>("");

  const fetchQrOnce = React.useCallback(async (): Promise<string> => {
    const url = `${baseURL()}/conexoes/whatsapp/qr`;
    const { data } = await requestWithLog<WhatsMeowQrResp>({
      label: "Obter QRCode WhatsApp",
      method: "GET",
      url,
    });

    return normalizeQr(data?.data?.QRCode);
  }, []);

  const criarConexaoWhatsApp = React.useCallback(async () => {
    try {
      setWaCreating(true);

      const url = `${baseURL()}/conexoes/whatsapp`;
      const { data } = await requestWithLog<ConexaoWhatsappResponse>({
        label: "Criar Conexão WhatsApp",
        method: "POST",
        url,
        body: {},
      });

      setWaConnected(true);
      setWaSessionOk(false);
      setWaJid("");

      notify(
        "success",
        "WhatsApp",
        data?.alreadyExists
          ? "Instância já existia. Agora clique em 'Ver QR Code'."
          : "Instância criada. Agora clique em 'Ver QR Code'."
      );

      console.log("conexao_whatsapp (parsed):", data);
    } catch (e: any) {
      console.error(e);
      notify("error", "WhatsApp", e?.message || "Falha ao criar conexão WhatsApp.");
    } finally {
      setWaCreating(false);
    }
  }, [notify]);

  const carregarQRCode = React.useCallback(async () => {
    setWaQrLoading(true);
    setWaQr("");
    setWaQrError("");

    try {
      const qr = await fetchQrOnce();

      if (!qr) {
        setWaQr("");
        setWaQrError(
          "QR vazio no momento. Clique em 'Atualizar QR' para tentar novamente."
        );
        notify("warn", "WhatsApp", "QR Code ainda não está disponível.");
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
  }, [fetchQrOnce, notify]);

  const abrirECarregarQRCode = React.useCallback(async () => {
    if (!waConnected) {
      notify("warn", "WhatsApp", "Crie a instância antes de gerar o QR Code.");
      return;
    }

    setWaDialogVisible(true);
    setWaSessionLoading(true);
    setWaQrLoading(true);
    setWaQr("");
    setWaQrError("");

    try {
      const connectUrl = `${baseURL()}/conexoes/whatsapp/connect`;
      const { data } = await requestWithLog<WhatsMeowConnectResp>({
        label: "Conectar Sessão WhatsApp",
        method: "POST",
        url: connectUrl,
        body: {},
      });

      console.log("whatsapp_connect_raw (parsed):", data?.raw);

      let qr = "";
      const maxAttempts = 5;

      for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        try {
          qr = await fetchQrOnce();
          if (qr) break;
        } catch (err) {
          if (attempt === maxAttempts) throw err;
        }

        if (attempt < maxAttempts) {
          await sleep(900);
        }
      }

      if (!qr) {
        setWaQr("");
        setWaQrError(
          "Não foi possível gerar o QR agora. Clique em 'Atualizar QR' para tentar novamente."
        );
        notify("warn", "WhatsApp", "Sessão iniciada, mas o QR ainda não ficou disponível.");
        return;
      }

      setWaQr(qr);
      notify("success", "WhatsApp", "QR Code carregado. Escaneie no WhatsApp.");
    } catch (e: any) {
      console.error(e);
      setWaQr("");
      setWaQrError(e?.message || "Falha ao iniciar a sessão e gerar o QR Code.");
      notify(
        "error",
        "WhatsApp",
        e?.message || "Falha ao iniciar a sessão e gerar o QR Code."
      );
    } finally {
      setWaSessionLoading(false);
      setWaQrLoading(false);
    }
  }, [fetchQrOnce, notify, waConnected]);

  const verificarStatusWhatsapp = React.useCallback(async (): Promise<boolean> => {
    try {
      setWaSessionLoading(true);

      const url = `${baseURL()}/conexoes/whatsapp/status`;
      const { data } = await requestWithLog<WhatsMeowStatusResp>({
        label: "Check WA Status",
        method: "GET",
        url,
      });

      if (data?.connected) {
        const jid = data.jid ?? "";
        setWaSessionOk(true);
        setWaJid(jid);
        setWaDialogVisible(false);

        const phone = jid ? `+${jid.split("@")[0]}` : "";
        notify(
          "success",
          "WhatsApp conectado!",
          phone ? `Número: ${phone}` : "Conexão estabelecida!"
        );

        return true;
      }

      notify("info", "WhatsApp", "Ainda aguardando a leitura do QR Code.");
      return false;
    } catch (e: any) {
      console.error(e);
      notify("error", "WhatsApp", e?.message || "Falha ao verificar status da conexão.");
      return false;
    } finally {
      setWaSessionLoading(false);
    }
  }, [notify]);

  const checkConnected = React.useCallback(async (): Promise<boolean> => {
    try {
      const url = `${baseURL()}/conexoes/whatsapp`;
      const { data } = await requestWithLog<GetConexaoWhatsappResp>({
        label: "Get Conexão WhatsApp",
        method: "GET",
        url,
      });

      // A conexão pode existir mesmo sem payload de `whatsapp` em algumas respostas.
      // Nesse caso, mantemos "instância criada" para evitar desaparecer da UI.
      const hasConexao = !!data?.exists && !!data?.conexao;
      setWaConnected(hasConexao);

      const jid: string = data?.whatsapp?.device_jid ?? "";
      const conectadoEm: string = data?.whatsapp?.conectado_em ?? "";

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

  const disconnectWhatsApp = React.useCallback(async (): Promise<void> => {
    try {
      const url = `${baseURL()}/conexoes/whatsapp/disconnect`;
      await requestWithLog({
        label: "Disconnect WhatsApp",
        method: "POST",
        url,
        body: {},
      });

      setWaDialogVisible(false);
      setWaQr("");
      setWaQrError("");
      setWaSessionOk(false);
      setWaJid("");

      await checkConnected();

      notify("info", "WhatsApp", "Sessão desconectada.");
    } catch (e: any) {
      console.error("disconnect error:", e);
      notify("error", "WhatsApp", e?.message || "Falha ao desconectar.");
    }
  }, [checkConnected, notify]);

  const sendTestMessage = React.useCallback(
    async (phone: string, body: string): Promise<void> => {
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
    },
    [notify]
  );

  const loadWhatsapp = checkConnected;

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
    abrirECarregarQRCode,
    carregarQRCode,
    verificarStatusWhatsapp,
    disconnectWhatsApp,
    sendTestMessage,
  };
}