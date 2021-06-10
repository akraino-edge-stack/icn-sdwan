/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package manager

const (
	NameSpaceName        = "sdewan-system"
	RootIssuerName       = "sdewan-controller"
	RootCAIssuerName     = "sdewan-controller-ca"
	RootCertName         = "sdewan-controller"
	SCCCertName          = "sdewan-controller-base"
	StoreName            = "centralcontroller"
	OverlayCollection    = "overlays"
	OverlayResource      = "overlay-name"
	ProposalCollection   = "proposals"
	ProposalResource     = "proposal-name"
	HubCollection        = "hubs"
	HubResource          = "hub-name"
	ConnectionCollection = "connections"
	ConnectionResource   = "connection-name"
	CNFCollection        = "cnfs"
	CNFResource          = "cnf-name"
	DeviceCollection     = "devices"
	DeviceResource       = "device-name"
	IPRangeCollection    = "ipranges"
	IPRangeResource      = "iprange-name"
	CertCollection       = "certificates"
	CertResource         = "certificate-name"
)
