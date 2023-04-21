package bidutil

import (
	"github.com/prebid/openrtb/v19/openrtb2"
)

type MTypePriority []openrtb2.MarkupType

func FallbackToMTypeFromImpWithDefault(bid *openrtb2.Bid, imps []openrtb2.Imp, typePriority MTypePriority, typeDefault openrtb2.MarkupType) {
	// use mtype from bid, if available
	if bid.MType != 0 {
		return
	}

	// use mtype from impression, if found (should be)
	for _, imp := range imps {
		if imp.ID == bid.ImpID {
			// check for mtype per given priority
			for _, p := range typePriority {
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
	bid.MType = typeDefault
}
