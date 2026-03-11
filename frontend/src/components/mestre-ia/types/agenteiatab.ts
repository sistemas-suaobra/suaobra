import type { Intent } from "../tabs/IntentDialog";

export type AgenteIaTabProps = {
  intents: Intent[];
  onNew: () => void;
  onEdit: (intent: Intent) => void;
  onDelete: (id: string) => void;
  onToggleActive: (id: string, next: boolean) => void;
};
