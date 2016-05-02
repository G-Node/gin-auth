// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package conf

import "testing"

func TestMakeUrl(t *testing.T) {
	u := MakeUrl("/foo/%d", 200)
	if u != "http://localhost:8080/foo/200" {
		t.Error("Wrong url")
	}
	u = MakeUrl("/foo/%s", "some string")
	if u != "http://localhost:8080/foo/some+string" {
		t.Error("Wrong url")
	}
}
