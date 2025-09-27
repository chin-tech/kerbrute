package cmd

import (
	"context"
	"fmt"
	"github.com/chin-tech/kerbrute/util"
	"sync"
	"sync/atomic"
)

// func makeWorker(ctx context.Context, inChan interface{}, wg *sync.WaitGroup, execFunc func(context.Context, ...interface{}), args ...interface{}) {
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			break
// 		case data, ok := <-inChan.(chan interface{}):
// 			if !ok {
// 				return
// 			}
// 			switch v := data.(type) {
// 			case string:
// 				execFunc(ctx, v, args[0])
// 			case [2]string:
// 				execFunc(ctx, v[0], v[1])
// 			}
// 		}
// 	}
// }

func makeSprayWorker(ctx context.Context, usernames <-chan string, wg *sync.WaitGroup, cred util.SecureCredential, userAsPass bool) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case username, ok := <-usernames:
			if !ok {
				return
			}
			if userAsPass {
				cred.Type = util.CredPassword
				cred.Cred = username
				testCred(ctx, username, cred)
			} else {
				testCred(ctx, username, cred)
			}
		}
	}
}

func makeBruteWorker(ctx context.Context, passwords <-chan util.SecureCredential, wg *sync.WaitGroup, username string) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case password, ok := <-passwords:
			if !ok {
				return
			}
			testCred(ctx, username, password)
		}
	}
}

func makeEnumWorker(ctx context.Context, usernames <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case username, ok := <-usernames:
			if !ok {
				return
			}
			TestUsername(ctx, username)
		}
	}
}

func makeBruteComboWorker(ctx context.Context, combos <-chan util.Combo, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			break
		case combo, ok := <-combos:
			if !ok {
				return
			}
			testCred(ctx, combo.Username, combo.Cred)
		}
	}
}

func testCred(ctx context.Context, username string, cred util.SecureCredential) {
	atomic.AddInt32(&counter, 1)
	login := fmt.Sprintf("%v@%v:%v", username, domain, cred.Cred)
	if ok, err := kSession.TestCredential(username, cred); ok {
		atomic.AddInt32(&successes, 1)
		if err != nil { // it's a valid login, but there's an error we should display
			logger.Log.Noticef("[+] VALID LOGIN WITH ERROR:\t %s\t (%s)", login, err)
		} else {
			logger.Log.Noticef("[+] VALID LOGIN:\t %s", login)
		}
		if stopOnSuccess {
			cancel()
		}
	} else {
		// This is to determine if the error is "okay" or if we should abort everything
		ok, errorString := kSession.HandleKerbError(err)
		if !ok {
			logger.Log.Errorf("[!] %v - %v", login, errorString)
			cancel()
		} else {
			logger.Log.Debugf("[!] %v - %v", login, errorString)
		}
	}
}

func TestLogin(ctx context.Context, username string, password string) {
	atomic.AddInt32(&counter, 1)
	login := fmt.Sprintf("%v@%v:%v", username, domain, password)
	if ok, err := kSession.TestLogin(username, password); ok {
		atomic.AddInt32(&successes, 1)
		if err != nil { // it's a valid login, but there's an error we should display
			logger.Log.Noticef("[+] VALID LOGIN WITH ERROR:\t %s\t (%s)", login, err)
		} else {
			logger.Log.Noticef("[+] VALID LOGIN:\t %s", login)
		}
		if stopOnSuccess {
			cancel()
		}
	} else {
		// This is to determine if the error is "okay" or if we should abort everything
		ok, errorString := kSession.HandleKerbError(err)
		if !ok {
			logger.Log.Errorf("[!] %v - %v", login, errorString)
			cancel()
		} else {
			logger.Log.Debugf("[!] %v - %v", login, errorString)
		}
	}
}

func TestUsername(ctx context.Context, username string) {
	atomic.AddInt32(&counter, 1)
	usernamefull := fmt.Sprintf("%v@%v", username, domain)
	valid, err := kSession.TestUsername(username)
	if valid {
		atomic.AddInt32(&successes, 1)
		if err != nil {
			logger.Log.Noticef("[+] VALID USERNAME WITH ERROR:\t %s\t (%s)", username, err)
		} else {
			logger.Log.Noticef("[+] VALID USERNAME:\t %s", usernamefull)
		}

	} else if err != nil {
		// This is to determine if the error is "okay" or if we should abort everything
		ok, errorString := kSession.HandleKerbError(err)
		if !ok {
			logger.Log.Errorf("[!] %v - %v", usernamefull, errorString)
			cancel()
		} else {
			logger.Log.Debugf("[!] %v - %v", usernamefull, errorString)
		}
	} else {
		logger.Log.Debug("[!] Unknown behavior - %v", usernamefull)
	}
}
