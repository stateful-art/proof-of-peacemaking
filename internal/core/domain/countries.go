package domain

import (
	"encoding/json"
	"os"
)

// CountryData represents the structure of each country in the countries.json file
type CountryData struct {
	Name string `json:"name"`
	Flag string `json:"flag"`
}

// CountriesMap stores the mapping of country codes to their data
var CountriesMap map[string]CountryData

// LoadCountries loads the country data from the countries.json file
func LoadCountries(jsonPath string) error {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	CountriesMap = make(map[string]CountryData)
	if err := json.Unmarshal(data, &CountriesMap); err != nil {
		return err
	}

	return nil
}

// GetCountryList returns a slice of CountryInfo sorted by country name
func GetCountryList() []CountryInfo {
	countries := make([]CountryInfo, 0, len(CountriesMap))
	for code, data := range CountriesMap {
		countries = append(countries, CountryInfo{
			Code: code,
			Name: data.Name,
			Flag: data.Flag,
		})
	}
	return countries
}

// GetCountryInfo returns the CountryInfo for a given country code
func GetCountryInfo(code string) (CountryInfo, bool) {
	if data, ok := CountriesMap[code]; ok {
		return CountryInfo{
			Code: code,
			Name: data.Name,
			Flag: data.Flag,
		}, true
	}
	return CountryInfo{}, false
}
