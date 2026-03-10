import * as React from "react";
import { SelectButton } from "primereact/selectbutton";
import { isLoaded } from "../store/store.js";

interface Props {}

export default function Menu(props: Props) {
  const items = [
    { url: "/dashboard", label: "Dashboard", icon: "pi pi-fw pi-home" },
    { url: "/obras-plus", label: "Obras+", icon: "pi pi-fw pi-database" },
    { url: "/venda-mais", label: "Venda Mais", icon: "pi pi-fw pi-dollar" },
    { url: "/mestre-ia", label: "Mestre IA", icon: "pi pi-fw pi-android" },
  ];

  const [value, setValue] = React.useState(items[0]);

  React.useEffect(() => {
    const current =
      items.find(
        (i) => i.url.replaceAll("/", "") === location.pathname.replaceAll("/", "")
      ) || items[0];

    setValue(current);
    isLoaded.set(true);
  }, []);

  const itemTemplate = (item: any) => {
    return (
      <div className="menu-item-content">
        <i className={`menu-icons ${item.icon}`} />
        <strong className="menu-label">{item.label}</strong>
      </div>
    );
  };

  return (
    <>
      <div className="menu-wrapper">
        <SelectButton
          className="menu-buttons"
          value={value}
          itemTemplate={itemTemplate}
          onChange={(e) => {
            if (e?.value?.url) {
              setValue(e.value);
              location.assign(e.value.url);
            }
          }}
          options={items}
          optionLabel="label"
        />
      </div>

      <style>{`
        .menu-wrapper {
          width: 100%;
          margin: 0 0 1rem 0;
        }

        .menu-buttons {
          width: 100%;
        }

        .menu-buttons .p-selectbutton {
          width: 100%;
          display: flex;
          flex-wrap: wrap;
          gap: 0.75rem;
          background: transparent;
        }

        .menu-buttons .p-button {
          flex: 1 1 calc(25% - 0.75rem);
          min-width: 180px;
          border-radius: 1rem !important;
          border: 1px solid rgba(0, 0, 0, 0.08) !important;
          background: #ffffff !important;
          color: #4b5563 !important;
          box-shadow: none !important;
          padding: 1rem 1.1rem !important;
          transition: all 0.2s ease;
          justify-content: flex-start !important;
        }

        .menu-buttons .p-button:not(.p-highlight):hover {
          background: #f8fafc !important;
          border-color: rgba(91, 92, 226, 0.18) !important;
        }

        .menu-buttons .p-button.p-highlight {
          background: linear-gradient(90deg, #6d5dfc 0%, #5b5ce2 100%) !important;
          border-color: transparent !important;
          color: #ffffff !important;
        }

        .menu-buttons .p-button:focus {
          box-shadow: none !important;
        }

        .menu-item-content {
          width: 100%;
          display: flex;
          align-items: center;
          justify-content: flex-start;
          gap: 0.65rem;
          text-align: left;
        }

        .menu-icons {
          font-size: 1rem;
          line-height: 1;
        }

        .menu-label {
          font-size: 0.98rem;
          font-weight: 700;
          line-height: 1.2;
        }

        @media screen and (max-width: 992px) {
          .menu-buttons .p-button {
            flex: 1 1 calc(50% - 0.75rem);
            min-width: unset;
          }
        }

        @media screen and (max-width: 768px) {
          .menu-wrapper {
            margin-bottom: 0.85rem;
          }

          .menu-buttons .p-selectbutton {
            flex-direction: column;
            gap: 0.5rem;
          }

          .menu-buttons .p-button {
            width: 100%;
            flex: 1 1 100%;
            min-width: 100%;
            border-radius: 0.85rem !important;
            padding: 0.95rem 1rem !important;
          }

          .menu-item-content {
            gap: 0.75rem;
          }

          .menu-icons {
            font-size: 0.95rem;
          }

          .menu-label {
            font-size: 0.95rem;
          }
        }

        @media screen and (max-width: 480px) {
          .menu-buttons .p-button {
            padding: 0.9rem 0.95rem !important;
          }

          .menu-label {
            font-size: 0.92rem;
          }
        }
      `}</style>
    </>
  );
}