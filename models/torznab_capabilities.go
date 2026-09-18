package models

import "encoding/xml"

type TorznabCategory struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type SupportedSearchCapability struct {
	Text            string `xml:",chardata"`
	Available       string `xml:"available,attr"`
	SupportedParams string `xml:"supportedParams,attr"`
}

type CapabilitiesServer struct {
	Text      string `xml:",chardata"`
	Version   string `xml:"version,attr"`
	Title     string `xml:"title,attr"`
	Strapline string `xml:"strapline,attr"`
	Email     string `xml:"email,attr"`
	URL       string `xml:"url,attr"`
	Image     string `xml:"image,attr"`
}

type CapabilitiesLimits struct {
	Text    string `xml:",chardata"`
	Max     string `xml:"max,attr"`
	Default string `xml:"default,attr"`
}

type CapabilitiesRetention struct {
	Text string `xml:",chardata"`
	Days string `xml:"days,attr,omitempty"`
}

type CapabilitiesRegistration struct {
	Text      string `xml:",chardata"`
	Available string `xml:"available,attr"`
	Open      string `xml:"open,attr"`
}

type CategoryWithSubcat struct {
	Text   string            `xml:",chardata"`
	ID     string            `xml:"id,attr"`
	Name   string            `xml:"name,attr"`
	Subcat []TorznabCategory `xml:"subcat"`
}

type CapabilitiesCategories struct {
	Text     string               `xml:",chardata"`
	Category []CategoryWithSubcat `xml:"category"`
}

type CapabilitiesGroup struct {
	Text        string `xml:",chardata"`
	ID          string `xml:"id,attr"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,attr"`
	Lastupdate  string `xml:"lastupdate,attr"`
}

type CapabilitiesGroups struct {
	Text  string              `xml:",chardata"`
	Group []CapabilitiesGroup `xml:"group"`
}

type CapabilitiesGenre struct {
	Text       string `xml:",chardata"`
	ID         string `xml:"id,attr"`
	Categoryid string `xml:"categoryid,attr"`
	Name       string `xml:"name,attr"`
}

type CapabilitiesGenres struct {
	Text  string              `xml:",chardata"`
	Genre []CapabilitiesGenre `xml:"genre"`
}

type CapabilitiesTag struct {
	Text        string `xml:",chardata"`
	Name        string `xml:"name,attr"`
	Description string `xml:"description,attr"`
}

type CapabilitiesTags struct {
	Text string            `xml:",chardata"`
	Tag  []CapabilitiesTag `xml:"tag"`
}

type CapabilitiesSearching struct {
	Text        string                    `xml:",chardata"`
	Search      SupportedSearchCapability `xml:"search"`
	TvSearch    SupportedSearchCapability `xml:"tv-search"`
	MovieSearch SupportedSearchCapability `xml:"movie-search"`
	AudioSearch SupportedSearchCapability `xml:"audio-search"`
	BookSearch  SupportedSearchCapability `xml:"book-search"`
}

type Capabilities struct {
	XMLName xml.Name           `xml:"caps"`
	Text    string             `xml:",chardata"`
	Server  CapabilitiesServer `xml:"server"`
	Limits  CapabilitiesLimits `xml:"limits"`
	// Omitted in torznab, newznab only
	Retention    *CapabilitiesRetention   `xml:"retention,omitempty"`
	Registration CapabilitiesRegistration `xml:"registration"`
	Searching    CapabilitiesSearching    `xml:"searching"`
	Categories   CapabilitiesCategories   `xml:"categories"`
	// No current application specified in torznab.
	Groups CapabilitiesGroups `xml:"groups"`
	Genres CapabilitiesGenres `xml:"genres"`
	Tags   CapabilitiesTags   `xml:"tags"`
}
