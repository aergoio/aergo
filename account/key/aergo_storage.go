/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package key

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	crypto "github.com/aergoio/aergo/v2/account/key/crypto"
	"github.com/aergoio/aergo/v2/types"
)

var (
	version2Strategy = map[string]crypto.KeyCryptoStrategy{
		"v1": crypto.NewV1Strategy(),
	}
)

const (
	// EncryptVersion should be always higher version
	keystorDirectory = "keystore"
	encryptVersion   = "v1"
	fileNamePostFix  = "__keystore.txt"
	fileNamePattern  = "[a-zA-Z0-9]+" + fileNamePostFix
	fileNameTemplate = "%s" + fileNamePostFix
)

type AergoStorage struct {
	sync.RWMutex
	storePath string
}

func NewAergoStorage(storePath string) (*AergoStorage, error) {
	absPath, err := filepath.Abs(filepath.Join(storePath, keystorDirectory))
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(absPath)
	if nil == err && fileInfo.Mode().IsRegular() {
		return nil, errors.New("Provided path is a file")
	}

	if os.IsNotExist(err) {
		os.MkdirAll(absPath, os.ModePerm)
	}

	return &AergoStorage{
		storePath: absPath,
	}, nil
}

func (ks *AergoStorage) Save(identity Identity, passphrase string, key *PrivateKey) (Identity, error) {
	ks.RWMutex.RLock()
	defer ks.RWMutex.RUnlock()

	// FIXME: save itself.... need to refactor store.go
	encodedIdentity := types.EncodeAddress(identity)

	fileName := fmt.Sprintf(fileNameTemplate, string(encodedIdentity))
	absFilePath := filepath.Join(ks.storePath, fileName)

	fileInfo, err := os.Stat(absFilePath)
	if nil != fileInfo {
		return nil, errors.New("already exists")
	}

	encrypted, err := GetKeystore(key, passphrase)
	if nil != err {
		return nil, err
	}
	err = ioutil.WriteFile(absFilePath, encrypted, os.FileMode(0644))
	if nil != err {
		return nil, err
	}

	return identity, nil
}

// GetKeystore encrypts a keystore file
func GetKeystore(key *PrivateKey, passphrase string) ([]byte, error) {
	// TODO: dispatch per keystore version
	strategy := version2Strategy[encryptVersion]
	encrypted, err := strategy.Encrypt(key, passphrase)
	if nil != err {
		return nil, err
	}
	return encrypted, nil
}

// LoadKeystore decrypts a keystore file
func LoadKeystore(keystore []byte, passphrase string) (*PrivateKey, error) {
	// TODO: dispatch per keystore version
	strategy := version2Strategy[encryptVersion]
	privateKey, err := strategy.Decrypt(keystore, passphrase)
	if nil != err {
		return nil, err
	}
	return privateKey, nil
}

func (ks *AergoStorage) Load(identity Identity, passphrase string) (*PrivateKey, error) {
	// FIXME: save itself. need to refactor store.go
	encodedIdentity := types.EncodeAddress(identity)

	fileName := fmt.Sprintf(fileNameTemplate, string(encodedIdentity))
	absFilePath := filepath.Join(ks.storePath, fileName)

	encrypted, err := ioutil.ReadFile(absFilePath)
	if nil != err {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("account with address %s does not exist", encodedIdentity)
		}
		return nil, fmt.Errorf("failed to read account with address %s: %v", encodedIdentity, err)
	}

	privateKey, err := LoadKeystore(encrypted, passphrase)
	if nil != err {
		return nil, err
	}

	return privateKey, nil
}

func (ks *AergoStorage) List() ([]Identity, error) {
	ks.RWMutex.RLock()
	defer ks.RWMutex.RUnlock()

	files, err := ioutil.ReadDir(ks.storePath)
	if nil != err {
		return nil, err
	}

	var ret = make([][]byte, 0)
	for _, file := range files {
		fileName := file.Name()
		matched, _ := regexp.MatchString(fileNamePattern, fileName)
		if matched {
			if index := strings.Index(fileName, fileNamePostFix); -1 != index {
				identity, err := types.DecodeAddress(fileName[0:index])
				if err == nil {
					ret = append(ret, identity)
				}
			}
		}
	}

	return ret, nil
}

func (ks *AergoStorage) Close() {
	// do nothing
}
