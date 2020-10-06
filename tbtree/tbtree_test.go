/*
Copyright 2019-2020 vChain, Inc.

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
package tbtree

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func monotonicInsertions(t *testing.T, tbtree *TBtree, itCount int, kCount int, ascMode bool) {
	for i := 0; i < itCount; i++ {
		for j := 0; j < kCount; j++ {
			k := make([]byte, 4)
			if ascMode {
				binary.BigEndian.PutUint32(k, uint32(j))
			} else {
				binary.BigEndian.PutUint32(k, uint32(kCount-j))
			}

			v := make([]byte, 8)
			binary.BigEndian.PutUint64(v, uint64(i<<4+j))

			ts := uint64(i*kCount+j) + 1

			snapshot, err := tbtree.Snapshot()
			require.NoError(t, err)
			snapshotTs := snapshot.Ts()
			require.Equal(t, ts-1, snapshotTs)

			v1, ts1, err := snapshot.Get(k)

			if i == 0 {
				require.Equal(t, ErrKeyNotFound, err)
			} else {
				require.NoError(t, err)

				expectedV := make([]byte, 8)
				binary.BigEndian.PutUint64(expectedV, uint64((i-1)<<4+j))
				require.Equal(t, expectedV, v1)

				expectedTs := uint64((i-1)*kCount+j) + 1
				require.Equal(t, expectedTs, ts1)
			}

			if i%2 == 1 {
				err = snapshot.Close()
				require.NoError(t, err)
			}

			err = tbtree.Insert(k, v, uint64(ts))
			require.NoError(t, err)

			if i%2 == 0 {
				err = snapshot.Close()
				require.NoError(t, err)
			}

			snapshot, err = tbtree.Snapshot()
			require.NoError(t, err)
			snapshotTs = snapshot.Ts()
			require.Equal(t, ts, snapshotTs)

			v1, ts1, err = snapshot.Get(k)
			require.NoError(t, err)
			require.Equal(t, v, v1)
			require.Equal(t, ts, ts1)

			err = snapshot.Close()
			require.NoError(t, err)
		}
	}
}

func checkAfterMonotonicInsertions(t *testing.T, tbtree *TBtree, itCount int, kCount int, ascMode bool) {
	snapshot, err := tbtree.Snapshot()
	require.NoError(t, err)

	i := itCount

	for j := 0; j < kCount; j++ {
		k := make([]byte, 4)
		if ascMode {
			binary.BigEndian.PutUint32(k, uint32(j))
		} else {
			binary.BigEndian.PutUint32(k, uint32(kCount-j))
		}

		v := make([]byte, 8)
		binary.BigEndian.PutUint64(v, uint64(i<<4+j))

		v1, ts1, err := snapshot.Get(k)

		require.NoError(t, err)

		expectedV := make([]byte, 8)
		binary.BigEndian.PutUint64(expectedV, uint64((i-1)<<4+j))
		require.Equal(t, expectedV, v1)

		expectedTs := uint64((i-1)*kCount+j) + 1
		require.Equal(t, expectedTs, ts1)
	}

	err = snapshot.Close()
	require.NoError(t, err)
}

func randomInsertions(t *testing.T, tbtree *TBtree, kCount int, override bool) {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)

	for i := 0; i < kCount; i++ {
		k := make([]byte, 4)
		binary.BigEndian.PutUint32(k, rnd.Uint32())

		if !override {
			snapshot, err := tbtree.Snapshot()
			require.NoError(t, err)

			for {
				_, _, err = snapshot.Get(k)
				if err == ErrKeyNotFound {
					break
				}
				binary.BigEndian.PutUint32(k, rnd.Uint32())
			}

			err = snapshot.Close()
			require.NoError(t, err)
		}

		v := make([]byte, 8)
		binary.BigEndian.PutUint64(v, uint64(i))

		ts := uint64(i + 1)

		err := tbtree.Insert(k, v, uint64(ts))
		require.NoError(t, err)

		snapshot, err := tbtree.Snapshot()
		require.NoError(t, err)
		snapshotTs := snapshot.Ts()
		require.Equal(t, ts, snapshotTs)

		v1, ts1, err := snapshot.Get(k)
		require.NoError(t, err)
		require.Equal(t, v, v1)
		require.Equal(t, ts, ts1)

		err = snapshot.Close()
		require.NoError(t, err)
	}
}

func TestTBTreeInsertionInAscendingOrder(t *testing.T) {
	tbtree, err := Open("tbtree.idb", DefaultOptions().SetMaxNodeSize(256).SetInsertionCountThreshold(100))
	require.NoError(t, err)
	defer os.Remove("tbtree.idb")

	n, err := tbtree.Flush()
	require.NoError(t, err)
	require.Equal(t, int64(0), n)

	itCount := 10
	keyCount := 1000
	monotonicInsertions(t, tbtree, itCount, keyCount, true)

	_, err = tbtree.Flush()
	require.NoError(t, err)

	err = tbtree.Close()
	require.NoError(t, err)

	n, err = tbtree.Flush()
	require.Equal(t, err, ErrAlreadyClosed)

	err = tbtree.Close()
	require.Equal(t, err, ErrAlreadyClosed)

	tbtree, err = Open("tbtree.idb", DefaultOptions().SetMaxNodeSize(256))
	require.NoError(t, err)

	require.Equal(t, tbtree.root.ts(), uint64(itCount*keyCount))

	checkAfterMonotonicInsertions(t, tbtree, itCount, keyCount, true)
}

func TestTBTreeInsertionInDescendingOrder(t *testing.T) {
	tbtree, err := Open("tbtree.idb", DefaultOptions().SetMaxNodeSize(256))
	require.NoError(t, err)
	defer os.Remove("tbtree.idb")

	itCount := 10
	keyCount := 1000

	monotonicInsertions(t, tbtree, itCount, keyCount, false)

	err = tbtree.Close()
	require.NoError(t, err)

	tbtree, err = Open("tbtree.idb", DefaultOptions().SetMaxNodeSize(256))
	require.NoError(t, err)

	require.Equal(t, tbtree.root.ts(), uint64(itCount*keyCount))

	checkAfterMonotonicInsertions(t, tbtree, itCount, keyCount, false)

	snapshot, err := tbtree.Snapshot()
	require.NotNil(t, snapshot)
	require.NoError(t, err)

	rspec := &ReaderSpec{
		initialKey: nil,
		isPrefix:   false,
		ascOrder:   true,
	}
	reader, err := snapshot.Reader(rspec)
	require.NoError(t, err)

	i := 0
	prevk := reader.initialKey
	for {
		k, _, _, err := reader.Read()
		if err != nil {
			require.Equal(t, ErrNoMoreEntries, err)
			break
		}

		require.True(t, bytes.Compare(prevk, k) < 1)
		prevk = k
		i++
	}
	require.Equal(t, keyCount, i)

	err = reader.Close()
	require.NoError(t, err)

	err = snapshot.Close()
	require.NoError(t, err)

	err = tbtree.Insert(prevk, prevk, uint64(itCount*keyCount+1))
	require.NoError(t, err)

	snapshot, err = tbtree.Snapshot()
	require.NoError(t, err)

	v, ts, err := snapshot.Get(prevk)
	require.NoError(t, err)
	require.Equal(t, uint64(itCount*keyCount+1), ts)
	require.Equal(t, prevk, v)

	snapshot.Close()
}

func TestTBTreeInsertionInRandomOrder(t *testing.T) {
	tbtree, err := Open("tbtree.idb", DefaultOptions().SetMaxNodeSize(DefaultMaxNodeSize).SetCacheSize(100_000))
	require.NoError(t, err)
	defer os.Remove("tbtree.idb")

	randomInsertions(t, tbtree, 100_000, true)

	err = tbtree.Close()
	require.NoError(t, err)
}

func BenchmarkRandomInsertion(b *testing.B) {
	seed := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(seed)

	for i := 0; i < b.N; i++ {
		opts := DefaultOptions().
			SetMaxNodeSize(DefaultMaxNodeSize).
			SetCacheSize(10_000).
			SetInsertionCountThreshold(100_000)

		tbtree, _ := Open("tbtree.idb", opts)

		kCount := 1_000_000

		for i := 0; i < kCount; i++ {
			k := make([]byte, 4)
			binary.BigEndian.PutUint32(k, rnd.Uint32())

			v := make([]byte, 8)
			binary.BigEndian.PutUint64(v, uint64(i))

			ts := uint64(i + 1)

			tbtree.Insert(k, v, uint64(ts))
		}

		tbtree.Close()

		os.Remove("tbtree.idb")
	}
}
