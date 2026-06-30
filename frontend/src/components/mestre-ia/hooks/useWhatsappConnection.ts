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
  owned?: boolean;
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
  error?: string;
};

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export function useWhatsappConnection(notify: NotifyFn) {
  const [waConnected, setWaConnected] = React.useState(false);
  // waOwned: a conexão exibida é do próprio usuário (true) ou é a conexão
  // legada/compartilhada do time emprestada como fallback (false).
  const [waOwned, setWaOwned] = React.useState(false);
  const [waSessionOk, setWaSessionOk] = React.useState(false);
  const [waJid, setWaJid] = React.useState<string>("");

  const [waDialogVisible, setWaDialogVisible] = React.useState(false);

  const [waCreating, setWaCreating] = React.useState(false);
  const [waDeleting, setWaDeleting] = React.useState(false);
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

      // A criação é sempre do número do PRÓPRIO usuário (o backend isola por
      // usuário), então marcamos como conexão própria.
      setWaConnected(true);
      setWaOwned(true);
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

  // Verificador unificado do status real (consulta o wuzapi via backend).
  // Quando conectado, atualiza o estado e fecha o modal automaticamente.
  // Opções controlam o nível de notificações (silencioso no polling / load).
  const checkStatus = React.useCallback(
    async (opts?: {
      notifyOnSuccess?: boolean;
      notifyWaiting?: boolean;
      notifyError?: boolean;
    }): Promise<boolean> => {
      try {
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

          if (opts?.notifyOnSuccess) {
            const phone = jid ? `+${jid.split("@")[0]}` : "";
            notify(
              "success",
              "WhatsApp conectado!",
              phone ? `Número: ${phone}` : "Conexão estabelecida!"
            );
          }

          return true;
        }

        // connected = false. Se houve erro transitório (wuzapi indisponível), NÃO
        // rebaixamos o estado atual; apenas reportamos quando solicitado.
        if (data?.error) {
          if (opts?.notifyError) {
            notify("error", "WhatsApp", "Falha ao verificar status da conexão.");
          }
          return false;
        }

        if (opts?.notifyWaiting) {
          notify("info", "WhatsApp", "Ainda aguardando a leitura do QR Code.");
        }
        return false;
      } catch (e: any) {
        console.error(e);
        if (opts?.notifyError) {
          notify("error", "WhatsApp", e?.message || "Falha ao verificar status da conexão.");
        }
        return false;
      }
    },
    [notify]
  );

  // Ação manual do botão "Já escaneei" (mantida como fallback).
  const verificarStatusWhatsapp = React.useCallback(async (): Promise<boolean> => {
    try {
      setWaSessionLoading(true);
      return await checkStatus({
        notifyOnSuccess: true,
        notifyWaiting: true,
        notifyError: true,
      });
    } finally {
      setWaSessionLoading(false);
    }
  }, [checkStatus]);

  // Polling automático: enquanto o modal do QR estiver aberto, verificamos o
  // status a cada 3s. Assim que o webhook/wuzapi confirmar a conexão, o modal
  // fecha sozinho — sem precisar clicar em "Já escaneei".
  React.useEffect(() => {
    if (!waDialogVisible) return;

    let cancelled = false;
    const intervalId = setInterval(async () => {
      if (cancelled) return;
      await checkStatus({ notifyOnSuccess: true });
    }, 3000);

    return () => {
      cancelled = true;
      clearInterval(intervalId);
    };
  }, [waDialogVisible, checkStatus]);

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
      setWaOwned(hasConexao ? !!data?.owned : false);

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

  const removeWhatsAppInstance = React.useCallback(async (): Promise<void> => {
    try {
      setWaDeleting(true);

      const url = `${baseURL()}/conexoes/whatsapp/delete`;
      await requestWithLog({
        label: "Excluir Instância WhatsApp",
        method: "POST",
        url,
        body: {},
      });

      setWaDialogVisible(false);
      setWaQr("");
      setWaQrError("");
      setWaConnected(false);
      setWaOwned(false);
      setWaSessionOk(false);
      setWaJid("");

      notify("success", "WhatsApp", "Instância excluída com sucesso.");
    } catch (e: any) {
      console.error(e);
      notify("error", "WhatsApp", e?.message || "Falha ao excluir instância.");
    } finally {
      setWaDeleting(false);
    }
  }, [notify]);

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

  // Ao carregar a página: primeiro lemos o estado persistido (rápido) e, em
  // seguida, confirmamos com o wuzapi de forma silenciosa. Isso "auto-cura" o
  // banco — se a sessão estiver realmente ativa, o conectado_em é regravado e a
  // tela deixa de mostrar "desconectado" indevidamente após alguns dias.
  const loadWhatsapp = React.useCallback(async (): Promise<boolean> => {
    const persisted = await checkConnected();
    const live = await checkStatus();
    return persisted || live;
  }, [checkConnected, checkStatus]);

  return {
    waConnected,
    waOwned,
    waSessionOk,
    waJid,
    waDialogVisible,
    setWaDialogVisible,

    waCreating,
    waDeleting,
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
    removeWhatsAppInstance,
    sendTestMessage,
  };
}