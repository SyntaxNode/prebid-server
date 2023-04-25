package adapters

import (
	"encoding/json"
	"fmt"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/errortypes"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type FallbackToMTypeFromImpWithDefault struct {
	Imps         []openrtb2.Imp
	TypePriority []openrtb2.MarkupType
	TypeDefault  openrtb2.MarkupType
}

func (f FallbackToMTypeFromImpWithDefault) Apply(bid *openrtb2.Bid) {
	// use mtype from bid, if available
	if bid.MType != 0 {
		return
	}

	// use mtype from impression, if found (should be)
	for _, imp := range f.Imps {
		if imp.ID == bid.ImpID {
			// check for mtype per given priority
			for _, p := range f.TypePriority {
				switch p {
				case openrtb2.MarkupBanner:
					if imp.Banner != nil {
						bid.MType = openrtb2.MarkupBanner
						return
					}
				case openrtb2.MarkupVideo:
					if imp.Video != nil {
						bid.MType = openrtb2.MarkupVideo
						return
					}
				case openrtb2.MarkupAudio:
					if imp.Audio != nil {
						bid.MType = openrtb2.MarkupAudio
						return
					}
				case openrtb2.MarkupNative:
					if imp.Native != nil {
						bid.MType = openrtb2.MarkupNative
						return
					}
				}
			}
		}
	}

	// fallback to default
	bid.MType = f.TypeDefault
}

type FallbackToMTypeFromPrebidExt struct {
}

func (f FallbackToMTypeFromPrebidExt) Apply(bid *openrtb2.Bid) error {
	// use mtype from bid, if available
	if bid.MType != 0 {
		return nil
	}

	// use mtype from ext.prebid.type, if available
	var bidExt openrtb_ext.ExtBid
	if err := json.Unmarshal(bid.Ext, &bidExt); err != nil || bidExt.Prebid == nil {
		return &errortypes.BadServerResponse{
			Message: fmt.Sprintf("missing ext.prebid.type in bid for impression: %s", bid.ImpID),
		}
	}

	mType, err := MTypeFromBidType(bidExt.Prebid.Type)
	if err != nil {
		return &errortypes.BadServerResponse{
			Message: fmt.Sprintf("invalid ext.prebid.type '%s' in bid for impression: %s", bidExt.Prebid.Type, bid.ImpID),
		}
	}

	bid.MType = mType
	return nil
}

// var bidExt openrtb_ext.ExtBid
func MTypeFromBidType(p openrtb_ext.BidType) (openrtb2.MarkupType, error) {
	switch p {
	case openrtb_ext.BidTypeBanner:
		return openrtb2.MarkupBanner, nil
	case openrtb_ext.BidTypeVideo:
		return openrtb2.MarkupVideo, nil
	case openrtb_ext.BidTypeAudio:
		return openrtb2.MarkupAudio, nil
	case openrtb_ext.BidTypeNative:
		return openrtb2.MarkupNative, nil
	}

	return 0, fmt.Errorf("invalid BidType: %s", p)
}
