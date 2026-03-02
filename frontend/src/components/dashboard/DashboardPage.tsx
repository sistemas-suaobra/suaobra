import { Dropdown } from "primereact/dropdown";
import React, { useState, useEffect, useRef } from "react";
import { Chart } from 'primereact/chart';
import { Chip } from 'primereact/chip';
import { api, makeURL } from "../../store/api";
import { useVariable, type ObjectAny } from "../../utils/interfaces";
import { Lead, User, UserProperties, loadUserState, user, userS } from "../../store/store";
import { makeLeadTooltip } from "../crm/VendaMaisPage";
import { Carousel } from 'primereact/carousel';
import { GetQuotes, type MotivationalQuote } from "../../store/quotes";
import { VerifyPanel, InfoInputPanel, type LoginUserState } from "../login/LoginPage";
import { useHookstate } from "@hookstate/core";
import { Galleria } from 'primereact/galleria';

declare global {
  interface Window {
    rudderAnalytics: any
  }
}

interface Props { }

interface Funnel {
  month: string;
  visit_cnt: number;
  lead_cnt: number;
  opportunity_cnt: number;
  sold_cnt: number;
}

interface HistoryEntry {
  date: string;
  visit_cnt: number;
  visit_cnt_cumulative: number;
  lead_cnt: number;
  lead_cnt_cumulative: number;
}

interface ChartData {
  data: ObjectAny;
  options: ObjectAny;
}

interface UserOption {
  code: string;
  label: string;
  id?: string;
  name?: string;
  email?: string;
}

interface UserData {
  id: string;
  name?: string;
  email?: string;
  [key: string]: any;
}

export default function DashboardPage(props: Props) {
  ///////////////////////////  VARIABLES  ///////////////////////////
  interface File { title: string; url: string }
  const videos: File[] = [
    {
      title: 'Como fazer seu primeiro acesso',
      url: 'https://files.suaobra.com.br/videos/suaobra-primeiro-acesso.mp4',
    },
    {
      title: 'Abordagens: como conseguir resultados usando a abordagem ideal',
      url: 'https://files.suaobra.com.br/videos/suaobra-abordagens.mp4',
    },
    {
      title: 'Como agendar atividade no VENDA MAIS: follow up de orçamentos',
      url: 'https://files.suaobra.com.br/videos/suaobra-agendar-atividade.mp4',
    },
    {
      title: '6 dicas para prospectar',
      url: 'https://files.suaobra.com.br/videos/suaobra-dicas-prospectar.mp4',
    },
    {
      title: '4 Dicas para Prospectar obras nas Redes Sociais',
      url: 'https://files.suaobra.com.br/videos/suaobra-prospectar-redes-sociais.mp4',
    },
    {
      title: 'Persistencia e as Vendas',
      url: 'https://files.suaobra.com.br/videos/suaobra-persistencia.mp4',
    },
    {
      title: 'Taxa de conversao - Script Simples e Eficiente',
      url: 'https://files.suaobra.com.br/videos/suaobra-taxa-conversao.mp4',
    },
    {
      title: 'IA Script com Inteligencia Artificial',
      url: 'https://files.suaobra.com.br/videos/suaobra-ia-script.mp4',
    },
    {
      title: 'Login e Master Equipe',
      url: 'https://files.suaobra.com.br/videos/login-master-equipe.mp4',
    }
  ]

  const files: File[] = [
    {
      title: 'Treinamento básico PROSPECÇÃO DE OBRAS',
      url: 'https://files.suaobra.com.br/pdf/Treinamento-Basico-Prospeccao.pdf',
    },
    {
      title: 'Metodologia da Negociação de Harvard',
      url: 'https://files.suaobra.com.br/pdf/Metodologia-da-Negociacao-de-Harvard.pdf',
    },
    // {
    //   title: 'SCRIPTS PRE PRONTOS 2025',
    //   url: 'https://app.suaobra.com.br/pdfs/Script-2025.pdf',
    // },
    {
      title: 'Funil cheio, vendas explosivas - Prospecção sem desculpas',
      url: 'https://files.suaobra.com.br/pdf/Funil-cheio-vendas-explosivas-Prospeccao-sem-desculpas.pdf',
    },
    {
      title: 'Como Utilizar a Metodologia Funil de Vendas',
      url: 'https://files.suaobra.com.br/pdf/Como-Utilizar-a-Metodologia-Funil-de-Vendas.pdf',
    },
    {
      title: 'Script e Abordagem Profissional e Cliente Final - 5 Modelos',
      url: 'https://files.suaobra.com.br/pdf/script-e-abordagem-profissional-e-cliente-final-5-modelos.pdf',
    }
  ]
  ///////////////////////////  HOOKS  ///////////////////////////
  const [selectedPeriod, setSelectedPeriod] = useState(makePeriodOption(thisMonth()))
  const [selectedUser, setSelectedUser] = useState(makeUserOption("all"))
  const [periodOptions, setPeriodOptions] = useState<string[]>([thisMonth(), prevMonth(), prevMonth2()])
  const [userOptions, setUserOptions] = useState<UserData[]>([])
  const quotes = useVariable(GetQuotes())

  const [leads, setLeads] = useState<Lead[]>([]);
  const [histData, setHistData] = useState<HistoryEntry[]>([]);
  const [funnel, setFunnel] = useState<Funnel>({} as Funnel);
  const [chartData, setChartData] = useState<ChartData>({ data: {}, options: {} });
  const localUserS = useHookstate<LoginUserState>({
    id: '',
    loaded: false,
    verified: false,
    is_manager: false,
    properties: new UserProperties(),
  })

  ///////////////////////////  EFFECTS  ///////////////////////////

  useEffect(() => {
    // randomize quotes
    quotes.set(
      quotes.get()
        .map(value => ({ value, sort: Math.random() }))
        .sort((a, b) => a.sort - b.sort)
        .map(({ value }) => value)
        .filter((_, i) => i < 3)
    )

    window.rudderAnalytics?.page({
      userId: user.get().id,
      category: "dashboard",
      name: window.document.title,
    })

    // Fetch users for dropdown
    getUsers()
  }, []);

  useEffect(() => {
    loadUserState().then(
      user => {
        localUserS.set(
          u => {
            u.id = user.id
            u.verified = user.verified
            u.is_manager = user.manager
            u.properties = user.properties
            u.loaded = true
            return u
          }
        )
      }
    )
  }, []);

  useEffect(() => {
    if (!localUserS.id.get()) return
    getHistory()
    getFunnel()
    getLatestLeads()
  }, [localUserS, selectedPeriod, selectedUser]);

  useEffect(() => {
    setTimeout(() => {
      setChart()
    }, 300);
  }, [histData]);

  ///////////////////////////  FUNCTIONS  ///////////////////////////

  const getUsers = async () => {
    let resp = await api().get(makeURL('/query/dashboard/users'), {})
    if (resp.error) return
    let users = await resp.records() as UserData[]
    if (users) {
      console.log("Fetched users:", users);
      setUserOptions(users)

      persistSelectedUser(users);
    }
  }

  // Helper function to maintain selected user across refreshes
  const persistSelectedUser = (users: UserData[]) => {
    // Skip if "all" is selected or user is not a manager
    if (selectedUser.code === 'all' || !localUserS.is_manager?.get()) return;

    // Find the user in the updated options list
    const foundUser = users.find(u => u.id === selectedUser.code);

    // If found, make sure the selected user data is current
    if (foundUser) {
      setSelectedUser(makeUserOption(foundUser.id, foundUser));
    } else {
      // If user no longer exists in options, reset to "all"
      console.log("Selected user not found in refreshed options, resetting to 'all'");
      setSelectedUser(makeUserOption('all'));
    }
  }

  const getLatestLeads = async () => {
    // For non-managers, always force user_id to be the current user
    const userId = !localUserS.is_manager?.get()
      ? localUserS.id.get()
      : (selectedUser.code === 'all' ? null : selectedUser.code);

    let resp = await api().get(makeURL('/query/dashboard/leads'), {
      month: selectedPeriod.code,
      user_id: userId
    })
    if (resp.error) return
    let leads = (await resp.records() as any[]).map(v => new Lead(v))
    if (leads) setLeads(leads)
  }

  const getFunnel = async () => {
    // For non-managers, always force user_id to be the current user
    const userId = !localUserS.is_manager?.get()
      ? localUserS.id.get()
      : (selectedUser.code === 'all' ? null : selectedUser.code);

    let resp = await api().get(makeURL('/query/dashboard/funnel'), {
      month: selectedPeriod.code,
      user_id: userId
    })
    if (resp.error) return
    let funnels = await resp.records() as Funnel[]
    if (funnels) setFunnel(funnels[0])
  }

  const getHistory = async () => {
    let start_date = selectedPeriod.code
    let end_date = nextMonth()
    if (start_date === prevMonth()) end_date = thisMonth()
    if (start_date === prevMonth2()) end_date = prevMonth()

    // For non-managers, always force user_id to be the current user
    const userId = !localUserS.is_manager?.get()
      ? localUserS.id.get()
      : (selectedUser.code === 'all' ? null : selectedUser.code);

    let resp = await api().get(makeURL('/query/dashboard/history'), {
      start_date,
      end_date,
      user_id: userId
    })
    if (resp.error) return
    let entries = await resp.records() as HistoryEntry[]
    if (entries) {
      if (start_date === thisMonth()) {
        // set null for entries after today
        for (let i = 0; i < entries.length; i++) {
          const element = entries[i];
          if (i > (new Date()).getDate()) {
            entries[i].lead_cnt_cumulative = null
            entries[i].visit_cnt_cumulative = null
          }
        }
      }
      setHistData(entries)
    }
  }

  const setChart = () => {
    const documentStyle = getComputedStyle(document.documentElement);
    const textColor = documentStyle.getPropertyValue('--text-color');
    const textColorSecondary = documentStyle.getPropertyValue('--text-color-secondary');
    const surfaceBorder = documentStyle.getPropertyValue('--surface-border');
    const data = {
      labels: histData.map(v => v.date.replaceAll("T00:00:00Z", "")),
      datasets: [
        {
          label: 'Vizualizações',
          data: histData.map(v => v.visit_cnt_cumulative).filter(v => v !== null),
          fill: true,
          borderColor: documentStyle.getPropertyValue('--orange-500'),
          tension: 0.4,
          backgroundColor: 'rgba(255,167,38,0.2)'
        },
        {
          label: 'Leads',
          data: histData.map(v => v.lead_cnt_cumulative).filter(v => v !== null),
          fill: true,
          borderColor: documentStyle.getPropertyValue('--blue-500'),
          tension: 0.4,
          backgroundColor: 'rgba(255,167,38,0.2)'
        },
      ]
    };

    const options = {
      maintainAspectRatio: true,
      aspectRatio: 4,
      plugins: {
        legend: {
          labels: {
            color: textColor
          }
        }
      },
      scales: {
        x: {
          ticks: {
            color: textColorSecondary
          },
          grid: {
            color: surfaceBorder
          }
        },
        y: {
          min: 0,
          ticks: {
            color: textColorSecondary
          },
          grid: {
            color: surfaceBorder
          }
        }
      }
    };

    setChartData({ data, options });
  }

  const funnelNumber = (n: number) => {
    return !n || n === 0 ? '-' : n.toString()
  }

  ///////////////////////////  JSX  ///////////////////////////

  const ArrowNumber = (numerator: number, denominator: number) => {
    let percent = (!denominator || !numerator || denominator === 0 ? 0 : numerator / denominator) * 100
    return <span
      style={{
        position: 'absolute',
        right: '-8%',
        top: '50%',
        marginTop: '-15px',
        zIndex: 1,
      }}
    >
      <div className="bg-black-alpha-90 text-white py-2 px-2 arrow-box text-xs w-3rem">
        {Math.round(percent)}%
      </div>
    </span>

  }

  const leadChip = (lead: Lead) => {
    const id = lead.obra_id
    return <span key={id}>
      {makeLeadTooltip(lead)}

      <Chip
        id={id}
        onClick={() => window.location.assign('venda-mais')}
        icon="pi pi-home"
        label={lead.title}
        style={{ cursor: 'pointer' }}
        className="ml-2 mt-2"
      />
    </span>
  }

  const quoteTemplate = (item: MotivationalQuote) => {
    return (
      <div className="border-1 surface-border border-round m-2 text-center py-5 px-3">
        <h3>
          {item.frase}
        </h3>
        <h4>
          - {item.autor}
        </h4>
      </div>
    );
  };

  const videoTemplate = (video: File) => {
    return <div>
      <div className="text-center">
        <h3>{video.title}</h3>
      </div>
      <video width="100%" height="100%" loop controls preload="metadata" key={video.url}>
        <source src={video.url} type="video/mp4" />
      </video>
    </div>
  };

  return (
    <div className="justify-content-center w-full border-circle bg-tr">
      <InfoInputPanel lus={localUserS} />
      <VerifyPanel lus={localUserS} />

      <div className=" bg-white border-round-3xl mx-2 p-4">
        <div className="flex md:col-6 col-12">
          <h2 className="mt-0">Resumo</h2>
        </div>

        <div className="formgrid grid">
          <div className="field md:col-6 col-12">
            <Dropdown
              id='date-dropdown'
              value={selectedPeriod.code}
              onChange={(e) => {
                setSelectedPeriod(makePeriodOption(e.value))
              }}
              options={makePeriodOptions(periodOptions)}
              optionValue="code"
              optionLabel="label"
              className="w-full"
            />
          </div>
          {(localUserS.is_manager?.get() === true) && (
            <div className="field md:col-6 col-12">
              <Dropdown
                id='user-dropdown'
                value={selectedUser.code}
                onChange={(e) => {
                  const selectedId = e.value;
                  console.log("Dropdown selection changed to:", selectedId);
                  const foundUser = userOptions.find(u => u.id === selectedId);
                  console.log("Found user:", foundUser);
                  setSelectedUser(makeUserOption(selectedId, foundUser));
                }}
                options={makeUserOptions(userOptions)}
                optionValue="code"
                optionLabel="label"
                className="w-full"
              />
            </div>
          )}
        </div>


        <div className="grid gap-2">
          <div className="col relative">
            <div className="text-center py-5 p-3 border-round-lg bg-primary font-bold">
              <div className="text-lg uppercase pb-2">Vizualizações</div>
              <div className="text-5xl">{funnelNumber(funnel.visit_cnt)}</div>
            </div>

            {ArrowNumber(funnel.lead_cnt, funnel.visit_cnt)}
          </div>

          <div className="col relative">
            <div className="text-center py-5 p-3 border-round-lg bg-primary font-bold ">
              <div className="text-lg uppercase pb-2">Leads</div>
              <div className="text-5xl">{funnelNumber(funnel.lead_cnt)}</div>
            </div>

            {ArrowNumber(funnel.opportunity_cnt, funnel.lead_cnt)}
          </div>

          <div className="col relative">
            <div className="text-center py-5 p-3 border-round-lg bg-primary font-bold ">
              <div className="text-lg uppercase pb-2">Oportunidades</div>
              <div className="text-5xl">{funnelNumber(funnel.opportunity_cnt)}</div>
            </div>

            {ArrowNumber(funnel.sold_cnt, funnel.opportunity_cnt)}
          </div>

          <div className="col relative">
            <div className="text-center py-5 p-3 border-round-lg bg-primary font-bold ">
              <div className="text-lg uppercase pb-2">Vendas</div>
              <div className="text-5xl">{funnelNumber(funnel.sold_cnt)}</div>
            </div>
          </div>

        </div>

      </div>

      <div className="mx-2 my-4 border-bottom-0" />

      <div className="mx-2 bg-white border-round-3xl px-5 pb-3 col-12">
        <div className="flex col-12">
          <h2>Historico</h2>
        </div>

        <Chart type="line" data={chartData.data} options={chartData.options} />
      </div>

      <div className="mx-2 my-4 border-bottom-0" />

      <div className="mx-2 bg-white border-round-3xl px-5 pb-3 col-12">
        <div className="flex col-12">
          <h2>Vídeos de Treinamento</h2>
        </div>


        <Galleria
          value={videos}
          // style={{ maxWidth: '640px' }} 
          item={videoTemplate}
          showThumbnails={false}
          showIndicators
        />
      </div>

      <div className="mx-2 my-4 border-bottom-0" />

      <div className="mx-2 bg-white border-round-3xl px-5 pb-3 col-12">
        <div className="flex col-12">
          <h2>Material de Apoio</h2>
        </div>

        <div className="grid w-full">
          {
            files.map((file, index) => {
              return <div key={index} className="flex align-items-center lg:col-4 md:col-6 col-12">
                <div><i className="pi pi-file-pdf" style={{ fontSize: '2.5rem' }}></i></div>
                <div> <a href={file.url} target="_blank">{file.title}</a></div>
              </div>
            })
          }
        </div>
      </div>

      <div className="mx-2 my-4 border-bottom-0" />

      <div className="mx-2 bg-white border-round-3xl px-5 pb-3">
        <div className='pt-2'>
          <h2>Últimas Obras Convertidas em Leads</h2>
        </div>
        {
          leads?.map(lead => leadChip(lead))
        }
      </div>

      <div className="mx-2 bg-white border-round-3xl px-5 pb-3 mt-4">
        <Carousel
          value={quotes.get()}
          numVisible={1}
          className="pt-4"
          circular
          autoplayInterval={5000}
          itemTemplate={quoteTemplate}
        />
      </div>
    </div>
  );
};


const makePeriodOption = (options: string) => {
  return { code: options, label: dateLongString(options) }
}

const makePeriodOptions = (options: string[]) => {
  let opts = []
  for (let o of options) {
    opts.push(makePeriodOption(o))
  }
  return opts
}

const nextMonth = () => {
  let d = (new Date())
  d.setDate(32)
  d.setHours(1)
  return dateString(d)
}

const thisMonth = () => {
  let d = (new Date())
  d.setDate(1)
  d.setHours(0)
  return dateString(d)
}

const prevMonth = () => {
  let d = (new Date())
  d.setDate(0)
  d.setDate(1)
  d.setHours(0)
  return dateString(d)
}

const prevMonth2 = () => {
  let d = (new Date())
  d.setDate(0)
  d.setDate(0)
  d.setDate(1)
  d.setHours(0)
  return dateString(d)
}

const dateString = (d: Date) => {
  return d.toISOString().split('T')[0]
}

const dateLongString = (ds: string) => {
  let d = new Date(ds)
  let month = monthMap[d.getUTCMonth() + 1] as string
  let year = d.getFullYear().toString()
  return `${month} ${year}`
}


const monthMap = {
  1: 'Janeiro',
  2: 'Fevereiro',
  3: 'Março',
  4: 'Abril',
  5: 'Maio',
  6: 'Junho',
  7: 'Julho',
  8: 'Agosto',
  9: 'Setembro',
  10: 'Outubro',
  11: 'Novembro',
  12: 'Dezembro',
}

const makeUserOption = (code: string = 'all', user?: UserData): UserOption => {
  if (code === 'all' || !user) {
    return { code: 'all', label: 'Todos' }
  }
  const name = user.name || user.email || 'User'
  const email = user.email || ''
  return {
    code: user.id,
    label: email ? `${name} (${email})` : name,
    id: user.id,
    name: name,
    email: email
  }
}

const makeUserOptions = (users: UserData[]): UserOption[] => {
  // Always start with the "All" option
  let opts = [makeUserOption('all')];

  // Add each user option
  if (users && users.length > 0) {
    users.forEach(user => {
      if (user && user.id) {
        opts.push(makeUserOption(user.id, user));
      }
    });
  }

  console.log("Generated user options:", opts);
  return opts;
} 