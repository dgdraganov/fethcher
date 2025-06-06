// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"context"
	"fethcher/internal/core"
	"fethcher/internal/ethereum"
	"sync"
)

type EthereumService struct {
	FetchTransactionsStub        func(context.Context, []string) ([]*ethereum.Transaction, error)
	fetchTransactionsMutex       sync.RWMutex
	fetchTransactionsArgsForCall []struct {
		arg1 context.Context
		arg2 []string
	}
	fetchTransactionsReturns struct {
		result1 []*ethereum.Transaction
		result2 error
	}
	fetchTransactionsReturnsOnCall map[int]struct {
		result1 []*ethereum.Transaction
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *EthereumService) FetchTransactions(arg1 context.Context, arg2 []string) ([]*ethereum.Transaction, error) {
	var arg2Copy []string
	if arg2 != nil {
		arg2Copy = make([]string, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.fetchTransactionsMutex.Lock()
	ret, specificReturn := fake.fetchTransactionsReturnsOnCall[len(fake.fetchTransactionsArgsForCall)]
	fake.fetchTransactionsArgsForCall = append(fake.fetchTransactionsArgsForCall, struct {
		arg1 context.Context
		arg2 []string
	}{arg1, arg2Copy})
	stub := fake.FetchTransactionsStub
	fakeReturns := fake.fetchTransactionsReturns
	fake.recordInvocation("FetchTransactions", []interface{}{arg1, arg2Copy})
	fake.fetchTransactionsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EthereumService) FetchTransactionsCallCount() int {
	fake.fetchTransactionsMutex.RLock()
	defer fake.fetchTransactionsMutex.RUnlock()
	return len(fake.fetchTransactionsArgsForCall)
}

func (fake *EthereumService) FetchTransactionsCalls(stub func(context.Context, []string) ([]*ethereum.Transaction, error)) {
	fake.fetchTransactionsMutex.Lock()
	defer fake.fetchTransactionsMutex.Unlock()
	fake.FetchTransactionsStub = stub
}

func (fake *EthereumService) FetchTransactionsArgsForCall(i int) (context.Context, []string) {
	fake.fetchTransactionsMutex.RLock()
	defer fake.fetchTransactionsMutex.RUnlock()
	argsForCall := fake.fetchTransactionsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *EthereumService) FetchTransactionsReturns(result1 []*ethereum.Transaction, result2 error) {
	fake.fetchTransactionsMutex.Lock()
	defer fake.fetchTransactionsMutex.Unlock()
	fake.FetchTransactionsStub = nil
	fake.fetchTransactionsReturns = struct {
		result1 []*ethereum.Transaction
		result2 error
	}{result1, result2}
}

func (fake *EthereumService) FetchTransactionsReturnsOnCall(i int, result1 []*ethereum.Transaction, result2 error) {
	fake.fetchTransactionsMutex.Lock()
	defer fake.fetchTransactionsMutex.Unlock()
	fake.FetchTransactionsStub = nil
	if fake.fetchTransactionsReturnsOnCall == nil {
		fake.fetchTransactionsReturnsOnCall = make(map[int]struct {
			result1 []*ethereum.Transaction
			result2 error
		})
	}
	fake.fetchTransactionsReturnsOnCall[i] = struct {
		result1 []*ethereum.Transaction
		result2 error
	}{result1, result2}
}

func (fake *EthereumService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.fetchTransactionsMutex.RLock()
	defer fake.fetchTransactionsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *EthereumService) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ core.EthereumService = new(EthereumService)
