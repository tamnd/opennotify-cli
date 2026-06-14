package opennotify

// Position holds the current coordinates of the ISS.
type Position struct {
	Latitude  string `kit:"id" json:"latitude"`
	Longitude string `json:"longitude"`
	Timestamp int64  `json:"timestamp"`
}

// Astronaut is one person currently in space.
type Astronaut struct {
	Name  string `kit:"id" json:"name"`
	Craft string `json:"craft"`
}

// --- unexported wire types for JSON decoding ---

type wirePosition struct {
	ISSPosition struct {
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	} `json:"iss_position"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type wireAstros struct {
	People []struct {
		Name  string `json:"name"`
		Craft string `json:"craft"`
	} `json:"people"`
	Number  int    `json:"number"`
	Message string `json:"message"`
}
