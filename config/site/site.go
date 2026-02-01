package site

// Target sites to fetch data from
var TargetSites = []SiteConfig{
	{ID: "NE=50143101", Name: "SHUNDAO 1"},
	{ID: "NE=50127734", Name: "SHUNDAO 2"},
}

// SiteConfig represents a site configuration
type SiteConfig struct {
	ID   string
	Name string
}

// EMICDevice represents an EMIC device configuration
type EMICDevice struct {
	ID       string
	NodeName string
	SiteID   string
	SiteName string
}

// EMICDevices are the specific EMIC devices to fetch data from
var EMICDevices = []EMICDevice{
	// SHUNDAO 1
	{ID: "NE=62950968", NodeName: "EMIC_Station1", SiteID: "NE=50143101", SiteName: "SHUNDAO1"},
	{ID: "NE=62974616", NodeName: "EMIC_Station2", SiteID: "NE=50143101", SiteName: "SHUNDAO1"},
	{ID: "NE=50184285", NodeName: "EMIC_Station3", SiteID: "NE=50143101", SiteName: "SHUNDAO1"},
	// SHUNDAO 2
	{ID: "NE=50145414", NodeName: "EMIC_Master", SiteID: "NE=50127734", SiteName: "SHUNDAO2"},
}

// PowerMeterDevice represents a Power Meter device configuration
type PowerMeterDevice struct {
	ID       string
	NodeName string
	SiteName string
}

// PowerMeterDevices are the Power Meter devices to fetch data from
var PowerMeterDevices = []PowerMeterDevice{
	// SHUNDAO 1 (12 devices)
	{ID: "NE=50143100", NodeName: "PowerMeter_01", SiteName: "SHUNDAO1"},
	{ID: "NE=50184269", NodeName: "PowerMeter_02", SiteName: "SHUNDAO1"},
	{ID: "NE=50173724", NodeName: "PowerMeter_03", SiteName: "SHUNDAO1"},
	{ID: "NE=50184278", NodeName: "PowerMeter_04", SiteName: "SHUNDAO1"},
	{ID: "NE=50184272", NodeName: "PowerMeter_05", SiteName: "SHUNDAO1"},
	{ID: "NE=50184271", NodeName: "PowerMeter_06", SiteName: "SHUNDAO1"},
	{ID: "NE=50196767", NodeName: "PowerMeter_07", SiteName: "SHUNDAO1"},
	{ID: "NE=50164162", NodeName: "PowerMeter_08", SiteName: "SHUNDAO1"},
	{ID: "NE=50184308", NodeName: "PowerMeter_09", SiteName: "SHUNDAO1"},
	{ID: "NE=50184307", NodeName: "PowerMeter_10", SiteName: "SHUNDAO1"},
	{ID: "NE=50184306", NodeName: "PowerMeter_11", SiteName: "SHUNDAO1"},
	{ID: "NE=50184305", NodeName: "PowerMeter_12", SiteName: "SHUNDAO1"},
	// SHUNDAO 2 (13 devices)
	{ID: "NE=50145427", NodeName: "PowerMeter_01", SiteName: "SHUNDAO2"},
	{ID: "NE=50185915", NodeName: "PowerMeter_02", SiteName: "SHUNDAO2"},
	{ID: "NE=50185595", NodeName: "PowerMeter_03", SiteName: "SHUNDAO2"},
	{ID: "NE=50185594", NodeName: "PowerMeter_04", SiteName: "SHUNDAO2"},
	{ID: "NE=50185923", NodeName: "PowerMeter_05", SiteName: "SHUNDAO2"},
	{ID: "NE=50185851", NodeName: "PowerMeter_06", SiteName: "SHUNDAO2"},
	{ID: "NE=50196837", NodeName: "PowerMeter_07", SiteName: "SHUNDAO2"},
	{ID: "NE=50196862", NodeName: "PowerMeter_08", SiteName: "SHUNDAO2"},
	{ID: "NE=50185301", NodeName: "PowerMeter_09", SiteName: "SHUNDAO2"},
	{ID: "NE=50185288", NodeName: "PowerMeter_10", SiteName: "SHUNDAO2"},
	{ID: "NE=50185269", NodeName: "PowerMeter_11", SiteName: "SHUNDAO2"},
	{ID: "NE=50185353", NodeName: "PowerMeter_12", SiteName: "SHUNDAO2"},
	{ID: "NE=50185294", NodeName: "PowerMeter_13", SiteName: "SHUNDAO2"},
}

// OutputDir is the directory to save JSON files
const OutputDir = "output"

// APIEndpoint for fetching site tree data
const APIEndpoint = "https://intl.fusionsolar.huawei.com/rest/dp/pvms/organization/v1/locate-tree"

// SubNodeTypeIDs for API request
const SubNodeTypeIDs = "20801,20800,20811,20814,20815,20816,20819,20821,20822,20824,20835,20836,20837,20838,20847,60014,60022,60066,20844,60080,60026,59986,60044,60043,60001,60002,60003,60010,60015,69999,60092,60067"

// TypeIDInclude for API request
const TypeIDInclude = "23089,23091"


