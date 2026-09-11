/*
Copyright (c) Arm Limited and Contributors.

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

package util

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
)

const volumeContextFileName = "volume-context.json"

func ParseJSONFile(fileName string, result interface{}) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, result)
}

func ToMiB(bytes int64) int64 {
	const mi = 1024 * 1024
	return (bytes + mi - 1) / mi
}

func FromEnv(env, def string) string {
	s := os.Getenv(env)
	if s != "" {
		return s
	}
	return def
}

type TryLock struct {
	locked int32
}

func (lock *TryLock) Lock() bool {
	return atomic.CompareAndSwapInt32(&lock.locked, 0, 1)
}

func (lock *TryLock) Unlock() {
	atomic.StoreInt32(&lock.locked, 0)
}

func ConvertInterfaceToMap(data interface{}) (map[string]string, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("the data is not a map[string]interface{}")
	}

	strMap := make(map[string]string)
	for key, value := range dataMap {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("the value for key %s is not a string", key)
		}
		strMap[key] = strValue
	}

	return strMap, nil
}

func stashContext(data interface{}, folder, fileName string) error {
	encodedBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshall context JSON: %w", err)
	}
	if _, err = os.Stat(folder); os.IsNotExist(err) {
		err = os.MkdirAll(folder, 0o755)
		if err != nil {
			return err
		}
	}
	fPath := filepath.Join(folder, fileName)
	err = os.WriteFile(fPath, encodedBytes, 0o600)
	if err != nil {
		return fmt.Errorf("failed to marshall context JSON at path (%s): %w", fPath, err)
	}
	return nil
}

func lookupContext(folder, fileName string) (interface{}, error) {
	var data interface{}
	fPath := filepath.Join(folder, fileName)
	encodedBytes, err := os.ReadFile(fPath) // #nosec - intended reading from fPath
	if err != nil {
		if !os.IsNotExist(err) {
			return data, fmt.Errorf("failed to read stashed context JSON from path (%s): %w", fPath, err)
		}
		return data, fmt.Errorf("volume context JSON file not found")
	}
	err = json.Unmarshal(encodedBytes, &data)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshall stashed context JSON from path (%s): %w", fPath, err)
	}
	return data, nil
}

func cleanUpContext(folder, fileName string) error {
	fPath := filepath.Join(folder, fileName)
	if err := os.Remove(fPath); err != nil {
		return fmt.Errorf("failed to cleanup volume context stash (%s): %w", fPath, err)
	}
	return nil
}

func StashVolumeContext(volumeContext map[string]string, path string) error {
	return stashContext(volumeContext, path, volumeContextFileName)
}

func LookupVolumeContext(path string) (map[string]string, error) {
	data, err := lookupContext(path, volumeContextFileName)
	if err != nil {
		return nil, err
	}
	return ConvertInterfaceToMap(data)
}

func CleanUpVolumeContext(path string) error {
	return cleanUpContext(path, volumeContextFileName)
}
