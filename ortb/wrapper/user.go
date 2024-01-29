package wrapper

import (
	"encoding/json"

	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/openrtb_ext"
	"github.com/prebid/prebid-server/v2/util/jsonutil"
	"github.com/prebid/prebid-server/v2/util/maputil"
	"github.com/prebid/prebid-server/v2/util/sliceutil"
)

type Trackable[T any] struct {
	Defined bool
	Dirty   bool
	Value   T
}

// all fields must be pointer to understand the concept of "not here", or wrapped in nillable struct
// always return a copy, next direct - to avoid
type UserExt struct {
	ext         map[string]json.RawMessage
	extDirty    bool
	consent     Trackable[string]
	prebid      *openrtb_ext.ExtUserPrebid
	prebidDirty bool
	eids        *[]openrtb2.EID
	eidsDirty   bool
	// consentedProvidersSettingsIn       *ConsentedProvidersSettingsIn
	// consentedProvidersSettingsInDirty  bool
	// consentedProvidersSettingsOut      *ConsentedProvidersSettingsOut
	// consentedProvidersSettingsOutDirty bool
}

func parseUserExt(ext json.RawMessage) (UserExt, error) {
	w := UserExt{
		ext: make(map[string]json.RawMessage),
	}

	if err := jsonutil.Unmarshal(ext, &w.ext); err != nil {
		return UserExt{}, err
	}

	if consentJson, hasConsent := w.ext["consent"]; hasConsent && consentJson != nil {
		if err := jsonutil.Unmarshal(consentJson, &w.consent); err != nil {
			return UserExt{}, err
		}
	}

	// prebidJson, hasPrebid := ue.ext[prebidKey]
	// if hasPrebid {
	// 	ue.prebid = &ExtUserPrebid{}
	// }
	// if prebidJson != nil {
	// 	if err := jsonutil.Unmarshal(prebidJson, ue.prebid); err != nil {
	// 		return err
	// 	}
	// }

	// eidsJson, hasEids := ue.ext[eidsKey]
	// if hasEids {
	// 	ue.eids = &[]openrtb2.EID{}
	// }
	// if eidsJson != nil {
	// 	if err := jsonutil.Unmarshal(eidsJson, ue.eids); err != nil {
	// 		return err
	// 	}
	// }

	// consentedProviderSettingsInJson, hasCPSettingsIn := ue.ext[consentedProvidersSettingsStringKey]
	// if hasCPSettingsIn {
	// 	ue.consentedProvidersSettingsIn = &ConsentedProvidersSettingsIn{}
	// }
	// if consentedProviderSettingsInJson != nil {
	// 	if err := jsonutil.Unmarshal(consentedProviderSettingsInJson, ue.consentedProvidersSettingsIn); err != nil {
	// 		return err
	// 	}
	// }

	// consentedProviderSettingsOutJson, hasCPSettingsOut := ue.ext[consentedProvidersSettingsListKey]
	// if hasCPSettingsOut {
	// 	ue.consentedProvidersSettingsOut = &ConsentedProvidersSettingsOut{}
	// }
	// if consentedProviderSettingsOutJson != nil {
	// 	if err := jsonutil.Unmarshal(consentedProviderSettingsOutJson, ue.consentedProvidersSettingsOut); err != nil {
	// 		return err
	// 	}
	// }

	return w, nil
}

// create a new ext json
func (ue *UserExt) marshal() (json.RawMessage, error) {
	if ue.consentDirty {
		if ue.consent != nil && len(*ue.consent) > 0 {
			consentJson, err := jsonutil.Marshal(ue.consent)
			if err != nil {
				return nil, err
			}
			ue.ext[consentKey] = json.RawMessage(consentJson)
		} else {
			delete(ue.ext, consentKey)
		}
		ue.consentDirty = false
	}

	if ue.prebidDirty {
		if ue.prebid != nil {
			prebidJson, err := jsonutil.Marshal(ue.prebid)
			if err != nil {
				return nil, err
			}
			if len(prebidJson) > jsonEmptyObjectLength {
				ue.ext[prebidKey] = json.RawMessage(prebidJson)
			} else {
				delete(ue.ext, prebidKey)
			}
		} else {
			delete(ue.ext, prebidKey)
		}
		ue.prebidDirty = false
	}

	if ue.consentedProvidersSettingsInDirty {
		if ue.consentedProvidersSettingsIn != nil {
			cpSettingsJson, err := jsonutil.Marshal(ue.consentedProvidersSettingsIn)
			if err != nil {
				return nil, err
			}
			if len(cpSettingsJson) > jsonEmptyObjectLength {
				ue.ext[consentedProvidersSettingsStringKey] = json.RawMessage(cpSettingsJson)
			} else {
				delete(ue.ext, consentedProvidersSettingsStringKey)
			}
		} else {
			delete(ue.ext, consentedProvidersSettingsStringKey)
		}
		ue.consentedProvidersSettingsInDirty = false
	}

	if ue.consentedProvidersSettingsOutDirty {
		if ue.consentedProvidersSettingsOut != nil {
			cpSettingsJson, err := jsonutil.Marshal(ue.consentedProvidersSettingsOut)
			if err != nil {
				return nil, err
			}
			if len(cpSettingsJson) > jsonEmptyObjectLength {
				ue.ext[consentedProvidersSettingsListKey] = json.RawMessage(cpSettingsJson)
			} else {
				delete(ue.ext, consentedProvidersSettingsListKey)
			}
		} else {
			delete(ue.ext, consentedProvidersSettingsListKey)
		}
		ue.consentedProvidersSettingsOutDirty = false
	}

	if ue.eidsDirty {
		if ue.eids != nil && len(*ue.eids) > 0 {
			eidsJson, err := jsonutil.Marshal(ue.eids)
			if err != nil {
				return nil, err
			}
			ue.ext[eidsKey] = json.RawMessage(eidsJson)
		} else {
			delete(ue.ext, eidsKey)
		}
		ue.eidsDirty = false
	}

	ue.extDirty = false
	if len(ue.ext) == 0 {
		return nil, nil
	}
	return jsonutil.Marshal(ue.ext)
}

func (ue *UserExt) Dirty() bool {
	return ue.extDirty || ue.eidsDirty || ue.prebidDirty || ue.consentDirty || ue.consentedProvidersSettingsInDirty || ue.consentedProvidersSettingsOutDirty
}

// return a copy to the map, so changes have to through SetExt so we can track Dirty state
func (ue *UserExt) GetExt() map[string]json.RawMessage {
	return maputil.Clone(ue.ext)
}

func (ue *UserExt) SetExt(ext map[string]json.RawMessage) {
	ue.ext = ext
	ue.extDirty = true
}

// determine between "missing" vs "empty"
// - pointer, but now have to deal with nul. how about optional?

func (ue *UserExt) GetConsent() *string {
	if ue.consent == nil {
		return nil
	}
	consent := *ue.consent
	return &consent
}

func (ue *UserExt) SetConsent(consent *string) {
	ue.consent = consent
	ue.consentDirty = true
}

// GetConsentedProvidersSettingsIn() returns a reference to a copy of ConsentedProvidersSettingsIn, a struct that
// has a string field formatted as a Google's Additional Consent string
func (ue *UserExt) GetConsentedProvidersSettingsIn() *ConsentedProvidersSettingsIn {
	if ue.consentedProvidersSettingsIn == nil {
		return nil
	}
	consentedProvidersSettingsIn := *ue.consentedProvidersSettingsIn
	return &consentedProvidersSettingsIn
}

// SetConsentedProvidersSettingsIn() sets ConsentedProvidersSettingsIn, a struct that
// has a string field formatted as a Google's Additional Consent string
func (ue *UserExt) SetConsentedProvidersSettingsIn(cpSettings *ConsentedProvidersSettingsIn) {
	ue.consentedProvidersSettingsIn = cpSettings
	ue.consentedProvidersSettingsInDirty = true
}

// GetConsentedProvidersSettingsOut() returns a reference to a copy of ConsentedProvidersSettingsOut, a struct that
// has an int array field listing Google's Additional Consent string elements
func (ue *UserExt) GetConsentedProvidersSettingsOut() *ConsentedProvidersSettingsOut {
	if ue.consentedProvidersSettingsOut == nil {
		return nil
	}
	consentedProvidersSettingsOut := *ue.consentedProvidersSettingsOut
	return &consentedProvidersSettingsOut
}

// SetConsentedProvidersSettingsIn() sets ConsentedProvidersSettingsOut, a struct that
// has an int array field listing Google's Additional Consent string elements. This
// function overrides an existing ConsentedProvidersSettingsOut object, if any
func (ue *UserExt) SetConsentedProvidersSettingsOut(cpSettings *ConsentedProvidersSettingsOut) {
	if cpSettings == nil {
		return
	}

	ue.consentedProvidersSettingsOut = cpSettings
	ue.consentedProvidersSettingsOutDirty = true
	return
}

func (ue *UserExt) GetPrebid() *ExtUserPrebid {
	if ue.prebid == nil {
		return nil
	}
	prebid := *ue.prebid
	return &prebid
}

func (ue *UserExt) SetPrebid(prebid *ExtUserPrebid) {
	ue.prebid = prebid
	ue.prebidDirty = true
}

func (ue *UserExt) GetEid() *[]openrtb2.EID {
	if ue.eids == nil {
		return nil
	}
	eids := *ue.eids
	return &eids
}

func (ue *UserExt) SetEid(eid *[]openrtb2.EID) {
	ue.eids = eid
	ue.eidsDirty = true
}

func (ue *UserExt) Clone() *UserExt {
	if ue == nil {
		return nil
	}
	clone := *ue
	clone.ext = maputil.Clone(ue.ext)

	if ue.consent != nil {
		clonedConsent := *ue.consent
		clone.consent = &clonedConsent
	}

	if ue.prebid != nil {
		clone.prebid = &ExtUserPrebid{}
		clone.prebid.BuyerUIDs = maputil.Clone(ue.prebid.BuyerUIDs)
	}

	if ue.eids != nil {
		clonedEids := make([]openrtb2.EID, len(*ue.eids))
		for i, eid := range *ue.eids {
			newEid := eid
			newEid.UIDs = sliceutil.Clone(eid.UIDs)
			clonedEids[i] = newEid
		}
		clone.eids = &clonedEids
	}

	if ue.consentedProvidersSettingsIn != nil {
		clone.consentedProvidersSettingsIn = &ConsentedProvidersSettingsIn{ConsentedProvidersString: ue.consentedProvidersSettingsIn.ConsentedProvidersString}
	}
	if ue.consentedProvidersSettingsOut != nil {
		clone.consentedProvidersSettingsOut = &ConsentedProvidersSettingsOut{ConsentedProvidersList: sliceutil.Clone(ue.consentedProvidersSettingsOut.ConsentedProvidersList)}
	}

	return &clone
}
