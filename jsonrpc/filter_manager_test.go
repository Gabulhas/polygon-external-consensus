package jsonrpc

import (
	"context"
	"errors"
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/Gabulhas/polygon-external-consensus/blockchain"
	"github.com/Gabulhas/polygon-external-consensus/types"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func Test_GetLogsForQuery(t *testing.T) {
	t.Parallel()

	blockHash := types.StringToHash("1")

	// Topics we're searching for
	topic1 := types.StringToHash("4")
	topic2 := types.StringToHash("5")
	topic3 := types.StringToHash("6")

	var topics = [][]types.Hash{{topic1}, {topic2}, {topic3}}

	testTable := []struct {
		name           string
		query          *LogQuery
		expectedLength int
		expectedError  error
	}{
		{
			"Found matching logs, fromBlock < toBlock",
			&LogQuery{
				fromBlock: 1,
				toBlock:   3,
				Topics:    topics,
			},
			3,
			nil,
		},
		{
			"Found matching logs, fromBlock == toBlock",
			&LogQuery{
				fromBlock: 2,
				toBlock:   2,
				Topics:    topics,
			},
			1,
			nil,
		},
		{
			"Found matching logs, BlockHash present",
			&LogQuery{
				BlockHash: &blockHash,
				Topics:    topics,
			},
			1,
			nil,
		},
		{
			"No logs found",
			&LogQuery{
				fromBlock: 4,
				toBlock:   5,
				Topics:    topics,
			},
			0,
			nil,
		},
		{
			"Invalid block range",
			&LogQuery{
				fromBlock: 10,
				toBlock:   5,
				Topics:    topics,
			},
			0,
			ErrIncorrectBlockRange,
		},
		{
			"Block range too high",
			&LogQuery{
				fromBlock: 10,
				toBlock:   1012,
				Topics:    topics,
			},
			0,
			ErrBlockRangeTooHigh,
		},
	}

	// setup test
	store := &mockBlockStore{
		topics: []types.Hash{topic1, topic2, topic3},
	}
	store.setupLogs()

	blocks := make([]*types.Block, 5)

	for i := range blocks {
		blocks[i] = &types.Block{
			Header: &types.Header{
				Number: uint64(i),
				Hash:   types.StringToHash(strconv.Itoa(i)),
			},
			Transactions: []*types.Transaction{
				{
					Value: big.NewInt(10),
				},
				{
					Value: big.NewInt(11),
				},
				{
					Value: big.NewInt(12),
				},
			},
		}
	}

	store.appendBlocksToStore(blocks)

	f := NewFilterManager(hclog.NewNullLogger(), store, 1000)

	t.Cleanup(func() {
		defer f.Close()
	})

	for _, testCase := range testTable {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			foundLogs, logError := f.GetLogsForQuery(testCase.query)

			if logError != nil && testCase.expectedError == nil {
				// If there is an error and test isn't expected to fail
				t.Fatalf("Error: %v", logError)
			}

			if testCase.expectedError != nil {
				assert.Lenf(t, foundLogs, testCase.expectedLength, "Invalid number of logs found")
			}

			assert.ErrorIs(t, logError, testCase.expectedError)
		})
	}
}

func Test_GetLogFilterFromID(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	logFilter := &LogQuery{
		Addresses: []types.Address{addr1},
		toBlock:   10,
		fromBlock: 0,
	}

	retrivedLogFilter, err := m.GetLogFilterFromID(
		m.NewLogFilter(logFilter, &MockClosedWSConnection{}),
	)
	assert.NoError(t, err)
	assert.Equal(t, logFilter, retrivedLogFilter.query)
}

func TestFilterLog(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	id := m.NewLogFilter(&LogQuery{
		Topics: [][]types.Hash{
			{hash1},
		},
	}, nil)

	store.emitEvent(&mockEvent{
		NewChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: hash1,
				},
				receipts: []*types.Receipt{
					{
						Logs: []*types.Log{
							{
								Topics: []types.Hash{
									hash1,
								},
							},
						},
						TxHash: hash3,
					},
				},
			},
		},
		OldChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: hash2,
				},
				receipts: []*types.Receipt{
					{
						Logs: []*types.Log{
							{
								Topics: []types.Hash{
									hash1,
								},
							},
						},
						TxHash: hash3,
					},
				},
			},
		},
	})

	time.Sleep(500 * time.Millisecond)

	if _, fetchErr := m.GetFilterChanges(id); fetchErr != nil {
		t.Fatalf("Unable to get filter changes, %v", fetchErr)
	}
}

func TestFilterBlock(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	// add block filter
	id := m.NewBlockFilter(nil)

	// emit two events
	store.emitEvent(&mockEvent{
		NewChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: types.StringToHash("1"),
				},
			},
			{
				header: &types.Header{
					Hash: types.StringToHash("2"),
				},
			},
		},
	})

	store.emitEvent(&mockEvent{
		NewChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: types.StringToHash("3"),
				},
			},
		},
	})

	// we need to wait for the manager to process the data
	time.Sleep(500 * time.Millisecond)

	if _, fetchErr := m.GetFilterChanges(id); fetchErr != nil {
		t.Fatalf("Unable to get filter changes, %v", fetchErr)
	}

	// emit one more event, it should not return the
	// first three hashes
	store.emitEvent(&mockEvent{
		NewChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: types.StringToHash("4"),
				},
			},
		},
	})

	time.Sleep(500 * time.Millisecond)

	if _, fetchErr := m.GetFilterChanges(id); fetchErr != nil {
		t.Fatalf("Unable to get filter changes, %v", fetchErr)
	}
}

func TestFilterTimeout(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	m.timeout = 2 * time.Second

	go m.Run()

	// add block filter
	id := m.NewBlockFilter(nil)

	assert.True(t, m.Exists(id))
	time.Sleep(3 * time.Second)
	assert.False(t, m.Exists(id))
}

func TestRemoveFilterByWebsocket(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	mock, _ := newMockWsConnWithMsgCh()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	id := m.NewBlockFilter(mock)

	m.RemoveFilterByWs(mock)

	// false because filter was removed
	assert.False(t, m.Exists(id))
}

func Test_flushWsFilters(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)

	t.Cleanup(func() {
		m.Close()
	})

	go m.Run()

	runTest := func(t *testing.T, flushErr error, shouldExist bool) {
		t.Helper()

		var (
			filterID string
		)

		mock := &mockWsConn{
			SetFilterIDFn: func(s string) {
				filterID = s
			},
			GetFilterIDFn: func() string {
				return filterID
			},
			WriteMessageFn: func(i int, b []byte) error {
				return flushErr
			},
		}

		id := m.NewBlockFilter(mock)

		// emit event
		store.emitEvent(&mockEvent{
			NewChain: []*mockHeader{
				{
					header: &types.Header{
						Hash: types.StringToHash("1"),
					},
				},
			},
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				t.Errorf("timeout for filter existence check, expected=%t, actual=%t", shouldExist, m.Exists(id))

				return
			default:
				if shouldExist == m.Exists(id) {
					return
				}
			}
		}
	}

	t.Run("should remove if sendUpdates returns websocket.ErrCloseSent", func(t *testing.T) {
		t.Parallel()

		runTest(t, websocket.ErrCloseSent, false)
	})

	t.Run("should remove if sendUpdates returns net.ErrClosed", func(t *testing.T) {
		t.Parallel()

		runTest(t, net.ErrClosed, false)
	})

	t.Run("should keep if sendUpdates returns unknown error", func(t *testing.T) {
		t.Parallel()

		runTest(t, errors.New("hoge"), true)
	})

	t.Run("should keep if sendUpdates doesn't return error", func(t *testing.T) {
		t.Parallel()

		runTest(t, nil, true)
	})
}

func TestFilterWebsocket(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	mock, msgCh := newMockWsConnWithMsgCh()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	id := m.NewBlockFilter(mock)

	// we cannot call get filter changes for a websocket filter
	_, err := m.GetFilterChanges(id)
	assert.Equal(t, err, ErrWSFilterDoesNotSupportGetChanges)

	// emit two events
	store.emitEvent(&mockEvent{
		NewChain: []*mockHeader{
			{
				header: &types.Header{
					Hash: types.StringToHash("1"),
				},
			},
		},
	})

	select {
	case <-msgCh:
	case <-time.After(2 * time.Second):
		t.Fatal("bad")
	}
}

type mockWsConn struct {
	SetFilterIDFn  func(string)
	GetFilterIDFn  func() string
	WriteMessageFn func(int, []byte) error
}

func (m *mockWsConn) SetFilterID(filterID string) {
	m.SetFilterIDFn(filterID)
}

func (m *mockWsConn) GetFilterID() string {
	return m.GetFilterIDFn()
}

func (m *mockWsConn) WriteMessage(messageType int, b []byte) error {
	return m.WriteMessageFn(messageType, b)
}

func newMockWsConnWithMsgCh() (*mockWsConn, <-chan []byte) {
	var (
		filterID string
		msgCh    = make(chan []byte, 1)
	)

	mock := &mockWsConn{
		SetFilterIDFn: func(s string) {
			filterID = s
		},
		GetFilterIDFn: func() string {
			return filterID
		},
		WriteMessageFn: func(i int, b []byte) error {
			msgCh <- b

			return nil
		},
	}

	return mock, msgCh
}

func TestHeadStream(t *testing.T) {
	t.Parallel()

	b := &blockStream{}

	b.push(&types.Header{Hash: types.StringToHash("1")})
	b.push(&types.Header{Hash: types.StringToHash("2")})

	cur := b.Head()

	b.push(&types.Header{Hash: types.StringToHash("3")})
	b.push(&types.Header{Hash: types.StringToHash("4")})

	// get the updates, there are two new entries
	updates, next := cur.getUpdates()

	assert.Equal(t, updates[0].Hash.String(), types.StringToHash("3").String())
	assert.Equal(t, updates[1].Hash.String(), types.StringToHash("4").String())

	// there are no new entries
	updates, _ = next.getUpdates()
	assert.Len(t, updates, 0)
}

type MockClosedWSConnection struct{}

func (m *MockClosedWSConnection) SetFilterID(_filterID string) {}

func (m *MockClosedWSConnection) GetFilterID() string {
	return ""
}

func (m *MockClosedWSConnection) WriteMessage(_messageType int, _data []byte) error {
	return websocket.ErrCloseSent
}

func TestClosedFilterDeletion(t *testing.T) {
	t.Parallel()

	store := newMockStore()

	m := NewFilterManager(hclog.NewNullLogger(), store, 1000)
	defer m.Close()

	go m.Run()

	// add block filter
	id := m.NewBlockFilter(&MockClosedWSConnection{})

	assert.True(t, m.Exists(id))

	// event is sent to the filter but writing to connection should fail
	err := m.dispatchEvent(&blockchain.Event{
		NewChain: []*types.Header{
			{
				Hash: types.StringToHash("1"),
			},
		},
	})

	// should not return error when the error is websocket.ErrCloseSen because filter is removed instead
	assert.NoError(t, err)

	// false because filter was removed automatically
	assert.False(t, m.Exists(id))
}
