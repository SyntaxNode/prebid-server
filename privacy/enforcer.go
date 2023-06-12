package privacy

import "github.com/prebid/prebid-server/openrtb_ext"

type EnforceResult int // or maybe ActivityResult?

const (
	EnforceAbstain EnforceResult = iota
	EnforceAllow
	EnforceDeny
)

type ActivityControl struct {
	plans map[Activity]EnforcementPlan
}

func (e ActivityControl) Allow(activity Activity, request openrtb_ext.RequestWrapper, target ScopedName) EnforceResult {
	plan, planDefined := e.plans[activity]

	if !planDefined {
		return EnforceAbstain
	}

	return plan.Enforce(request, target)
}

// allow this to be created from acitivty config, which veronika will get from the account config root object
// maybe call this ActivityPlan?
type EnforcementPlan struct {
	defaultResult EnforceResult
	rules         []EnforcementRule
}

func (p EnforcementPlan) Allow(request openrtb_ext.RequestWrapper, target ScopedName) EnforceResult {
	for _, rule := range p.rules {
		result := rule.Allow(request, target) // exit on first non-abstain response
		if result == EnforceAllow || result == EnforceDeny {
			return result
		}
	}
	return p.defaultResult
}

// maybe call this ActivityRule?
type EnforcementRule interface {
	Allow(request openrtb_ext.RequestWrapper, target ScopedName) EnforceResult
}

type ComponentEnforcementRule struct {
	componentName []ScopedName
	componentType []string
	// include gppSectionId from 3.5
	// include geo from 3.5
	allowed bool // behavior if rule matches. can be either true=allow or false=deny. result is abstain if the rule doesn't match
}

func (r EnforcementRule) Allow(request openrtb_ext.RequestWrapper, target ScopedName) EnforceResult {
	// all string comparisons in this section are case sensitive
	// doc: https://docs.google.com/document/d/1dRxFUFmhh2jGanzGZvfkK_6jtHPpHXWD7Qsi6KEugeE/edit
	// the doc details the boolean operations.
	//  - "or" within each field (componentName, componentType
	//  - "and" between the rules present. empty fields are ignored (refer to doc for details)

	// componentName
	// check for matching scoped name. a wildcard is allowed for the name in which any target with the same scope is matched

	// componentType
	// can either act as a scope wildcard or meta targeting. can be scope "bidder", "analytics", maybe others.
	// may also be "rtd" meta. you need to pass that through somehow, perhaps as targetMeta? targetMeta can be a slice. should be small enough that search speed isn't a concern.

	// gppSectionId
	// check if id is present in the gppsid slice. no parsing of gpp should happen here.

	// geo
	// simple filter on the req.user section
	return 0
}

// the default scope should be hardcoded as bidder
// ex: "bidder.appnexus", "bidder.*", "appnexus", "analytics.pubmatic"
// TODO: add parsing helpers
type ScopedName struct {
	Scope string
	Name  string
}

// ex: "USA.VA", "USA". see all comments in https://github.com/prebid/prebid-server/issues/2622
// TODO: add parsing helpers
type Geo struct {
	Country string
	Region  string
}
