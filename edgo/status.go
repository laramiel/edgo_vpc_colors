package edgo

type Base struct {
	RAW       string `json:"-"`
	Timestamp string `json:"timestamp"`
	Event     string `json:"event"`
}

type Name_Localized struct {
	Name          string `json:"Name"`
	NameLocalised string `json:"Name_Localised,omitempty"`
}

type Category_Localized struct {
	Category          string `json:"Category"`
	CategoryLocalised string `json:"Category_Localised,omitempty"`
}

type ShipType_Localized struct {
	ShipType          string `json:"ShipType"`
	ShipTypeLocalised string `json:"ShipType_Localised,omitempty"`
}

// 3.1 cargo.json
type Cargo struct {
	Base
	Vessel    string `json:"Vessel"` // Ship | SRV
	Inventory []struct {
		Name_Localized
		Count     int64 `json:"Count"`
		Stolen    int64 `json:"Stolen,omitempty"`
		MissionId int64 `json:"MissionID,omitempty"`
	} `json:"Inventory,omitempty"`
}

// 8.17 market.json
type Market struct {
	Base

	MarketID    int64  `json:"MarketID"`
	StationName string `json:"StationName"`
	StationType string `json:"StationType"`
	StarSystem  string `json:"StarSystem"`
	Items       []struct {
		Name_Localized
		Category_Localized
		ID            int `json:"ID"`
		BuyPrice      int
		SellPrice     int
		MeanPrice     int
		StockBracket  int
		DemandBracket int
		Stock         int
		Demand        int
		Consumer      bool
		Producer      bool
		Rare          bool
	} `json:"MarketItem,omitempty"`
}

// 11.28 modulesinfo.json
type ModulesInfo struct {
	Base
	Modules []struct {
		Slot     string  `json:"Slot"`
		Item     string  `json:"Item"`
		Power    float64 `json:"Power"`
		Priority int64   `json:"Priority"`
	} `json:"Modules"`
}

// navroute.json
type NavRoute struct {
	Base
	Route []struct {
		StarSystem    string     `json:"StarSystem"`
		SystemAddress int64      `json:"SystemAddress"`
		StarPos       [3]float64 `json:"StarPos"`
		StarClass     string     `json:"StarClass"`
	} `json:"Route"`
}

// 8.31 outfitting.json
type Outfitting struct {
	Base
	MarketID    int
	StationName string
	StarSystem  string
	Horizons    bool
	Items       []struct {
		ID       int
		Name     string
		BuyPrice int
	} `json:"Items"`
}

// 8.46 shipyard.json
type Shipyard struct {
	Base
	MarketID       int
	StationName    string
	StarSystem     string
	Horizons       bool
	AllowCobraMkIV bool
	PriceList      []struct {
		ID int
		ShipType_Localized
		ShipPrice int
	} `json:"PriceList"`
}

// 12 status.json
type Status struct {
	Base
	Flags        int     `json:"Flags"`
	Pips         [3]int  `json:"Pips"`
	FireGroup    int     `json:"FireGroup"`
	GuiFocus     int     `json:"GuiFocus"`
	Cargo        float64 `json:"Cargo"`
	LegalState   string  `json:"LegalState"`
	Latitude     float64 `json:"Latitude"`
	Longitude    float64 `json:"Longitude"`
	Altitude     float64 `json:"Altitude"`
	Heading      int64   `json:"Heading"`
	BodyName     string  `json:"BodyName,omitempty"`
	PlanetRadius float64 `json:"PlanetRadius"`
	Fuel         struct {
		FuelMain      float64 `json:"FuelMain"`
		FuelReservoir float64 `json:"FuelReservoir"`
	} `json:"Fuel"`
}
