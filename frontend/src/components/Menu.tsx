import * as React from "react";
import { SelectButton } from 'primereact/selectbutton';
import { isLoaded } from '../store/store.js';

interface Props {}

export default function Menu(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  const items = [
      {url: '/dashboard', label: 'Dashboard', icon: 'pi pi-fw pi-home'},
      {url: '/obras-plus', label: 'Obras+', icon: 'pi pi-fw pi-database'},
      // {url: '/info-obras', label: 'InfoObras', icon: 'pi pi-fw pi-chart-line'},
      {url: '/venda-mais', label: 'Venda Mais', icon: 'pi pi-fw pi-dollar'},
      // {url: '/automacao', label: 'Automação', icon: 'pi pi-fw pi-sync'},
      {url: '/mestre-ia', label: 'Mestre IA', icon: 'pi pi-fw pi-android'}
      // {url: '/clube', label: 'Clube', icon: 'pi pi-fw pi-car'}
  ];

  ///////////////////////////  HOOKS  ///////////////////////////
  const [value, setValue] = React.useState(items[0]);


  ///////////////////////////  EFFECTS  ///////////////////////////
  React.useEffect(() => {
    setValue(items.filter(i => i.url.replaceAll('/','') === location.pathname.replaceAll('/',''))[0])
    isLoaded.set(true)
  }, [])

  ///////////////////////////  FUNCTIONS  ///////////////////////////
  ///////////////////////////  JSX  ///////////////////////////

  const itemTemplate = (item: any) => {
    return <>
      <i className={'menu-icons ' + item.icon}/>
      <strong className="ml-1">{item.label}</strong>
    </>
  }

  return (
  <>  
    <br/>
      <div className="flex flex-wrap justify-content-center">
      <SelectButton 
        className="menu-buttons"
        value={value}
        itemTemplate={itemTemplate}
        onChange={(e) => {
          if(e?.value?.url) location.assign(e.value.url)
          // setValue(e.value)
        }}
        options={items}
      />
    </div>
    {/* <br/> */}
  </>
  )
};