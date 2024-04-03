package openTripMap

var Otm *OtmClient

func NewClient(apiKey string) {
	Otm = &OtmClient{
		ApiKey: apiKey,
		Cache:  make(map[string]FeatureCollection),
	}
}
