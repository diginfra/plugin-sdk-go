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

package source

import (
	"context"
	"io"
	"math"
	"time"

	"github.com/diginfra/plugin-sdk-go/pkg/sdk"
)

var (
	defaultInstanceTimeout = 30 * time.Millisecond
)

type builtinInstance struct {
	BaseInstance
	shutdown      func()
	progress      func() (float64, string)
	ctx           context.Context
	timeout       time.Duration
	timeoutTicker *time.Ticker
	eof           bool
	eventSize     uint32
	batchSize     uint32
}

func (s *builtinInstance) Close() {
	// this cancels the context and calls the optional callback
	s.shutdown()

	// stop timeout ticker
	s.timeoutTicker.Stop()
}

func (s *builtinInstance) Progress(pState sdk.PluginState) (float64, string) {
	if s.progress != nil {
		return s.progress()
	}
	return 0, ""
}

// WithInstanceContext sets a custom context in the opened event source.
// If the context is cancelled, the event source is closed and sdk.ErrEOF
// is returned by the current invocation of NextBatch() and by any subsequent
// invocation.
func WithInstanceContext(ctx context.Context) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.ctx = ctx
	}
}

// WithInstanceTimeout sets a custom timeout in the opened event source.
// When the timeout is reached, the current invocation of NextBatch() returns
// sdk.ErrTimeout.
func WithInstanceTimeout(timeout time.Duration) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.timeout = timeout
	}
}

// WithInstanceClose sets a custom closing callback in the opened event source.
// The passed-in function is invoked when the event source gets closed.
func WithInstanceClose(close func()) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.shutdown = close
	}
}

// WithInstanceBatchSize sets a custom size for the pre-allocated event batch
// used by NextBatch()
func WithInstanceBatchSize(size uint32) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.batchSize = size
	}
}

// WithInstanceEventSize sets a custom maximum size for each event returned
// by NextBatch()
func WithInstanceEventSize(size uint32) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.eventSize = size
	}
}

// WithInstanceProgress sets a custom callback for the framework to request
// a the progress state of the opened event stream
func WithInstanceProgress(progress func() (float64, string)) func(*builtinInstance) {
	return func(s *builtinInstance) {
		s.progress = progress
	}
}

// PullFunc produces a new event and returns a non-nil error in case of failure.
//
// The event data is produced through the sdk.EventWriter interface.
// The context argument can be used to check for termination signals, which
// happen when the framework closes the event source or when the optional
// context passed-in by the user gets cancelled.
type PullFunc func(context.Context, sdk.EventWriter) error

type pullInstance struct {
	builtinInstance
	pull PullFunc
}

// NewPullInstance opens a new event source and starts a capture session,
// filling the event batches with a pull model.
//
// The PullFunc required argument is a function that creates a new event and
// returns a non-nil error in case of success. The returned source.Instance
// provides a pre-built implementation of NextBatch() that correctly handles
// termination and timeouts. This should be used by developers to open an event
// source without defining a new type and by using a functional design.
//
// The pull function is invoked sequentially and is blocking for the event
// source, meaning that it must not be a suspensive function. This implies
// avoiding suspending an execution through a select or through synchronization
// primitives.
//
// Users can pass option parameters to influence the behavior of the opened
// event source, such as passing a context or setting a custom timeout duration.
//
// The context passed-in to the pull function is cancelled automatically
// when the framework invokes Close() on the event source, or when the
// user-configured context is cancelled.
func NewPullInstance(pull PullFunc, options ...func(*builtinInstance)) (Instance, error) {
	res := &pullInstance{
		pull: pull,
		builtinInstance: builtinInstance{
			ctx:       context.Background(),
			timeout:   defaultInstanceTimeout,
			shutdown:  func() {},
			eof:       false,
			batchSize: sdk.DefaultBatchSize,
			eventSize: sdk.DefaultEvtSize,
		},
	}

	// apply options
	for _, opt := range options {
		opt(&res.builtinInstance)
	}

	// create custom-sized event batch
	batch, err := sdk.NewEventWriters(int64(res.batchSize), int64(res.eventSize))
	if err != nil {
		return nil, err
	}
	res.SetEvents(batch)

	// init timer
	res.timeoutTicker = time.NewTicker(res.timeout)

	// setup internally-cancellable context
	prevCancel := res.shutdown
	cancelableCtx, cancelCtx := context.WithCancel(res.ctx)
	res.ctx = cancelableCtx
	res.shutdown = func() {
		cancelCtx()
		prevCancel()
	}

	// return opened instance
	return res, nil
}

func (s *pullInstance) NextBatch(pState sdk.PluginState, evts sdk.EventWriters) (n int, err error) {
	// once EOF has been hit, we should return it at each new call of NextBatch
	if s.eof {
		return 0, sdk.ErrEOF
	}

	// timeout needs to be resetted for this batch
	s.timeoutTicker.Reset(s.timeout)

	// attempt filling the event batch
	n = 0
	for n < evts.Len() {
		// check if we should return before pulling another event
		select {
		// timeout hits, so we flush a partial batch
		case <-s.timeoutTicker.C:
			return n, sdk.ErrTimeout
		// context has been canceled, so we exit
		case <-s.ctx.Done():
			s.eof = true
			return n, sdk.ErrEOF
		default:
		}

		// pull a new event
		if err = s.pull(s.ctx, evts.Get(n)); err != nil {
			// in case of non-timeout error, we consider the event source ended
			if err != sdk.ErrTimeout {
				s.eof = true
			}
			return n, err
		}
		n++
	}

	// return a full batch
	return n, nil
}

// PushEvent represents an event produced from an event source with the push model.
//
// If the event source produced the event successfully, then Data must be non-nil
// and Err must be ni. If the event source encountered a failure, Data must be
// nil and Err should contain an error describing the failure.
//
// Timestamp can be optionally set to indicate a specific timestamp for the
// produced event.
type PushEvent struct {
	Err       error
	Data      []byte
	Timestamp time.Time
}

type pushInstance struct {
	builtinInstance
	evtC <-chan PushEvent
}

// NewPushInstance opens a new event source and starts a capture session,
// filling the event batches with a push model.
//
// In this model, events are produced through a channel in the form of
// source.PushEvent messages. This is suitable for cases in which event
// production is suspensive, meaning that the time elapsed waiting for a
// new event to be produced is not deterministic or has no guaranteed limit.
//
// Users can pass option parameters to influence the behavior of the opened
// event source, such as passing a context or setting a custom timeout duration.
//
// The opened event source can be manually closed by cancelling the optional
// passed-in context, by closing the event cannel, or by sending
// source.PushEvent containing a non-nil Err.
func NewPushInstance(evtC <-chan PushEvent, options ...func(*builtinInstance)) (Instance, error) {
	res := &pushInstance{
		evtC: evtC,
		builtinInstance: builtinInstance{
			ctx:       context.Background(),
			timeout:   defaultInstanceTimeout,
			shutdown:  func() {},
			eof:       false,
			batchSize: sdk.DefaultBatchSize,
			eventSize: sdk.DefaultEvtSize,
		},
	}

	// apply options
	for _, opt := range options {
		opt(&res.builtinInstance)
	}

	// create custom-sized event batch
	batch, err := sdk.NewEventWriters(int64(res.batchSize), int64(res.eventSize))
	if err != nil {
		return nil, err
	}
	res.SetEvents(batch)

	// init timer
	res.timeoutTicker = time.NewTicker(res.timeout)

	// setup internally-cancellable context
	prevCancel := res.shutdown
	cancelableCtx, cancelCtx := context.WithCancel(res.ctx)
	res.ctx = cancelableCtx
	res.shutdown = func() {
		cancelCtx()
		prevCancel()
	}

	return res, nil
}

func (s *pushInstance) NextBatch(pState sdk.PluginState, evts sdk.EventWriters) (int, error) {
	// once EOF has been hit, we should return it at each new call of NextBatch
	if s.eof {
		return 0, sdk.ErrEOF
	}

	// timeout needs to be resetted for this batch
	s.timeoutTicker.Reset(s.timeout)

	// attempt filling the event batch
	n := 0
	for n < evts.Len() {
		select {
		// an event is received, so we add it in the batch
		case evt, ok := <-s.evtC:
			// event channel is closed, we reached EOF
			if !ok {
				evt.Err = sdk.ErrEOF
			}
			// if there are no errors so far, try writing the event
			if evt.Err == nil {
				if l, wErr := evts.Get(n).Writer().Write(evt.Data); wErr != nil {
					evt.Err = wErr
				} else if l < len(evt.Data) {
					evt.Err = io.ErrShortWrite
				}
			}
			// an error occurred, so we need to exit
			if evt.Err != nil {
				// in case of non-timeout error, we consider the event source ended
				if evt.Err != sdk.ErrTimeout {
					s.eof = true
				}
				return n, evt.Err
			}
			// event added to the batch successfully
			if evt.Timestamp.IsZero() {
				evts.Get(n).SetTimestamp(math.MaxUint64)
			} else {
				evts.Get(n).SetTimestamp(uint64(evt.Timestamp.UnixNano()))
			}
			n++
		// timeout hits, so we flush a partial batch
		case <-s.timeoutTicker.C:
			return n, sdk.ErrTimeout
		// context has been canceled, so we exit
		case <-s.ctx.Done():
			s.eof = true
			return n, sdk.ErrEOF
		}
	}
	return n, nil
}
