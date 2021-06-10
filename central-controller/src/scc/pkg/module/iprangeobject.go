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

package module

import (
	pkgerrors "github.com/pkg/errors"
	"math"
	"strconv"
	"strings"
)

// App contains metadata for Apps
type IPRangeObject struct {
	Metadata      ObjectMetaData      `json:"metadata"`
	Specification IPRangeObjectSpec   `json:"spec"`
	Status        IPRangeObjectStatus `json:"-"`
}

//IPRangeObjectSpec contains the parameters
type IPRangeObjectSpec struct {
	Subnet string `json:"subnet" validate:"required,ipv4"`
	MinIp  int    `json:"minIp" validate:"gte=1,lte=255"`
	MaxIp  int    `json:"maxIp" validate:"gte=1,lte=255"`
}

type IPRangeObjectStatus struct {
	Masks [32]byte
	Data  map[string]string
}

func (c *IPRangeObject) GetMetadata() ObjectMetaData {
	return c.Metadata
}

func (c *IPRangeObject) GetType() string {
	return "IPRange"
}

func (c *IPRangeObject) base() string {
	index := strings.LastIndex(c.Specification.Subnet, ".")
	if index == -1 {
		return c.Specification.Subnet
	} else {
		return c.Specification.Subnet[0 : index+1]
	}
}

func (c *IPRangeObject) IsConflict(o *IPRangeObject) bool {
	if strings.Compare(c.base(), o.base()) != 0 {
		return false
	}

	if c.Specification.MinIp > o.Specification.MaxIp || c.Specification.MaxIp < o.Specification.MinIp {
		return false
	}

	return true
}

func (c *IPRangeObject) InUsed() bool {
	return (len(c.Status.Data) != 0)
}

func (c *IPRangeObject) Allocate(name string) (string, error) {
	i := c.Specification.MinIp
	index := (c.Specification.MinIp - 1) / 8
	b := byte(math.Exp2(float64(7 - (c.Specification.MinIp-1)%8)))
	for i <= c.Specification.MaxIp {
		if c.Status.Masks[index]&b == 0 {
			c.Status.Masks[index] |= b
			c.Status.Data[strconv.Itoa(i)] = name
			return c.base() + strconv.Itoa(i), nil
		}
		if i%8 == 0 {
			b = 0x80
			index += 1
			for c.Status.Masks[index] == 0xff {
				//    log.Println("by pass", index)
				i += 8
				index += 1
			}
		} else {
			b = b / 2
		}
		i = i + 1
	}

	return "", pkgerrors.New("No available IP")
}

func (c *IPRangeObject) Free(sip string) error {
	ip := 0
	i := strings.LastIndex(sip, ".")
	if i == -1 {
		return pkgerrors.New("invalid ip")
	} else {
		base_ip := sip[0 : i+1]
		if c.base() != base_ip {
			return pkgerrors.New("ip is not in range")
		}

		ip, _ = strconv.Atoi(sip[i+1 : len(sip)])
	}

	if ip < c.Specification.MinIp || ip > c.Specification.MaxIp {
		return pkgerrors.New("ip is not in range")
	}

	index := (ip - 1) / 8
	b := byte(math.Exp2(float64(7 - (ip-1)%8)))
	if c.Status.Masks[index]&b == 0 {
		return pkgerrors.New("ip is not allocated")
	}

	delete(c.Status.Data, strconv.Itoa(ip))
	c.Status.Masks[index] &= (^b)
	return nil
}

func (c *IPRangeObject) FreeAll() error {
	for sip, _ := range c.Status.Data {
		ip, _ := strconv.Atoi(sip)
		index := (ip - 1) / 8
		b := byte(math.Exp2(float64(7 - (ip-1)%8)))
		delete(c.Status.Data, sip)
		c.Status.Masks[index] &= (^b)
	}
	return nil
}
