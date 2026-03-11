import React from "react"
import { useStore } from "@nanostores/react"
import { obrasPlusCity, loadUserState } from "../../../store/store"
import { makeCity, type City } from "../../../store/cities.js"
import { api, baseURL } from "../../../store/api"

export function useObrasPlusFilters(visible: boolean) {
  const $obrasPlusCity = useStore(obrasPlusCity)

  const [citiesOptions, setCitiesOptions] = React.useState<City[]>([])
  const [selectedCity, setSelectedCity] = React.useState<City | null>($obrasPlusCity || null)

  const [selectedNeighborhood, setSelectedNeighborhood] = React.useState<any[]>([])
  const [neighborhoodsOptions, setNeighborhoodsOptions] = React.useState<any[]>([])

  const [filterValue, setFilterValue] = React.useState("")
  const [startDateFrom, setStartDateFrom] = React.useState("")
  const [startDateTo, setStartDateTo] = React.useState("")
  const [endDateFrom, setEndDateFrom] = React.useState("")
  const [endDateTo, setEndDateTo] = React.useState("")

  const resetFilters = React.useCallback(() => {
    setSelectedNeighborhood([])
    setFilterValue("")
    setStartDateFrom("")
    setStartDateTo("")
    setEndDateFrom("")
    setEndDateTo("")
  }, [])

  React.useEffect(() => {
    loadUserState().then((userData) => {
      const cities = userData.team?.cities?.sort().map((id: string) => makeCity(id)) || []
      setCitiesOptions(cities)
      if (!cities.length) return

      let savedCity = localStorage.getItem("obrasPlusCity") || ""
      if (!cities.map((c: City) => c.id).includes(savedCity)) {
        savedCity = cities[0]?.id
        if (savedCity) localStorage.setItem("obrasPlusCity", savedCity)
      }

      if (!savedCity) return
      const city = makeCity(savedCity)
      setSelectedCity(city)
      obrasPlusCity.set(city)
    })
  }, [])

  React.useEffect(() => {
    if (!selectedCity) return
    const fetchNeighborhoods = async () => {
      try {
        const resp = await api().get(`${baseURL()}/query/obras-plus-neighborhood`, { city: selectedCity.city || "" })
        if (resp.error) throw new Error(resp.error)
        const data = await resp.response.json()
        const barrios = (data.barrios as string[]) || []
        setNeighborhoodsOptions(barrios.map((b) => ({ bairro: b })))
      } catch {
        setNeighborhoodsOptions([])
      }
    }
    fetchNeighborhoods()
  }, [selectedCity])

  React.useEffect(() => {
    if (!visible) return
    resetFilters()
  }, [visible, resetFilters])

  const onCityChange = (city: City) => {
    setSelectedCity(city)
    setSelectedNeighborhood([])
    if (city?.id) localStorage.setItem("obrasPlusCity", city.id)
    obrasPlusCity.set(city)
  }

  return {
    citiesOptions,
    selectedCity,
    onCityChange,
    selectedNeighborhood,
    setSelectedNeighborhood,
    neighborhoodsOptions,
    filterValue,
    setFilterValue,
    startDateFrom,
    setStartDateFrom,
    startDateTo,
    setStartDateTo,
    endDateFrom,
    setEndDateFrom,
    endDateTo,
    setEndDateTo,
  }
}