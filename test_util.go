// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import "io"

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }
