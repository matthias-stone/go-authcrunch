// Copyright 2024 Paul Greenberg greenpau@outlook.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authn

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/matthias-stone/go-authcrunch/pkg/identity"
	"github.com/matthias-stone/go-authcrunch/pkg/ids"
	"github.com/matthias-stone/go-authcrunch/pkg/requests"
	"github.com/matthias-stone/go-authcrunch/pkg/user"
)

var gpgKeyRegexPattern1 = regexp.MustCompile(`^[-]{3,5}\s*BEGIN\sPGP\sPUBLIC\sKEY\sBLOCK[-]{3,5}\s*`)
var gpgKeyRegexPattern2 = regexp.MustCompile(`[-]{3,5}\s*END\sPGP\sPUBLIC\sKEY\sBLOCK[-]{3,5}\s*$`)

// TestUserGPGKey tests GPG key.
func (p *Portal) TestUserGPGKey(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	rr *requests.Request,
	parsedUser *user.User,
	resp map[string]interface{},
	usr *user.User,
	backend ids.IdentityStore,
	bodyData map[string]interface{}) error {

	rr.Key.Usage = "gpg"

	// Extract data.
	if v, exists := bodyData["content"]; exists {
		switch keyContent := v.(type) {
		case string:
			rr.Key.Payload = strings.TrimSpace(keyContent)
		default:
			resp["message"] = "Profile API did find key content in the request payload, but it is malformed"
			return handleAPIProfileResponse(w, rr, http.StatusBadRequest, resp)
		}
	} else {
		resp["message"] = "Profile API did not find key content in the request payload"
		return handleAPIProfileResponse(w, rr, http.StatusBadRequest, resp)
	}

	// Validate data.
	if !gpgKeyRegexPattern1.MatchString(rr.Key.Payload) || !gpgKeyRegexPattern2.MatchString(rr.Key.Payload) {
		resp["message"] = "Profile API found non-compliant GPG public key value"
		return handleAPIProfileResponse(w, rr, http.StatusBadRequest, resp)
	}

	pk, err := identity.NewPublicKey(rr)
	if err != nil {
		var errMsg string = fmt.Sprintf("the Profile API failed to convert input into GPG key: %v", err)
		resp["message"] = errMsg
		return handleAPIProfileResponse(w, rr, http.StatusBadRequest, resp)
	}

	respData := make(map[string]interface{})
	if pk != nil {
		respData["success"] = true
		if pk.FingerprintMD5 != "" {
			respData["fingerprint_md5"] = pk.FingerprintMD5
		}
		if pk.Fingerprint != "" {
			respData["fingerprint"] = pk.Fingerprint
		}
		if pk.Comment != "" {
			respData["comment"] = pk.Comment
		}
	} else {
		respData["success"] = false
	}
	resp["entry"] = respData
	return handleAPIProfileResponse(w, rr, http.StatusOK, resp)
}
