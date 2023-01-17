package config

type Adapter struct {
	Endpoint         string
	ExtraAdapterInfo string

	// needed for Rubicon
	XAPI AdapterXAPI

	// needed for Facebook
	PlatformID string
	AppSecret  string

	// IX
	SamplingEnabled    bool `mapstructure:"sampling_enabled"`
	SamplingInitial    int  `mapstructure:"sampling_initial"`
	SamplingThereafter int  `mapstructure:"sampling_thereafter"`
}
