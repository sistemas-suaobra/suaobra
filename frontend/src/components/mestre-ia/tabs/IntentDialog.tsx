import React from "react";
import { Dialog } from "primereact/dialog";
import { Button } from "primereact/button";
import { InputText } from "primereact/inputtext";
import { InputTextarea } from "primereact/inputtextarea";
import { Chip } from "primereact/chip";
import { InputSwitch } from "primereact/inputswitch";
import { Toast } from "primereact/toast";

export type Intent = {
  id: string;
  titulo: string;
  ativo: boolean;
  keywords: string[];
  respostaBase: string;
};

function normalizeText(s: string) {
  return (s || "")
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
}

function uniqKeywords(words: string[]) {
  const map = new Map<string, string>();
  for (const w of words) {
    const trimmed = (w || "").trim();
    if (!trimmed) continue;
    map.set(normalizeText(trimmed), trimmed);
  }
  return Array.from(map.values());
}

type IntentDialogProps = {
  visible: boolean;
  editing: Intent | null;
  initialValues?: Partial<Intent>;
  onClose: () => void;
  onSave: (payload: Omit<Intent, "id">) => void;
  toastRef?: React.RefObject<Toast>;
};

export function IntentDialog(props: IntentDialogProps) {
  const { visible, editing, onClose, onSave, toastRef } = props;

  const [titulo, setTitulo] = React.useState("");
  const [ativo, setAtivo] = React.useState(true);
  const [keywordInput, setKeywordInput] = React.useState("");
  const [keywords, setKeywords] = React.useState<string[]>([]);
  const [respostaBase, setRespostaBase] = React.useState("");

  const notify = (
    severity: "success" | "info" | "warn" | "error",
    summary: string,
    detail: string
  ) => {
    toastRef?.current?.show({ severity, summary, detail, life: 3000 });
  };

  const reset = React.useCallback(() => {
    setTitulo("");
    setAtivo(true);
    setKeywordInput("");
    setKeywords([]);
    setRespostaBase("");
  }, []);

  React.useEffect(() => {
    if (!visible) return;

    if (editing) {
      setTitulo(editing.titulo);
      setAtivo(editing.ativo);
      setKeywords(editing.keywords || []);
      setRespostaBase(editing.respostaBase || "");
      setKeywordInput("");
      return;
    }

    reset();
    if (props.initialValues) {
      if (props.initialValues.titulo) setTitulo(props.initialValues.titulo);
      if (typeof props.initialValues.ativo === "boolean") setAtivo(props.initialValues.ativo);
      if (props.initialValues.keywords) setKeywords(props.initialValues.keywords);
      if (props.initialValues.respostaBase) setRespostaBase(props.initialValues.respostaBase);
    }
  }, [visible, editing, reset]);

  const addKeyword = () => {
    const raw = keywordInput.trim();
    if (!raw) return;

    const parts = raw
      .split(",")
      .map((p) => p.trim())
      .filter(Boolean);

    setKeywords((prev) => uniqKeywords([...prev, ...parts]));
    setKeywordInput("");
  };

  const removeKeyword = (word: string) => {
    setKeywords((prev) => prev.filter((k) => normalizeText(k) !== normalizeText(word)));
  };

  const handleSave = () => {
    if (!titulo.trim()) return notify("warn", "Faltando", "Informe o título da intenção.");
    if (!respostaBase.trim()) return notify("warn", "Faltando", "Informe a resposta base.");
    if (!keywords.length) return notify("warn", "Faltando", "Adicione pelo menos 1 palavra-chave.");

    onSave({
      titulo: titulo.trim(),
      ativo,
      keywords: uniqKeywords(keywords),
      respostaBase: respostaBase.trim(),
    });
  };

  return (
    <Dialog
      visible={visible}
      header={editing ? "Editar intenção" : "Nova intenção"}
      style={{ width: "760px", maxWidth: "95vw" }}
      onHide={onClose}
      draggable={false}
      dismissableMask
      footer={
        <div className="flex justify-content-end gap-2">
          <Button label="Salvar" icon="pi pi-check" onClick={handleSave} />
          <Button label="Cancelar" icon="pi pi-times" severity="secondary" onClick={onClose} />
        </div>
      }
    >
      <div className="formgrid grid">
        <div className="field col-12 md:col-8">
          <label>Título</label>
          <InputText className="w-full" value={titulo} onChange={(e) => setTitulo(e.target.value)} />
        </div>

        <div className="field col-12 md:col-4">
          <label>&nbsp;</label>
          <div className="flex align-items-center gap-2" style={{ marginTop: 6 }}>
            <InputSwitch checked={ativo} onChange={(e) => setAtivo(!!e.value)} />
            <span className="text-secondary">{ativo ? "Ativo" : "Inativo"}</span>
          </div>
        </div>

        <div className="field col-12">
          <label>Palavras-chave</label>
          <div className="p-inputgroup">
            <InputText
              className="w-full"
              placeholder="Digite e pressione Enter (ou separe por vírgula)"
              value={keywordInput}
              onChange={(e) => setKeywordInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  addKeyword();
                }
              }}
            />
            <Button icon="pi pi-plus" onClick={addKeyword} />
          </div>

          <div className="mt-2 flex flex-wrap gap-2">
            {keywords.map((k, idx) => (
              <Chip
                key={`kw-${idx}`}
                label={k}
                removable
                onRemove={() => removeKeyword(k)}
                className="border-round-xl"
              />
            ))}
          </div>
        </div>

        <div className="field col-12">
          <label>Resposta base</label>
          <InputTextarea
            className="w-full"
            rows={6}
            placeholder="Escreva a resposta que o agente deve usar quando esta intenção for detectada..."
            value={respostaBase}
            onChange={(e) => setRespostaBase(e.target.value)}
          />
        </div>
      </div>
    </Dialog>
  );
}