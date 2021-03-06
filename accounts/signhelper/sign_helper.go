// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package signhelper

import (
	"crypto/ecdsa"
	"github.com/matrix/go-matrix/accounts"
	"github.com/matrix/go-matrix/common"
	"github.com/matrix/go-matrix/core/types"
	"github.com/matrix/go-matrix/crypto"
	"github.com/pkg/errors"
	"math/big"

	"sync"
)

var (
	ErrNilAccountManager = errors.New("account manager is nil")
	ErrEmptySignAddress  = errors.New("sign address is empty")
	ErrUnSetSignAccount  = errors.New("The sign account not set yet!")
)

type SignHelper struct {
	mu           sync.RWMutex
	am           *accounts.Manager
	signWallet   accounts.Wallet
	signAccount  accounts.Account
	signPassword string
	testMode     bool
	testKey      *ecdsa.PrivateKey
}

func NewSignHelper() *SignHelper {
	return &SignHelper{
		am:          nil,
		signWallet:  nil,
		signAccount: accounts.Account{},
		testMode:    false,
	}
}

func (sh *SignHelper) SetAccountManager(am *accounts.Manager, signAddress common.Address, signPassword string) error {
	if am == nil {
		return ErrNilAccountManager
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.am = am
	sh.signWallet = nil
	sh.signAccount = accounts.Account{}

	if (signAddress != common.Address{}) {
		return sh.resetSignAccount(signAddress, signPassword)
	}

	return nil
}

func (sh *SignHelper) ResetSignAccount(signAddress common.Address, signPassword string) error {
	if (signAddress == common.Address{}) {
		return ErrEmptySignAddress
	}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	if sh.am == nil {
		return ErrNilAccountManager
	}

	return sh.resetSignAccount(signAddress, signPassword)
}

func (sh *SignHelper) SetTestMode(prvKey *ecdsa.PrivateKey) {
	sh.testMode = true
	sh.testKey = prvKey
}

func (sh *SignHelper) resetSignAccount(signAddress common.Address, signPassword string) error {
	if signAddress == sh.signAccount.Address {
		sh.signPassword = signPassword
		return nil
	}

	sh.signAccount.Address = signAddress
	sh.signWallet = nil
	wallet, err := sh.am.Find(sh.signAccount)
	if err != nil {
		return err
	}
	sh.signWallet = wallet
	sh.signPassword = signPassword
	return nil
}

func (sh *SignHelper) SignHashWithValidate(hash []byte, validate bool) (common.Signature, error) {
	if sh.testMode {
		sign, err := crypto.SignWithValidate(hash, validate, sh.testKey)
		if err != nil {
			return common.Signature{}, err
		}
		return common.BytesToSignature(sign), nil
	}

	sh.mu.RLock()
	defer sh.mu.RUnlock()

	if nil == sh.signWallet {
		return common.Signature{}, ErrUnSetSignAccount
	}

	// Sign the requested hash with the wallet
	sign, err := sh.signWallet.SignHashValidateWithPass(sh.signAccount, sh.signPassword, hash, validate)
	if err != nil {
		return common.Signature{}, err
	}
	return common.BytesToSignature(sign), nil
}

func (sh *SignHelper) SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	if nil == sh.signWallet {
		return nil, ErrUnSetSignAccount
	}

	// Sign the requested hash with the wallet
	return sh.signWallet.SignTxWithPassphrase(sh.signAccount, sh.signPassword, tx, chainID)
}
