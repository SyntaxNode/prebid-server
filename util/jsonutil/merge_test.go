package jsonutil

import (
	"encoding/json"
	"testing"

	"github.com/prebid/prebid-server/v2/util/sliceutil"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeClonePtr(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		var (
			banner      = &openrtb2.Banner{ID: "1"}
			imp         = &openrtb2.Imp{Banner: banner}
			impOriginal = imp
		)

		// root objects are not cloned
		err := MergeClone(imp, []byte(`{"banner":{"id":"4"}}`))
		require.NoError(t, err)

		assert.Same(t, impOriginal, imp, "imp-ref")
		assert.NotSame(t, imp.Banner, banner, "banner-ref")
	})

	t.Run("embedded-nil", func(t *testing.T) {
		var (
			banner = &openrtb2.Banner{ID: "1"}
			video  = &openrtb2.Video{PodID: "a"}
			imp    = &openrtb2.Imp{Banner: banner, Video: video}
		)

		err := MergeClone(imp, []byte(`{"banner":null}`))
		require.NoError(t, err)

		assert.NotSame(t, banner, imp.Banner, "banner-ref")
		assert.Same(t, video, imp.Video, "video")
		assert.Nil(t, imp.Banner, "banner-nil")
	})

	t.Run("embedded-struct", func(t *testing.T) {
		var (
			banner = &openrtb2.Banner{ID: "1"}
			video  = &openrtb2.Video{PodID: "a"}
			imp    = &openrtb2.Imp{Banner: banner, Video: video}
		)

		err := MergeClone(imp, []byte(`{"banner":{"id":"2"}}`))
		require.NoError(t, err)

		assert.NotSame(t, banner, imp.Banner, "banner-ref")
		assert.Same(t, video, imp.Video, "video-ref")
		assert.Equal(t, "1", banner.ID, "id-original")
		assert.Equal(t, "2", imp.Banner.ID, "id-clone")
	})

	t.Run("embedded-int", func(t *testing.T) {
		var (
			clickbrowser = int8(1)
			imp          = &openrtb2.Imp{ClickBrowser: &clickbrowser}
		)

		err := MergeClone(imp, []byte(`{"clickbrowser":2}`))
		require.NoError(t, err)

		require.NotNil(t, imp.ClickBrowser, "clickbrowser-nil")
		assert.NotSame(t, clickbrowser, imp.ClickBrowser, "clickbrowser-ref")
		assert.Equal(t, int8(2), *imp.ClickBrowser, "clickbrowser-val")
	})

	t.Run("invalid-null", func(t *testing.T) {
		var (
			banner = &openrtb2.Banner{ID: "1"}
			imp    = &openrtb2.Imp{Banner: banner}
		)

		err := MergeClone(imp, []byte(`{"banner":nul}`))
		require.EqualError(t, err, "cannot unmarshal openrtb2.Imp.Banner: expect ull")
	})

	t.Run("invalid-malformed", func(t *testing.T) {
		var (
			banner = &openrtb2.Banner{ID: "1"}
			imp    = &openrtb2.Imp{Banner: banner}
		)

		err := MergeClone(imp, []byte(`{"banner":malformed}`))
		require.EqualError(t, err, "cannot unmarshal openrtb2.Imp.Banner: expect { or n, but found m")
	})
}

func TestMergeCloneSlice(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		var (
			iframeBuster = []string{"a", "b"}
			imp          = &openrtb2.Imp{IframeBuster: iframeBuster}
		)

		err := MergeClone(imp, []byte(`{"iframeBuster":null}`))
		require.NoError(t, err)

		assert.Equal(t, []string{"a", "b"}, iframeBuster, "iframeBuster-val")
		assert.Nil(t, imp.IframeBuster, "iframeBuster-nil")
	})

	t.Run("one", func(t *testing.T) {
		var (
			iframeBuster = []string{"a"}
			imp          = &openrtb2.Imp{IframeBuster: iframeBuster}
		)

		err := MergeClone(imp, []byte(`{"iframeBuster":["b"]}`))
		require.NoError(t, err)

		assert.NotSame(t, iframeBuster, imp.IframeBuster, "ref")
		assert.Equal(t, []string{"a"}, iframeBuster, "original-val")
		assert.Equal(t, []string{"b"}, imp.IframeBuster, "new-val")
	})

	t.Run("many", func(t *testing.T) {
		var (
			iframeBuster = []string{"a"}
			imp          = &openrtb2.Imp{IframeBuster: iframeBuster}
		)

		err := MergeClone(imp, []byte(`{"iframeBuster":["b", "c"]}`))
		require.NoError(t, err)

		assert.NotSame(t, iframeBuster, imp.IframeBuster, "ref")
		assert.Equal(t, []string{"a"}, iframeBuster, "original-val")
		assert.Equal(t, []string{"b", "c"}, imp.IframeBuster, "new-val")
	})

	t.Run("invalid-null", func(t *testing.T) {
		var (
			iframeBuster = []string{"a"}
			imp          = &openrtb2.Imp{IframeBuster: iframeBuster}
		)

		err := MergeClone(imp, []byte(`{"iframeBuster":nul}`))
		require.EqualError(t, err, "cannot unmarshal openrtb2.Imp.IframeBuster: expect ull")
	})

	t.Run("invalid-malformed", func(t *testing.T) {
		var (
			iframeBuster = []string{"a"}
			imp          = &openrtb2.Imp{IframeBuster: iframeBuster}
		)

		err := MergeClone(imp, []byte(`{"iframeBuster":malformed}`))
		require.EqualError(t, err, "cannot unmarshal openrtb2.Imp.IframeBuster: decode slice: expect [ or n, but found m")
	})
}

func TestMergeCloneMap(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		var (
			testMap = map[string]int{"a": 1, "b": 2}
			test    = &struct {
				Foo map[string]int `json:"foo"`
			}{Foo: testMap}
		)

		err := MergeClone(test, []byte(`{"foo":null}`))
		require.NoError(t, err)

		assert.NotSame(t, testMap, test.Foo, "ref")
		assert.Equal(t, map[string]int{"a": 1, "b": 2}, testMap, "val")
		assert.Nil(t, test.Foo, "nil")
	})

	t.Run("key-string", func(t *testing.T) {
		var (
			testMap = map[string]int{"a": 1, "b": 2}
			test    = &struct {
				Foo map[string]int `json:"foo"`
			}{Foo: testMap}
		)

		err := MergeClone(test, []byte(`{"foo":{"c":3}}`))
		require.NoError(t, err)

		assert.NotSame(t, testMap, test.Foo)
		assert.Equal(t, map[string]int{"a": 1, "b": 2}, testMap, "original-val")
		assert.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, test.Foo, "new-val")

		// verify modifications don't corrupt original
		testMap["a"] = 10
		assert.Equal(t, map[string]int{"a": 10, "b": 2}, testMap, "mod-original-val")
		assert.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, test.Foo, "mod-ew-val")
	})

	t.Run("key-numeric", func(t *testing.T) {
		var (
			testMap = map[int]string{1: "a", 2: "b"}
			test    = &struct {
				Foo map[int]string `json:"foo"`
			}{Foo: testMap}
		)

		err := MergeClone(test, []byte(`{"foo":{"3":"c"}}`))
		require.NoError(t, err)

		assert.NotSame(t, testMap, test.Foo)
		assert.Equal(t, map[int]string{1: "a", 2: "b"}, testMap, "original-val")
		assert.Equal(t, map[int]string{1: "a", 2: "b", 3: "c"}, test.Foo, "new-val")

		// verify modifications don't corrupt original
		testMap[1] = "z"
		assert.Equal(t, map[int]string{1: "z", 2: "b"}, testMap, "mod-original-val")
		assert.Equal(t, map[int]string{1: "a", 2: "b", 3: "c"}, test.Foo, "mod-ew-val")
	})

	t.Run("invalid-null", func(t *testing.T) {
		var (
			testMap = map[int]string{1: "a", 2: "b"}
			test    = &struct {
				Foo map[int]string `json:"foo"`
			}{Foo: testMap}
		)

		err := MergeClone(test, []byte(`{"foo":nul}`))
		require.EqualError(t, err, "cannot unmarshal Foo: expect ull")
	})

	t.Run("invalid-malformed", func(t *testing.T) {
		var (
			testMap = map[int]string{1: "a", 2: "b"}
			test    = &struct {
				Foo map[int]string `json:"foo"`
			}{Foo: testMap}
		)

		err := MergeClone(test, []byte(`{"foo":malformed}`))
		require.EqualError(t, err, "cannot unmarshal Foo: expect { or n, but found m")
	})
}

func TestMergeCloneExt(t *testing.T) {
	testCases := []struct {
		name          string
		givenExisting json.RawMessage
		givenIncoming json.RawMessage
		expectedExt   json.RawMessage
		expectedErr   string
	}{
		{
			name:          "both-populated",
			givenExisting: json.RawMessage(`{"a":1,"b":2}`),
			givenIncoming: json.RawMessage(`{"b":200,"c":3}`),
			expectedExt:   json.RawMessage(`{"a":1,"b":200,"c":3}`),
		},
		{
			name:          "both-omitted",
			givenExisting: nil,
			givenIncoming: nil,
			expectedExt:   nil,
		},
		{
			name:          "both-nil",
			givenExisting: nil,
			givenIncoming: json.RawMessage(`null`),
			expectedExt:   nil,
		},
		{
			name:          "both-empty",
			givenExisting: nil,
			givenIncoming: json.RawMessage(`{}`),
			expectedExt:   json.RawMessage(`{}`),
		},
		{
			name:          "ext-omitted",
			givenExisting: json.RawMessage(`{"b":2}`),
			givenIncoming: nil,
			expectedExt:   json.RawMessage(`{"b":2}`),
		},
		{
			name:          "ext-nil",
			givenExisting: json.RawMessage(`{"b":2}`),
			givenIncoming: json.RawMessage(`null`),
			expectedExt:   json.RawMessage(`{"b":2}`),
		},
		{
			name:          "ext-empty",
			givenExisting: json.RawMessage(`{"b":2}`),
			givenIncoming: json.RawMessage(`{}`),
			expectedExt:   json.RawMessage(`{"b":2}`),
		},
		{
			name:          "ext-malformed",
			givenExisting: json.RawMessage(`{"b":2}`),
			givenIncoming: json.RawMessage(`malformed`),
			expectedErr:   "openrtb2.BidRequest.Ext",
		},
		{
			name:          "existing-nil",
			givenExisting: nil,
			givenIncoming: json.RawMessage(`{"a":1}`),
			expectedExt:   json.RawMessage(`{"a":1}`),
		},
		{
			name:          "existing-empty",
			givenExisting: json.RawMessage(`{}`),
			givenIncoming: json.RawMessage(`{"a":1}`),
			expectedExt:   json.RawMessage(`{"a":1}`),
		},
		{
			name:          "existing-omitted",
			givenExisting: nil,
			givenIncoming: json.RawMessage(`{"b":2}`),
			expectedExt:   json.RawMessage(`{"b":2}`),
		},
		{
			name:          "existing-malformed",
			givenExisting: json.RawMessage(`malformed`),
			givenIncoming: json.RawMessage(`{"a":1}`),
			expectedErr:   "cannot unmarshal openrtb2.BidRequest.Ext: invalid json on existing object",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// copy original values to check at the end for no modification
			originalExisting := sliceutil.Clone(test.givenExisting)
			originalIncoming := sliceutil.Clone(test.givenIncoming)

			// build request
			request := &openrtb2.BidRequest{Ext: test.givenExisting}

			// build data
			data := test.givenIncoming
			if len(data) > 0 {
				data = []byte(`{"ext":` + string(data) + `}`) // wrap in ext
			} else {
				data = []byte(`{}`) // omit ext
			}

			err := MergeClone(request, data)

			// assert error
			if test.expectedErr == "" {
				assert.NoError(t, err, "err")
			} else {
				assert.ErrorContains(t, err, test.expectedErr, "err")
			}

			// assert ext
			if test.expectedErr != "" {
				// expect no change in case of error
				assert.Equal(t, string(test.givenExisting), string(request.Ext), "json")
			} else {
				// compare as strings instead of json in case of nil or malformed ext
				assert.Equal(t, string(test.expectedExt), string(request.Ext), "json")
			}

			// assert no modifications
			// - can't use `assert.Same`` comparison checks since that's expected if
			//   either existing or incoming are nil / omitted / empty.
			assert.Equal(t, originalExisting, []byte(test.givenExisting), "existing")
			assert.Equal(t, originalIncoming, []byte(test.givenIncoming), "incoming")
		})
	}
}

func TestMergeCloneCombinations(t *testing.T) {
	t.Run("slice-of-ptr", func(t *testing.T) {
		var (
			imp      = &openrtb2.Imp{ID: "1"}
			impSlice = []*openrtb2.Imp{imp}
			test     = &struct {
				Imps []*openrtb2.Imp `json:"imps"`
			}{Imps: impSlice}
		)

		err := MergeClone(test, []byte(`{"imps":[{"id":"2"}]}`))
		require.NoError(t, err)

		assert.NotSame(t, impSlice, test.Imps, "slice-ref")
		require.Len(t, test.Imps, 1, "slice-len")

		assert.NotSame(t, imp, test.Imps[0], "item-ref")
		assert.Equal(t, "1", imp.ID, "original-val")
		assert.Equal(t, "2", test.Imps[0].ID, "new-val")
	})

	// special case of "slice-of-ptr"
	t.Run("jsonrawmessage-ptr", func(t *testing.T) {
		var (
			testJson = json.RawMessage(`{"a":1}`)
			test     = &struct {
				Foo *json.RawMessage `json:"foo"`
			}{Foo: &testJson}
		)

		err := MergeClone(test, []byte(`{"foo":{"b":2}}`))
		require.NoError(t, err)

		assert.NotSame(t, &testJson, test.Foo, "ref")
		assert.Equal(t, json.RawMessage(`{"a":1}`), testJson)
		assert.Equal(t, json.RawMessage(`{"a":1,"b":2}`), *test.Foo)
	})

	t.Run("struct-ptr", func(t *testing.T) {
		var (
			imp  = &openrtb2.Imp{ID: "1"}
			test = &struct {
				Imp *openrtb2.Imp `json:"imp"`
			}{Imp: imp}
		)

		err := MergeClone(test, []byte(`{"imp":{"id":"2"}}`))
		require.NoError(t, err)

		assert.NotSame(t, imp, test.Imp, "ref")
		assert.Equal(t, "1", imp.ID, "original-val")
		assert.Equal(t, "2", test.Imp.ID, "new-val")
	})

	t.Run("map-of-ptrs", func(t *testing.T) {
		var (
			imp    = &openrtb2.Imp{ID: "1"}
			impMap = map[string]*openrtb2.Imp{"a": imp}
			test   = &struct {
				Imps map[string]*openrtb2.Imp `json:"imps"`
			}{Imps: impMap}
		)

		err := MergeClone(test, []byte(`{"imps":{"a":{"id":"2"}}}`))
		require.NoError(t, err)

		assert.NotSame(t, impMap, test.Imps, "map-ref")
		assert.NotSame(t, imp, test.Imps["a"], "imp-ref")

		assert.Same(t, impMap["a"], imp, "imp-map-ref")

		assert.Equal(t, "1", imp.ID, "original-val")
		assert.Equal(t, "2", test.Imps["a"].ID, "new-val")
	})
}
