package adapters

import (
	"github.com/prebid/openrtb/v19/openrtb2"
)

func FallbackToMTypeFromImpWithDefault(bid *openrtb2.Bid, imps []openrtb2.Imp, typePriority []openrtb2.MarkupType, typeDefault openrtb2.MarkupType) {
	// use mtype from bid, if available
	if bid.MType != 0 {
		return
	}

	// use mtype from impression, if found (should be)
	for _, imp := range imps {
		if imp.ID == bid.ImpID {
			// find match from priority
		}
	}

	// fallback to default
	bid.MType = typeDefault
}
