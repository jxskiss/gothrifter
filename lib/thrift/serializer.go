/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thrift

import "github.com/thrift-iterator/go"

type Serializer struct {
	cfg thrifter.API
}

func NewSerializer() *Serializer {
	return &Serializer{cfg: thrifter.DefaultConfig}
}

func (t *Serializer) WriteString(msg interface{}) (s string, err error) {
	b, err := t.cfg.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *Serializer) Write(msg interface{}) (b []byte, err error) {
	if b, err = t.cfg.Marshal(msg); err != nil {
		return nil, err
	}
	return b, nil
}
