import type { Intent } from "../tabs/IntentDialog";

export type AgenteIaTabProps = {
  intents: Intent[];
  loading?: boolean;
  onNew: () => void;
  onEdit: (intent: Intent) => void;
  onDelete: (id: string) => void;
  onToggleActive: (id: string, next: boolean) => void;
};
