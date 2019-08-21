package main

type Location struct {
	Latitude    float64 `json:"Lat"`
	Longitude   float64 `json:"Lng"`
	Istat       string  `json:"IstatComune"`
	Name        string  `json:"Nome"`
	Description string  `json:"Descrizione"`
	Address     string  `json:"Indirizzo"`
}

func (loc Location) Point() map[string]interface{} {
	return map[string]interface{}{
		"Formato": 0,
		"Lat":     loc.Latitude,
		"Lng":     loc.Longitude,
	}
}
