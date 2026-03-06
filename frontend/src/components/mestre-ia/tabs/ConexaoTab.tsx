import React from "react";
import { Toast } from "primereact/toast";

import { WhatsAppCard } from "../whatsapp/WhatsAppCard";
import { EmailCard } from "../email/EmailCard";
import { useWhatsappConnection } from "../hooks/useWhatsappConnection";

export default function ConexaoTab() {
  const toast = React.useRef<Toast>(null);

  const notify = React.useCallback(
    (
      severity: "success" | "info" | "warn" | "error",
      summary: string,
      detail: string
    ) => {
      toast.current?.show({ severity, summary, detail, life: 3500 });
    },
    []
  );

  const wa = useWhatsappConnection(notify);

  // ✅ no F5: busca no backend e reidrata o estado (exists, etc.)
  React.useEffect(() => {
    wa.loadWhatsapp?.();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="w-full">
      <Toast ref={toast} />

      <div className="flex align-items-center justify-content-between mb-3">
        <div>
          <div style={{ fontSize: 18, fontWeight: 700 }}>Conexões</div>
          <div className="text-secondary" style={{ marginTop: 4 }}>
            Configure as conexões com WhatsApp e E-mail para automações
          </div>
        </div>
      </div>

      <div className="grid align-items-stretch">
        <WhatsAppCard {...wa} />
        <EmailCard />
      </div>
    </div>
  );
}