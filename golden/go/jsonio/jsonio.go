// Package jsonio contains the routines necessary to consume and emit JSON to be processed
// by the Gold ingester.

package jsonio

// The JSON output from DM looks like this:
//
//  {
//     "build_number" : "20",
//     "gitHash" : "abcd",
//     "key" : {
//        "arch" : "x86",
//        "configuration" : "Debug",
//        "gpu" : "nvidia",
//        "model" : "z620",
//        "os" : "Ubuntu13.10"
//     },
//     "results" : [
//        {
//           "key" : {
//              "config" : "565",
//              "name" : "ninepatch-stretch",
//              "source_type" : "gm"
//           },
//           "md5" : "f78cfafcbabaf815f3dfcf61fb59acc7",
//           "options" : {
//              "ext" : "png"
//           }
//        },
//        {
//           "key" : {
//              "config" : "8888",
//              "name" : "ninepatch-stretch",
//              "source_type" : "gm"
//           },
//           "md5" : "3e8a42f35a1e76f00caa191e6310d789",
//           "options" : {
//              "ext" : "png"
//           }
//

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
)

var (
	rawGoldResultsJsonMap map[string]string
	goldResultsJsonMap    map[string]string
	resultJsonMap         map[string]string
)

func init() {
	rawGoldResultsJsonMap = jsonNameMap(rawGoldResults{})
	goldResultsJsonMap = jsonNameMap(GoldResults{})
	resultJsonMap = jsonNameMap(Result{})
}

// ParseGoldResults parses JSON encoded Gold results. This needs to be called
// instead of parsing directly into an instance of GoldResult.
func ParseGoldResults(r io.Reader) (*GoldResults, []string, error) {
	// Decode JSON into a type that is more tolerant to failures. If there is
	// a failure we just return the failure.
	raw := &rawGoldResults{}
	if err := json.NewDecoder(r).Decode(raw); err != nil {
		return nil, nil, err
	}

	// parse and validate the raw input from the previous step, i.e.
	// parse string encoded integers.
	var errMessages []string = nil
	if errMsg := raw.parseValidate(); errMsg != nil {
		errMessages = append(errMessages, errMsg...)
	}

	// Extract the embedded Gold result and validate it.
	ret := raw.GoldResults
	if errMsg, err := ret.Validate(); err != nil {
		errMessages = append(errMessages, errMsg...)
	}

	if len(errMessages) > 0 {
		return nil, errMessages, messagesToError(errMessages)
	}
	return &ret, nil, nil
}

// GoldResults is the top level structure to capture the the results of a
// rendered test to be processed by Gold.
type GoldResults struct {
	GitHash string            `json:"gitHash"  validate:"required"`
	Key     map[string]string `json:"key"      validate:"required,min=1"`
	Results []*Result         `json:"results"  validate:"min=1"`

	// Required fields for tryjobs.
	Issue         int64 `json:"issue,string"`
	BuildBucketID int64 `json:"buildbucket_build_id,string"`
	Patchset      int64 `json:"patchset,string"`

	// Optional fields
	SwarmingTaskID string `json:"swarming_task_id"`
	SwarmingBotID  string `json:"swarming_bot_id"`
	Builder        string `json:"builder"`
}

type rawGoldResults struct {
	GoldResults

	// Override the fields that represent integers as strings.
	Issue         string `json:"issue"`
	BuildBucketID string `json:"buildbucket_build_id"`
	Patchset      string `json:"patchset"`
}

// parseValidate validates the rawGoldResult instance and parses integers
// that are encoded as strings.
func (r *rawGoldResults) parseValidate() []string {
	jn := rawGoldResultsJsonMap
	var ret []string
	issueValid := (r.Issue == "") || (r.Issue != "" && r.BuildBucketID != "" && r.Patchset != "")
	addErrMessage(&ret, issueValid, "fields '%s', '%s' must not be empty if field '%s' contains a value", jn["Patchset"], jn["BuildBucketID"], jn["Issue"])

	f := []string{"Issue", r.Issue, "Patchset", r.Patchset, "BuildBucketID", r.BuildBucketID}
	for i := 0; i < len(f); i += 2 {
		valid := f[i+1] == "" || govalidator.IsInt(f[i+1])
		addErrMessage(&ret, valid, "field '%s' must be empty or contain a valid integer", jn[f[i]])
	}

	if len(ret) == 0 {
		// If there was no error we can just parse the strings to int64.
		r.GoldResults.Issue, _ = strconv.ParseInt(r.Issue, 10, 64)
		r.GoldResults.BuildBucketID, _ = strconv.ParseInt(r.BuildBucketID, 10, 64)
		r.GoldResults.Patchset, _ = strconv.ParseInt(r.Patchset, 10, 64)
	}
	return ret
}

// Validate validates the instance of GoldResult. If there are no errors
// both return values will be nil. Otherwise the first return value contains
// error messages (one for each field) and the returned error contains a
// concatenation of these error messages.
func (g *GoldResults) Validate() ([]string, error) {
	jn := goldResultsJsonMap
	errMsg := []string{}

	// Validate the fields
	addErrMessage(&errMsg, govalidator.IsHexadecimal(g.GitHash), "field '%s' must be hexadecimal. Received '%s'", jn["GitHash"], g.GitHash)
	addErrMessage(&errMsg, len(g.Key) > 0 && hasNonEmptyKV(g.Key), "field '%s' must not be empty and must not have empty keys or values", jn["Key"])
	addErrMessage(&errMsg, len(g.Results) > 0, "field '%s' must not be empty.", jn["Results"])

	validIssue := g.Issue == 0 || (g.Issue > 0 && g.Patchset > 0 && g.BuildBucketID > 0)
	addErrMessage(&errMsg, validIssue, "fields '%s', '%s', '%s' must all be zero or all not be zero", jn["Issue"], jn["Patchset"], jn["BuildBucketID"])
	for _, r := range g.Results {
		r.validate(&errMsg, jn["Results"])
	}

	// If we have an error construct an error object from the error messages.
	if len(errMsg) > 0 {
		return errMsg, messagesToError(errMsg)
	}
	return nil, nil
}

// Result is used by DMResults hand holds the individual result of one test.
type Result struct {
	Key     map[string]string `json:"key"      validate:"required"`
	Options map[string]string `json:"options"  validate:"required"`
	Digest  string            `json:"md5"      validate:"required"`
}

// validate the Result instance.
func (r *Result) validate(errMsg *[]string, parentField string) {
	jn := resultJsonMap
	addErrMessage(errMsg, len(r.Key) > 0 && hasNonEmptyKV(r.Key), "field '%s' must be non-empty and must not have empty keys or values", parentField+"."+jn["Key"])
	addErrMessage(errMsg, hasNonEmptyKV(r.Options), "field '%s' must not have empty keys or values", parentField+"."+jn["Options"])
	addErrMessage(errMsg, govalidator.IsHexadecimal(r.Digest), "field '%s' must be hexadecimal", parentField+"."+jn["Digest"])
}

// addErrMessage adds an error message to errMsg if isValid is false. The
// error message is created using formatStr and args.
func addErrMessage(errMsg *[]string, isValid bool, formatStr string, args ...interface{}) {
	if isValid {
		return
	}
	*errMsg = append(*errMsg, fmt.Sprintf(formatStr, args...))
}

// messagesToError concatenates the error messages into a single error
func messagesToError(errMessages []string) error {
	return fmt.Errorf("%s", strings.Join(errMessages, "\n")+"\n")
}

// returns true if all keys and values in the map are not empty strings
func hasNonEmptyKV(kvMap map[string]string) bool {
	for k, v := range kvMap {
		if strings.TrimSpace(k) == "" && strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}

// jsonNameMap returns a map that maps a field name of the given struct to
// the name specified in the json tag.
func jsonNameMap(structType interface{}) map[string]string {
	sType := reflect.TypeOf(structType)
	nFields := sType.NumField()
	ret := make(map[string]string, nFields)
	for i := 0; i < nFields; i++ {
		f := sType.Field(i)
		jsonName := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}
		ret[f.Name] = jsonName
	}
	return ret
}