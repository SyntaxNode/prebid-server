package aax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type adapter struct {
	endpoint string
}

type aaxResponseBidExt struct {
	AdCodeType string `json:"adCodeType"`
}

func (a *adapter) MakeRequests(request *openrtb2.BidRequest, reqInfo *adapters.ExtraRequestInfo) ([]*adapters.RequestData, []error) {
	var errs []error

	reqJson, err := json.Marshal(request)
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json;charset=utf-8")

	return []*adapters.RequestData{{
		Method:  "POST",
		Uri:     a.endpoint,
		Body:    reqJson,
		Headers: headers,
	}}, errs
}

func (a *adapter) MakeBids(internalRequest *openrtb2.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) (*adapters.BidderResponse, []error) {
	var errs []error

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode == http.StatusBadRequest {
		return nil, []error{&errortypes.BadInput{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{&errortypes.BadServerResponse{
			Message: fmt.Sprintf("Unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode),
		}}
	}

	var bidResp openrtb2.BidResponse

	if err := json.Unmarshal(response.Body, &bidResp); err != nil {
		return nil, []error{err}
	}

	bidResponse := adapters.NewBidderResponse()

	for _, seatBid := range bidResp.SeatBid {
		for i := range seatBid.Bid {
			if err := fallbackToMTypeFromExt(&seatBid.Bid[i], internalRequest.Imp); err != nil {
				errs = append(errs, err)
			} else {
				b := &adapters.TypedBid{
					Bid: &seatBid.Bid[i],
				}
				bidResponse.Bids = append(bidResponse.Bids, b)
			}
		}
	}
	return bidResponse, errs
}

// Builder builds a new instance of the Aax adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter, server config.Server) (adapters.Bidder, error) {
	url := buildEndpoint(config.Endpoint, config.ExtraAdapterInfo)
	return &adapter{
		endpoint: url,
	}, nil
}

func fallbackToMTypeFromExt(bid *openrtb2.Bid, imps []openrtb2.Imp) error {
	// use mtype from bid, if available
	if bid.MType != 0 {
		return nil
	}

	var bidExt aaxResponseBidExt
	if err := json.Unmarshal(bid.Ext, &bidExt); err == nil {
		switch bidExt.AdCodeType {
		case "banner":
			bid.MType = openrtb2.MarkupBanner
			return nil
		case "native":
			bid.MType = openrtb2.MarkupNative
			return nil
		case "video":
			bid.MType = openrtb2.MarkupVideo
			return nil
		}
	}

	var mType openrtb2.MarkupType
	var typeCnt = 0
	for _, imp := range imps {
		if imp.ID == bid.ImpID {
			if imp.Banner != nil {
				typeCnt += 1
				mType = openrtb2.MarkupBanner
			}
			if imp.Native != nil {
				typeCnt += 1
				mType = openrtb2.MarkupNative
			}
			if imp.Video != nil {
				typeCnt += 1
				mType = openrtb2.MarkupVideo
			}
		}
	}
	if typeCnt == 1 {
		bid.MType = mType
		return nil
	}
	return fmt.Errorf("unable to fetch mediaType in multi-format: %s", bid.ImpID)
}

func buildEndpoint(aaxUrl, hostUrl string) string {

	if len(hostUrl) == 0 {
		return aaxUrl
	}
	urlObject, err := url.Parse(aaxUrl)
	if err != nil {
		return aaxUrl
	}
	values := urlObject.Query()
	values.Add("src", hostUrl)
	urlObject.RawQuery = values.Encode()
	return urlObject.String()
}
