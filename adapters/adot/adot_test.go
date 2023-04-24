package adot

import (
	"encoding/json"
	"testing"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/adapters/adapterstest"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

const testsBidderEndpoint = "https://dsp.adotmob.com/headerbidding{PUBLISHER_PATH}/bidrequest"

func TestJsonSamples(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderAdot, config.Adapter{
		Endpoint: testsBidderEndpoint}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	adapterstest.RunJSONBidderTest(t, "adottest", bidder)
}

func TestFallbackToMTypeFromExt(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := fallbackToMTypeFromExt(nil)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		byteInvalid, _ := json.Marshal(&adotBidExt{Adot: bidExt{"invalid"}})
		err := fallbackToMTypeFromExt(&openrtb2.Bid{Ext: json.RawMessage(byteInvalid)})
		assert.Error(t, err)
	})

	t.Run("banner", func(t *testing.T) {
		byteBanner, _ := json.Marshal(&adotBidExt{Adot: bidExt{"banner"}})
		bid := &openrtb2.Bid{Ext: json.RawMessage(byteBanner)}
		err := fallbackToMTypeFromExt(bid)
		assert.NoError(t, err)
		assert.Equal(t, openrtb2.MarkupBanner, bid.MType)
	})

	t.Run("video", func(t *testing.T) {
		byteVideo, _ := json.Marshal(&adotBidExt{Adot: bidExt{"video"}})
		bid := &openrtb2.Bid{Ext: json.RawMessage(byteVideo)}
		err := fallbackToMTypeFromExt(bid)
		assert.NoError(t, err)
		assert.Equal(t, openrtb2.MarkupVideo, bid.MType)
	})

	t.Run("native", func(t *testing.T) {
		byteNative, _ := json.Marshal(&adotBidExt{Adot: bidExt{"native"}})
		bid := &openrtb2.Bid{Ext: json.RawMessage(byteNative)}
		err := fallbackToMTypeFromExt(bid)
		assert.NoError(t, err)
		assert.Equal(t, openrtb2.MarkupNative, bid.MType)
	})
}

func TestBidResponseNoContent(t *testing.T) {
	bidder, buildErr := Builder(openrtb_ext.BidderAdot, config.Adapter{
		Endpoint: "https://dsp.adotmob.com/headerbidding{PUBLISHER_PATH}/bidrequest"}, config.Server{ExternalUrl: "http://hosturl.com", GvlID: 1, DataCenter: "2"})

	if buildErr != nil {
		t.Fatalf("Builder returned unexpected error %v", buildErr)
	}

	bidResponse, err := bidder.MakeBids(nil, nil, &adapters.ResponseData{StatusCode: 204})
	if bidResponse != nil {
		t.Fatalf("the bid response should be nil since the bidder status is No Content")
	} else if err != nil {
		t.Fatalf("the error should be nil since the bidder status is 204 : No Content")
	}
}

func TestResolveMacros(t *testing.T) {
	bid := &openrtb2.Bid{AdM: "adm:imp_${AUCTION_PRICE} amd:creativeview_${AUCTION_PRICE}", NURL: "nurl_${AUCTION_PRICE}", Price: 123.45}
	resolveMacros(bid)
	assert.Equal(t, "adm:imp_123.45 amd:creativeview_123.45", bid.AdM)
	assert.Equal(t, "nurl_123.45", bid.NURL)
}

func TestGetImpAdotExt(t *testing.T) {
	ext := &openrtb2.Imp{Ext: json.RawMessage(`{"bidder":{"publisherPath": "/hubvisor"}}`)}
	adotExt := getImpAdotExt(ext)
	assert.Equal(t, adotExt.PublisherPath, "/hubvisor")

	emptyBidderExt := &openrtb2.Imp{Ext: json.RawMessage(`{"bidder":{}}`)}
	emptyAdotBidderExt := getImpAdotExt(emptyBidderExt)
	assert.NotNil(t, emptyAdotBidderExt)
	assert.Equal(t, emptyAdotBidderExt.PublisherPath, "")

	emptyExt := &openrtb2.Imp{Ext: json.RawMessage(`{}`)}
	emptyAdotExt := getImpAdotExt(emptyExt)
	assert.Nil(t, emptyAdotExt)
}
