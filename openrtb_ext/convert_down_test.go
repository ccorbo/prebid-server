package openrtb_ext

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v17/adcom1"
	"github.com/prebid/openrtb/v17/openrtb2"
	"github.com/stretchr/testify/assert"
)

func TestConvertDownTo25(t *testing.T) {
	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name: "2.6-to-2.5",
			givenRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Imp:    []openrtb2.Imp{{Rwdd: 1}},
				Source: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}, Ver: "2"}},
				Regs:   &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(1), USPrivacy: "3", GPP: "gpp", GPPSID: []int8{1, 2}},
				User:   &openrtb2.User{Consent: "1", EIDs: []openrtb2.EID{{Source: "42"}}},
			},
			expectedRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Imp:    []openrtb2.Imp{{Ext: json.RawMessage(`{"prebid":{"is_rewarded_inventory":1}}`)}},
				Source: &openrtb2.Source{Ext: json.RawMessage(`{"schain":{"complete":1,"nodes":[],"ver":"2"}}`)},
				Regs:   &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":1,"gpp":"gpp","gpp_sid":[1,2],"us_privacy":"3"}`)},
				User:   &openrtb2.User{Ext: json.RawMessage(`{"consent":"1","eids":[{"source":"42"}]}`)},
			},
		},
		{
			name: "2.6-dropped", // integration with clear26Fields
			givenRequest: openrtb2.BidRequest{
				ID:     "anyID",
				CatTax: adcom1.CatTaxIABContent10,
				Device: &openrtb2.Device{LangB: "anyLang"},
			},
			expectedRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Device: &openrtb2.Device{},
			},
		},
		{
			name: "2.6-202211-dropped", // integration with clear202211Fields
			givenRequest: openrtb2.BidRequest{
				ID:  "anyID",
				App: &openrtb2.App{InventoryPartnerDomain: "anyDomain"},
			},
			expectedRequest: openrtb2.BidRequest{
				ID:  "anyID",
				App: &openrtb2.App{},
			},
		},
		{
			name: "2.6-to-2.5-OtherExtFields",
			givenRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Imp:    []openrtb2.Imp{{Rwdd: 1, Ext: json.RawMessage(`{"other":"otherImp"}`)}},
				Ext:    json.RawMessage(`{"other":"otherExt"}`),
				Source: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}, Ver: "2"}, Ext: json.RawMessage(`{"other":"otherSource"}`)},
				Regs:   &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(1), USPrivacy: "3", Ext: json.RawMessage(`{"other":"otherRegs"}`)},
				User:   &openrtb2.User{Consent: "1", EIDs: []openrtb2.EID{{Source: "42"}}, Ext: json.RawMessage(`{"other":"otherUser"}`)},
			},
			expectedRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Imp:    []openrtb2.Imp{{Ext: json.RawMessage(`{"other":"otherImp","prebid":{"is_rewarded_inventory":1}}`)}},
				Ext:    json.RawMessage(`{"other":"otherExt"}`),
				Source: &openrtb2.Source{Ext: json.RawMessage(`{"other":"otherSource","schain":{"complete":1,"nodes":[],"ver":"2"}}`)},
				Regs:   &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":1,"other":"otherRegs","us_privacy":"3"}`)},
				User:   &openrtb2.User{Ext: json.RawMessage(`{"consent":"1","eids":[{"source":"42"}],"other":"otherUser"}`)},
			},
		},
		{
			name: "malformed-schain",
			givenRequest: openrtb2.BidRequest{
				ID:     "anyID",
				Source: &openrtb2.Source{SChain: &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}, Ver: "2"}, Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-gdpr",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				Regs: &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(1), Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-consent",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				User: &openrtb2.User{Consent: "1", Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-usprivacy",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				Regs: &openrtb2.Regs{USPrivacy: "3", Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-eid",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				User: &openrtb2.User{EIDs: []openrtb2.EID{{Source: "42"}}, Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-imp",
			givenRequest: openrtb2.BidRequest{
				ID:  "anyID",
				Imp: []openrtb2.Imp{{Rwdd: 1, Ext: json.RawMessage(`malformed`)}},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-gpp",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				Regs: &openrtb2.Regs{GPP: "gpp", Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
		{
			name: "malformed-gppsid",
			givenRequest: openrtb2.BidRequest{
				ID:   "anyID",
				Regs: &openrtb2.Regs{GPPSID: []int8{1, 2}, Ext: json.RawMessage(`malformed`)},
			},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := ConvertDownTo25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveSupplyChainFrom26To25(t *testing.T) {
	var (
		schain1     = &openrtb2.SupplyChain{Complete: 1, Nodes: []openrtb2.SupplyChainNode{}, Ver: "1"}
		schain1Json = json.RawMessage(`{"schain":{"complete":1,"nodes":[],"ver":"1"}}`)
		schain2Json = json.RawMessage(`{"schain":{"complete":1,"nodes":[],"ver":"2"}}`)
	)

	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "notpresent-source",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "notpresent-source-schain",
			givenRequest:    openrtb2.BidRequest{Source: &openrtb2.Source{}},
			expectedRequest: openrtb2.BidRequest{Source: &openrtb2.Source{}},
		},
		{
			name:            "2.6-migratedto-2.5",
			givenRequest:    openrtb2.BidRequest{Source: &openrtb2.Source{SChain: schain1}},
			expectedRequest: openrtb2.BidRequest{Source: &openrtb2.Source{Ext: schain1Json}},
		},
		{
			name:            "2.5-overwritten",
			givenRequest:    openrtb2.BidRequest{Source: &openrtb2.Source{SChain: schain1, Ext: schain2Json}},
			expectedRequest: openrtb2.BidRequest{Source: &openrtb2.Source{Ext: schain1Json}},
		},
		{
			name:         "malformed",
			givenRequest: openrtb2.BidRequest{Source: &openrtb2.Source{SChain: schain1, Ext: json.RawMessage(`malformed`)}},
			expectedErr:  "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := moveSupplyChainFrom26To25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveGDPRFrom26To25(t *testing.T) {
	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "notpresent-regs",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "notpresent-regs-gdpr",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{}},
		},
		{
			name:            "2.6-migratedto-2.5",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(0)}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":0}`)}},
		},
		{
			name:            "2.5-overwritten",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(0), Ext: json.RawMessage(`{"gdpr":1}`)}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"gdpr":0}`)}},
		},
		{
			name:         "malformed",
			givenRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{GDPR: openrtb2.Int8Ptr(0), Ext: json.RawMessage(`malformed`)}},
			expectedErr:  "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := moveGDPRFrom26To25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveConsentFrom26To25(t *testing.T) {
	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "notpresent-user",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "notpresent-user-consent",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{}},
		},
		{
			name:            "2.6-migratedto-2.5",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{Consent: "1"}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{Ext: json.RawMessage(`{"consent":"1"}`)}},
		},
		{
			name:            "2.5-overwritten",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{Consent: "1", Ext: json.RawMessage(`{"consent":"2"}`)}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{Ext: json.RawMessage(`{"consent":"1"}`)}},
		},
		{
			name:         "malformed",
			givenRequest: openrtb2.BidRequest{User: &openrtb2.User{Consent: "1", Ext: json.RawMessage(`malformed`)}},
			expectedErr:  "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := moveConsentFrom26To25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveUSPrivacyFrom26To25(t *testing.T) {
	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "notpresent-regs",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "notpresent-regs-usprivacy",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{}},
		},
		{
			name:            "2.6-migratedto-2.5",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{USPrivacy: "1"}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"us_privacy":"1"}`)}},
		},
		{
			name:            "2.5-overwritten",
			givenRequest:    openrtb2.BidRequest{Regs: &openrtb2.Regs{USPrivacy: "1", Ext: json.RawMessage(`{"us_privacy":"2"}`)}},
			expectedRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{Ext: json.RawMessage(`{"us_privacy":"1"}`)}},
		},
		{
			name:         "malformed",
			givenRequest: openrtb2.BidRequest{Regs: &openrtb2.Regs{USPrivacy: "1", Ext: json.RawMessage(`malformed`)}},
			expectedErr:  "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := moveUSPrivacyFrom26To25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveEIDFrom26To25(t *testing.T) {
	var (
		eid1     = []openrtb2.EID{{Source: "1"}}
		eid1Json = json.RawMessage(`{"eids":[{"source":"1"}]}`)
		eid2Json = json.RawMessage(`{"eids":[{"source":"2"}]}`)
	)

	testCases := []struct {
		name            string
		givenRequest    openrtb2.BidRequest
		expectedRequest openrtb2.BidRequest
		expectedErr     string
	}{
		{
			name:            "notpresent-user",
			givenRequest:    openrtb2.BidRequest{},
			expectedRequest: openrtb2.BidRequest{},
		},
		{
			name:            "notpresent-user-eids",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{}},
		},
		{
			name:            "2.6-migratedto-2.5",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{EIDs: eid1}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{Ext: eid1Json}},
		},
		{
			name:            "2.6-migratedto-2.5-empty",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{EIDs: []openrtb2.EID{}}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{}},
		},
		{
			name:            "2.5-overwritten",
			givenRequest:    openrtb2.BidRequest{User: &openrtb2.User{EIDs: eid1, Ext: eid2Json}},
			expectedRequest: openrtb2.BidRequest{User: &openrtb2.User{Ext: eid1Json}},
		},
		{
			name:         "malformed",
			givenRequest: openrtb2.BidRequest{User: &openrtb2.User{EIDs: eid1, Ext: json.RawMessage(`malformed`)}},
			expectedErr:  "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &RequestWrapper{BidRequest: &test.givenRequest}
			err := moveEIDFrom26To25(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildRequest(), "error")
				assert.Equal(t, test.expectedRequest, *w.BidRequest, "result")
			}
		})
	}
}

func TestMoveRewardedFrom26ToPrebidExt(t *testing.T) {
	testCases := []struct {
		name        string
		givenImp    openrtb2.Imp
		expectedImp openrtb2.Imp
		expectedErr string
	}{
		{
			name:        "notpresent-prebid",
			givenImp:    openrtb2.Imp{},
			expectedImp: openrtb2.Imp{},
		},
		{
			name:        "2.6-migratedto-2.5",
			givenImp:    openrtb2.Imp{Rwdd: 1},
			expectedImp: openrtb2.Imp{Ext: json.RawMessage(`{"prebid":{"is_rewarded_inventory":1}}`)},
		},
		{
			name:        "2.5-overwritten",
			givenImp:    openrtb2.Imp{Rwdd: 1, Ext: json.RawMessage(`{"prebid":{"is_rewarded_inventory":2}}`)},
			expectedImp: openrtb2.Imp{Ext: json.RawMessage(`{"prebid":{"is_rewarded_inventory":1}}`)},
		},
		{
			name:        "Malformed",
			givenImp:    openrtb2.Imp{Rwdd: 1, Ext: json.RawMessage(`malformed`)},
			expectedErr: "invalid character 'm' looking for beginning of value",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			w := &ImpWrapper{Imp: &test.givenImp}
			err := moveRewardedFrom26ToPrebidExt(w)

			if len(test.expectedErr) > 0 {
				assert.EqualError(t, err, test.expectedErr, "error")
			} else {
				assert.NoError(t, w.RebuildImp(), "error")
				assert.Equal(t, test.expectedImp, *w.Imp, "result")
			}
		})
	}
}

func TestClear26Fields(t *testing.T) {
	var int8_1 int8 = 1

	given := &openrtb2.BidRequest{
		ID:     "anyID",
		WLangB: []string{"anyLang"},
		CatTax: adcom1.CatTaxIABAudience11,
		App: &openrtb2.App{
			CatTax:  adcom1.CatTaxIABAudience11,
			KwArray: []string{"anyKeyword"},
			Content: &openrtb2.Content{
				ID:      "anyContent",
				CatTax:  adcom1.CatTaxIABAudience11,
				KwArray: []string{"anyKeyword"},
				LangB:   "anyLang",
				Network: &openrtb2.Network{
					ID: "anyNetwork",
				},
				Channel: &openrtb2.Channel{
					ID: "anyChannel",
				},
				Producer: &openrtb2.Producer{
					ID:     "anyProcedure",
					CatTax: adcom1.CatTaxIABAudience11,
				},
			},
			Publisher: &openrtb2.Publisher{
				ID:     "anyPublisher",
				CatTax: adcom1.CatTaxIABAudience11,
			},
		},
		Site: &openrtb2.Site{
			CatTax:  adcom1.CatTaxIABAudience11,
			KwArray: []string{"anyKeyword"},
			Content: &openrtb2.Content{
				ID:      "anyContent",
				CatTax:  adcom1.CatTaxIABAudience11,
				KwArray: []string{"anyKeyword"},
				LangB:   "anyLang",
				Network: &openrtb2.Network{
					ID: "anyNetwork",
				},
				Channel: &openrtb2.Channel{
					ID: "anyChannel",
				},
				Producer: &openrtb2.Producer{
					ID:     "anyProcedure",
					CatTax: adcom1.CatTaxIABAudience11,
				},
			},
			Publisher: &openrtb2.Publisher{
				ID:     "anyPublisher",
				CatTax: adcom1.CatTaxIABAudience11,
			},
		},
		Device: &openrtb2.Device{
			IP:    "1.2.3.4",
			LangB: "anyLang",
			SUA: &openrtb2.UserAgent{
				Model: "PBS 2000",
			},
		},
		Regs: &openrtb2.Regs{
			COPPA:     1,
			GDPR:      &int8_1,
			USPrivacy: "anyCCPA",
		},
		Source: &openrtb2.Source{
			TID: "anyTransactionID",
			SChain: &openrtb2.SupplyChain{
				Complete: 1,
			},
		},
		User: &openrtb2.User{
			ID:      "anyUser",
			KwArray: []string{"anyKeyword"},
			Consent: "anyConsent",
			EIDs:    []openrtb2.EID{{Source: "anySource"}},
		},
		Imp: []openrtb2.Imp{{
			ID:   "imp1",
			Rwdd: 1,
			SSAI: openrtb2.AdInsertServer,
			Audio: &openrtb2.Audio{
				MIMEs:        []string{"any/audio"},
				PodDur:       30,
				RqdDurs:      []int64{15, 60},
				PodID:        1,
				PodSeq:       adcom1.PodSeqFirst,
				SlotInPod:    adcom1.SlotPosFirst,
				MinCPMPerSec: 100.0,
			},
			Video: &openrtb2.Video{
				MIMEs:        []string{"any/video"},
				MaxSeq:       30,
				PodDur:       30,
				PodID:        1,
				PodSeq:       adcom1.PodSeqFirst,
				RqdDurs:      []int64{15, 60},
				SlotInPod:    adcom1.SlotPosFirst,
				MinCPMPerSec: 100.0,
			},
		}},
	}

	expected := &openrtb2.BidRequest{
		ID: "anyID",
		App: &openrtb2.App{
			Content: &openrtb2.Content{
				ID: "anyContent",
				Producer: &openrtb2.Producer{
					ID: "anyProcedure",
				},
			},
			Publisher: &openrtb2.Publisher{
				ID: "anyPublisher",
			},
		},
		Site: &openrtb2.Site{
			Content: &openrtb2.Content{
				ID: "anyContent",
				Producer: &openrtb2.Producer{
					ID: "anyProcedure",
				},
			},
			Publisher: &openrtb2.Publisher{
				ID: "anyPublisher",
			},
		},
		Device: &openrtb2.Device{
			IP: "1.2.3.4",
		},
		Regs: &openrtb2.Regs{
			COPPA: 1,
		},
		Source: &openrtb2.Source{
			TID: "anyTransactionID",
		},
		User: &openrtb2.User{
			ID: "anyUser",
		},
		Imp: []openrtb2.Imp{{
			ID: "imp1",
			Audio: &openrtb2.Audio{
				MIMEs: []string{"any/audio"},
			},
			Video: &openrtb2.Video{
				MIMEs: []string{"any/video"},
			},
		}},
	}

	r := &RequestWrapper{BidRequest: given}
	clear26Fields(r)
	assert.Equal(t, expected, r.BidRequest)

}

func TestClear202211Fields(t *testing.T) {
	testCases := []struct {
		name     string
		given    openrtb2.BidRequest
		expected openrtb2.BidRequest
	}{
		{
			name: "app",
			given: openrtb2.BidRequest{
				ID:   "anyID",
				App:  &openrtb2.App{InventoryPartnerDomain: "anyDomain"},
				Imp:  []openrtb2.Imp{{ID: "imp1", Qty: &openrtb2.Qty{Multiplier: 2.0}, DT: 42}},
				Regs: &openrtb2.Regs{GPP: "anyGPP", GPPSID: []int8{1, 2, 3}},
			},
			expected: openrtb2.BidRequest{
				ID:   "anyID",
				App:  &openrtb2.App{},
				Imp:  []openrtb2.Imp{{ID: "imp1"}},
				Regs: &openrtb2.Regs{},
			},
		},
		{
			name: "site",
			given: openrtb2.BidRequest{
				ID:   "anyID",
				Site: &openrtb2.Site{InventoryPartnerDomain: "anyDomain"},
				Imp:  []openrtb2.Imp{{ID: "imp1", Qty: &openrtb2.Qty{Multiplier: 2.0}, DT: 42}},
				Regs: &openrtb2.Regs{GPP: "anyGPP", GPPSID: []int8{1, 2, 3}},
			},
			expected: openrtb2.BidRequest{
				ID:   "anyID",
				Site: &openrtb2.Site{},
				Imp:  []openrtb2.Imp{{ID: "imp1"}},
				Regs: &openrtb2.Regs{},
			},
		},
		{
			name: "dooh",
			given: openrtb2.BidRequest{
				ID:   "anyID",
				DOOH: &openrtb2.DOOH{ID: "anyDOOH"},
				Imp:  []openrtb2.Imp{{ID: "imp1", Qty: &openrtb2.Qty{Multiplier: 2.0}, DT: 42}},
				Regs: &openrtb2.Regs{GPP: "anyGPP", GPPSID: []int8{1, 2, 3}},
			},
			expected: openrtb2.BidRequest{
				ID:   "anyID",
				Imp:  []openrtb2.Imp{{ID: "imp1"}},
				Regs: &openrtb2.Regs{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			r := &RequestWrapper{BidRequest: &test.given}
			clear202211Fields(r)
			assert.Equal(t, &test.expected, r.BidRequest)
		})
	}
}

func TestCreateImpressions(t *testing.T) {
	testCases := []struct {
		name     string
		given    *openrtb2.BidRequest
		expected *openrtb2.BidRequest
	}{
		{
			name: "base test",
			given: &openrtb2.BidRequest{
				ID: "1",
				Imp: []openrtb2.Imp{
					{
						ID: "1",
						Video: &openrtb2.Video{
							PodDur: int64(60),
							MaxSeq: int64(4),
							W:      600,
							H:      500,
						},
					},
				},
			},
			expected: &openrtb2.BidRequest{
				ID: "1",
				Imp: []openrtb2.Imp{
					{
						ID: "0_0",
						Video: &openrtb2.Video{
							MaxDuration: 15,
							W:           600,
							H:           500,
						},
					},
					{
						ID: "0_1",
						Video: &openrtb2.Video{
							MaxDuration: 15,
							W:           600,
							H:           500,
						},
					},
					{
						ID: "0_2",
						Video: &openrtb2.Video{
							MaxDuration: 15,
							W:           600,
							H:           500,
						},
					},
					{
						ID: "0_3",
						Video: &openrtb2.Video{
							MaxDuration: 15,
							W:           600,
							H:           500,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.given
			CreateImpressions(r)
			assert.Equal(t, &tc.expected, r)
		})
	}
}
