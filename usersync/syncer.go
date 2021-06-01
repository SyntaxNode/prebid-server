package usersync

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	validator "github.com/asaskevich/govalidator"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/macros"
	"github.com/prebid/prebid-server/privacy"
)

var (
	errNoSyncTypesProvided        = errors.New("no sync types provided")
	errNoSyncTypesSupported       = errors.New("no sync types supported")
	errDefaultTypeMissingIFrame   = errors.New("default is set to iframe but no iframe endpoint is configured")
	errDefaultTypeMissingRedirect = errors.New("default is set to redirect but no redirect endpoint is configured")
)

// Syncer represents the user sync configuration for a bidder or a shared set of bidders.
type Syncer interface {
	// Key is the name of the syncer as stored in the user's cookie. This is often, but not
	// necessarily, a one-to-one mapping with a bidder.
	Key() string

	// SupportsType returns true if the syncer supports at least one of the specified sync types.
	SupportsType(syncTypes []SyncType) bool

	// GetSync returns a user sync for the user's device to perform, or an error if the none of the
	// sync types are supported or if macro substitution fails.
	GetSync(syncTypes []SyncType, privacyPolicies privacy.Policies) (Sync, error)
}

// Sync represents a user sync to be performed by the user's device.
type Sync struct {
	URL         string
	Type        SyncType
	SupportCORS bool
}

type standardSyncer struct {
	key             string
	defaultSyncType SyncType
	iframe          *template.Template
	redirect        *template.Template
	supportCORS     bool
}

const (
	setuidSyncTypeIFrame   = "b"
	setuidSyncTypeRedirect = "i"
)

var errEndpointRequired = errors.New("at least one endpoint (iframe or redirect) is required")
var errDefaultSyncTypeRequired = errors.New("default sync type is required when more then one sync endpoint is configured")

// NewSyncer creates a new Syncer from the provided configuration, or an error if macro substition
// fails or an endpoint url is invalid.
func NewSyncer(hostConfig config.UserSync, syncerConfig config.Syncer) (Syncer, error) {
	if syncerConfig.IFrame == nil && syncerConfig.Redirect == nil {
		return nil, errEndpointRequired
	}

	syncer := standardSyncer{
		key:         syncerConfig.Key,
		supportCORS: syncerConfig.SupportCORS,
	}

	if defaultSyncType, err := resolveDefaultSyncType(syncerConfig); err != nil {
		return nil, err
	} else {
		syncer.defaultSyncType = defaultSyncType
	}

	if syncerConfig.IFrame != nil {
		var err error
		syncer.iframe, err = buildTemplate(syncerConfig.Key, setuidSyncTypeIFrame, hostConfig, *syncerConfig.IFrame)
		if err != nil {
			return nil, fmt.Errorf("iframe %v", err)
		}
		if err := validateTemplate(syncer.iframe); err != nil {
			return nil, fmt.Errorf("iframe %v", err)
		}
	}

	if syncerConfig.Redirect != nil {
		var err error
		syncer.redirect, err = buildTemplate(syncerConfig.Key, setuidSyncTypeRedirect, hostConfig, *syncerConfig.Redirect)
		if err != nil {
			return nil, fmt.Errorf("redirect %v", err)
		}
		if err := validateTemplate(syncer.redirect); err != nil {
			return nil, fmt.Errorf("redirect %v", err)
		}
	}

	return syncer, nil
}

func resolveDefaultSyncType(syncerConfig config.Syncer) (SyncType, error) {
	if syncerConfig.Default == "" {
		if syncerConfig.IFrame != nil && syncerConfig.Redirect != nil {
			return SyncTypeUnknown, errDefaultSyncTypeRequired
		} else if syncerConfig.IFrame != nil {
			return SyncTypeIFrame, nil
		} else {
			return SyncTypeRedirect, nil
		}
	}

	if syncType, err := SyncTypeParse(syncerConfig.Default); err == nil {
		switch syncType {
		case SyncTypeIFrame:
			if syncerConfig.IFrame == nil {
				return SyncTypeUnknown, errDefaultTypeMissingIFrame
			}
		case SyncTypeRedirect:
			if syncerConfig.Redirect == nil {
				return SyncTypeUnknown, errDefaultTypeMissingRedirect
			}
		}
		return syncType, nil
	}

	return SyncTypeUnknown, fmt.Errorf("invalid default sync type '%s'", syncerConfig.Default)
}

var (
	macroRegexExternalHost = regexp.MustCompile(`{{\s*.ExternalURL\s*}}`)
	macroRegexSyncerKey    = regexp.MustCompile(`{{\s*.SyncerKey\s*}}`)
	macroRegexSyncType     = regexp.MustCompile(`{{\s*.SyncType\s*}}`)
	macroRegexUserMacro    = regexp.MustCompile(`{{\s*.UserMacro\s*}}`)
	macroRegexRedirect     = regexp.MustCompile(`{{\s*.RedirectURL\s*}}`)
	macroRegex             = regexp.MustCompile(`{{.*?}}`)
)

func buildTemplate(key, syncTypeValue string, hostConfig config.UserSync, syncerEndpoint config.SyncerEndpoint) (*template.Template, error) {
	redirectTemplate := syncerEndpoint.RedirectURL
	if redirectTemplate == "" {
		redirectTemplate = hostConfig.RedirectURL
	}

	externalURL := syncerEndpoint.ExternalURL
	if externalURL == "" {
		externalURL = hostConfig.ExternalURL
	}

	redirectURL := macroRegexExternalHost.ReplaceAllLiteralString(redirectTemplate, externalURL)
	redirectURL = macroRegexSyncerKey.ReplaceAllLiteralString(redirectURL, key)
	redirectURL = macroRegexSyncType.ReplaceAllLiteralString(redirectURL, syncTypeValue)
	redirectURL = macroRegexUserMacro.ReplaceAllLiteralString(redirectURL, syncerEndpoint.UserMacro)
	redirectURL = escapeTemplate(redirectURL)

	url := macroRegexRedirect.ReplaceAllString(syncerEndpoint.URL, redirectURL)

	templateName := strings.ToLower(key) + "_usersync_url"
	return template.New(templateName).Parse(url)
}

// escapeTemplate url encodes a string template leaving the macro tags unaffected.
func escapeTemplate(x string) string {
	escaped := strings.Builder{}

	i := 0
	for _, m := range macroRegex.FindAllStringIndex(x, -1) {
		escaped.WriteString(url.QueryEscape(x[i:m[0]]))
		escaped.WriteString(x[m[0]:m[1]])
		i = m[1]
	}
	escaped.WriteString(url.QueryEscape(x[i:]))

	return escaped.String()
}

var templateTestValues = macros.UserSyncTemplateParams{
	GDPR:        "anyGDPR",
	GDPRConsent: "anyGDPRConsent",
	USPrivacy:   "anyCCPAConsent",
}

func validateTemplate(template *template.Template) error {
	url, err := macros.ResolveMacros(template, templateTestValues)
	if err != nil {
		return err
	}

	if !validator.IsURL(url) || !validator.IsRequestURL(url) {
		return fmt.Errorf(`composed url: "%s" is invalid`, url)
	}

	return nil
}

func (s standardSyncer) Key() string {
	return s.key
}

func (s standardSyncer) SupportsType(syncTypes []SyncType) bool {
	supported := s.filterSupportedSyncTypes(syncTypes)
	return len(supported) > 0
}

func (s standardSyncer) filterSupportedSyncTypes(syncTypes []SyncType) []SyncType {
	supported := make([]SyncType, 0, len(syncTypes))
	for _, syncType := range syncTypes {
		switch syncType {
		case SyncTypeIFrame:
			if s.iframe != nil {
				supported = append(supported, SyncTypeIFrame)
			}
		case SyncTypeRedirect:
			if s.redirect != nil {
				supported = append(supported, SyncTypeRedirect)
			}
		}
	}
	return supported
}

func (s standardSyncer) GetSync(syncTypes []SyncType, privacyPolicies privacy.Policies) (Sync, error) {
	syncType, err := s.chooseSyncType(syncTypes)
	if err != nil {
		return Sync{}, err
	}

	syncTemplate := s.chooseTemplate(syncType)

	url, err := macros.ResolveMacros(syncTemplate, macros.UserSyncTemplateParams{
		GDPR:        privacyPolicies.GDPR.Signal,
		GDPRConsent: privacyPolicies.GDPR.Consent,
		USPrivacy:   privacyPolicies.CCPA.Consent,
	})
	if err != nil {
		return Sync{}, err
	}

	sync := Sync{
		URL:         url,
		Type:        syncType,
		SupportCORS: s.supportCORS,
	}
	return sync, nil
}

func (s standardSyncer) chooseSyncType(syncTypes []SyncType) (SyncType, error) {
	if len(syncTypes) == 0 {
		return SyncTypeUnknown, errNoSyncTypesProvided
	}

	supported := s.filterSupportedSyncTypes(syncTypes)
	if len(supported) == 0 {
		return SyncTypeUnknown, errNoSyncTypesSupported
	}

	// prefer default type
	for _, syncType := range supported {
		if syncType == s.defaultSyncType {
			return syncType, nil
		}
	}

	return syncTypes[0], nil
}

func (s standardSyncer) chooseTemplate(syncType SyncType) *template.Template {
	switch syncType {
	case SyncTypeIFrame:
		return s.iframe
	case SyncTypeRedirect:
		return s.redirect
	default:
		return nil
	}
}

type bidderInfoWithName struct {
	bidderName string
	bidderInfo config.BidderInfo
}

// func NewSyncersFromBidderInfos(hostConfig config.UserSync, bidderInfos config.BidderInfos) (map[string]Syncer, error) {
// 	bidderInfosBySyncerKey := make(map[string][]bidderInfoWithName)
// 	for bidderName, bidderInfo := range bidderInfos {
// 		s := bidderInfo.Syncer
// 		if s != nil && (s.IFrame != nil || s.Redirect != nil) {
// 			bidderInfosBySyncerKey[s.Key] = append(bidderInfosBySyncerKey[s.Key], bidderInfoWithName{bidderName, bidderInfo})
// 		}
// 	}

// 	// build syncers
// 	syncersByBidder := map[string]Syncer{}
// 	var errs []error
// 	for _, bidderInfos := range bidderInfosBySyncerKey {
// 		syncerConfig, err := getPrimarySyncerConfig(bidderInfos)
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}

// 		syncer, err := NewSyncer(hostConfig, syncerConfig)
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}

// 		for _, b := range bidderInfos {
// 			syncersByBidder[b.bidderName] = syncer
// 		}
// 	}

// 	if len(errs) > 0 {
// 		return nil, errortypes.NewAggregateError("msg", errs)
// 	}

// 	return syncersByBidder, nil
// }

// func getPrimarySyncerConfig(bidderInfos []bidderInfoWithName) (*config.Syncer, error) {
// 	if len(bidderInfos) == 1 {
// 		return bidderInfos[0].bidderInfo.Syncer, nil
// 	}

// 	// if multiple, ensure just one has endpoints defined + return it
// 	// if not just one, return error

// 	return nil, nil
// }
