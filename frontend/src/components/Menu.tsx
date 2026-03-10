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
      <>
        <i className={"menu-icons " + item.icon} />
        <strong className="ml-1">{item.label}</strong>
      </>
    );
  };

  return (
    <>
      <br />

      <div className="flex flex-wrap justify-content-center menu-wrapper">
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
        />
      </div>

      <style>{`
        /* desktop: não mexe em nada */

        @media screen and (max-width: 768px) {
          .menu-wrapper {
            width: 100%;
            justify-content: stretch !important;
          }

          .menu-buttons {
            width: 100%;
          }

          .menu-buttons .p-selectbutton {
            width: 100%;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
          }

          .menu-buttons .p-button {
            width: 100%;
            display: flex;
            justify-content: flex-start;
            align-items: center;
            border-radius: 0.9rem !important;
            padding: 0.9rem 1rem !important;
          }

          .menu-buttons .p-button .p-button-label {
            flex: unset;
          }

          .menu-buttons .menu-icons {
            margin-right: 0.15rem;
          }
        }

        @media screen and (max-width: 480px) {
          .menu-buttons .p-button {
            padding: 0.82rem 0.9rem !important;
          }

          .menu-buttons strong {
            font-size: 0.94rem;
          }
        }
      `}</style>
    </>
  );
}