// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"strconv"
	"testing"
)

func TestPatchDelta(t *testing.T) {
	base := []string{
		"tree 16e3f83e622db3b3a6de764f7d3dcd2888d1146c\nparent aa9384566161a242ad0ca2563e613736edf38fe9\nauthor Felipe Contreras <felipe.contreras@gmail.com> 1367010755 -0500\ncommitter Junio C Hamano <gitster@pobox.com> 1367014827 -0700\n\nremote-hg: use hashlib instead of hg sha1 util\n\nTo be in sync with remote-bzr.\n\nSigned-off-by: Felipe Contreras <felipe.contreras@gmail.com>\nSigned-off-by: Junio C Hamano <gitster@pobox.com>\n",
		"tree 6e48d1e480899bd1ad8f5512979c27fe4392d7ae\nparent 11ee57bc4c44763b7ea92c5f583e27a5fbbff76b\nauthor Brandon Casey <casey@nrlssc.navy.mil> 1216761811 -0500\ncommitter Junio C Hamano <gitster@pobox.com> 1217657702 -0700\n\nt/t4202-log.sh: add newline at end of file\n\nSome shells hang when parsing the script if the last statement is not\nfollowed by a newline. So add one.\n\nSigned-off-by: Brandon Casey <casey@nrlssc.navy.mil>\nSigned-off-by: Junio C Hamano <gitster@pobox.com>\n"}

	delta := []string{
		"\xa2\x03\x91\x03]tree 0bb83c46690d6b136b1e02c90b91eb7488b6a505\nparent d6bb9136c93baddf0ee5052638591bd881b19f77\x91]?\x014\x91\x9dM5bzr: add support to push URLs\n\nJust like in remote-hg\x930\x01r",
		"\xd8\x03\xd8\x03]tree e29606187ff772f3cd2cac848d1c139591865898\nparent 09b78bc1fc4e525bc68fa0ce76521457717fe675\x91]o\a6838201\xb1\xd3\x05\x01"}

	expected := []string{`tree 0bb83c46690d6b136b1e02c90b91eb7488b6a505
parent d6bb9136c93baddf0ee5052638591bd881b19f77
author Felipe Contreras <felipe.contreras@gmail.com> 1367010754 -0500
committer Junio C Hamano <gitster@pobox.com> 1367014827 -0700

remote-bzr: add support to push URLs

Just like in remote-hg.

Signed-off-by: Felipe Contreras <felipe.contreras@gmail.com>
Signed-off-by: Junio C Hamano <gitster@pobox.com>
`, `tree e29606187ff772f3cd2cac848d1c139591865898
parent 09b78bc1fc4e525bc68fa0ce76521457717fe675
author Brandon Casey <casey@nrlssc.navy.mil> 1216761811 -0500
committer Junio C Hamano <gitster@pobox.com> 1216838201 -0700

t/t4202-log.sh: add newline at end of file

Some shells hang when parsing the script if the last statement is not
followed by a newline. So add one.

Signed-off-by: Brandon Casey <casey@nrlssc.navy.mil>
Signed-off-by: Junio C Hamano <gitster@pobox.com>
`}

	for i := range base {
		buf, err := patchDelta([]byte(base[i]), []byte(delta[i]))

		if err != nil {
			t.Error(err)
		}

		if string(buf) != expected[i] {
			t.Errorf("does not match, got %s", strconv.Quote(string(buf)))
		}
	}
}
