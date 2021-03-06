// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package peers

import (
	"sort"
	"testing"
	"time"

	"github.com/m3db/m3db/client"
	"github.com/m3db/m3db/retention"
	"github.com/m3db/m3db/storage/block"
	"github.com/m3db/m3db/storage/namespace"
	"github.com/m3db/m3x/ident"
	xtime "github.com/m3db/m3x/time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type testSeriesMetadata struct {
	id   string
	tags map[string]string
}

func (s testSeriesMetadata) ID() ident.ID {
	return ident.StringID(s.id)
}

func (s testSeriesMetadata) Tags() ident.Tags {
	if s.tags == nil {
		return ident.Tags{}
	}

	// Return in sorted order for deterministic order
	var keys []string
	for key := range s.tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var tags ident.Tags
	for _, key := range keys {
		tags.Append(ident.StringTag(key, s.tags[key]))
	}

	return tags
}

func TestBootstrapIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	opts := testDefaultOpts.
		SetFetchBlocksMetadataEndpointVersion(client.FetchBlocksMetadataEndpointV2)

	blockSize := 2 * time.Hour
	indexBlockSize := 2 * blockSize

	ropts := retention.NewOptions().
		SetBlockSize(blockSize).
		SetRetentionPeriod(24 * blockSize)

	nsMetadata := testNamespaceMetadata(t, func(opts namespace.Options) namespace.Options {
		return opts.
			SetRetentionOptions(ropts).
			SetIndexOptions(opts.IndexOptions().
				SetEnabled(true).
				SetBlockSize(indexBlockSize))
	})

	at := time.Now()
	start := at.Add(-ropts.RetentionPeriod()).Truncate(blockSize)
	indexStart := start.Truncate(indexBlockSize)
	for !start.Equal(indexStart) {
		// make sure data blocks overlap, test block size is 2h
		// and test index block size is 4h
		start = start.Add(blockSize)
		indexStart = start.Truncate(indexBlockSize)
	}

	fooSeries := struct {
		id   string
		tags map[string]string
	}{
		"foo",
		map[string]string{"aaa": "bbb", "ccc": "ddd"},
	}
	dataBlocks := []struct {
		blockStart time.Time
		series     []testSeriesMetadata
	}{
		{
			blockStart: start,
			series: []testSeriesMetadata{
				{fooSeries.id, fooSeries.tags},
				{"bar", map[string]string{"eee": "fff", "ggg": "hhh"}},
				{"baz", map[string]string{"iii": "jjj", "kkk": "lll"}},
			},
		},
		{
			blockStart: start.Add(blockSize),
			series: []testSeriesMetadata{
				{fooSeries.id, fooSeries.tags},
				{"qux", map[string]string{"mmm": "nnn", "ooo": "ppp"}},
				{"qaz", map[string]string{"qqq": "rrr", "sss": "ttt"}},
			},
		},
		{
			blockStart: start.Add(2 * blockSize),
			series: []testSeriesMetadata{
				{fooSeries.id, fooSeries.tags},
				{"qan", map[string]string{"uuu": "vvv", "www": "xxx"}},
				{"qam", map[string]string{"yyy": "zzz", "000": "111"}},
			},
		},
	}

	end := start.Add(ropts.RetentionPeriod())

	shardTimeRanges := map[uint32]xtime.Ranges{
		0: xtime.Ranges{}.AddRange(xtime.Range{
			Start: start,
			End:   end,
		}),
	}

	nsID := nsMetadata.ID().String()

	mockAdminSession := client.NewMockAdminSession(ctrl)
	mockAdminSessionCalls := []*gomock.Call{}

	for blockStart := start; blockStart.Before(end); blockStart = blockStart.Add(blockSize) {
		// Find and expect calls for blocks
		matchedBlock := false
		for _, dataBlock := range dataBlocks {
			if !dataBlock.blockStart.Equal(blockStart) {
				continue
			}

			matchedBlock = true
			mockIter := client.NewMockPeerBlockMetadataIter(ctrl)
			mockIterCalls := []*gomock.Call{}
			for _, elem := range dataBlock.series {
				mockIterCalls = append(mockIterCalls,
					mockIter.EXPECT().Next().Return(true))

				metadata := block.NewMetadata(elem.ID(), elem.Tags(),
					blockStart, 1, nil, time.Time{})

				mockIterCalls = append(mockIterCalls,
					mockIter.EXPECT().Current().Return(nil, metadata))
			}

			mockIterCalls = append(mockIterCalls,
				mockIter.EXPECT().Next().Return(false),
				mockIter.EXPECT().Err().Return(nil))

			gomock.InOrder(mockIterCalls...)

			rangeStart := blockStart
			rangeEnd := rangeStart.Add(blockSize)
			version := opts.FetchBlocksMetadataEndpointVersion()

			call := mockAdminSession.EXPECT().
				FetchBootstrapBlocksMetadataFromPeers(ident.NewIDMatcher(nsID),
					uint32(0), rangeStart, rangeEnd, gomock.Any(), version).
				Return(mockIter, nil)
			mockAdminSessionCalls = append(mockAdminSessionCalls, call)
			break
		}

		if !matchedBlock {
			mockIter := client.NewMockPeerBlockMetadataIter(ctrl)
			gomock.InOrder(
				mockIter.EXPECT().Next().Return(false),
				mockIter.EXPECT().Err().Return(nil),
			)

			rangeStart := blockStart
			rangeEnd := rangeStart.Add(blockSize)
			version := opts.FetchBlocksMetadataEndpointVersion()

			call := mockAdminSession.EXPECT().
				FetchBootstrapBlocksMetadataFromPeers(ident.NewIDMatcher(nsID),
					uint32(0), rangeStart, rangeEnd, gomock.Any(), version).
				Return(mockIter, nil)
			mockAdminSessionCalls = append(mockAdminSessionCalls, call)
		}
	}

	gomock.InOrder(mockAdminSessionCalls...)

	mockAdminClient := client.NewMockAdminClient(ctrl)
	mockAdminClient.EXPECT().DefaultAdminSession().Return(mockAdminSession, nil)

	opts = opts.SetAdminClient(mockAdminClient)

	src := newPeersSource(opts)
	res, err := src.ReadIndex(nsMetadata, shardTimeRanges,
		testDefaultRunOpts)
	require.NoError(t, err)

	indexResults := res.IndexResults()
	require.Equal(t, 2, len(indexResults))

	for _, expected := range []struct {
		indexBlockStart time.Time
		series          map[string]testSeriesMetadata
	}{
		{
			indexBlockStart: indexStart,
			series: map[string]testSeriesMetadata{
				dataBlocks[0].series[0].id: dataBlocks[0].series[0],
				dataBlocks[0].series[1].id: dataBlocks[0].series[1],
				dataBlocks[0].series[2].id: dataBlocks[0].series[2],
				dataBlocks[1].series[1].id: dataBlocks[1].series[1],
				dataBlocks[1].series[2].id: dataBlocks[1].series[2],
			},
		},
		{
			indexBlockStart: indexStart.Add(indexBlockSize),
			series: map[string]testSeriesMetadata{
				dataBlocks[2].series[0].id: dataBlocks[2].series[0],
				dataBlocks[2].series[1].id: dataBlocks[2].series[1],
				dataBlocks[2].series[2].id: dataBlocks[2].series[2],
			},
		},
	} {
		expectedAt := xtime.ToUnixNano(expected.indexBlockStart)
		indexBlock, ok := indexResults[expectedAt]
		require.True(t, ok)
		require.Equal(t, 1, len(indexBlock.Segments()))
		for _, seg := range indexBlock.Segments() {
			reader, err := seg.Reader()
			require.NoError(t, err)

			docs, err := reader.AllDocs()
			require.NoError(t, err)

			matches := map[string]struct{}{}
			for docs.Next() {
				curr := docs.Current()

				_, ok := matches[string(curr.ID)]
				require.False(t, ok)
				matches[string(curr.ID)] = struct{}{}

				series, ok := expected.series[string(curr.ID)]
				require.True(t, ok)

				matchingTags := map[string]struct{}{}
				for _, tag := range curr.Fields {
					_, ok := matchingTags[string(tag.Name)]
					require.False(t, ok)
					matchingTags[string(tag.Name)] = struct{}{}

					tagValue, ok := series.tags[string(tag.Name)]
					require.True(t, ok)

					require.Equal(t, tagValue, string(tag.Value))
				}
				require.Equal(t, len(series.tags), len(matchingTags))
			}
			require.NoError(t, docs.Err())
			require.NoError(t, docs.Close())

			require.Equal(t, len(expected.series), len(matches))
		}
	}
}
