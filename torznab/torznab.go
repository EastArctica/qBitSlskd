package torznab

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/EastArctica/qbitslskd/models"
)

// Static response telling Lidar the capabilities. The only one supported is search, since thats all this fake torznab instance needs to do.
func CapabilitiesHandler(w http.ResponseWriter, req *http.Request) {
	var capabilities models.Capabilities = models.Capabilities{
		Server: models.CapabilitiesServer{
			Version:   "1.0",
			Title:     "qBitSlskd",
			Strapline: "qBitSlskd",
			Email:     "example@example.com",
			// I think this is where the actual server is but we don't have much of a choice for that
			URL: "https://github.com/EastArctica/qBitSlskd",
			// slskd icon
			Image: "https://avatars.githubusercontent.com/u/76762370",
		},
		Limits: models.CapabilitiesLimits{
			Max:     "10000",
			Default: "250",
		},
		Registration: models.CapabilitiesRegistration{
			Available: "yes",
			Open:      "yes",
		},
		Searching: models.CapabilitiesSearching{
			Search: models.SupportedSearchCapability{
				Available:       "yes",
				SupportedParams: "q",
			},
			TvSearch: models.SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			MovieSearch: models.SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			AudioSearch: models.SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
			BookSearch: models.SupportedSearchCapability{
				Available:       "no",
				SupportedParams: "",
			},
		},
	}

	data, err := xml.Marshal(capabilities)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("response-type", "application/xml")
	fmt.Fprint(w, string(data))
}
