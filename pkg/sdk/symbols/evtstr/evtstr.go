// SPDX-License-Identifier: Apache-2.0
/*
Copyright (C) 2023 The Diginfra Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This package exports the following C function:
// - char* plugin_event_to_string(ss_plugin_t *s, const ss_plugin_event *evt)
//
// The exported plugin_event_to_string requires that s to be a handle
// of cgo.Handle from this SDK. The value of the s handle must implement
// the sdk.Stringer and sdk.StringerBuffer interfaces.
//
// This function is part of the plugin_api interface as defined in plugin_api.h.
// In almost all cases, your plugin should import this module,
// unless your plugin exports those symbols by other means.
package evtstr

/*
#include "../../plugin_types.h"
*/
import "C"
import (
	"unsafe"

	"github.com/diginfra/plugin-sdk-go/pkg/cgo"
	"github.com/diginfra/plugin-sdk-go/pkg/sdk"
)

//export plugin_event_to_string
func plugin_event_to_string(pState C.uintptr_t, evt *C.ss_plugin_event_input) *C.char {
	buf := cgo.Handle(pState).Value().(sdk.StringerBuffer).StringerBuffer()
	stringer, ok := cgo.Handle(pState).Value().(sdk.Stringer)
	if ok {
		if str, err := stringer.String(sdk.NewEventReader(unsafe.Pointer(evt))); err == nil {
			buf.Write(str)
		} else {
			buf.Write(err.Error())
		}
	} else {
		buf.Write("")
	}
	return (*C.char)(buf.CharPtr())
}
