package wrapper

import "github.com/prebid/openrtb/v19/openrtb2"

// getting a wrapper will always just return the wrapper (so.. give direct access?)
// accessing a field of a wrapper will require get / set to track Dirty flag

// pass BidRequest by reference is almost an equivlanet of the BidRequest itself. no more nil

// rules:
// - we already need the callers to not modify the .ext directly, as changes will later be ignore
//    - failure to comply harms caller, maintains object integrity

// - if we return a pointer to a value, it may be modified

func example() {
	r, _ := NewBidRequest("{}")
	r.UserExt.GetPrebid() // returns a prebid working copy. changes local until calling SetPrebid()
	r.UserExt.consent.Defined
}

type BidRequest struct {
	openrtb2.BidRequest
	UserExt UserExt
}

// wr, err := wrapper.WrapBidRequest(r)
// - after this, if we get no error we're safe until we get to marshal

func NewBidRequest(r openrtb2.BidRequest) (*BidRequest, error) {
	w := &BidRequest{BidRequest: r}

	if err := w.wrapUser(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *BidRequest) wrapUser() error {
	if w.User == nil {
		return nil
	}

	parsed, err := parseUserExt(w.User.Ext)
	if err != nil {
		return err
	}

	w.UserExt = parsed
	return nil
}
