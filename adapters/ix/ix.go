package ix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/prebid/openrtb/v17/native1"
	native1response "github.com/prebid/openrtb/v17/native1/response"
	"github.com/prebid/openrtb/v17/openrtb2"
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type IxAdapter struct {
	URI         string
	maxRequests int
	logger      log.Logger
}

func (a *IxAdapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	nImp := len(request.Imp)
	if nImp > a.maxRequests {
		request.Imp = request.Imp[:a.maxRequests]
		nImp = a.maxRequests
	}
	requests := make([]*adapters.RequestData, 0, a.maxRequests)
	errs := make([]error, 0)

	headers := http.Header{
		"Content-Type": {"application/json;charset=utf-8"},
		"Accept":       {"application/json"}}

	if request.Site != nil {
		site := *request.Site
		if site.Publisher == nil {
			site.Publisher = &openrtb2.Publisher{}
		} else {
			publisher := *site.Publisher
			site.Publisher = &publisher
		}
		request.Site = &site
	}

	siteIds := make(map[string]bool)
	sanitizedImps := make([]openrtb2.Imp, 0, len(request.Imp))
	for _, imp := range request.Imp {
		if err := parseSiteId(&imp, siteIds); err != nil {
			errs = append(errs, err)
			continue
		}

		if imp.Banner != nil {
			banner := *imp.Banner
			imp.Banner = &banner

			if len(banner.Format) == 0 && banner.W != nil && banner.H != nil {
				banner.Format = []openrtb2.Format{{W: *banner.W, H: *banner.H}}
			}

			if len(banner.Format) == 1 {
				banner.W = openrtb2.Int64Ptr(banner.Format[0].W)
				banner.H = openrtb2.Int64Ptr(banner.Format[0].H)
			}
		}
		sanitizedImps = append(sanitizedImps, imp)
	}
	if request.Site != nil && len(siteIds) == 1 {
		for siteId, _ := range siteIds {
			request.Site.Publisher.ID = siteId
		}
	}
	if len(siteIds) > 1 {
		var siteIdStringBuffer bytes.Buffer
		for siteId, _ := range siteIds {
			siteIdStringBuffer.WriteString(siteId)
			siteIdStringBuffer.WriteString(", ")
		}
		a.logger.Warn(fmt.Sprintf("Multiple SiteIDs found. %s", siteIdStringBuffer.String()))
	}

	request.Imp = sanitizedImps
	if len(request.Imp) != 0 {
		if requestData, err := createRequestData(a, request, &headers); err == nil {
			requests = append(requests, requestData)
		} else {
			errs = append(errs, err)
		}
	}
	return requests, errs
}

func parseSiteId(imp *openrtb2.Imp, siteIds map[string]bool) error {
	var bidderExt adapters.ExtImpBidder
	if err := json.Unmarshal(imp.Ext, &bidderExt); err != nil {
		return err
	}

	var ixExt openrtb_ext.ExtImpIx
	if err := json.Unmarshal(bidderExt.Bidder, &ixExt); err != nil {
		return err
	}

	if ixExt.SiteId != "" {
		siteIds[ixExt.SiteId] = true
	}
	return nil
}

func createRequestData(a *IxAdapter, request *openrtb2.BidRequest, headers *http.Header) (*adapters.RequestData, error) {
	body, err := json.Marshal(request)
	return &adapters.RequestData{
		Method:  "POST",
		Uri:     a.URI,
		Body:    body,
		Headers: *headers,
	}, err
}

func (a *IxAdapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	switch {
	case response.StatusCode == http.StatusNoContent:
		return nil, nil
	case response.StatusCode == http.StatusBadRequest:
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	case response.StatusCode != http.StatusOK:
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	var bidResponse openrtb2.BidResponse
	if err := json.Unmarshal(response.Body, &bidResponse); err != nil {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("JSON parsing error: %v", err),
		}}
	}

	// Store media type per impression in a map for later use to set in bid.ext.prebid.type
	// Won't work for multiple bid case with a multi-format ad unit. We expect to get type from exchange on such case.
	impMediaTypeReq := map[string]openrtb_ext.BidType{}
	for _, imp := range internalRequest.Imp {
		if imp.Banner != nil {
			impMediaTypeReq[imp.ID] = openrtb_ext.BidTypeBanner
		} else if imp.Video != nil {
			impMediaTypeReq[imp.ID] = openrtb_ext.BidTypeVideo
		} else if imp.Native != nil {
			impMediaTypeReq[imp.ID] = openrtb_ext.BidTypeNative
		} else if imp.Audio != nil {
			impMediaTypeReq[imp.ID] = openrtb_ext.BidTypeAudio
		}
	}

	// capacity 0 will make channel unbuffered
	bidderResponse := adapters.NewBidderResponseWithBidsCapacity(0)
	bidderResponse.Currency = bidResponse.Cur

	var errs []error

	for _, seatBid := range bidResponse.SeatBid {
		for _, bid := range seatBid.Bid {

			bidType, err := getMediaTypeForBid(bid, impMediaTypeReq)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			var bidExtVideo *openrtb_ext.ExtBidPrebidVideo
			var bidExt openrtb_ext.ExtBid
			if bidType == openrtb_ext.BidTypeVideo {
				unmarshalExtErr := json.Unmarshal(bid.Ext, &bidExt)
				if unmarshalExtErr == nil && bidExt.Prebid != nil && bidExt.Prebid.Video != nil {
					bidExtVideo = &openrtb_ext.ExtBidPrebidVideo{
						Duration: bidExt.Prebid.Video.Duration,
					}
					if len(bid.Cat) == 0 {
						bid.Cat = []string{bidExt.Prebid.Video.PrimaryCategory}
					}
				}
			}

			var bidNative1v1 *Native11Wrapper
			if bidType == openrtb_ext.BidTypeNative {
				err := json.Unmarshal([]byte(bid.AdM), &bidNative1v1)
				if err == nil && len(bidNative1v1.Native.EventTrackers) > 0 {
					mergeNativeImpTrackers(&bidNative1v1.Native)
					if json, err := marshalJsonWithoutUnicode(bidNative1v1); err == nil {
						bid.AdM = string(json)
					}
				}
			}

			var bidNative1v2 *native1response.Response
			if bidType == openrtb_ext.BidTypeNative {
				err := json.Unmarshal([]byte(bid.AdM), &bidNative1v2)
				if err == nil && len(bidNative1v2.EventTrackers) > 0 {
					mergeNativeImpTrackers(bidNative1v2)
					if json, err := marshalJsonWithoutUnicode(bidNative1v2); err == nil {
						bid.AdM = string(json)
					}
				}
			}

			bidderResponse.Bids = append(bidderResponse.Bids, &adapters.TypedBid{
				Bid:      &bid,
				BidType:  bidType,
				BidVideo: bidExtVideo,
			})
		}
	}

	a.logger.Info(fmt.Sprintf("Returning Bid Response: RequestID=%s, NumberOfImp=%d, NumberOfBids=%d", internalRequest.ID, len(internalRequest.Imp), len(bidderResponse.Bids)))
	return bidderResponse, errs
}

func getMediaTypeForBid(bid openrtb2.Bid, impMediaTypeReq map[string]openrtb_ext.BidType) (openrtb_ext.BidType, error) {
	switch bid.MType {
	case openrtb2.MarkupBanner:
		return openrtb_ext.BidTypeBanner, nil
	case openrtb2.MarkupVideo:
		return openrtb_ext.BidTypeVideo, nil
	case openrtb2.MarkupAudio:
		return openrtb_ext.BidTypeAudio, nil
	case openrtb2.MarkupNative:
		return openrtb_ext.BidTypeNative, nil
	}

	if bid.Ext != nil {
		var bidExt openrtb_ext.ExtBid
		err := json.Unmarshal(bid.Ext, &bidExt)
		if err == nil && bidExt.Prebid != nil {
			prebidType := string(bidExt.Prebid.Type)
			if prebidType != "" {
				return openrtb_ext.ParseBidType(prebidType)
			}
		}
	}

	if bidType, ok := impMediaTypeReq[bid.ImpID]; ok {
		return bidType, nil
	} else {
		return "", fmt.Errorf("unmatched impression id: %s", bid.ImpID)
	}
}

// Builder builds a new instance of the Ix adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	loggerConfig := log.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "time"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	if config.SamplingEnabled {
		loggerConfig = getLoggerSamplingConfig(loggerConfig, config)
	}

	logger, _ := loggerConfig.Build()

	defer logger.Sync()

	bidder := &IxAdapter{
		URI:         config.Endpoint,
		maxRequests: 20,
		logger:      *logger,
	}
	return bidder, nil
}

// native 1.2 to 1.1 tracker compatibility handling

type Native11Wrapper struct {
	Native native1response.Response `json:"native,omitempty"`
}

func mergeNativeImpTrackers(bidNative *native1response.Response) {

	// create unique list of imp pixels urls from `imptrackers` and `eventtrackers`
	uniqueImpPixels := map[string]struct{}{}
	for _, v := range bidNative.ImpTrackers {
		uniqueImpPixels[v] = struct{}{}
	}

	for _, v := range bidNative.EventTrackers {
		if v.Event == native1.EventTypeImpression && v.Method == native1.EventTrackingMethodImage {
			uniqueImpPixels[v.URL] = struct{}{}
		}
	}

	// rewrite `imptrackers` with new deduped list of imp pixels
	bidNative.ImpTrackers = make([]string, 0)
	for k := range uniqueImpPixels {
		bidNative.ImpTrackers = append(bidNative.ImpTrackers, k)
	}

	// sort so tests pass correctly
	sort.Strings(bidNative.ImpTrackers)
}

func marshalJsonWithoutUnicode(v interface{}) (string, error) {
	// json.Marshal uses HTMLEscape for strings inside JSON which affects URLs
	// this is a problem with Native responses that embed JSON within JSON
	// a custom encoder can be used to disable this encoding.
	// https://pkg.go.dev/encoding/json#Marshal
	// https://pkg.go.dev/encoding/json#Encoder.SetEscapeHTML
	sb := &strings.Builder{}
	encoder := json.NewEncoder(sb)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	// json.Encode also writes a newline, need to remove
	// https://pkg.go.dev/encoding/json#Encoder.Encode
	return strings.TrimSuffix(sb.String(), "\n"), nil
}

// getLoggerSamplingConfig applies sampling configs to logger config and returns the updated config.
func getLoggerSamplingConfig(loggerConfig log.Config, config config.Adapter) log.Config {
	samplingConfig := &log.SamplingConfig{
		Initial:    config.SamplingInitial,
		Thereafter: config.SamplingThereafter,
	}
	loggerConfig.Sampling = samplingConfig
	return loggerConfig
}
