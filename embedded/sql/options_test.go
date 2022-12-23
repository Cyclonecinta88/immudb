/*
Copyright 2022 Codenotary Inc. All rights reserved.

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

package sql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	var opts *Options

	require.Error(t, opts.Validate())

	opts = &Options{}
	require.Error(t, opts.Validate())

	opts.WithDistinctLimit(0)
	require.Error(t, opts.Validate())

	opts.WithDistinctLimit(defaultDistinctLimit)
	require.Equal(t, defaultDistinctLimit, opts.distinctLimit)

	opts.WithPrefix([]byte("sqlPrefix"))
	require.Equal(t, []byte("sqlPrefix"), opts.prefix)

	opts.WithAutocommit(true)
	require.True(t, opts.autocommit)

	require.NoError(t, opts.Validate())
}
