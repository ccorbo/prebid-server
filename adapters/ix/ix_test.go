package ix

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"

	"github.com/prebid/openrtb/v17/adcom1"
	"github.com/prebid/openrtb/v17/openrtb2"

	log "go.uber.org/zap"
)

const endpoint string = "http://host/endpoint"

func TestJsonSamples(t *testing.T) {
	if bidder, err := Builder(openrtb_ext.BidderIx, config.Adapter{Endpoint: endpoint}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"}); err == nil {
		ixBidder := bidder.(*IxAdapter)
		ixBidder.maxRequests = 2
		adapterstest.RunJSONBidderTest(t, "ixtest", bidder)
	} else {
		t.Fatalf("Builder returned unexpected error %v", err)
	}
}

func TestIxMakeBidsWithCategoryDuration(t *testing.T) {
	loggerConfig := log.NewDevelopmentConfig()
	logger, _ := loggerConfig.Build()

	bidder := &IxAdapter{logger: *logger}

	mockedReq := &openrtb2.BidRequest{
		Imp: []openrtb2.Imp{{
			ID: "1_1",
			Video: &openrtb2.Video{
				W:           640,
				H:           360,
				MIMEs:       []string{"video/mp4"},
				MaxDuration: 60,
				Protocols:   []adcom1.MediaCreativeSubtype{2, 3, 5, 6},
			},
			Ext: json.RawMessage(
				`{
					"prebid": {},
					"bidder": {
						"siteID": 123456
					}
				}`,
			)},
		},
	}
	mockedExtReq := &adapters.RequestData{}
	mockedBidResponse := &openrtb2.BidResponse{
		ID: "test-1",
		SeatBid: []openrtb2.SeatBid{{
			Seat: "Buyer",
			Bid: []openrtb2.Bid{{
				ID:    "1",
				ImpID: "1_1",
				Price: 1.23,
				AdID:  "123",
				Ext: json.RawMessage(
					`{
						"prebid": {
							"video": {
								"duration": 60,
								"primary_category": "IAB18-1"
							}
						}
					}`,
				),
			}},
		}},
	}
	body, _ := json.Marshal(mockedBidResponse)
	mockedRes := &adapters.ResponseData{
		StatusCode: 200,
		Body:       body,
	}

	expectedBidCount := 1
	expectedBidType := openrtb_ext.BidTypeVideo
	expectedBidDuration := 60
	expectedBidCategory := "IAB18-1"
	expectedErrorCount := 0

	bidResponse, errors := bidder.MakeBids(mockedReq, mockedExtReq, mockedRes)

	if len(bidResponse.Bids) != expectedBidCount {
		t.Errorf("should have 1 bid, bids=%v", bidResponse.Bids)
	}
	if bidResponse.Bids[0].BidType != expectedBidType {
		t.Errorf("bid type should be video, bidType=%s", bidResponse.Bids[0].BidType)
	}
	if bidResponse.Bids[0].BidVideo.Duration != expectedBidDuration {
		t.Errorf("video duration should be set")
	}
	if bidResponse.Bids[0].Bid.Cat[0] != expectedBidCategory {
		t.Errorf("bid category should be set")
	}
	if len(errors) != expectedErrorCount {
		t.Errorf("should not have any errors, errors=%v", errors)
	}
}

func TestIxMakeBidsWithInvalidJson(t *testing.T) {
	bidder := &IxAdapter{}
	bidder.maxRequests = 1

	mockedReq := &openrtb2.BidRequest{
		Imp: []openrtb2.Imp{{
			ID:     "1_1",
			Banner: &openrtb2.Banner{},
			Ext: json.RawMessage(
				`{
					"bidder": {
						"siteID": "123456"
					}
				}`,
			)},
		},
		Ext: json.RawMessage(
			`X-invalidValueForJsonField`,
		),
	}

	actualAdapterRequests, errs := bidder.MakeRequests(mockedReq, &adapters.ExtraRequestInfo{})

	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "json: error calling MarshalJSON for type json.RawMessage: invalid character 'X' looking for beginning of value")
	assert.Len(t, actualAdapterRequests, 0)
}

func TestIxBuilderWithLoggerSamplingOff(t *testing.T) {
	if bidder, err := Builder(openrtb_ext.BidderIx, config.Adapter{Endpoint: endpoint, SamplingEnabled: false}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"}); err == nil {
		ixBidder := bidder.(*IxAdapter)
		assert.NotNil(t, ixBidder.logger)
	} else {
		t.Fatalf("Builder returned unexpected error %v", err)
	}
}

func TestIxBuilderWithLoggerSamplingOn(t *testing.T) {
	if bidder, err := Builder(openrtb_ext.BidderIx, config.Adapter{Endpoint: endpoint, SamplingEnabled: true}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"}); err == nil {
		ixBidder := bidder.(*IxAdapter)
		assert.NotNil(t, ixBidder.logger)
	} else {
		t.Fatalf("Builder returned unexpected error %v", err)
	}
}

func TestGetLoggerSamplingConfig(t *testing.T) {
	loggerConfig := log.NewDevelopmentConfig()
	updatedLoggerConfig := getLoggerSamplingConfig(loggerConfig, config.Adapter{SamplingInitial: 1, SamplingThereafter: 0})
	assert.Equal(t, 1, updatedLoggerConfig.Sampling.Initial)
	assert.Equal(t, 0, updatedLoggerConfig.Sampling.Thereafter)
}

func TestMakeBidsMultiBid(t *testing.T) {
	bidder := &IxAdapter{}

	mockedReq := &openrtb2.BidRequest{
		ID: "test-1",
		Imp: []openrtb2.Imp{{
			ID: "1_0",
			Video: &openrtb2.Video{
				W:           640,
				H:           360,
				MIMEs:       []string{"video/mp4"},
				MaxDuration: 60,
				Protocols:   []adcom1.MediaCreativeSubtype{2, 3, 5, 6},
			},
			Ext: json.RawMessage(
				`{
					"prebid": {},
					"bidder": {
						"siteID": 123456
					}
				}`,
			)},
		},
	}
	mockedExtReq := &adapters.RequestData{}
	mockedBidResponse := &openrtb2.BidResponse{
		ID: "test-1",
		SeatBid: []openrtb2.SeatBid{{
			Seat: "Buyer",
			Bid: []openrtb2.Bid{
				{
					ID:    "1",
					ImpID: "1_0",
					Price: 1.23,
					AdID:  "123",
					Ext: json.RawMessage(
						`{
							"prebid": {
								"video": {
									"duration": 60,
									"primary_category": "IAB18-1"
								}
							}
						}`,
					),
				},
				{
					ID:    "2",
					ImpID: "1_0",
					Price: 1.53,
					AdID:  "123",
					Ext: json.RawMessage(
						`{
							"prebid": {
								"video": {
									"duration": 60,
									"primary_category": "IAB1-1"
								}
							}
						}`,
					),
				},
			},
		}},
	}
	body, _ := json.Marshal(mockedBidResponse)
	mockedRes := &adapters.ResponseData{
		StatusCode: 200,
		Body:       body,
	}

	expectedBidCount := 2

	bidResponse, _ := bidder.MakeBids(mockedReq, mockedExtReq, mockedRes)
	assert.Equal(t, expectedBidCount, len(bidResponse.Bids))
	assert.Equal(t, "1", bidResponse.Bids[0].Bid.ID)
	assert.Equal(t, "2", bidResponse.Bids[1].Bid.ID)
}
