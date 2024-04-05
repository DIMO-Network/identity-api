package shared

// VINs are not allowed to contain I, O, or Q.
// In the year digit we additionally disallow U, Z, and 0.
// https://en.wikibooks.org/wiki/Vehicle_Identification_Numbers_(VIN_codes)/Model_year
var years = map[byte]int{
	'5': 2005,
	'6': 2006,
	'7': 2007,
	'8': 2008,
	'9': 2009,
	'A': 2010,
	'B': 2011,
	'C': 2012,
	'D': 2013,
	'E': 2014,
	'F': 2015,
	'G': 2016,
	'H': 2017,
	'J': 2018,
	'K': 2019,
	'L': 2020,
	'M': 2021,
	'N': 2022,
	'P': 2023,
	'R': 2024,
	'S': 2025,
	'T': 2026,
	'V': 2027,
	'W': 2028,
	'X': 2029,
	'Y': 2030,
	'1': 2031,
	'2': 2032,
	'3': 2033,
	'4': 2034,
}

type VIN string

// Year will decode the year portion from the VIN from 2005 to 2034. returns 0 if out of range. VIN nomenclature only allows for 30 year timespans, then repeats, ie. 2023 = 1993
func (v VIN) Year() int {
	return years[v[9]]
}

func (v VIN) String() string {
	return string(v)
}

var teslaModelByDigit = map[byte]string{
	'C': "Cybertruck",
	'R': "Roadster",
	'S': "Model S",
	'X': "Model X",
	'Y': "Model Y",
	'3': "Model 3",
}

func (v VIN) TeslaModel() string {
	// Maybe some error handling?
	return teslaModelByDigit[v[3]]
}

func (v VIN) Wmi() string {
	return string(v[0:3])
}

func (v VIN) VDS() string {
	return string(v[3:8])
}

func (v VIN) VIS() string {
	return string(v[9:17])
}

func (v VIN) CheckDigit() string {
	return string(v[8:9])
}

func (v VIN) SerialNumber() string {
	return string(v[11:17])
}
