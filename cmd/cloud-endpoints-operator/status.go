// Copyright 2018 Google LLC

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     https://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

func makeStatus(parent *CloudEndpoint, children *CloudEndpointControllerRequestChildren) *CloudEndpointControllerStatus {
	status := CloudEndpointControllerStatus{
		StateCurrent: "IDLE",
		JWTAudiences: make([]string, 0),
	}

	changed := false
	sig := calcParentSig(parent, "")

	if parent.Status.LastAppliedSig != "" {
		if parent.Status.StateCurrent == StateIdle && sig != parent.Status.LastAppliedSig {
			changed = true
			status.LastAppliedSig = ""
		} else {
			status.LastAppliedSig = parent.Status.LastAppliedSig
		}
	}

	if parent.Status.StateCurrent != "" && changed == false {
		status.StateCurrent = parent.Status.StateCurrent
	}

	if parent.Status.Endpoint != "" && changed == false {
		status.Endpoint = parent.Status.Endpoint
	}

	if parent.Status.Config != "" && changed == false {
		status.Config = parent.Status.Config
	}

	if parent.Status.ConfigSubmit != "" && changed == false {
		status.ConfigSubmit = parent.Status.ConfigSubmit
	}

	if parent.Status.ServiceRollout != "" && changed == false {
		status.ServiceRollout = parent.Status.ServiceRollout
	}

	if parent.Status.IngressIP != "" && changed == false {
		status.IngressIP = parent.Status.IngressIP
	}

	if parent.Status.JWTAudiences != nil && changed == false {
		status.JWTAudiences = parent.Status.JWTAudiences
	}

	if parent.Status.ConfigMapHash != "" && changed == false {
		status.ConfigMapHash = parent.Status.ConfigMapHash
	}

	return &status
}
