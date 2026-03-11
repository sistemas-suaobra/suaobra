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

  React.useEffect(() => {
    wa.loadWhatsapp?.();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="conexao-tab w-full">
      <Toast ref={toast} />

      <div className="flex flex-column md:flex-row md:align-items-center md:justify-content-between gap-3 mb-3">
        <div className="min-w-0">
          <div style={{ fontSize: 18, fontWeight: 700 }}>Conexões</div>
          <div className="text-secondary" style={{ marginTop: 4, lineHeight: 1.5 }}>
            Configure as conexões com WhatsApp e E-mail para automações
          </div>
        </div>
      </div>

      <div className="grid align-items-stretch">
        <WhatsAppCard {...wa} />
        <EmailCard />
      </div>

      <style>{`
        .conexao-tab .grid {
          margin-top: 0;
        }

        @media screen and (max-width: 768px) {
          .conexao-tab {
            overflow-x: hidden;
          }
        }
      `}</style>
    </div>
  );
}