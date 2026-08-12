import * as React from "react"
import { baseURL } from "../../store/api"
import { user } from "../../store/store"
import { reminderBannerText, type DueReminderRecord } from "./leadReminders"

const POLL_MS = 15_000
const AUTO_ACK_MS = 25_000

async function fetchDueReminders(): Promise<DueReminderRecord[]> {
  const token = user.get().token
  if (!token) return []

  const resp = await fetch(`${baseURL()}/query/crm/reminders-due`, {
    method: "GET",
    headers: {
      Accept: "application/json",
      Authorization: token,
    },
  })
  if (!resp.ok) return []

  const data = await resp.json().catch(() => null)
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.records)) return data.records
  return []
}

async function ackReminder(leadId: string): Promise<void> {
  const token = user.get().token
  if (!token || !leadId) return

  await fetch(`${baseURL()}/crm/reminders/ack`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      Authorization: token,
    },
    body: JSON.stringify({ lead_id: leadId }),
  }).catch(() => undefined)
}

export default function ReminderBanner() {
  const [items, setItems] = React.useState<DueReminderRecord[]>([])
  const ackingRef = React.useRef(false)

  const dismiss = React.useCallback(async (records: DueReminderRecord[]) => {
    if (ackingRef.current || !records.length) return
    ackingRef.current = true
    setItems([])
    try {
      for (const rec of records) {
        const id = String(rec?.lead_id ?? "").trim()
        if (id) await ackReminder(id)
      }
    } finally {
      ackingRef.current = false
    }
  }, [])

  React.useEffect(() => {
    let cancelled = false

    const poll = async () => {
      if (ackingRef.current) return
      const due = await fetchDueReminders()
      if (cancelled) return
      setItems(due)
    }

    poll()
    const timer = window.setInterval(poll, POLL_MS)
    return () => {
      cancelled = true
      window.clearInterval(timer)
    }
  }, [])

  React.useEffect(() => {
    if (!items.length) return
    const timer = window.setTimeout(() => {
      void dismiss(items)
    }, AUTO_ACK_MS)
    return () => window.clearTimeout(timer)
  }, [items, dismiss])

  if (!items.length) return null

  const first = items[0]
  const extra = items.length - 1

  return (
    <div
      className="mx-3 mb-3 p-3 border-round flex align-items-start justify-content-between gap-3"
      style={{
        background: "#fff8e1",
        border: "1px solid #ffe082",
        color: "#5d4037",
      }}
      role="status"
    >
      <div style={{ minWidth: 0 }}>
        <div className="flex align-items-center gap-2 mb-1">
          <i className="pi pi-clock" />
          <strong>Hora de retornar o contato</strong>
        </div>
        <div style={{ fontSize: 14 }}>
          {reminderBannerText(first)}
          {extra > 0 ? ` (+${extra} outro${extra === 1 ? "" : "s"})` : ""}
        </div>
      </div>

      <div className="flex align-items-center gap-2 flex-shrink-0">
        {window.location.pathname.replaceAll("/", "") !== "venda-mais" ? (
          <button
            type="button"
            className="p-button p-button-sm p-button-text"
            onClick={() => window.location.assign("/venda-mais")}
          >
            Abrir Venda Mais
          </button>
        ) : null}
        <button
          type="button"
          className="p-button p-button-sm"
          onClick={() => void dismiss(items)}
        >
          Ok
        </button>
      </div>
    </div>
  )
}
